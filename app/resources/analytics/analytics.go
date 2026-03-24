package analytics

type ChannelStat struct {
	ChannelName string
	Count       int
}

type AdminReplyTimeStat struct {
	AdminHandle    string
	AvgTimeMinutes float64
}

type Report struct {
	ChannelOnboardings []ChannelStat
	ChannelPosts       []ChannelStat
	AdminReplyTimes    []AdminReplyTimeStat
}
