package main

import (
	"log"
	"os"
	"path"
	"sync"

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
		if arg == "node_id" {
			cfg.id = os.Args[i+2]
			i++
			continue
		}

		if arg == "http_port" {
			cfg.httpPort = os.Args[i+2]
			i++
			continue
		}

		if arg == "raft_port" {
			cfg.raftPort = os.Args[i+2]
			i++
			continue
		}
	}

	if cfg.id == "" {
		log.Fatal("Missing required parameter: node_id")
	}

	if cfg.raftPort == "" {
		log.Fatal("Missing required parameter: raft_port")
	}

	if cfg.httpPort == "" {
		log.Fatal("Missing required parameter: http_port")
	}

	return cfg
}


func main(){
	cfg := getConfig()

	dataDir := "data"
	db := &sync.Map{}
	kf := raft.NewFsm(db)
	r, err := raft.SetupRaft(path.Join(dataDir, "raft"+cfg.id), cfg.id, "localhost:"+cfg.raftPort, kf)
	if err != nil {
		log.Fatal(err)
	}

	hs := server.NewServer(r, db)
	httpServerAddr := ":" + cfg.httpPort
	log.Printf("HTTP server listening on %s\n", httpServerAddr)
	if err := hs.ListenAndServe(httpServerAddr); err != nil {
		log.Fatalf("HTTP server error: %v", err)
	}
}