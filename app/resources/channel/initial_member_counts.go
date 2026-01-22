package channel

// InitialMemberCounts maps channel slugs to their initial member count values.
// This is used to conditionally add initial values to the actual member count
// when displaying to users, based on the channel slug.
var InitialMemberCounts = map[string]int{
	// Male cohort channels - 3000 members each
	"month-1-warriors":      3000,
	"month-2-titans":        3000,
	"month-3-warriors":      3000,
	"month-4-warriors":      3000,
	"month-5-warriors":      3000,
	"month-6-8-warriors":    3000,
	"month-8-plus-warriors": 3000,

	// Female cohort channels - 3000 members each
	"month-1-heroines":      3000,
	"month-2-heroines":      3000,
	"month-3-heroines":      3000,
	"month-4-heroines":      3000,
	"month-5-heroines":      3000,
	"month-6-8-heroines":    3000,
	"month-8-plus-queens":   3000,

	// Topic channels - Varying member counts
	"stress-sleep-nutrition":          850,
	"digestion-metabolism-gut-health": 763,
	"dandruff-hair-health":            882,
	"hormones-pcos":                   965,
	"stress-sleep-nutrition-female":   453,
	"digestion-metabolism-gut-female": 882,
	"dandruff-hair-health-female":     950,

	// Other channels
	"admin-announcements": 0,
}

// GetInitialMemberCount returns the initial member count for a channel by name.
// If the channel name is not found in the configuration, it returns 0 (no inflation).
func GetInitialMemberCount(channelName string) int {
	if count, ok := InitialMemberCounts[channelName]; ok {
		return count
	}
	return 0
}
