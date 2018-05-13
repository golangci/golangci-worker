package state

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/golangci/golangci-worker/app/utils/runmode"
)

type APIStorage struct {
	Host string
}

func NewAPIStorage() *APIStorage {
	return &APIStorage{
		Host: os.Getenv("API_URL"),
	}
}

func getHTTPClient() *http.Client {
	if runmode.IsProduction() {
		return http.DefaultClient
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // nolint
		},
	}
	return &http.Client{Transport: tr}
}

func (s APIStorage) getStatusURL(analysisID string) string {
	return fmt.Sprintf("%s/v1/repos/repo/owner/analyzes/%s/state", s.Host, analysisID)
}

func (s APIStorage) UpdateState(ctx context.Context, analysisID string, state *State) error {
	body, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("can't marshal payload json: %s", err)
	}

	req, err := http.NewRequest("PUT", s.getStatusURL(analysisID), bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("can't build http request: %s", err)
	}
	req = req.WithContext(ctx)

	resp, err := getHTTPClient().Do(req)
	if err != nil {
		return fmt.Errorf("can't make http request: %s", err)
	}
	if resp.Body != nil {
		defer resp.Body.Close()
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status code %d", resp.StatusCode)
	}

	return nil
}

func (s APIStorage) GetState(ctx context.Context, analysisID string) (*State, error) {
	req, err := http.NewRequest("GET", s.getStatusURL(analysisID), nil)
	if err != nil {
		return nil, fmt.Errorf("can't build http request: %s", err)
	}
	req = req.WithContext(ctx)

	resp, err := getHTTPClient().Do(req)
	if err != nil {
		return nil, fmt.Errorf("can't make http request: %s", err)
	}
	if resp.Body != nil {
		defer resp.Body.Close()
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status code %d", resp.StatusCode)
	}

	var state State
	if err = json.NewDecoder(resp.Body).Decode(&state); err != nil {
		return nil, fmt.Errorf("can't read json body: %s", err)
	}

	return &state, nil
}
