package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"search/config"
	"search/embeddings"
	"search/protocol"
	"search/utils"
)

// Where the corpus is stored
// var preamble = flag.String("preamble", "/home/ubuntu", "Preamble")

func printUsage() {
	fmt.Println("Usage:\n\"go run . all-servers\" or\n\"go run . client coordinator-ip\" or\n\"go run . coordinator numEmbServers numUrlServers ip1 ip2 ...\" or\n\"go run . emb-server index\" or\n\"go run . url-server index\" or\n\"go run . client-latency coordinator-ip\" or\n\"go run . client-tput-embed coordinator-ip\" or\n\"go run . client-tput-url coordinator-ip\" or\n\"go run . client-tput-offline coordinator-ip\"")
}

func main() {
	preamble := flag.String("preamble", "/home/lianzheng", "Preamble")
	flag.Parse()
	coordinatorIP := "0.0.0.0"
	if len(os.Args) < 2 {
		return
	}

	conf := config.MakeConfig(*preamble + "/data")

	if os.Args[1] == "emb-server" {
		_, embAddrs, _ := protocol.NewEmbeddingServers(conf.EMBEDDINGS_CLUSTERS_PER_SERVER(), conf.DEFAULT_EMBEDDINGS_HINT_SZ(), true, false, true, conf)
		fmt.Println("Set up embedding server")
		fmt.Println(embAddrs)

	} else if os.Args[1] == "url-server" {
		_, urlAddrs, _ := protocol.NewUrlServers(conf.URL_CLUSTERS_PER_SERVER(), conf.DEFAULT_URL_HINT_SZ(), true, false, true, conf)
		fmt.Println("Set up url server")
		fmt.Println(urlAddrs)

	} else if os.Args[1] == "client" {
		if len(os.Args) >= 3 {
			coordinatorIP = os.Args[2]
		}

		protocol.RunClient(utils.RemoteAddr(coordinatorIP, utils.EmbServerPort), conf)

		// protocol.
	} else if os.Args[1] == "test" {
		in, out := embeddings.SetupEmbeddingProcess(1280, conf)
		var query struct {
			Cluster_index uint64
			Emb           []int8
		}
		io.WriteString(in, "text"+"\n")                             // send query to embedding process
		if err := json.NewDecoder(out).Decode(&query); err != nil { // get back embedding + cluster
			panic(err)
		}
		fmt.Println(query)
	}
}