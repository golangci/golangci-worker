package status

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

type Updater interface {
	UpdateStatus(ctx context.Context, analysisID, status string) error
}

type APIUpdater struct {
	Host string
}

func NewAPIUpdater() *APIUpdater {
	return &APIUpdater{
		Host: os.Getenv("API_URL"),
	}
}

func getHTTPClient() *http.Client {
	if runmode.IsProduction() {
		return http.DefaultClient
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	return &http.Client{Transport: tr}
}

func (u APIUpdater) UpdateStatus(ctx context.Context, analysisID, status string) error {
	payload := struct {
		Status string
	}{
		Status: status,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("can't marshal payload json: %s", err)
	}

	url := fmt.Sprintf("%s/v1/repos/repo/owner/analyzes/%s/status", u.Host, analysisID)
	req, err := http.NewRequest("PUT", url, bytes.NewReader(body))
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
