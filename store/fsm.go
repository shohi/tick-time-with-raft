package store

import (
	"io"
	"log"
	"regexp"
	"strings"

	"github.com/hashicorp/raft"
)

type fsm Store

func (f *fsm) Apply(l *raft.Log) interface{} {

	// log.Printf("apply [%v] [%v]", l, l.Data)

	stats := f.raft.Stats()

	config := stats["latest_configuration"]

	peers := peersList(config)
	f.peersLength = len(peers)

	ID := f.serverID
	f.numericalID = getNumericalID(ID, peers)

	log.Printf("apply ID [%s] [%d], content: [%s]", ID, f.numericalID, string(l.Data))

	return nil
}

func (f *fsm) Snapshot() (raft.FSMSnapshot, error) {

	log.Printf("snapshot")

	return &fsmSnapshot{}, nil
}

func (f *fsm) Restore(rc io.ReadCloser) error {

	log.Printf("restore [%v]", rc)

	return nil
}

type fsmSnapshot struct{}

func (f *fsmSnapshot) Persist(sink raft.SnapshotSink) error {
	err := func() error {
		b := []byte("hello from persist")

		if _, err := sink.Write(b); err != nil {
			return err
		}

		return sink.Close()
	}()

	if err != nil {
		sink.Cancel()
	}

	return err
}

func (f *fsmSnapshot) Release() {}

func getNumericalID(ID string, peers []string) int {
	for i, value := range peers {
		if value == ID {
			return i
		}
	}
	return -1
}

func peersList(rawConfig string) []string {
	peers := []string{}

	re := regexp.MustCompile(`ID:[0-9A-z]*`)

	for _, peer := range re.FindAllString(rawConfig, -1) {
		peers = append(peers, strings.Replace(peer, "ID:", "", -1))
	}

	return peers
}
