package updating

type Feed struct {
	Name     string
	IsPublic bool
}

type FeedRepository interface {
	UpdateFeed(feedId int, data Feed) error
	UpdateFeedSources(feedId int, sourceIds ...int) error
}
