package traya

import (
	"strings"
	"testing"
	"time"
)

func TestIsRunningKitEligible(t *testing.T) {
	tests := []struct {
		name                string
		runningKitStartDate string
		want                bool
	}{
		{
			name:                "empty date is not eligible",
			runningKitStartDate: "",
			want:                false,
		},
		{
			name:                "started today is eligible",
			runningKitStartDate: time.Now().Format("2006-01-02"),
			want:                true,
		},
		{
			name:                "started 20 days ago is eligible",
			runningKitStartDate: time.Now().AddDate(0, 0, -20).Format(time.RFC3339),
			want:                true,
		},
		{
			name:                "started 30 days ago is not eligible",
			runningKitStartDate: time.Now().AddDate(0, 0, -31).Format(time.RFC3339),
			want:                false,
		},
		{
			name:                "started 90 days ago is not eligible",
			runningKitStartDate: time.Now().AddDate(0, 0, -90).Format(time.RFC3339),
			want:                false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := isRunningKitEligible(tt.runningKitStartDate)
			if err != nil {
				t.Fatalf("isRunningKitEligible(%q) returned error: %v", tt.runningKitStartDate, err)
			}
			if got != tt.want {
				t.Errorf("isRunningKitEligible(%q) = %v, want %v", tt.runningKitStartDate, got, tt.want)
			}
		})
	}

	if _, err := isRunningKitEligible("not-a-date"); err == nil {
		t.Errorf("expected error for invalid date, got nil")
	}
}

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
			wantPresent:     []string{"month-1-warriors"},
			wantAbsent:      []string{"traya-warriors", "traya-heroines", "stress-sleep-nutrition", "general"},
		},
		{
			name:            "active female month 3",
			gender:          "female",
			orderCount:      3,
			isWithin60Days:  true,
			latestOrderDate: "2026-05-01",
			wantPresent:     []string{"month-3-icons"},
			wantAbsent:      []string{"traya-heroines", "traya-warriors", "hormones-pcos", "dandruff-hair-health-female", "general"},
		},
		{
			name:            "active male month 10 maps to legends",
			gender:          "male",
			orderCount:      10,
			isWithin60Days:  true,
			latestOrderDate: "2026-05-01",
			wantPresent:     []string{"month-8-plus-legends"},
			wantAbsent:      []string{"traya-warriors", "general"},
		},
		{
			name:            "lost male",
			gender:          "male",
			orderCount:      3,
			isWithin60Days:  false,
			latestOrderDate: "2026-01-01",
			wantPresent:     []string{"traya-warriors"},
			wantAbsent:      []string{"month-3-pioneers", "stress-sleep-nutrition", "traya-heroines", "general"},
		},
		{
			name:            "lost female",
			gender:          "female",
			orderCount:      5,
			isWithin60Days:  false,
			latestOrderDate: "2026-01-01",
			wantPresent:     []string{"traya-heroines"},
			wantAbsent:      []string{"month-5-elites", "hormones-pcos", "traya-warriors", "general"},
		},
		{
			name:            "no-order user gets nothing",
			gender:          "male",
			orderCount:      0,
			isWithin60Days:  false,
			latestOrderDate: "",
			wantAbsent:      []string{"traya-warriors", "traya-heroines", "month-1-warriors", "stress-sleep-nutrition", "general"},
		},
		{
			name:            "no-order female gets nothing",
			gender:          "female",
			orderCount:      0,
			isWithin60Days:  false,
			latestOrderDate: "",
			wantAbsent:      []string{"traya-heroines", "traya-warriors", "hormones-pcos", "general"},
		},
		{
			name:                "female lead older than 30 days gets womens community",
			gender:              "female",
			orderCount:          0,
			isWithin60Days:      false,
			latestOrderDate:     "",
			customerType:        "lead",
			leadOlderThan30Days: true,
			wantPresent:         []string{"traya-womens-community"},
			wantAbsent:          []string{"traya-heroines", "general"},
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
			wantPresent:         []string{trayaExplorersChannelSlug},
			wantAbsent:          []string{"traya-warriors", "general"},
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
			wantAbsent:          []string{trayaExplorersChannelSlug, "traya-warriors", "general"},
		},
		{
			name:            "deprecated channels never appear for active male",
			gender:          "male",
			orderCount:      2,
			isWithin60Days:  true,
			latestOrderDate: "2026-05-01",
			wantAbsent: []string{
				"general",
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
				"general",
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
		"general",
		"hormones-pcos",
		"stress-sleep-nutrition-female",
		"digestion-metabolism-gut-female",
		"dandruff-hair-health-female",
		"stress-sleep-nutrition",
		"digestion-metabolism-gut-health",
		"dandruff-hair-health",
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

func TestGenerateHandle(t *testing.T) {
	tests := []struct {
		name        string
		firstName   string
		phoneNumber string
		wantPrefix  string
		wantLen     int // name[:4] + phone[:2] + 4 random digits
	}{
		{
			name:        "typical name and indian number",
			firstName:   "Chinmayee",
			phoneNumber: "+919876543210",
			wantPrefix:  "chin91",
			wantLen:     10,
		},
		{
			name:        "short name",
			firstName:   "Jo",
			phoneNumber: "+447911123456",
			wantPrefix:  "jo44",
			wantLen:     8,
		},
		{
			name:        "no phone",
			firstName:   "Chithra",
			phoneNumber: "",
			wantPrefix:  "chit",
			wantLen:     8,
		},
		{
			name:        "empty name",
			firstName:   "",
			phoneNumber: "+919876543210",
			wantPrefix:  "91",
			wantLen:     6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateHandle(tt.firstName, tt.phoneNumber)
			if !strings.HasPrefix(got, tt.wantPrefix) {
				t.Errorf("generateHandle(%q, %q) = %q, want prefix %q", tt.firstName, tt.phoneNumber, got, tt.wantPrefix)
			}
			if len(got) != tt.wantLen {
				t.Errorf("generateHandle(%q, %q) = %q, want length %d, got %d", tt.firstName, tt.phoneNumber, got, tt.wantLen, len(got))
			}
		})
	}
}

func keys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
