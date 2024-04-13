package protocol

import (
	"bytes"
	"encoding/gob"
	"os"
)

type TiptoeServer interface {
	Server
}

func (s *Server) GobEncode() ([]byte, error) {
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	err := enc.Encode(s.hint)
	if err != nil {
		return buf.Bytes(), err
	}

	if s.hint.ServeEmbeddings {
		err = enc.Encode(s.embeddingsServer)
		if err != nil {
			return buf.Bytes(), err
		}
	}

	if s.hint.ServeUrls {
		err = enc.Encode(s.urlsServer)
		if err != nil {
			return buf.Bytes(), err
		}
	}

	return buf.Bytes(), err
}

func (s *Server) GobDecode(buf []byte) error {
	b := bytes.NewBuffer(buf)
	dec := gob.NewDecoder(b)
	err := dec.Decode(&s.hint)
	if err != nil {
		return err
	}

	if s.hint.ServeEmbeddings {
		err := dec.Decode(&s.embeddingsServer)
		if err != nil {
			return err
		}
	}

	if s.hint.ServeUrls {
		err := dec.Decode(&s.urlsServer)
		if err != nil {
			return err
		}
	}

	return nil
}

// func DumpStateToFile[S TiptoeServer](s *S, filename string) {
func DumpStateToFile[S TiptoeServer](s *S, filename string) {
	f, err := os.Create(filename) // deletes prior contents
	if err != nil {
		panic(err)
	}
	defer f.Close()

	enc := gob.NewEncoder(f)
	err = enc.Encode(s)
	if err != nil {
		panic(err)
	}
}

// func LoadStateFromFile[S TiptoeServer](s *S, filename string) {
func LoadStateFromFile[S TiptoeServer](s *S, filename string) {
	f, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	dec := gob.NewDecoder(f)
	err = dec.Decode(&s)
	if err != nil {
		panic(err)
	}
}
