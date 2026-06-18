package traya

import (
	"strings"
	"testing"
	"time"
)

func TestComputeTargetChannels(t *testing.T) {
	tests := []struct {
		name                 string
		gender               string
		orderCount           int
		isWithinActiveWindow bool
		customerType         string
		caseID               string
		leadOlderThan30Days  bool
		leadOlderThan15Days  bool
		wantPresent          []string
		wantAbsent           []string
	}{
		{
			name:                 "active male month 1",
			gender:               "male",
			orderCount:           1,
			isWithinActiveWindow: true,
			wantPresent:          []string{"month-1-warriors"},
			wantAbsent:           []string{"traya-warriors", "traya-heroines", "stress-sleep-nutrition", "general"},
		},
		{
			name:                 "active female month 3",
			gender:               "female",
			orderCount:           3,
			isWithinActiveWindow: true,
			wantPresent:          []string{"month-3-icons"},
			wantAbsent:           []string{"traya-heroines", "traya-warriors", "hormones-pcos", "dandruff-hair-health-female", "general"},
		},
		{
			name:                 "active male month 10 maps to legends",
			gender:               "male",
			orderCount:           10,
			isWithinActiveWindow: true,
			wantPresent:          []string{"month-8-plus-legends"},
			wantAbsent:           []string{"traya-warriors", "general"},
		},
		{
			name:                 "lost male",
			gender:               "male",
			orderCount:           3,
			isWithinActiveWindow: false,
			wantPresent:          []string{"traya-warriors"},
			wantAbsent:           []string{"month-3-pioneers", "stress-sleep-nutrition", "traya-heroines", "general"},
		},
		{
			name:                 "lost female",
			gender:               "female",
			orderCount:           5,
			isWithinActiveWindow: false,
			wantPresent:          []string{"traya-heroines"},
			wantAbsent:           []string{"month-5-elites", "hormones-pcos", "traya-warriors", "general"},
		},
		{
			name:                 "outside active window male gets lost channel",
			gender:               "male",
			orderCount:           0,
			isWithinActiveWindow: false,
			wantPresent:          []string{"traya-warriors"},
			wantAbsent:           []string{"traya-heroines", "month-1-warriors", "stress-sleep-nutrition", "general"},
		},
		{
			name:                 "outside active window female gets lost channel",
			gender:               "female",
			orderCount:           0,
			isWithinActiveWindow: false,
			wantPresent:          []string{"traya-heroines"},
			wantAbsent:           []string{"traya-warriors", "hormones-pcos", "general"},
		},
		{
			name:                 "female lead older than 30 days gets womens community",
			gender:               "female",
			orderCount:           0,
			isWithinActiveWindow: false,
			customerType:         "lead",
			leadOlderThan30Days:  true,
			wantPresent:          []string{"traya-womens-community", "traya-heroines"},
			wantAbsent:           []string{"general"},
		},
		{
			name:                 "male lead with case_id prefix 1 older than 15 days gets explorers",
			gender:               "male",
			orderCount:           0,
			isWithinActiveWindow: false,
			customerType:         "lead",
			caseID:               "12345",
			leadOlderThan15Days:  true,
			wantPresent:          []string{trayaExplorersChannelSlug, "traya-warriors"},
			wantAbsent:           []string{"general"},
		},
		{
			name:                 "male lead with non-1/2 prefix does not get explorers",
			gender:               "male",
			orderCount:           0,
			isWithinActiveWindow: false,
			customerType:         "lead",
			caseID:               "98765",
			leadOlderThan15Days:  true,
			wantPresent:          []string{"traya-warriors"},
			wantAbsent:           []string{trayaExplorersChannelSlug, "general"},
		},
		{
			name:                 "deprecated channels never appear for active male",
			gender:               "male",
			orderCount:           2,
			isWithinActiveWindow: true,
			wantAbsent: []string{
				"general",
				"stress-sleep-nutrition",
				"digestion-metabolism-gut-health",
				"dandruff-hair-health",
				"hormones-pcos",
			},
		},
		{
			name:                 "deprecated channels never appear for active female",
			gender:               "female",
			orderCount:           2,
			isWithinActiveWindow: true,
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
			got := computeTargetChannels(tt.gender, tt.orderCount, tt.isWithinActiveWindow, tt.customerType, tt.caseID, tt.leadOlderThan30Days, tt.leadOlderThan15Days)

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

func TestIsLastOrderWithinActiveWindow(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name          string
		daysAgo       int
		kitExpireDays int
		want          bool
	}{
		{"within kit window", 10, 30, true},
		{"within grace period after kit expiry", 80, 30, true},
		{"just inside boundary", 89, 30, true},
		{"past kit plus grace", 100, 30, false},
		{"zero kit days still has 60 day grace", 50, 0, true},
		{"zero kit days past grace", 70, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orderDate := now.AddDate(0, 0, -tt.daysAgo).Format(time.RFC3339)
			got, err := isLastOrderWithinActiveWindow(orderDate, tt.kitExpireDays)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("isLastOrderWithinActiveWindow(%d days ago, kitExpireDays=%d) = %v, want %v", tt.daysAgo, tt.kitExpireDays, got, tt.want)
			}
		})
	}

	t.Run("empty date returns false", func(t *testing.T) {
		got, err := isLastOrderWithinActiveWindow("", 30)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got {
			t.Errorf("expected false for empty date")
		}
	})
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
