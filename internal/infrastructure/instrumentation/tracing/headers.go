package tracing

import "strings"

// parseOTLPHeaders parses the OTel spec's OTEL_EXPORTER_OTLP_HEADERS format:
// comma-separated key=value pairs, e.g. "Authorization=Basic xyz,x-tenant=foo".
// Malformed pairs (missing '=') are skipped rather than failing the whole value,
// since a single typo in an ops-managed env var shouldn't disable tracing headers entirely.
func parseOTLPHeaders(raw string) map[string]string {
	headers := map[string]string{}

	raw = strings.TrimSpace(raw)
	if raw == "" {
		return headers
	}

	for _, pair := range strings.Split(raw, ",") {
		key, value, ok := strings.Cut(pair, "=")
		if !ok {
			continue
		}

		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if key == "" {
			continue
		}

		headers[key] = value
	}

	return headers
}
