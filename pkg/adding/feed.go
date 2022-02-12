package adding

type (
	FeedData struct {
		Name     string
		UserId   string
		IsPublic bool
	}
	Feed interface {
		Id() int
		Name() string
		UserId() string
		IsPublic() bool
	}
	FeedRepository interface {
		AddFeed(data FeedData) (Feed, error)
	}
)
