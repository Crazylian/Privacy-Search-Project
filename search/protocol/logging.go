package protocol

import (
	"encoding/csv"
	"encoding/gob"
	"fmt"
	"os"
	"search/corpus"
	"search/database"
	"search/embeddings"
	"search/utils"
	"strconv"
	"time"

	"github.com/ahenzinger/underhood/underhood"
	"github.com/henrycg/simplepir/matrix"
	"github.com/henrycg/simplepir/pir"
)

type Perf struct {
	clientTotal   float64
	t1            float64
	t2            float64
	clientPreproc float64
	clientSetup   float64
	up1           float64
	up2           float64
	down1         float64
	down2         float64
	tput1         float64
	tput2         float64
	tOffline      float64
	upOffline     float64
	downOffline   float64
	tputOffline   float64
}

func logHintSize(hint *TiptoeHint) float64 {
	gob.Register(corpus.Params{})
	total := utils.MessageSizeMB(hint.CParams)

	if hint.ServeEmbeddings {
		gob.Register(database.ClusterMap{})
		h := utils.MessageSizeMB(hint.EmbeddingsHint)
		m := utils.MessageSizeMB(hint.EmbeddingsIndexMap)
		total += (h + m)

		fmt.Printf("\t\tEmbeddings hint: %.2f MB\n", h)
		fmt.Printf("\t\tEmbeddings map: %.2f MB\n", m)
	}

	if hint.ServeUrls {
		gob.Register(database.SubclusterMap{})
		h := utils.MessageSizeMB(hint.UrlsHint)
		m := utils.MessageSizeMB(hint.UrlsIndexMap)
		total += (h + m)

		fmt.Printf("\t\tUrls hint: %.2f MB\n", h)
		fmt.Printf("\t\tUrls map: %.2f MB\n", m)
	}

	fmt.Printf("\tTotal metadata: %.2f MB\n", total)
	return total
}

func logOfflineStats(numDocs uint64,
	start time.Time,
	up *underhood.HintQuery,
	down *UnderhoodAnswer) (float64, float64, float64) {
	elapsed := time.Since(start)
	upSz := utils.MessageSizeMB(*up)
	downSz := utils.MessageSizeMB(down.EmbAnswer) + utils.MessageSizeMB(down.UrlAnswer)

	fmt.Printf("\tPreprocessed query to %d-document corpus in: %s\n", numDocs, elapsed)
	fmt.Printf("\tUpload: %.2f MB\n", upSz)
	fmt.Printf("\tDownload: %.2f MB\n", downSz)

	return elapsed.Seconds(), upSz, downSz
}

func logStats[T matrix.Elem](numDocs uint64,
	start time.Time,
	up *pir.Query[T],
	down *pir.Answer[T]) (float64, float64, float64) {
	elapsed := time.Since(start)
	upSz := utils.MessageSizeMB(*up)
	downSz := utils.MessageSizeMB(*down)

	fmt.Printf("\tAnswered query to %d-document corpus in: %s\n", numDocs, elapsed)
	fmt.Printf("\tUpload: %.2f MB\n", upSz)
	fmt.Printf("\tDownload: %.2f MB\n", downSz)

	return elapsed.Seconds(), upSz, downSz
}

func checkAnswer(got, index, p uint64, emb []int8, corp *corpus.Corpus) {
	docEmb := corp.GetEmbedding(index)
	shouldBe := embeddings.InnerProduct(docEmb, emb)
	res := embeddings.SmoothResult(got, p)

	if res != shouldBe {
		fmt.Printf("Recovering doc %d: got %d instead of %d\n",
			index/corp.GetEmbeddingSlots(), res, shouldBe)
		panic("Bad answer")
	}
}

func checkAnswers(got []uint64, cluster uint, p uint64, emb []int8, corp *corpus.Corpus) {
	clusterSz := corp.NumDocsInCluster(cluster)
	index := uint64(corp.ClusterToIndex(cluster))

	for j := uint64(0); j < uint64(len(got)); j++ {
		if (j >= clusterSz) && (got[j] != 0) {
			fmt.Printf("Row %d of %d (actually %d)\n", j, len(got), clusterSz)
			fmt.Printf("Got %d instead of %d\n", got[j], 0)
			panic("Bad answer")
		} else if j < clusterSz {
			checkAnswer(got[j], index, p, emb, corp)
			index += corp.GetEmbeddingSlots()
		}
	}
}

func checkSubclusterSize(recoveredUrl string,
	clusterIndex uint64,
	chunkIndex int,
	corp *corpus.Corpus) {
	occ := corpus.CountUrls(recoveredUrl)
	shouldBe := corp.SizeOfSubclusterByIndex(uint(clusterIndex), chunkIndex)

	if occ != shouldBe {
		fmt.Println(recoveredUrl)
		fmt.Printf("Num URLS is %d -- expected %d\n", occ, shouldBe)
		fmt.Printf("Query to cluster %d, chunk %d\n", clusterIndex, chunkIndex)
		panic("Should not happen")
	}
}

func initCsv(fn string) {
	f, err := os.OpenFile(fn, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	writer := csv.NewWriter(f)
	defer writer.Flush()

	records := []string{"Trial", "NumClients", "NumDocs", "EmbeddingSlots", "SlotBits", "UrlBytes", "Num servers 1", "Num servers 2", "Hint (MB)", "Time (s)", "CP (s)", "CS (s)", "Time 1 (s)", "Time 2 (s)", "Q (MB)", "Q1 (MB)", "Q2 (MB)", "A (MB)", "A1 (MB)", "A2 (MB)", "Tput1 (queries/s)", "Tput2 (queries/s)", "Time offline (s)", "Q offline (MB)", "A offline (MB)", "Tput offline (queries/s)"}
	writer.Write(records)
}

func writeLatencyCsv(fn string,
	perf []Perf,
	corpus *corpus.Params,
	hintSz float64,
	numEmbServers, numUrlServers int) {
	f := utils.OpenAppendFile(fn)
	defer f.Close()

	writer := csv.NewWriter(f)
	defer writer.Flush()

	for i := 0; i < len(perf); i++ {
		writer.Write([]string{strconv.Itoa(i),
			"0",
			strconv.FormatUint(corpus.NumDocs, 10),
			strconv.FormatUint(corpus.EmbeddingSlots, 10),
			strconv.FormatUint(corpus.SlotBits, 10),
			strconv.FormatUint(corpus.UrlBytes, 10),
			strconv.Itoa(numEmbServers),
			strconv.Itoa(numUrlServers),
			strconv.FormatFloat(hintSz, 'f', 4, 64),
			strconv.FormatFloat(perf[i].clientTotal, 'f', 4, 64),
			strconv.FormatFloat(perf[i].clientPreproc, 'f', 4, 64),
			strconv.FormatFloat(perf[i].clientSetup, 'f', 4, 64),
			strconv.FormatFloat(perf[i].t1, 'f', 4, 64),
			strconv.FormatFloat(perf[i].t2, 'f', 4, 64),
			strconv.FormatFloat(perf[i].up1+perf[i].up2+perf[i].upOffline, 'f', 4, 64),
			strconv.FormatFloat(perf[i].up1, 'f', 4, 64),
			strconv.FormatFloat(perf[i].up2, 'f', 4, 64),
			strconv.FormatFloat(perf[i].down1+perf[i].down2+perf[i].downOffline, 'f', 4, 64),
			strconv.FormatFloat(perf[i].down1, 'f', 4, 64),
			strconv.FormatFloat(perf[i].down2, 'f', 4, 64),
			"0", "0",
			strconv.FormatFloat(perf[i].tOffline, 'f', 4, 64),
			strconv.FormatFloat(perf[i].upOffline, 'f', 4, 64),
			strconv.FormatFloat(perf[i].downOffline, 'f', 4, 64),
			"0",
		})

	}
}

func writeTputCsv(fn string,
	numClients []int,
	perf []Perf,
	corpus *corpus.Params,
	hintSz float64,
	numEmbServers, numUrlServers int) {
	if len(numClients) != len(perf) {
		panic("Should not happen")
	}

	f := utils.OpenAppendFile(fn)
	defer f.Close()

	writer := csv.NewWriter(f)
	defer writer.Flush()

	for i := 0; i < len(perf); i++ {
		writer.Write([]string{strconv.Itoa(i),
			strconv.Itoa(numClients[i]),
			strconv.FormatUint(corpus.NumDocs, 10),
			strconv.FormatUint(corpus.EmbeddingSlots, 10),
			strconv.FormatUint(corpus.SlotBits, 10),
			strconv.FormatUint(corpus.UrlBytes, 10),
			strconv.Itoa(numEmbServers),
			strconv.Itoa(numUrlServers),
			strconv.FormatFloat(hintSz, 'f', 4, 64),
			"0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0",
			strconv.FormatFloat(perf[i].tput1, 'f', 4, 64),
			strconv.FormatFloat(perf[i].tput2, 'f', 4, 64),
			"0", "0", "0",
			strconv.FormatFloat(perf[i].tputOffline, 'f', 4, 64),
		})

	}
}
