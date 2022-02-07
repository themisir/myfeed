package listing

type Feed interface {
	Id() int
	Name() string
	UserId() string
}

type FeedRepository interface {
	GetUserFeeds(userId string) ([]Feed, error)
	GetFeed(feedId int) (Feed, error)
}
