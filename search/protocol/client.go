// package protocol

// import (
// 	"encoding/gob"
// 	"encoding/json"
// 	"fmt"
// 	"io"
// 	"net/rpc"
// 	"search/config"
// 	"search/corpus"
// 	"search/database"
// 	"search/embeddings"
// 	"search/utils"
// 	"strings"
// 	"time"

// 	"github.com/ahenzinger/underhood/underhood"
// 	"github.com/fatih/color"
// 	"github.com/henrycg/simplepir/matrix"
// 	"github.com/henrycg/simplepir/pir"
// )

// type UnderhoodAnswer struct {
// 	EmbAnswer underhood.HintAnswer
// 	UrlAnswer underhood.HintAnswer
// }

// type QueryType interface {
// 	bool | underhood.HintQuery | pir.Query[matrix.Elem64] | pir.Query[matrix.Elem32]
// }

// type AnsType interface {
// 	TiptoeHint | UnderhoodAnswer | pir.Answer[matrix.Elem64] | pir.Answer[matrix.Elem32]
// }

// type Client struct {
// 	params corpus.Params

// 	embClient  *underhood.Client[matrix.Elem64]
// 	embInfo    *pir.DBInfo
// 	embMap     database.ClusterMap
// 	embIndices map[uint64]bool

// 	urlClient  *underhood.Client[matrix.Elem32]
// 	urlInfo    *pir.DBInfo
// 	urlMap     database.SubclusterMap
// 	urlIndices map[uint64]bool

// 	rpcClient *rpc.Client
// }

// func NewClient() *Client {
// 	c := new(Client)
// 	return c
// }

// func (c *Client) NumDocs() uint64 {
// 	return c.params.NumDocs
// }

// func (c *Client) NumClusters() int {
// 	if len(c.embMap) > 0 {
// 		return len(c.embMap)
// 	}
// 	return len(c.urlMap)
// }

// func (c *Client) Setup(hint *TiptoeHint) {
// 	if hint == nil {
// 		panic("Hint is empty")
// 	}
// 	if hint.CParams.NumDocs == 0 {
// 		panic("Corpus is empty")
// 	}

// 	c.params = hint.CParams
// 	c.embInfo = &hint.EmbeddingsHint.Info
// 	c.urlInfo = &hint.UrlsHint.Info

// 	if hint.ServeEmbeddings {
// 		if hint.EmbeddingsHint.IsEmpty() {
// 			panic("Embeddings hint is empty")
// 		}

// 		c.embClient = utils.NewUnderhoodClient(&hint.EmbeddingsHint)

// 		c.embMap = hint.EmbeddingsIndexMap
// 		c.embIndices = make(map[uint64]bool)
// 		for _, v := range c.embMap {
// 			c.embIndices[v] = true
// 		}

// 		fmt.Printf("\tEmbeddings client: %s\n", utils.PrintParams(c.embInfo))
// 	}

// 	if hint.ServeUrls {
// 		if hint.UrlsHint.IsEmpty() {
// 			panic("Urls hint is empty")
// 		}

// 		c.urlClient = utils.NewUnderhoodClient(&hint.UrlsHint)

// 		c.urlMap = hint.UrlsIndexMap
// 		c.urlIndices = make(map[uint64]bool)
// 		for _, vals := range c.urlMap {
// 			for _, v := range vals {
// 				c.urlIndices[v.Index()] = true
// 			}
// 		}

// 		fmt.Printf("\tURL client: %s\n", utils.PrintParams(c.urlInfo))
// 	}

// 	if hint.ServeUrls && hint.ServeEmbeddings &&
// 		(len(c.urlMap) != len(c.embMap)) {
// 		fmt.Printf("Both maps don't have the same length: %d %d\n", len(c.urlMap), len(c.embMap))
// 		//    panic("Both maps don't have same length.")
// 	}
// }

// func InitHint(embhint *TiptoeHint, urlhint *TiptoeHint) *TiptoeHint {
// 	var hint = new(TiptoeHint)
// 	hint.CParams.NumDocs = embhint.CParams.NumDocs + urlhint.CParams.NumDocs

// 	hint.CParams.EmbeddingSlots = embhint.CParams.EmbeddingSlots
// 	hint.CParams.SlotBits = embhint.CParams.SlotBits
// 	hint.EmbeddingsHint = embhint.EmbeddingsHint
// 	hint.EmbeddingsIndexMap = embhint.EmbeddingsIndexMap

// 	hint.CParams.UrlBytes = urlhint.CParams.UrlBytes
// 	hint.CParams.CompressUrl = urlhint.CParams.CompressUrl
// 	hint.UrlsHint = urlhint.UrlsHint
// 	hint.UrlsIndexMap = urlhint.UrlsIndexMap

// 	hint.ServeEmbeddings = true
// 	hint.ServeUrls = true
// 	return hint
// }

// func RunClient(EmbAddr string, UrlAddr string, conf *config.Config) {
// 	fmt.Println("Setting up client...")

// 	c := NewClient()
// 	fmt.Println("1.Getting metadata")
// 	embhint := c.getHint(false, EmbAddr)
// 	urlhint := c.getHint(false, UrlAddr)
// 	// fmt.Println(embhint.EmbeddingsHint)
// 	// fmt.Println(urlhint.UrlsHint)
// 	hint := InitHint(embhint, urlhint)

// 	c.Setup(hint)
// 	// logHintSize(hint)
// 	gob.Register(corpus.Params{})
// 	total := utils.MessageSizeMB(hint.CParams)

// 	if hint.ServeEmbeddings {
// 		gob.Register(database.ClusterMap{})
// 		h := utils.MessageSizeMB(hint.EmbeddingsHint)
// 		m := utils.MessageSizeMB(hint.EmbeddingsIndexMap)
// 		total += (h + m)

// 		fmt.Printf("\t\tEmbeddings hint: %.2f MB\n", h)
// 		fmt.Printf("\t\tEmbeddings map: %.2f MB\n", m)
// 	}

// 	if hint.ServeUrls {
// 		gob.Register(database.SubclusterMap{})
// 		h := utils.MessageSizeMB(hint.UrlsHint)
// 		m := utils.MessageSizeMB(hint.UrlsIndexMap)
// 		total += (h + m)

// 		fmt.Printf("\t\tUrls hint: %.2f MB\n", h)
// 		fmt.Printf("\t\tUrls map: %.2f MB\n", m)
// 	}
// 	fmt.Printf("\tTotal metadata: %.2f MB\n", total)

// 	in, out := embeddings.SetupEmbeddingProcess(c.NumClusters(), conf)
// 	fmt.Println("Running client preprocessing")
// 	defer in.Close()
// 	defer out.Close()

// 	for {
// 		fmt.Printf("Enter private search query: ")
// 		text := utils.ReadLineFromStdin()
// 		fmt.Printf("\n\n")
// 		if (strings.TrimSpace(text) == "") || (strings.TrimSpace(text) == "quit") {
// 			break
// 		}
// 		c.runRound(in, out, text, EmbAddr, UrlAddr, true, true)
// 	}

// 	if c.rpcClient != nil {
// 		c.rpcClient.Close()
// 	}
// }

// func (c *Client) runRound(in io.WriteCloser, out io.ReadCloser, text, EmbAddr string, UrlAddr string, verbose, keepConn bool) {
// 	fmt.Printf("Executing query \"%s\"\n", text)
// 	var query struct {
// 		Cluster_index uint64
// 		Emb           []int8
// 	}

// 	// Perform processing
// 	start := time.Now()
// 	ct := c.PreprocessQuery()
// 	EmbofflineAns := c.applyHint(ct, false, EmbAddr)
// 	fmt.Println(EmbofflineAns.EmbAnswer)
// 	toDrop := int(2048 - 1408)
// 	*ct = (*ct)[:len(*ct)-toDrop]
// 	UrlofflineAns := c.applyHint(ct, false, UrlAddr)
// 	var offlineAns = new(UnderhoodAnswer)
// 	offlineAns.EmbAnswer = EmbofflineAns.EmbAnswer
// 	offlineAns.UrlAnswer = UrlofflineAns.UrlAnswer
// 	c.ProcessHintApply(offlineAns)
// 	clientPreproc := time.Since(start).Seconds()
// 	if verbose {
// 		fmt.Printf("\tPreprocessing complete -- %fs\n\n", clientPreproc)
// 	}

// 	// Build embeddings query
// 	start = time.Now()
// 	if verbose {
// 		fmt.Println("2.Generating embeding of the query")
// 	}

// 	io.WriteString(in, text+"\n")
// 	if err := json.NewDecoder(out).Decode(&query); err != nil {
// 		panic(err)
// 	}

// 	if query.Cluster_index >= uint64(c.NumClusters()) {
// 		panic("Should not happen")
// 	}

// 	if verbose {
// 		fmt.Printf("3.Building PIR query for cluster %d\n", query.Cluster_index)
// 	}
// 	embQuery := c.QueryEmbeddings(query.Emb, query.Cluster_index)
// 	// p.clientSetup = time.Since(start).Seconds()
// 	clientSetup := time.Since(start).Seconds()

// 	// Send embeddings query to server
// 	if verbose {
// 		fmt.Println("4.Sending SimplePIR query to server")
// 	}
// 	networkingStartEmb := time.Now()
// 	embAns := c.getEmbeddingsAnswer(embQuery, false, EmbAddr)
// 	// p.t1, p.up1, p.down1 = logStats(c.params.NumDocs, networkingStart, embQuery, embAns)
// 	EmbTime := time.Since(networkingStartEmb).Seconds()

// 	// Recover document and URL chunk to query for
// 	fmt.Println("5.Decrypting server answer")
// 	embDec := c.ReconstructEmbeddingsWithinCluster(embAns, query.Cluster_index)
// 	scores := embeddings.SmoothResults(embDec, c.embInfo.P())
// 	indicesByScore := utils.SortByScores(scores)
// 	docIndex := indicesByScore[0]

// 	if verbose {
// 		fmt.Printf("\tDoc %d within cluster %d has the largest inner product with our query\n", docIndex, query.Cluster_index)
// 		fmt.Printf("Building PIR query for url/title of doc %d in cluster %d\n", docIndex, query.Cluster_index)
// 	}
// 	// Build URL query
// 	urlQuery, retrievedChunk := c.QueryUrls(query.Cluster_index, docIndex)

// 	// Send URL query to server
// 	if verbose {
// 		fmt.Printf("Sending PIR query to server for chunk %d\n", retrievedChunk)
// 	}
// 	networkingStartUrl := time.Now()
// 	urlAns := c.getUrlsAnswer(urlQuery, keepConn, UrlAddr)
// 	// p.t2, p.up2, p.down2 = logStats(c.params.NumDocs, networkingStart, urlQuery, urlAns)
// 	UrlTime := time.Since(networkingStartUrl).Seconds()

// 	// Recover URLs of top 10 docs in chunk
// 	urls := c.ReconstructUrls(urlAns, query.Cluster_index, docIndex)
// 	if verbose {
// 		fmt.Println("Reconstructed PIR answers.")
// 		fmt.Printf("\tThe top 10 retrieved urls are:\n")
// 	}
// 	j := 1
// 	for at := 0; at < len(indicesByScore); at++ {
// 		if scores[at] == 0 {
// 			break
// 		}

// 		doc := indicesByScore[at]
// 		_, chunk, index := c.urlMap.SubclusterToIndex(query.Cluster_index, doc)

// 		if chunk == retrievedChunk {
// 			if verbose {
// 				fmt.Printf("\t% 3d) [score %s] %s\n", j,
// 					color.YellowString(fmt.Sprintf("% 4d", scores[at])),
// 					color.BlueString(corpus.GetIthUrl(urls, index)))
// 			}
// 			j += 1
// 			if j > 10 {
// 				break
// 			}
// 		}
// 	}

// 	clientTotal := time.Since(start).Seconds()
// 	fmt.Printf("\tAnswered in:\n\t\t%v (preproc)\n\t\t%v (client)\n\t\t%v (round 1)\n\t\t%v (round 2)\n\t\t%v (total)\n---\n",
// 		clientPreproc, clientSetup, EmbTime, UrlTime, clientTotal)
// }

// func (c *Client) PreprocessQuery() *underhood.HintQuery {
// 	if c.params.NumDocs == 0 {
// 		panic("Not set up")
// 	}

// 	if c.embClient != nil {
// 		hintQuery := c.embClient.HintQuery()
// 		if c.urlClient != nil {
// 			c.urlClient.CopySecret(c.embClient)
// 		}
// 		return hintQuery
// 	} else if c.urlClient != nil {
// 		return c.urlClient.HintQuery()
// 	} else {
// 		panic("Should not happen")
// 	}
// }

// func (c *Client) ProcessHintApply(ans *UnderhoodAnswer) {
// 	if c.embClient != nil {
// 		c.embClient.HintRecover(&ans.EmbAnswer)
// 		c.embClient.PreprocessQueryLHE()
// 	}

// 	if c.urlClient != nil {
// 		c.urlClient.HintRecover(&ans.UrlAnswer)
// 		c.urlClient.PreprocessQuery()
// 	}
// }

// func (c *Client) QueryEmbeddings(emb []int8, clusterIndex uint64) *pir.Query[matrix.Elem64] {
// 	if c.params.NumDocs == 0 {
// 		panic("Not set up")
// 	}

// 	dbIndex := c.embMap.ClusterToIndex(uint(clusterIndex))
// 	m := c.embInfo.M
// 	dim := uint64(len(emb))

// 	if m%dim != 0 {
// 		panic("Should not happen")
// 	}
// 	if dbIndex%dim != 0 {
// 		panic("Should not happen")
// 	}

// 	_, colIndex := database.Decompose(dbIndex, m)
// 	arr := matrix.Zeros[matrix.Elem64](m, 1)
// 	for j := uint64(0); j < dim; j++ {
// 		arr.AddAt(colIndex+j, 0, matrix.Elem64(emb[j]))
// 	}

// 	return c.embClient.QueryLHE(arr)
// }

// func (c *Client) QueryUrls(clusterIndex, docIndex uint64) (*pir.Query[matrix.Elem32], uint64) {
// 	if c.params.NumDocs == 0 {
// 		panic("Not set up")
// 	}

// 	dbIndex, chunkIndex, _ := c.urlMap.SubclusterToIndex(clusterIndex, docIndex)

// 	return c.urlClient.Query(dbIndex), chunkIndex
// }

// func (c *Client) getEmbeddingsAnswer(query *pir.Query[matrix.Elem64], keepConn bool, tcp string) *pir.Answer[matrix.Elem64] {
// 	ans := pir.Answer[matrix.Elem64]{}
// 	c.rpcClient = makeRPC[pir.Query[matrix.Elem64], pir.Answer[matrix.Elem64]](query, &ans, keepConn, tcp, "GetEmbeddingsAnswer", c.rpcClient)
// 	return &ans
// }

// func (c *Client) getUrlsAnswer(query *pir.Query[matrix.Elem32], keepConn bool, tcp string) *pir.Answer[matrix.Elem32] {
// 	ans := pir.Answer[matrix.Elem32]{}
// 	c.rpcClient = makeRPC[pir.Query[matrix.Elem32], pir.Answer[matrix.Elem32]](query, &ans, keepConn, tcp, "GetUrlsAnswer", c.rpcClient)
// 	return &ans
// }

// func (c *Client) getHint(keepConn bool, tcp string) *TiptoeHint {
// 	query := true
// 	hint := TiptoeHint{}
// 	c.rpcClient = makeRPC[bool, TiptoeHint](&query, &hint, keepConn, tcp, "GetHint", c.rpcClient)
// 	return &hint
// }

// func (c *Client) applyHint(ct *underhood.HintQuery, keepConn bool, tcp string) *UnderhoodAnswer {
// 	ans := UnderhoodAnswer{}
// 	c.rpcClient = makeRPC[underhood.HintQuery, UnderhoodAnswer](ct, &ans, keepConn, tcp, "ApplyHint", c.rpcClient)
// 	return &ans
// }

// func (c *Client) ReconstructEmbeddingsWithinCluster(ans *pir.Answer[matrix.Elem64], clusterIndex uint64) []uint64 {
// 	dbIndex := c.embMap.ClusterToIndex(uint(clusterIndex))
// 	rowStart, colIndex := database.Decompose(dbIndex, c.embInfo.M)
// 	rowEnd := database.FindEnd(c.embIndices, rowStart, colIndex, c.embInfo.M, c.embInfo.L, 0)

// 	vals := c.embClient.RecoverLHE(ans)

// 	res := make([]uint64, rowEnd-rowStart)
// 	at := 0
// 	for j := rowStart; j < rowEnd; j++ {
// 		res[at] = uint64(vals.Get(j, 0))
// 		at += 1
// 	}

// 	return res
// }

// func (c *Client) ReconstructUrls(answer *pir.Answer[matrix.Elem32], clusterIndex, docIndex uint64) string {
// 	dbIndex, _, _ := c.urlMap.SubclusterToIndex(clusterIndex, docIndex)
// 	rowStart, colIndex := database.Decompose(dbIndex, c.urlInfo.M)
// 	rowEnd := database.FindEnd(c.urlIndices, rowStart, colIndex, c.urlInfo.M, c.urlInfo.L, c.params.UrlBytes)

// 	vals := c.urlClient.Recover(answer)

// 	out := make([]byte, rowEnd-rowStart)
// 	for i, e := range vals[rowStart:rowEnd] {
// 		out[i] = byte(e)
// 	}

// 	if c.params.CompressUrl {
// 		res, err := corpus.Decompress(out)
// 		for err != nil {
// 			out = out[:len(out)-1]
// 			if len(out) == 0 {
// 				panic("Should not happen")
// 			}
// 			res, err = corpus.Decompress(out)
// 		}
// 		return strings.TrimRight(res, "\x00")
// 	}

// 	return strings.TrimRight(string(out), "\x00")
// }

// func makeRPC[Q QueryType, A AnsType](query *Q, reply *A, keepConn bool, tcp, rpc string, client *rpc.Client) *rpc.Client {
// 	if client == nil {
// 		client = utils.DialTCP(tcp)
// 	}
// 	utils.CallTCP(client, "Server."+rpc, query, reply)
// 	if !keepConn {
// 		client.Close()
// 		client = nil
// 	}

// 	return client
// }

package protocol

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/rpc"
	"strings"
	"time"

	"search/config"
	"search/corpus"
	"search/database"
	"search/embeddings"
	"search/utils"

	"github.com/ahenzinger/underhood/underhood"
	"github.com/fatih/color"
	"github.com/henrycg/simplepir/matrix"
	"github.com/henrycg/simplepir/pir"
)

type UnderhoodAnswer struct {
	EmbAnswer underhood.HintAnswer
	UrlAnswer underhood.HintAnswer
}

type QueryType interface {
	bool | underhood.HintQuery | pir.Query[matrix.Elem64] | pir.Query[matrix.Elem32]
}

type AnsType interface {
	TiptoeHint | UnderhoodAnswer | pir.Answer[matrix.Elem64] | pir.Answer[matrix.Elem32]
}

type Client struct {
	params corpus.Params

	embClient  *underhood.Client[matrix.Elem64]
	embInfo    *pir.DBInfo
	embMap     database.ClusterMap
	embIndices map[uint64]bool

	urlClient  *underhood.Client[matrix.Elem32]
	urlInfo    *pir.DBInfo
	urlMap     database.SubclusterMap
	urlIndices map[uint64]bool

	rpcClient      *rpc.Client
	useCoordinator bool

	stepCount int
}

func NewClient(useCoordinator bool) *Client {
	c := new(Client)
	c.useCoordinator = useCoordinator
	return c
}

func (c *Client) Free() {
	c.urlClient.Free()
	c.embClient.Free()
}

func (c *Client) NumDocs() uint64 {
	return c.params.NumDocs
}

func (c *Client) NumClusters() int {
	if len(c.embMap) > 0 {
		return len(c.embMap)
	}
	return len(c.urlMap)
}

func (c *Client) printStep(text string) {
	col := color.New(color.FgGreen).Add(color.Bold)
	col.Printf("%d) %v\n", c.stepCount, text)
	c.stepCount += 1
}

func RunClient(coordinatorAddr string, conf *config.Config) {
	color.Yellow("Setting up client...")

	c := NewClient(true /* use coordinator */)
	c.printStep("Getting metadata")
	hint := c.getHint(true /* keep conn */, coordinatorAddr)
	c.Setup(hint)
	logHintSize(hint)

	in, out := embeddings.SetupEmbeddingProcess(c.NumClusters(), conf)
	col := color.New(color.FgYellow).Add(color.Bold)

	for {
		c.stepCount = 1
		c.printStep("Running client preprocessing")
		perf := c.preprocessRound(coordinatorAddr, true /* verbose */, true /* keep conn */)

		col.Printf("Enter private search query: ")
		text := utils.ReadLineFromStdin()
		fmt.Printf("\n\n")
		if (strings.TrimSpace(text) == "") || (strings.TrimSpace(text) == "quit") {
			break
		}
		c.runRound(perf, in, out, text, coordinatorAddr, true /* verbose */, true /* keep conn */)
	}

	if c.rpcClient != nil {
		c.rpcClient.Close()
	}
	in.Close()
	out.Close()
}

func (c *Client) preprocessRound(coordinatorAddr string, verbose, keepConn bool) Perf {
	var p Perf

	// Perform preprocessing
	start := time.Now()
	ct := c.PreprocessQuery()

	networkingStart := time.Now()
	offlineAns := c.applyHint(ct, keepConn, coordinatorAddr)
	fmt.Println(offlineAns.EmbAnswer)
	p.tOffline, p.upOffline, p.downOffline = logOfflineStats(c.NumDocs(), networkingStart, ct, offlineAns)

	c.ProcessHintApply(offlineAns)

	p.clientPreproc = time.Since(start).Seconds()

	if verbose {
		fmt.Printf("\tPreprocessing complete -- %fs\n\n", p.clientPreproc)
	}

	return p
}

func (c *Client) runRound(p Perf, in io.WriteCloser, out io.ReadCloser,
	text, coordinatorAddr string, verbose, keepConn bool) Perf {
	y := color.New(color.FgYellow, color.Bold)
	fmt.Printf("Executing query \"%s\"\n", y.Sprintf(text))

	// Build embeddings query
	start := time.Now()
	if verbose {
		c.printStep("Generating embedding of the query")
	}

	var query struct {
		Cluster_index uint64
		Emb           []int8
	}

	io.WriteString(in, text+"\n")                               // send query to embedding process
	if err := json.NewDecoder(out).Decode(&query); err != nil { // get back embedding + cluster
		log.Printf("Did you remember to set up your python venv?")
		panic(err)
	}

	if query.Cluster_index >= uint64(c.NumClusters()) {
		panic("Should not happen")
	}

	if verbose {
		c.printStep(fmt.Sprintf("Building PIR query for cluster %d", query.Cluster_index))
	}

	embQuery := c.QueryEmbeddings(query.Emb, query.Cluster_index)
	p.clientSetup = time.Since(start).Seconds()

	// Send embeddings query to server
	if verbose {
		c.printStep("Sending SimplePIR query to server")
	}
	networkingStart := time.Now()
	embAns := c.getEmbeddingsAnswer(embQuery, true /* keep conn */, coordinatorAddr)
	p.t1, p.up1, p.down1 = logStats(c.params.NumDocs, networkingStart, embQuery, embAns)

	// Recover document and URL chunk to query for
	c.printStep("Decrypting server answer")
	embDec := c.ReconstructEmbeddingsWithinCluster(embAns, query.Cluster_index)
	scores := embeddings.SmoothResults(embDec, c.embInfo.P())
	indicesByScore := utils.SortByScores(scores)
	docIndex := indicesByScore[0]

	if verbose {
		fmt.Printf("\tDoc %d within cluster %d has the largest inner product with our query\n",
			docIndex, query.Cluster_index)
		c.printStep(fmt.Sprintf("Building PIR query for url/title of doc %d in cluster %d",
			docIndex, query.Cluster_index))
	}

	// Build URL query
	urlQuery, retrievedChunk := c.QueryUrls(query.Cluster_index, docIndex)

	// Send URL query to server
	if verbose {
		c.printStep(fmt.Sprintf("Sending PIR query to server for chunk %d", retrievedChunk))
	}
	networkingStart = time.Now()
	urlAns := c.getUrlsAnswer(urlQuery, keepConn, coordinatorAddr)
	p.t2, p.up2, p.down2 = logStats(c.params.NumDocs, networkingStart, urlQuery, urlAns)

	// Recover URLs of top 10 docs in chunk
	urls := c.ReconstructUrls(urlAns, query.Cluster_index, docIndex)
	if verbose {
		c.printStep("Reconstructed PIR answers.")
		fmt.Printf("\tThe top 10 retrieved urls are:\n")
	}

	j := 1
	for at := 0; at < len(indicesByScore); at++ {
		if scores[at] == 0 {
			break
		}

		doc := indicesByScore[at]
		_, chunk, index := c.urlMap.SubclusterToIndex(query.Cluster_index, doc)

		if chunk == retrievedChunk {
			if verbose {
				fmt.Printf("\t% 3d) [score %s] %s\n", j,
					color.YellowString(fmt.Sprintf("% 4d", scores[at])),
					color.BlueString(corpus.GetIthUrl(urls, index)))
			}
			j += 1
			if j > 10 {
				break
			}
		}
	}

	p.clientTotal = time.Since(start).Seconds()
	fmt.Printf("\tAnswered in:\n\t\t%v (preproc)\n\t\t%v (client)\n\t\t%v (round 1)\n\t\t%v (round 2)\n\t\t%v (total)\n---\n",
		p.clientPreproc, p.clientSetup, p.t1, p.t2, p.clientTotal)

	return p
}

func (c *Client) Setup(hint *TiptoeHint) {
	if hint == nil {
		panic("Hint is empty")
	}

	if hint.CParams.NumDocs == 0 {
		panic("Corpus is empty")
	}

	c.params = hint.CParams
	c.embInfo = &hint.EmbeddingsHint.Info
	c.urlInfo = &hint.UrlsHint.Info

	if hint.ServeEmbeddings {
		if hint.EmbeddingsHint.IsEmpty() {
			panic("Embeddings hint is empty")
		}

		c.embClient = utils.NewUnderhoodClient(&hint.EmbeddingsHint)

		c.embMap = hint.EmbeddingsIndexMap
		c.embIndices = make(map[uint64]bool)
		for _, v := range c.embMap {
			c.embIndices[v] = true
		}

		fmt.Printf("\tEmbeddings client: %s\n", utils.PrintParams(c.embInfo))
	}

	if hint.ServeUrls {
		if hint.UrlsHint.IsEmpty() {
			panic("Urls hint is empty")
		}

		c.urlClient = utils.NewUnderhoodClient(&hint.UrlsHint)

		c.urlMap = hint.UrlsIndexMap
		c.urlIndices = make(map[uint64]bool)
		for _, vals := range c.urlMap {
			for _, v := range vals {
				c.urlIndices[v.Index()] = true
			}
		}

		fmt.Printf("\tURL client: %s\n", utils.PrintParams(c.urlInfo))
	}

	if hint.ServeUrls && hint.ServeEmbeddings &&
		(len(c.urlMap) != len(c.embMap)) {
		fmt.Printf("Both maps don't have the same length: %d %d\n", len(c.urlMap), len(c.embMap))
		//    panic("Both maps don't have same length.")
	}
}

func (c *Client) PreprocessQuery() *underhood.HintQuery {
	if c.params.NumDocs == 0 {
		panic("Not set up")
	}

	if c.embClient != nil {
		hintQuery := c.embClient.HintQuery()
		if c.urlClient != nil {
			c.urlClient.CopySecret(c.embClient)
		}
		return hintQuery
	} else if c.urlClient != nil {
		return c.urlClient.HintQuery()
	} else {
		panic("Should not happen")
	}
}

func (c *Client) ProcessHintApply(ans *UnderhoodAnswer) {
	if c.embClient != nil {
		c.embClient.HintRecover(&ans.EmbAnswer)
		c.embClient.PreprocessQueryLHE()
	}

	if c.urlClient != nil {
		c.urlClient.HintRecover(&ans.UrlAnswer)
		c.urlClient.PreprocessQuery()
	}
}

func (c *Client) QueryEmbeddings(emb []int8, clusterIndex uint64) *pir.Query[matrix.Elem64] {
	if c.params.NumDocs == 0 {
		panic("Not set up")
	}

	dbIndex := c.embMap.ClusterToIndex(uint(clusterIndex))
	m := c.embInfo.M
	dim := uint64(len(emb))

	if m%dim != 0 {
		panic("Should not happen")
	}
	if dbIndex%dim != 0 {
		panic("Should not happen")
	}

	_, colIndex := database.Decompose(dbIndex, m)
	arr := matrix.Zeros[matrix.Elem64](m, 1)
	for j := uint64(0); j < dim; j++ {
		arr.AddAt(colIndex+j, 0, matrix.Elem64(emb[j]))
	}

	return c.embClient.QueryLHE(arr)
}

func (c *Client) QueryUrls(clusterIndex, docIndex uint64) (*pir.Query[matrix.Elem32], uint64) {
	if c.params.NumDocs == 0 {
		panic("Not set up")
	}

	dbIndex, chunkIndex, _ := c.urlMap.SubclusterToIndex(clusterIndex, docIndex)

	return c.urlClient.Query(dbIndex), chunkIndex
}

func (c *Client) ReconstructEmbeddings(answer *pir.Answer[matrix.Elem64],
	clusterIndex uint64) uint64 {
	vals := c.embClient.RecoverLHE(answer)

	dbIndex := c.embMap.ClusterToIndex(uint(clusterIndex))
	rowIndex, _ := database.Decompose(dbIndex, c.embInfo.M)
	res := vals.Get(rowIndex, 0)

	return uint64(res)
}

func (c *Client) ReconstructEmbeddingsWithinCluster(answer *pir.Answer[matrix.Elem64],
	clusterIndex uint64) []uint64 {
	dbIndex := c.embMap.ClusterToIndex(uint(clusterIndex))
	rowStart, colIndex := database.Decompose(dbIndex, c.embInfo.M)
	rowEnd := database.FindEnd(c.embIndices, rowStart, colIndex,
		c.embInfo.M, c.embInfo.L, 0)

	vals := c.embClient.RecoverLHE(answer)

	res := make([]uint64, rowEnd-rowStart)
	at := 0
	for j := rowStart; j < rowEnd; j++ {
		res[at] = uint64(vals.Get(j, 0))
		at += 1
	}

	return res
}

func (c *Client) ReconstructUrls(answer *pir.Answer[matrix.Elem32],
	clusterIndex, docIndex uint64) string {
	dbIndex, _, _ := c.urlMap.SubclusterToIndex(clusterIndex, docIndex)
	rowStart, colIndex := database.Decompose(dbIndex, c.urlInfo.M)
	rowEnd := database.FindEnd(c.urlIndices, rowStart, colIndex,
		c.urlInfo.M, c.urlInfo.L, c.params.UrlBytes)

	vals := c.urlClient.Recover(answer)

	out := make([]byte, rowEnd-rowStart)
	for i, e := range vals[rowStart:rowEnd] {
		out[i] = byte(e)
	}

	if c.params.CompressUrl {
		res, err := corpus.Decompress(out)
		for err != nil {
			out = out[:len(out)-1]
			if len(out) == 0 {
				panic("Should not happen")
			}
			res, err = corpus.Decompress(out)
		}
		return strings.TrimRight(res, "\x00")
	}

	return strings.TrimRight(string(out), "\x00")
}

func makeRPC[Q QueryType, A AnsType](query *Q, reply *A, useCoordinator, keepConn bool,
	tcp, rpc string, client *rpc.Client) *rpc.Client {
	if !useCoordinator {
		conn := utils.DialTCP(tcp)
		utils.CallTCP(conn, "Server."+rpc, query, reply)
		conn.Close()
	} else {
		if client == nil {
			client = utils.DialTLS(tcp)
		}

		utils.CallTLS(client, "Coordinator."+rpc, query, reply)

		if !keepConn {
			client.Close()
			client = nil
		}
	}

	return client
}

func (c *Client) getHint(keepConn bool, tcp string) *TiptoeHint {
	query := true
	hint := TiptoeHint{}
	c.rpcClient = makeRPC[bool, TiptoeHint](&query, &hint, c.useCoordinator, keepConn,
		tcp, "GetHint", c.rpcClient)
	return &hint
}

func (c *Client) applyHint(ct *underhood.HintQuery,
	keepConn bool,
	tcp string) *UnderhoodAnswer {
	ans := UnderhoodAnswer{}
	c.rpcClient = makeRPC[underhood.HintQuery, UnderhoodAnswer](ct, &ans,
		c.useCoordinator, keepConn,
		tcp, "ApplyHint",
		c.rpcClient)
	return &ans
}

func (c *Client) getEmbeddingsAnswer(query *pir.Query[matrix.Elem64],
	keepConn bool,
	tcp string) *pir.Answer[matrix.Elem64] {
	ans := pir.Answer[matrix.Elem64]{}
	c.rpcClient = makeRPC[pir.Query[matrix.Elem64], pir.Answer[matrix.Elem64]](query, &ans,
		c.useCoordinator, keepConn,
		tcp, "GetEmbeddingsAnswer",
		c.rpcClient)
	return &ans
}

func (c *Client) getUrlsAnswer(query *pir.Query[matrix.Elem32],
	keepConn bool,
	tcp string) *pir.Answer[matrix.Elem32] {
	ans := pir.Answer[matrix.Elem32]{}
	c.rpcClient = makeRPC[pir.Query[matrix.Elem32], pir.Answer[matrix.Elem32]](query, &ans,
		c.useCoordinator, keepConn,
		tcp, "GetUrlsAnswer",
		c.rpcClient)
	return &ans
}

func (c *Client) closeConn() {
	if c.rpcClient != nil {
		c.rpcClient.Close()
		c.rpcClient = nil
	}
}
