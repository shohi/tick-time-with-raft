package main

import (
	"flag"
	"fmt"
	"log"
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

func validate() {
	if nodeID == "" {
		log.Fatal("node id must be set")
	}

	if raftDir == "" {
		log.Fatal("raft dir must be set")
	}

	os.MkdirAll(raftDir, 0700)
}

func createStore() store.Store {

	s := store.NewStoreWithOptions(store.Options{
		RaftDir:   raftDir,
		RaftBind:  raftAddr,
		ServerID:  nodeID,
		Bootstrap: joinAddr == "",
	})

	if err := s.Open(); err != nil {
		log.Fatalf("%v", err)
	}
	s.Start()

	return *s
}

func createServerAndJoin(s store.Store) *server.Service {
	service := server.NewService(server.Options{
		Addr:  httpAddr,
		Store: s,
	})
	if err := service.Start(); err != nil {
		log.Fatalf("failed to start HTTP service: %s", err.Error())
	}

	if err := server.Join(joinAddr, raftAddr, nodeID); err != nil {
		log.Fatalf("failed to join node at %s: %s", joinAddr, err.Error())

	}

	return service
}

func main() {
	flag.Parse()
	validate()

	_ = createServerAndJoin(createStore())
	log.Println("started successfully ...")

	terminate := make(chan os.Signal, 1)
	signal.Notify(terminate, os.Interrupt)
	<-terminate
	log.Println("exiting ...")
}
