package adding

type (
	FeedData struct {
		Name   string
		UserId string
	}
	Feed interface {
		Id() int
		Name() string
		UserId() string
	}
	FeedRepository interface {
		AddFeed(data FeedData) (Feed, error)
	}
)
