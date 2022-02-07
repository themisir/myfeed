package updating

type Feed struct {
	Name string
}

type FeedRepository interface {
	UpdateFeed(feedId int, data Feed) error
	UpdateFeedSources(feedId int, sourceIds ...int) error
}
