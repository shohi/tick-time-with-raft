package store

import (
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/hashicorp/raft"
)

const (
	retainSnapshotCount = 2
	raftTimeout         = 10 * time.Second
)

type Store struct {
	RaftDir     string
	RaftBind    string
	raft        *raft.Raft
	serverID    string
	numericalID int
	peersLength int
	notifyCh    chan bool // TODO: watch leader change
}

func NewStore() *Store {
	return &Store{numericalID: -1, peersLength: -1}
}

func (s *Store) Start() {
	c := time.Tick(5 * time.Second)
	go func() {
		for range c {
			confFuture := s.raft.GetConfiguration()
			if err := confFuture.Error(); err != nil {
				log.Printf("config error: %v", err)
			} else {
				conf := confFuture.Configuration()
				log.Printf("state: %v, config => [%v]", s.raft.State(), conf)

			}
			// log.Printf("ticker ID %d of %d, count %d", s.numericalID, s.peersLength, count)
		}
	}()
}

func (s *Store) Open(enableSingle bool, localID string) error {
	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID(localID)
	s.serverID = localID

	log.Printf("Open, local ID [%v]", config.LocalID)

	addr, err := net.ResolveTCPAddr("tcp", s.RaftBind)
	if err != nil {
		return err
	}

	transport, err := raft.NewTCPTransport(s.RaftBind, addr, 3, 10*time.Second, os.Stderr)
	if err != nil {
		return err
	}

	snapshots, err := raft.NewFileSnapshotStore(s.RaftDir, retainSnapshotCount, os.Stderr)
	if err != nil {
		return fmt.Errorf("file snapshot store: %s", err)
	}

	var logStore raft.LogStore
	var stableStore raft.StableStore

	logStore = raft.NewInmemStore()
	stableStore = raft.NewInmemStore()

	ra, err := raft.NewRaft(config, (*fsm)(s), logStore, stableStore, snapshots, transport)
	if err != nil {
		return fmt.Errorf("new raft: %s", err)
	}
	s.raft = ra

	if enableSingle {
		configuration := raft.Configuration{
			Servers: []raft.Server{
				{
					ID:      config.LocalID,
					Address: transport.LocalAddr(),
				},
			},
		}
		ra.BootstrapCluster(configuration)
	}

	return nil
}

// TODO: refactor error handling and logging
func (s *Store) Join(nodeID, addr string) error {
	log.Printf("received join request for remote node %s at %s", nodeID, addr)

	configFuture := s.raft.GetConfiguration()
	if err := configFuture.Error(); err != nil {
		log.Printf("failed to get raft configuration: %v", err)
		return err
	}

	for _, srv := range configFuture.Configuration().Servers {

		if srv.ID == raft.ServerID(nodeID) || srv.Address == raft.ServerAddress(addr) {
			if srv.Address == raft.ServerAddress(addr) && srv.ID == raft.ServerID(nodeID) {
				log.Printf("node %s at %s already member of cluster, ignoring join request", nodeID, addr)
				return nil
			}

			future := s.raft.RemoveServer(srv.ID, 0, 0)
			if err := future.Error(); err != nil {
				return fmt.Errorf("error removing existing node %s at %s: %s", nodeID, addr, err)
			}
		}
	}

	f := s.raft.AddVoter(raft.ServerID(nodeID), raft.ServerAddress(addr), 0, 0)
	if f.Error() != nil {
		return f.Error()
	}

	s.raft.Apply([]byte("new voter"), raftTimeout)

	log.Printf("node %s at %s joined successfully", nodeID, addr)
	return nil
}

func (s *Store) Remove(nodeID string) error {
	log.Printf("received remove request for remote node %s", nodeID)

	configFuture := s.raft.GetConfiguration()
	if err := configFuture.Error(); err != nil {
		err = fmt.Errorf("failed to get raft confgiuration: %w", err)
		log.Printf("%v", err)
		return err
	}

	var found = false
	var err error
	for _, srv := range configFuture.Configuration().Servers {
		if srv.ID == raft.ServerID(nodeID) {
			found = true
			future := s.raft.RemoveServer(srv.ID, 0, 0)
			err = future.Error()
			if err != nil {
				err = fmt.Errorf("removing existing node %s: %w", nodeID, err)
			}
		}
	}

	if !found {
		err = fmt.Errorf("no node %s", nodeID)
	}

	if err != nil {
		log.Printf("%v", err)
	} else {
		log.Printf("removing existing node %s successfully", nodeID)
	}

	return err
}

func (s *Store) Shutdown() error {
	// TODO: redirect to leader
	future := s.raft.Shutdown()
	err := future.Error()

	if err != nil {
		err = fmt.Errorf("shutting down node %s: %w", s.serverID, err)
		log.Printf("%v", err)
		return err
	}

	return nil
}

func (s *Store) Leader() string {
	return string(s.raft.Leader())
}
