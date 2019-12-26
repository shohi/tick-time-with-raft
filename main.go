package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/juanpabloaj/rafttick/server"
	"github.com/juanpabloaj/rafttick/store"
)

const (
	DefaultHTTPAddr = ":9001"
	DefaultRaftAddr = ":6001"
)

var (
	httpAddr string
	joinAddr string
	nodeID   string
	raftAddr string
	raftDir  string
)

func init() {
	flag.StringVar(&httpAddr, "addr", DefaultHTTPAddr, "Set the HTTP bind address")
	flag.StringVar(&joinAddr, "join", "", "Set join address, if any")

	flag.StringVar(&nodeID, "id", "", "raft Node ID")
	flag.StringVar(&raftAddr, "raft-addr", DefaultRaftAddr, "Set Raft bind address")
	flag.StringVar(&raftDir, "raft-dir", "", "raft dir")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] \n", os.Args[0])
		flag.PrintDefaults()
	}
}

func main() {
	flag.Parse()

	if nodeID == "" {
		log.Fatal("node id must be set")
	}

	if raftDir == "" {
		log.Fatal("raft dir must be set")
	}

	os.MkdirAll(raftDir, 0700)

	s := store.NewStore()

	s.RaftDir = raftDir
	s.RaftBind = raftAddr

	if err := s.Open(joinAddr == "", nodeID); err != nil {
		log.Fatalf("%v", err)
	}
	s.Start()

	service := server.NewService(httpAddr, *s)
	if err := service.Start(); err != nil {
		log.Fatalf("failed to start HTTP service: %s", err.Error())
	}

	if joinAddr != "" {
		if err := join(joinAddr, raftAddr, nodeID); err != nil {
			log.Fatalf("failed to join node at %s: %s", joinAddr, err.Error())
		}
	}

	log.Println("started successfully ...")

	terminate := make(chan os.Signal, 1)
	signal.Notify(terminate, os.Interrupt)
	<-terminate
	log.Println("exiting ...")
}

func join(joinAddr, raftAddr, nodeID string) error {
	b, err := json.Marshal(map[string]string{"addr": raftAddr, "id": nodeID})
	if err != nil {
		return err
	}

	log.Printf("join cluster => [%v]", string(b))

	urlPath := fmt.Sprintf("http://%s/join?id=%s&addr=%s",
		joinAddr, nodeID, raftAddr)

	resp, err := http.Post(urlPath, "application-type/json", nil)
	if err != nil {
		return err
	}

	if resp != nil && resp.Body != nil {
		resp.Body.Close()
	}

	return nil
}
