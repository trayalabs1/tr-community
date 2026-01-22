package channel

// InitialMemberCounts maps channel names to their initial member count values.
// This is used to conditionally add initial values to the actual member count
// when displaying to users, based on the channel name.
var InitialMemberCounts = map[string]int{
	// Warriors Channels - 3000 members each
	"Month 1 Warriors":     3000,
	"Month 2 Warriors":     3000,
	"Month 3 Warriors":     3000,
	"Month 4 Warriors":     3000,
	"Month 5 Warriors":     3000,
	"Month 6-8 Warriors":   3000,
	"Month 8+ Warriors":    3000,

	// Heroines Channels - 3000 members each
	"Month 1 Heroines":     3000,
	"Month 2 Heroines":     3000,
	"Month 3 Heroines":     3000,
	"Month 4 Heroines":     3000,
	"Month 5 Heroines":     3000,
	"Month 6-8 Heroines":   3000,
	"Month 8+ Heroines":    3000,

	// Topic Channels - Varying member counts
	"Stress, Sleep & Nutrition":           850,
	"Digestion, Metabolism & Gut Health":  763,
	"Dandruff & Hair Health":              882,
	"Hormones & PCOS":                     965,
	"Stress, Sleep & Nutrition - Female":  453,
	"Digestion, Metabolism & Gut - Female": 882,
	"Dandruff & Hair Health - Female":     950,
}

// GetInitialMemberCount returns the initial member count for a channel by name.
// If the channel name is not found in the configuration, it returns 0 (no inflation).
func GetInitialMemberCount(channelName string) int {
	if count, ok := InitialMemberCounts[channelName]; ok {
		return count
	}
	return 0
}
