package adding

import "time"

type (
	PostData struct {
		SourceId    int
		Title       string
		Description string
		Url         string
		PublishedAt *time.Time
		UpdatedAt   *time.Time
	}
	Post interface {
		Id() int
		Title() string
		Description() string
		Url() string
		PublishedAt() *time.Time
		UpdatedAt() *time.Time
	}
	PostRepository interface {
		AddPost(data PostData) (Post, error)
		AddManyPosts(items ...PostData) error
	}
)
