package coordination

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type Assignment struct {
	WorkerID    string    `json:"worker_id"`
	WorkerIndex int       `json:"worker_index"`
	WorkerCount int       `json:"worker_count"`
	Symbols     []string  `json:"symbols"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func RegisterAssignment(ctx context.Context, endpoint string, assignment Assignment) error {
	endpoint = strings.TrimRight(strings.TrimSpace(endpoint), "/")
	if endpoint == "" {
		return nil
	}
	if assignment.WorkerID == "" {
		assignment.WorkerID = fmt.Sprintf("worker-%d", assignment.WorkerIndex)
	}
	if assignment.UpdatedAt.IsZero() {
		assignment.UpdatedAt = time.Now().UTC()
	}

	value, err := json.Marshal(assignment)
	if err != nil {
		return err
	}

	requestBody, err := json.Marshal(map[string]string{
		"key":   base64.StdEncoding.EncodeToString([]byte("/lab14/crypto/workers/" + assignment.WorkerID)),
		"value": base64.StdEncoding.EncodeToString(value),
	})
	if err != nil {
		return err
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint+"/v3/kv/put", bytes.NewReader(requestBody))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("etcd returned status %s", response.Status)
	}
	return nil
}
