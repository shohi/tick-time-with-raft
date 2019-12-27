package store

import (
	"errors"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"github.com/hashicorp/raft"
)

type intAddr string

// intAddr implements the net.Addr interface. An address for the internal
// transport.

func (n intAddr) Network() string {
	return "internal"
}
func (n intAddr) String() string {
	return string(n)
}

var (
	errNotAdvertisable = errors.New("local bind address is not advertisable")
	errNotTCP          = errors.New("local address is not a TCP address")
)

// intStreamLayer is internal StreamLayer which implements raft.StreamLayer
// based on raft.TCPStreamLayer
// TODO: refer natsStreamLayer
type intStreamLayer struct {
	// advertise net.Addr
	advertise intAddr
	listener  *net.TCPListener

	localAddr string
	httpAddr  string
	raftAddr  string
}

// Dial implements the StreamLayer interface.
func (s *intStreamLayer) Dial(address raft.ServerAddress, timeout time.Duration) (net.Conn, error) {
	_, raftAddr := splitAddr(string(address))
	return net.DialTimeout("tcp", string(raftAddr), timeout)
}

// Accept implements the net.Listener interface.
func (s *intStreamLayer) Accept() (c net.Conn, err error) {
	return s.listener.Accept()
}

// Close implements the net.Listener interface.
func (s *intStreamLayer) Close() (err error) {
	return s.listener.Close()
}

// Addr implements the net.Listener interface.
func (s *intStreamLayer) Addr() net.Addr {
	// Use an advertise addr if provided
	if s.advertise != "" {
		return s.advertise
	}
	return s.listener.Addr()
}

func splitAddr(addr string) (httpAddr, raftAddr string) {
	fields := strings.Split(addr, "@")
	if len(fields) != 2 {
		err := fmt.Errorf("address must contain joinAddr and raftAddr, separator is '@' - [%s]", addr)
		panic(err)
	}

	return fields[0], fields[1]
}

func newIntStreamLayer(addr string, timeout time.Duration, log *log.Logger) (raft.StreamLayer, error) {
	httpAddr, raftAddr := splitAddr(addr)

	/*
		advertise, err := net.ResolveTCPAddr("tcp", raftAddr)
		if err != nil {
			return nil, err
		}
	*/

	// Try to bind
	list, err := net.Listen("tcp", raftAddr)
	if err != nil {
		return nil, err
	}

	// Create stream
	stream := &intStreamLayer{
		localAddr: addr,
		httpAddr:  httpAddr,
		raftAddr:  raftAddr,

		advertise: intAddr(addr), // use addr as advertise address
		listener:  list.(*net.TCPListener),
	}

	/*
		// Verify that we have a usable advertise address
		taddr, ok := stream.Addr().(*net.TCPAddr)
		if !ok {
			list.Close()
			return nil, errNotTCP
		}
		if taddr.IP.IsUnspecified() {
			list.Close()
			return nil, errNotAdvertisable
		}
	*/

	return stream, nil
}
