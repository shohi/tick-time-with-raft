package store

import (
	"log"
	"os"
	"time"

	"github.com/hashicorp/raft"
)

// [ServiceID, ServerAddress] -> [a, join_addr@raft_addr]
// intConn is internal connection
// TODO: refer natsConn
type intConn struct{}

func newIntTransport(addr string, timeout time.Duration) (*raft.NetworkTransport, error) {
	logger := log.New(os.Stderr, "", log.LstdFlags)
	return newIntTransportWithLogger(addr, timeout, logger)
}

func newIntTransportWithLogger(addr string, timeout time.Duration, logger *log.Logger) (*raft.NetworkTransport, error) {
	return createIntTransport(addr, timeout, logger, func(stream raft.StreamLayer) *raft.NetworkTransport {
		return raft.NewNetworkTransportWithLogger(stream, 3, timeout, logger)
	})
}

func createIntTransport(addr string, timeout time.Duration, logger *log.Logger,
	transportCreator func(stream raft.StreamLayer) *raft.NetworkTransport) (*raft.NetworkTransport, error) {

	stream, err := newIntStreamLayer(addr, timeout, logger)
	if err != nil {
		return nil, err
	}

	return transportCreator(stream), nil
}
