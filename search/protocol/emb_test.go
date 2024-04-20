package protocol

import (
	"fmt"
	"os"
	"runtime/pprof"
	"search/config"
	"search/corpus"
	"search/embeddings"
	"search/utils"
	"testing"
	"time"

	"github.com/ahenzinger/underhood/underhood"
	"github.com/henrycg/simplepir/matrix"
	"github.com/henrycg/simplepir/pir"
)

var s *Server

var conf *config.Config

const embNumQueries = 20

type testServer struct {
	emb *underhood.Server[matrix.Elem64]
	url *underhood.Server[matrix.Elem32]
}

func setupTestServer(h *TiptoeHint) *testServer {
	out := new(testServer)
	if !h.EmbeddingsHint.IsEmpty() {
		out.emb = underhood.NewServerHintOnly(&h.EmbeddingsHint.Hint)
	}

	if !h.UrlsHint.IsEmpty() {
		out.url = underhood.NewServerHintOnly(&h.UrlsHint.Hint)
	}

	return out
}

func applyHint(s *testServer, hq *underhood.HintQuery) *UnderhoodAnswer {
	out := new(UnderhoodAnswer)
	if s.emb != nil {
		out.EmbAnswer = *s.emb.HintAnswer(hq)
	}

	if s.url != nil {
		out.UrlAnswer = *s.url.HintAnswer(hq)
	}

	return out
}

func testRecoverCluster(s *Server, corp *corpus.Corpus) {
	c := NewClient()

	var h TiptoeHint
	s.GetHint(true, &h)
	c.Setup(&h)
	logHintSize(&h)

	p := h.EmbeddingsHint.Info.P()
	tserv := setupTestServer(&h)

	for iter := 0; iter < embNumQueries; iter++ {
		ct := c.PreprocessQuery()

		offlineStart := time.Now()
		uAns := applyHint(tserv, ct)
		logOfflineStats(c.NumDocs(), offlineStart, ct, uAns)
		c.ProcessHintApply(uAns)

		i := utils.RandomIndex(c.NumClusters())
		emb := embeddings.RandomEmbedding(c.params.EmbeddingSlots, (1 << (c.params.SlotBits - 1)))
		query := c.QueryEmbeddings(emb, i)

		start := time.Now()
		var ans pir.Answer[matrix.Elem64]
		s.GetEmbeddingsAnswer(query, &ans)
		logStats(c.NumDocs(), start, query, &ans)

		dec := c.ReconstructEmbeddingsWithinCluster(&ans, i)
		checkAnswers(dec, uint(i), p, emb, corp)
	}
}

func TestEmbeddingsRealData(t *testing.T) {
	s = Newserver()
	conf = config.MakeConfig("/home/ubuntu/data")
	f, _ := os.Create("emb_test.prof")
	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()

	corp := corpus.ReadEmbeddingsTxt(0, 10, conf)
	s.PreprocessEmbeddingsFromCorpus(corp, 25 /* hint size in MB */, conf)
	// k.Setup(1, 0, []string{serverTcp}, false, conf)

	fmt.Printf("Running embedding queries (over %d-doc real corpus)\n", corp.GetNumDocs())

	testRecoverCluster(s, corp)
}
