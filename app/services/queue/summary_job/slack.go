package summary_job

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fctx"
)

func formatMessage(c Counts, calculatedAt time.Time, loc *time.Location) string {
	return fmt.Sprintf(
		"*Daily Submission Queue Summary*\n"+
			"Calculated at: *%s*\n"+
			"Pending Review: *%d*\n"+
			"Pending Reply: *%d*\n"+
			"Pending Reply to Reply: *%d*",
		calculatedAt.In(loc).Format("2006-01-02 15:04:05 MST"),
		c.PendingReview, c.PendingReply, c.PendingReplyToReply,
	)
}

func postToSlack(ctx context.Context, client *http.Client, webhookURL, text string) error {
	body, err := json.Marshal(map[string]string{"text": text})
	if err != nil {
		return fault.Wrap(err, fctx.With(ctx))
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhookURL, bytes.NewReader(body))
	if err != nil {
		return fault.Wrap(err, fctx.With(ctx))
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		return fault.Wrap(err, fctx.With(ctx))
	}
	defer res.Body.Close()

	if res.StatusCode >= 300 {
		return fault.Wrap(fmt.Errorf("slack webhook returned status %d", res.StatusCode), fctx.With(ctx))
	}

	return nil
}
