package main

import (
	"log"
	"os"
	"path"

	"KVDatastore/server"
	"KVDatastore/raft"
)


type config struct {
	id       string
	httpPort string
	raftPort string
}

func getConfig() config {
	cfg := config{}
	for i, arg := range os.Args[1:] {
		if arg == "--node-id" {
			cfg.id = os.Args[i+2]
			i++
			continue
		}

		if arg == "--http-port" {
			cfg.httpPort = os.Args[i+2]
			i++
			continue
		}

		if arg == "--raft-port" {
			cfg.raftPort = os.Args[i+2]
			i++
			continue
		}
	}

	if cfg.id == "" {
		log.Fatal("Missing required parameter: --node-id")
	}

	if cfg.raftPort == "" {
		log.Fatal("Missing required parameter: --raft-port")
	}

	if cfg.httpPort == "" {
		log.Fatal("Missing required parameter: --http-port")
	}

	return cfg
}


func main(){
	cfg := getConfig()

	dataDir := "data"
	kf := &raft.Fsm{}
	r, err := raft.SetupRaft(path.Join(dataDir, "raft"+cfg.id), cfg.id, "localhost:"+cfg.raftPort, kf)
	if err != nil {
		log.Fatal(err)
	}

	hs := server.NewServer(r)
	httpServerAddr := ":" + cfg.httpPort
	log.Printf("HTTP server listening on %s\n", httpServerAddr)
	if err := hs.ListenAndServe(httpServerAddr); err != nil {
		log.Fatalf("HTTP server error: %v", err)
	}
}