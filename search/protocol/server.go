package protocol

import (
	"fmt"
	"net/rpc"
	"search/config"
	"search/corpus"
	"search/database"
	"search/utils"

	"github.com/ahenzinger/underhood/underhood"
	"github.com/henrycg/simplepir/matrix"
	"github.com/henrycg/simplepir/pir"
	"github.com/henrycg/simplepir/rand"
)

// Number of servers to launch at once
const BLOCK_SZ = 10

type TiptoeHint struct {
	CParams corpus.Params

	ServeEmbeddings    bool
	EmbeddingsHint     utils.PIR_hint[matrix.Elem64]
	EmbeddingsIndexMap database.ClusterMap

	ServeUrls    bool
	UrlsHint     utils.PIR_hint[matrix.Elem32]
	UrlsIndexMap database.SubclusterMap
}

type Server struct {
	hint             *TiptoeHint
	embeddingsServer *pir.Server[matrix.Elem64]
	urlsServer       *pir.Server[matrix.Elem32]

	embHintServer *underhood.Server[matrix.Elem64]
	urlHintServer *underhood.Server[matrix.Elem32]
}

func Newserver() *Server {
	s := new(Server)
	return s
}

func (s *Server) PreprocessEmbeddingsFromCorpus(c *corpus.Corpus, hintSz uint64, conf *config.Config) {
	embeddings_seed := rand.RandomPRGKey()
	s.preprocessEmbeddingsSeeded(c, embeddings_seed, hintSz, conf)
}

func (s *Server) preprocessEmbeddingsSeeded(c *corpus.Corpus, seed *rand.PRGKey, hintSz uint64, conf *config.Config) {
	// fmt.Printf("Preprocessing a corpus of %d embeddings of length %d\n", c.GetNumDocs(), c.GetEmbeddingSlots())
	fmt.Printf("Preprocessing a corpus of %d embeddings of length %d\n", c.GetEmbeddingSlots(), c.GetNumDocs())
	db, indexMap := database.BuildEmbeddingsDatabase(c, seed, hintSz, conf)
	s.embeddingsServer = pir.NewServerSeed(db, seed)

	s.hint = new(TiptoeHint)
	s.hint.ServeEmbeddings = true
	s.hint.CParams = c.GetParams()
	s.hint.EmbeddingsHint.Hint = *s.embeddingsServer.Hint()
	s.hint.EmbeddingsHint.Info = *s.embeddingsServer.DBInfo()
	s.hint.EmbeddingsHint.Seeds = []rand.PRGKey{*seed}
	s.hint.EmbeddingsHint.Offsets = []uint64{s.hint.EmbeddingsHint.Info.M}
	s.hint.EmbeddingsIndexMap = indexMap

	max_inner_prod := 2 * (1 << (2*c.GetSlotBits() - 2)) * c.GetEmbeddingSlots()
	if s.embeddingsServer.Params().P < max_inner_prod {
		fmt.Printf("%d < %d\n", s.embeddingsServer.Params().P, max_inner_prod)
		panic("Parameters not supported. Inner products may wrap around.")
	}

	fmt.Println("done")
}

func (s *Server) PreprocessUrlsFromCorpus(c *corpus.Corpus, hintSz uint64) {
	urls_seed := rand.RandomPRGKey()
	s.preprocessUrlsSeeded(c, urls_seed, hintSz)
}

func (s *Server) preprocessUrlsSeeded(c *corpus.Corpus, seed *rand.PRGKey, hintSz uint64) {
	fmt.Printf("Preprocessing a corpus of %d urls in chunks of length <= %d\n", c.GetNumDocs(), c.GetUrlBytes())

	db, indexMap := database.BuildUrlsDatabase(c, seed, hintSz)
	s.urlsServer = pir.NewServerSeed(db, seed)

	s.hint = new(TiptoeHint)
	s.hint.ServeUrls = true
	s.hint.CParams = c.GetParams()
	s.hint.UrlsHint.Hint = *s.urlsServer.Hint()
	s.hint.UrlsHint.Info = *s.urlsServer.DBInfo()
	s.hint.UrlsHint.Seeds = []rand.PRGKey{*seed}
	s.hint.UrlsHint.Offsets = []uint64{s.hint.UrlsHint.Info.M}
	s.hint.UrlsIndexMap = indexMap

	fmt.Println("done")
}

// func (s *Server) GetHint(request bool, hint *TiptoeHint) error {
// 	*hint = *s.hint
// 	return nil
// }

// func (s *Server) GetEmbeddingsAnswer(query *pir.Query[matrix.Elem64],
// 	ans *pir.Answer[matrix.Elem64]) error {
// 	*ans = *s.embeddingsServer.Answer(query)
// 	return nil
// }

// func (s *Server) GetUrlsAnswer(query *pir.Query[matrix.Elem32],
// 	ans *pir.Answer[matrix.Elem32]) error {
// 	*ans = *s.urlsServer.Answer(query)
// 	return nil
// }

func launchServers(corpusSetup func() *corpus.Corpus, serverSetup func(*Server, *corpus.Corpus)) (*Server, *corpus.Corpus) {
	server := Newserver()
	corpus := corpusSetup()
	serverSetup(server, corpus)
	fmt.Println("Task Finished")

	return server, corpus
}

func launchServersFromLogs(logs string, corpusSetup func() *corpus.Corpus, serverSetup func(*Server, *corpus.Corpus), wantCorpus bool) (*Server, *corpus.Corpus) {
	fmt.Println("Launch Servers From Logs...")
	var server *Server
	var corpus *corpus.Corpus
	if utils.FileExists(logs) {
		fmt.Printf("File %s exists ...\n", logs)
		// 是否需要读取语料
		if wantCorpus {
			corpus = corpusSetup()
		}
		// 设置服务器
		server = Newserver()
		LoadStateFromFile(server, logs)
	} else {
		// 生成语料，设置服务器，并顺序写入文件
		server = Newserver()
		corpus = corpusSetup()
		serverSetup(server, corpus)
		DumpStateToFile(server, logs)
	}

	fmt.Println("Task Finished")
	return server, corpus
}

func NewEmbeddingServers(clustersPerServer int, hintSz uint64, log, wantCorpus, serve bool, conf *config.Config) (*Server, string, *corpus.Corpus) {
	fmt.Println("Reading corpus...")
	var servers *Server
	var corpuses *corpus.Corpus
	var addrs string

	serverSetup := func(s *Server, c *corpus.Corpus) {
		s.PreprocessEmbeddingsFromCorpus(c, hintSz, conf)
	}

	corpusSetup := func() *corpus.Corpus {
		return corpus.ReadEmbeddingsTxt(0, clustersPerServer, conf)
	}

	if !log {
		s, c := launchServers(corpusSetup, serverSetup)
		servers = s
		corpuses = c
	} else {
		logs := conf.EmbeddingServerLog(0)
		s, c := launchServersFromLogs(logs, corpusSetup, serverSetup, wantCorpus)
		servers = s
		corpuses = c
	}

	if serve {
		addrs = Serve(servers, utils.EmbServerPort)
	}

	return servers, addrs, corpuses
}

func NewUrlServers(clustersPerServer int, hintSz uint64, log, wantCorpus, serve bool, conf *config.Config) (*Server, string, *corpus.Corpus) {
	fmt.Println("Reading URL corpus...")
	var servers *Server
	var corpuses *corpus.Corpus
	var addrs string

	serverSetup := func(s *Server, c *corpus.Corpus) {
		s.PreprocessUrlsFromCorpus(c, hintSz)
	}
	corpusSetup := func() *corpus.Corpus {
		return corpus.ReadUrlsTxt(0, clustersPerServer, conf)
	}

	if !log {
		s, c := launchServers(corpusSetup, serverSetup)
		servers = s
		corpuses = c
	} else {
		logs := conf.UrlServerlog(0)
		s, c := launchServersFromLogs(logs, corpusSetup, serverSetup, wantCorpus)
		servers = s
		corpuses = c
	}

	if serve {
		addrs = Serve(servers, utils.UrlServerPort)
	}

	return servers, addrs, corpuses

}

func (s *Server) Serve(port int) {
	rs := rpc.NewServer()
	rs.Register(s)
	utils.ListenAndServeTCP(rs, port)
}

func Serve(servers *Server, port int) string {
	addrs := utils.LocalAddr(port)
	go servers.Serve(port)
	// rs := rpc.NewServer()
	// rs.Register(servers)
	// utils.ListenAndServeTCP(rs, port)
	return addrs
}
