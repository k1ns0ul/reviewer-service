package load

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

var baseURL = "http://localhost:8080"

func Benchmark_AllEndpoints(b *testing.B) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	if err := warmup(client); err != nil {
		b.Fatalf("warmup failed: %v", err)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			callAllOnce(b, client)
		}
	})
}

func warmup(client *http.Client) error {
	if err := postJSON(client, baseURL+"/team/add", map[string]interface{}{
		"teamName": "backend",
		"members": []map[string]interface{}{
			{"userId": "u1", "username": "alice", "isActive": true},
			{"userId": "u2", "username": "bob", "isActive": true},
		},
	}); err != nil {
		return err
	}
	if err := postJSON(client, baseURL+"/pullRequest/create", map[string]interface{}{
		"pullRequestId":   "pr-base",
		"pullRequestName": "feature-base",
		"authorId":        "u1",
	}); err != nil {
		return err
	}
	return nil
}

func callAllOnce(b *testing.B, client *http.Client) {
	doReq(b, client, http.MethodGet, "/health", nil)
	doReq(b, client, http.MethodGet, "/stats", nil)
	doReq(b, client, http.MethodGet, "/team/get?teamName=backend", nil)
	doReq(b, client, http.MethodGet, "/users/getReview?userId=u1", nil)

	_ = postJSON(client, baseURL+"/pullRequest/create", map[string]interface{}{
		"pullRequestId":   "pr-bench",
		"pullRequestName": "feature-bench",
		"authorId":        "u1",
	})

	_ = postJSON(client, baseURL+"/pullRequest/merge", map[string]string{
		"pullRequestId": "pr-base",
	})
	_ = postJSON(client, baseURL+"/pullRequest/reassign", map[string]string{
		"pullRequestId": "pr-base",
		"oldUserId":     "u2",
	})
	_ = postJSON(client, baseURL+"/team/deactivate", map[string]string{
		"teamName": "backend",
	})
}

func doReq(b *testing.B, client *http.Client, method, path string, body []byte) {
	req, err := http.NewRequest(method, baseURL+path, bytes.NewReader(body))
	if err != nil {
		b.Fatalf("new request: %v", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	_ = resp.Body.Close()
}

func postJSON(client *http.Client, url string, payload interface{}) error {
	b, _ := json.Marshal(payload)
	req, _ := http.NewRequest(http.MethodPost, url, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	_ = resp.Body.Close()
	return nil
}
