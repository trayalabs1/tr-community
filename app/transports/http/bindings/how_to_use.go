package bindings

import "time"

const (
	howToUseSeenMetaKey = "community_htu_seen"
	howToUseShowMetaKey = "community_htu_show"
)

func withHowToUseTooltip(meta map[string]any, joined time.Time, cutoff time.Time) map[string]any {
	out := make(map[string]any, len(meta)+1)
	for k, v := range meta {
		out[k] = v
	}

	seen, _ := out[howToUseSeenMetaKey].(bool)
	out[howToUseShowMetaKey] = !cutoff.IsZero() && !joined.Before(cutoff) && !seen

	return out
}
