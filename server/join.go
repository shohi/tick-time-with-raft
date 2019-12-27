package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func (s *Service) Join(joinAddr, raftAddr string) error {
	if joinAddr == "" {
		return nil
	}

	serverAddr := fmt.Sprintf("%s@%s", s.opts.Addr, raftAddr)
	b, err := json.Marshal(map[string]string{"addr": serverAddr, "id": s.opts.ServerID})
	if err != nil {
		return err
	}

	log.Printf("join cluster => [%v]", string(b))

	urlPath := fmt.Sprintf("http://%s/join?id=%s&addr=%s",
		joinAddr, s.opts.ServerID, serverAddr)

	resp, err := http.Post(urlPath, "application-type/json", nil)
	if err != nil {
		return err
	}

	if resp != nil && resp.Body != nil {
		resp.Body.Close()
	}

	return nil
}
