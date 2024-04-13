package database

import (
	"fmt"
	"search/config"
	"search/corpus"
	"search/packing"
	"search/utils"

	"github.com/henrycg/simplepir/lwe"
	"github.com/henrycg/simplepir/matrix"
	"github.com/henrycg/simplepir/pir"
	"github.com/henrycg/simplepir/rand"
)

func BuildUrlsDatabase(c *corpus.Corpus, seed *rand.PRGKey, hintSz uint64) (*pir.Database[matrix.Elem32], SubclusterMap) {
	d := uint64(8) // 一个byte中的bit数
	l := uint64(c.GetUrlBytes())
	logQ := uint64(32)

	if hintSz*250 > l {
		fmt.Printf("Increasing L from %d to %d\n", l, hintSz*250)
		l = hintSz * 250
	} else {
		fmt.Printf("Hint size is %d MB\n", l/250)
	}

	// 将url字符串打包进database columns
	chunks, actualSz := packing.BuildUrlChunks(c)
	cols, colSzs := packing.PackChunks(chunks, l)

	m := uint64(len(cols))
	l = utils.Max(colSzs)
	fmt.Printf("DB size is %d -- best possible would be %d\n", l*m, actualSz)

	// 获取 SimplePIR Params
	p := lwe.NewParamsFixedP(logQ, m, 256)
	if p == nil || p.Logq != 32 {
		panic("Failure in picking SimplePIR DB parameters")
	}

	subclusterToCluster := c.SubclusterToClusterMap()

	// 将字符串储存于 database
	vals := make([]uint64, l*m)
	indexMap := make(map[uint][]corpus.Subcluster)
	for colIndex, colContents := range cols {
		rowIndex := uint64(0)
		for _, subcluster := range colContents {
			cluster := subclusterToCluster[subcluster]
			if _, ok := indexMap[cluster]; !ok {
				indexMap[cluster] = make([]corpus.Subcluster, c.NumSubclustersInCluster(cluster))
			}
			arr := c.GetSubcluster(subcluster)
			i := c.IndexOfSubclusterWithinCluster(cluster, subcluster)
			sz := uint64(c.SizeOfSubclusterByIndex(cluster, i))
			indexMap[cluster][i].SetIndex(DBIndex(rowIndex, uint64(colIndex), m))
			indexMap[cluster][i].SetSize(sz)

			for j := 0; j < len(arr); j++ {
				vals[DBIndex(rowIndex, uint64(colIndex), m)] = uint64(arr[j])
				rowIndex += 1
				if rowIndex > l {
					panic("Should not happen")
				}
			}
		}
	}
	db := pir.NewDatabaseFixedParams[matrix.Elem32](l*m, d, vals, p)
	if db.Info.L != l {
		panic("Should not happen")
	}

	return db, indexMap
}

func BuildEmbeddingsDatabase(c *corpus.Corpus, seed *rand.PRGKey, hintSz uint64, conf *config.Config) (*pir.Database[matrix.Elem64], ClusterMap) {
	l := hintSz * 125
	logQ := uint64(64)

	fmt.Printf("Building db with %d embedding\n", c.GetNumDocs())

	// 将聚类打包进 database columns
	chunks, actualSz := packing.BuildEmbChunks(c)
	cols, colSzs := packing.PackChunks(chunks, l)

	m := uint64(len(cols)) * c.GetEmbeddingSlots()
	l = utils.Max(colSzs)
	fmt.Printf("DB size is %d -- best possible would be %d\n", l*m, actualSz)

	// 获取SimplePIR params
	recordLen := conf.SIMPLEPIR_EMBEDDINGS_RECORD_LENGTH()
	p := lwe.NewParamsFixedP(logQ, m, (1 << recordLen))
	if (p == nil) || (p.P < uint64(1<<c.GetSlotBits())) || (p.Logq != 64) {
		fmt.Printf("P = %d; LogQ = %d\n", p.P, p.Logq)
		panic("Failure in picking SimplePIR DB parameters")
	}

	// 将嵌入储存到 database，便将聚类保存到一列中
	vals := make([]uint64, l*m)
	indexMap := make(map[uint]uint64)
	slots := c.GetEmbeddingSlots()

	for colIndex, colContents := range cols {
		rowIndex := uint64(0)
		for _, clusterIndex := range colContents {
			if _, ok := indexMap[clusterIndex]; ok {
				panic("Key should not yet exist")
			}

			indexMap[clusterIndex] = DBIndex(rowIndex, slots*uint64(colIndex), m)
			sz := c.NumDocsInCluster(clusterIndex)
			start := uint64(c.ClusterToIndex(clusterIndex))

			for x := uint64(0); x < sz; x++ {
				arr := c.GetEmbedding(start)
				for j := uint64(0); j < slots; j++ {
					vals[DBIndex(rowIndex, slots*uint64(colIndex)+j, m)] = uint64(arr[j])
				}
				start += slots
				rowIndex += 1
				if rowIndex > l {
					panic("Should not happen")
				}
			}
		}
	}
	db := pir.NewDatabaseFixedParams[matrix.Elem64](l*m, uint64(recordLen), vals, p)
	fmt.Printf("DB dimensions: %d by %d\n", db.Info.L, db.Info.M)

	if db.Info.L != l {
		panic("Should not happen")
	}

	return db, indexMap
}
