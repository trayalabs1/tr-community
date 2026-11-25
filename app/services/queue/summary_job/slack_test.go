package summary_job

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormatMessage(t *testing.T) {
	loc, err := time.LoadLocation("Asia/Kolkata")
	require.NoError(t, err)
	calculatedAt := time.Date(2026, 3, 5, 23, 55, 0, 0, loc)

	msg := formatMessage(Counts{PendingReview: 12, PendingReply: 50, PendingReplyToReply: 7}, calculatedAt, loc)

	assert.Contains(t, msg, "12")
	assert.Contains(t, msg, "50")
	assert.Contains(t, msg, "7")
	assert.Contains(t, msg, "Pending Review")
	assert.Contains(t, msg, "Pending Reply")
	assert.Contains(t, msg, "Pending Reply to Reply")
	assert.Contains(t, msg, "2026-03-05 23:55:00")
}

func TestPostToSlack(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		var receivedBody map[string]string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, http.MethodPost, r.Method)
			require.Equal(t, "application/json", r.Header.Get("Content-Type"))
			require.NoError(t, json.NewDecoder(r.Body).Decode(&receivedBody))
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		err := postToSlack(context.Background(), server.Client(), server.URL, "hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", receivedBody["text"])
	})

	t.Run("non-2xx response is an error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		err := postToSlack(context.Background(), server.Client(), server.URL, "hello")
		assert.Error(t, err)
	})
}
