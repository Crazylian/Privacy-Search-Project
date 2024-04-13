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

// func (s *Server) ApplyHint(ct *underhood.HintQuery, out *UnderhoodAnswer) error {
// 	if s.hint.ServeEmbeddings {
// 		if s.embHintServer == nil {
// 			s.preprocessEmbHint()
// 		}
// 		out.EmbAnswer = *s.embHintServer.HintAnswer(ct)

// 		if s.hint.ServeUrls {
// 			toDrop := int(s.hint.EmbeddingsHint.Info.Params.N - s.hint.UrlsHint.Info.Params.N)
// 			*ct = (*ct)[:len(*ct)-toDrop]
// 		}
// 	}

// 	if s.hint.ServeUrls {
// 		if s.urlHintServer == nil {
// 			s.preprocessUrlHint()
// 		}
// 		out.UrlAnswer = *s.urlHintServer.HintAnswer(ct)
// 	}

// 	return nil
// }

// func (s *Server) preprocessEmbHint() {
// 	// Decompose hint
// 	s.embHintServer = underhood.NewServerHintOnly(&s.hint.EmbeddingsHint.Hint)

// 	// Drop hint contents that shouldn't be sent back
// 	rows := s.hint.EmbeddingsHint.Hint.Rows()
// 	s.hint.EmbeddingsHint.Hint.DropLastrows(rows)
// }

// func (s *Server) preprocessUrlHint() {
// 	// Decompose hint
// 	s.urlHintServer = underhood.NewServerHintOnly(&s.hint.UrlsHint.Hint)

// 	// Drop hint contents that shouldn't be sent back
// 	rows := s.hint.UrlsHint.Hint.Rows()
// 	s.hint.UrlsHint.Hint.DropLastrows(rows)
// }
