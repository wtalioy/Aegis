package providers

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func newJSONRequest(ctx context.Context, method, url string, body any) (*http.Request, error) {
	payload, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

func decodeJSONResponse(resp *http.Response, target any) error {
	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return err
	}
	return nil
}

func newStreamScanner(resp *http.Response) *bufio.Scanner {
	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	return scanner
}

func unexpectedStatus(provider string, status int) error {
	return fmt.Errorf("%s returned status %d", provider, status)
}
