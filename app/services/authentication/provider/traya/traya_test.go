package traya

import (
	"testing"
)

func TestComputeTargetChannels(t *testing.T) {
	tests := []struct {
		name                string
		gender              string
		orderCount          int
		isWithin60Days      bool
		latestOrderDate     string
		customerType        string
		caseID              string
		leadOlderThan30Days bool
		leadOlderThan15Days bool
		wantPresent         []string
		wantAbsent          []string
	}{
		{
			name:            "active male month 1",
			gender:          "male",
			orderCount:      1,
			isWithin60Days:  true,
			latestOrderDate: "2026-05-01",
			wantPresent:     []string{"month-1-warriors", "general"},
			wantAbsent:      []string{"traya-warriors", "traya-heroines", "stress-sleep-nutrition"},
		},
		{
			name:            "active female month 3",
			gender:          "female",
			orderCount:      3,
			isWithin60Days:  true,
			latestOrderDate: "2026-05-01",
			wantPresent:     []string{"month-3-icons", "general"},
			wantAbsent:      []string{"traya-heroines", "traya-warriors", "hormones-pcos", "dandruff-hair-health-female"},
		},
		{
			name:            "active male month 10 maps to legends",
			gender:          "male",
			orderCount:      10,
			isWithin60Days:  true,
			latestOrderDate: "2026-05-01",
			wantPresent:     []string{"month-8-plus-legends", "general"},
			wantAbsent:      []string{"traya-warriors"},
		},
		{
			name:            "lost male",
			gender:          "male",
			orderCount:      3,
			isWithin60Days:  false,
			latestOrderDate: "2026-01-01",
			wantPresent:     []string{"traya-warriors", "general"},
			wantAbsent:      []string{"month-3-pioneers", "stress-sleep-nutrition", "traya-heroines"},
		},
		{
			name:            "lost female",
			gender:          "female",
			orderCount:      5,
			isWithin60Days:  false,
			latestOrderDate: "2026-01-01",
			wantPresent:     []string{"traya-heroines", "general"},
			wantAbsent:      []string{"month-5-elites", "hormones-pcos", "traya-warriors"},
		},
		{
			name:            "no-order user gets neither cohort nor lost channel",
			gender:          "male",
			orderCount:      0,
			isWithin60Days:  false,
			latestOrderDate: "",
			wantPresent:     []string{"general"},
			wantAbsent:      []string{"traya-warriors", "traya-heroines", "month-1-warriors", "stress-sleep-nutrition"},
		},
		{
			name:            "no-order female gets neither cohort nor lost channel",
			gender:          "female",
			orderCount:      0,
			isWithin60Days:  false,
			latestOrderDate: "",
			wantPresent:     []string{"general"},
			wantAbsent:      []string{"traya-heroines", "traya-warriors", "hormones-pcos"},
		},
		{
			name:                "female lead older than 30 days gets womens community",
			gender:              "female",
			orderCount:          0,
			isWithin60Days:      false,
			latestOrderDate:     "",
			customerType:        "lead",
			leadOlderThan30Days: true,
			wantPresent:         []string{"general", "traya-womens-community"},
			wantAbsent:          []string{"traya-heroines"},
		},
		{
			name:                "male lead with case_id prefix 1 older than 15 days gets explorers",
			gender:              "male",
			orderCount:          0,
			isWithin60Days:      false,
			latestOrderDate:     "",
			customerType:        "lead",
			caseID:              "12345",
			leadOlderThan15Days: true,
			wantPresent:         []string{"general", trayaExplorersChannelSlug},
			wantAbsent:          []string{"traya-warriors"},
		},
		{
			name:                "male lead with non-1/2 prefix does not get explorers",
			gender:              "male",
			orderCount:          0,
			isWithin60Days:      false,
			latestOrderDate:     "",
			customerType:        "lead",
			caseID:              "98765",
			leadOlderThan15Days: true,
			wantPresent:         []string{"general"},
			wantAbsent:          []string{trayaExplorersChannelSlug, "traya-warriors"},
		},
		{
			name:            "deprecated channels never appear for active male",
			gender:          "male",
			orderCount:      2,
			isWithin60Days:  true,
			latestOrderDate: "2026-05-01",
			wantAbsent: []string{
				"stress-sleep-nutrition",
				"digestion-metabolism-gut-health",
				"dandruff-hair-health",
				"hormones-pcos",
			},
		},
		{
			name:            "deprecated channels never appear for active female",
			gender:          "female",
			orderCount:      2,
			isWithin60Days:  true,
			latestOrderDate: "2026-05-01",
			wantAbsent: []string{
				"hormones-pcos",
				"stress-sleep-nutrition-female",
				"digestion-metabolism-gut-female",
				"dandruff-hair-health-female",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := computeTargetChannels(tt.gender, tt.orderCount, tt.isWithin60Days, tt.latestOrderDate, tt.customerType, tt.caseID, tt.leadOlderThan30Days, tt.leadOlderThan15Days)

			for _, slug := range tt.wantPresent {
				if !got[slug] {
					t.Errorf("expected slug %q to be present in target channels, got %v", slug, keys(got))
				}
			}
			for _, slug := range tt.wantAbsent {
				if got[slug] {
					t.Errorf("expected slug %q to be absent from target channels, got %v", slug, keys(got))
				}
			}
		})
	}
}

func TestGetAllManagedChannelSlugs_IncludesNewAndDeprecated(t *testing.T) {
	managed := getAllManagedChannelSlugs()

	mustInclude := []string{
		"traya-warriors",
		"traya-heroines",
		"hormones-pcos",
		"stress-sleep-nutrition-female",
		"digestion-metabolism-gut-female",
		"dandruff-hair-health-female",
		"stress-sleep-nutrition",
		"digestion-metabolism-gut-health",
		"dandruff-hair-health",
		"general",
		"month-1-warriors",
		"month-1-heroines",
		"traya-womens-community",
		trayaExplorersChannelSlug,
	}

	for _, slug := range mustInclude {
		if !managed[slug] {
			t.Errorf("expected managed channel set to include %q", slug)
		}
	}
}

func keys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
