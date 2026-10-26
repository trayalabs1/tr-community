package analytics

type ChannelStat struct {
	ChannelName string
	Count       int
}

type AdminReplyTimeStat struct {
	AdminHandle    string
	AvgTimeMinutes float64
}

type ThreadsRepliedCounts struct {
	Total   int
	Direct  int
	ToReply int
}

type AdminThreadsRepliedStat struct {
	AdminHandle string
	ThreadsRepliedCounts
}

type Report struct {
	ChannelOnboardings  []ChannelStat
	ChannelPosts        []ChannelStat
	AdminReplyTimes     []AdminReplyTimeStat
	AdminThreadsReplied []AdminThreadsRepliedStat
	ThreadsReplied      ThreadsRepliedCounts
}
