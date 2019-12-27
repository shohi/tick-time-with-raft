package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func Join(joinAddr, raftAddr, nodeID string) error {
	if joinAddr == "" {
		return nil
	}

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
