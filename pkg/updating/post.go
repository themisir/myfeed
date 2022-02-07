package updating

import "time"

type Post struct {
	Title       string
	Description string
	Url         string
	PublishedAt *time.Time
	UpdatedAt   *time.Time
}

type PostRepository interface {
	UpdateSourcePost(sourceId int, postId int, data Post) error
}
