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
		name           string
		expiryDaysFrom int
		want           bool
	}{
		{"expires in future", 30, true},
		{"expired today", 0, true},
		{"within grace period after expiry", -50, true},
		{"just inside grace boundary", -59, true},
		{"past grace period", -70, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expiryDate := now.AddDate(0, 0, tt.expiryDaysFrom).Format(time.RFC3339)
			got, err := isLastOrderWithinActiveWindow(expiryDate)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("isLastOrderWithinActiveWindow(expiry %d days from now) = %v, want %v", tt.expiryDaysFrom, got, tt.want)
			}
		})
	}

	t.Run("empty date returns false", func(t *testing.T) {
		got, err := isLastOrderWithinActiveWindow("")
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

func TestIsOlderThanNDays(t *testing.T) {
	now := time.Now().UTC()
	dateOnly := "2006-01-02"
	rfc3339 := time.RFC3339

	tests := []struct {
		name    string
		dateStr string
		days    int
		want    bool
		wantErr bool
	}{
		{
			name:    "empty string returns false",
			dateStr: "",
			days:    30,
			want:    false,
		},
		{
			name:    "date 31 days ago is older than 30",
			dateStr: now.AddDate(0, 0, -31).Format(dateOnly),
			days:    30,
			want:    true,
		},
		{
			name:    "date exactly 30 days ago is older than or equal to 30",
			dateStr: now.AddDate(0, 0, -30).Format(dateOnly),
			days:    30,
			want:    true,
		},
		{
			name:    "date 29 days ago is not older than 30",
			dateStr: now.AddDate(0, 0, -29).Format(dateOnly),
			days:    30,
			want:    false,
		},
		{
			name:    "future date is not older than 30",
			dateStr: now.AddDate(0, 0, 1).Format(dateOnly),
			days:    30,
			want:    false,
		},
		{
			name:    "today is not older than 1 day",
			dateStr: now.Format(dateOnly),
			days:    1,
			want:    false,
		},
		{
			name:    "RFC3339 timestamp older than N days",
			dateStr: now.AddDate(0, 0, -16).Format(rfc3339),
			days:    15,
			want:    true,
		},
		{
			name:    "RFC3339 timestamp not older than N days",
			dateStr: now.AddDate(0, 0, -14).Format(rfc3339),
			days:    15,
			want:    false,
		},
		{
			name:    "invalid date string returns error",
			dateStr: "not-a-date",
			days:    30,
			want:    false,
			wantErr: true,
		},
		{
			name:    "zero days: any past date qualifies",
			dateStr: now.AddDate(0, 0, -1).Format(dateOnly),
			days:    0,
			want:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := isOlderThanNDays(tt.dateStr, tt.days)
			if (err != nil) != tt.wantErr {
				t.Fatalf("isOlderThanNDays(%q, %d) error = %v, wantErr %v", tt.dateStr, tt.days, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("isOlderThanNDays(%q, %d) = %v, want %v", tt.dateStr, tt.days, got, tt.want)
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
