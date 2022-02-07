package removing

type FeedRepository interface {
	RemoveFeed(feedId int) error
}
