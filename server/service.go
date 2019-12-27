package server

import (
	"fmt"
	"log"
	"net/http"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/juanpabloaj/rafttick/store"
)

type Service struct {
	addr   string
	store  store.Store
	stopCh chan struct{}
}

func NewService(addr string, store store.Store) *Service {
	return &Service{
		addr:   addr,
		store:  store,
		stopCh: make(chan struct{}),
	}
}

func (s *Service) Start() error {
	router := mux.NewRouter()

	router.HandleFunc("/join", s.handleJoin)
	router.HandleFunc("/remove", s.handleRemove)
	// TODO
	router.HandleFunc("/shutdown", s.handleShutdown)

	router.HandleFunc("/leader", s.handleLeader)

	go func() {
		log.Printf("listening on %s ...", s.addr)
		http.ListenAndServe(s.addr, router)
	}()

	go func() {
		<-s.stopCh

		// wait response to be fully flushed
		time.Sleep(2 * time.Second)

		syscall.Kill(syscall.Getpid(), syscall.SIGINT)

	}()

	return nil
}

func (s *Service) handleJoin(w http.ResponseWriter, r *http.Request) {
	log.Printf("===> join request: [%v]", r.URL)
	nodeID := r.URL.Query().Get("id")
	raftAddr := r.URL.Query().Get("addr")
	if nodeID == "" || raftAddr == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := s.store.Join(nodeID, raftAddr); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)

	msg := fmt.Sprintf("node [%s] join cluster successfully with raft addr [%s]", nodeID, raftAddr)
	w.Write([]byte(msg))
}

func (s *Service) handleRemove(w http.ResponseWriter, r *http.Request) {
	nodeID := r.URL.Query().Get("id")
	if nodeID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := s.store.Remove(nodeID); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)

	msg := fmt.Sprintf("node [%s] remove from cluster successfully", nodeID)
	w.Write([]byte(msg))
}

func (s *Service) handleShutdown(w http.ResponseWriter, r *http.Request) {
	err := s.store.Shutdown()

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.WriteHeader(http.StatusOK)

	msg := fmt.Sprintf("server is shutting down")
	w.Write([]byte(msg))

	close(s.stopCh)
}

func (s *Service) handleLeader(w http.ResponseWriter, r *http.Request) {
	leader := s.store.Leader()
	fmt.Fprintf(w, "%s", leader)
}
