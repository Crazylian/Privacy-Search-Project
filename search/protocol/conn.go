package protocol

import (
	"github.com/henrycg/simplepir/matrix"
	"github.com/henrycg/simplepir/pir"
)

func (s *Server) GetHint(request bool, hint *TiptoeHint) error {
	*hint = *s.hint
	return nil
}

func (s *Server) GetEmbeddingsAnswer(query *pir.Query[matrix.Elem64], ans *pir.Answer[matrix.Elem64]) error {
	*ans = *s.embeddingsServer.Answer(query)
	return nil
}

func (s *Server) GetUrlsAnswer(query *pir.Query[matrix.Elem32], ans *pir.Answer[matrix.Elem32]) error {
	*ans = *s.urlsServer.Answer(query)
	return nil
}
