package common

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

var (
	ErrNoMoreSessions = fmt.Errorf("no more sessions")
)

// Call makes an HTTP request using the configured credentials
func Call(user, passwd, sessionId string, port int, method, path string) (*http.Response, []byte, error) {
	// Little delay so we don't trip the ratelimiter
	time.Sleep(100 * time.Millisecond)

	url := "http://localhost:" + strconv.Itoa(port) + path
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to create request: %w", err)
	}
	req.SetBasicAuth(user, passwd)
	req.Header.Set("session-id", sessionId)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to read response body: %w", err)
	}

	return resp, body, nil
}

// GetSession gets a session ID from the server
func GetSession(user, passwd string, port int) (string, error) {
	resp, _, err := Call(user, passwd, "", port, "GET", "/")
	if err != nil {
		return "", err
	}
	if resp.StatusCode == http.StatusTooManyRequests {
		return "", ErrNoMoreSessions
	}
	sessionId := resp.Header.Get("session-id")
	return sessionId, nil
}

// CallOK makes an HTTP request and expects a 200 OK response
func CallOK(user, passwd, sessionId string, port int, method, path string) ([]byte, error) {
	resp, body, err := Call(user, passwd, sessionId, port, method, path)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, resp.StatusCode, string(body))
	}
	return body, nil
}

// CallExpecting makes an HTTP request and expects a specific status code
func CallExpecting(user, passwd, sessionId string, port int, method, path string, expectedStatus int) ([]byte, error) {
	resp, body, err := Call(user, passwd, sessionId, port, method, path)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != expectedStatus {
		return nil, fmt.Errorf("Expected status %d, got %d. Body: %s", expectedStatus, resp.StatusCode, string(body))
	}
	return body, nil
}
