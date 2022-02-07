package listing

import "time"

type Post interface {
	Id() int
	Title() string
	Description() string
	Url() string
	PublishedAt() *time.Time
	UpdatedAt() *time.Time
}

type SourcePost interface {
	Post
	Source() Source
}

type PostRepository interface {
	GetSourcePosts(sourceId int) ([]Post, error)
	GetFeedPosts(feedId int) ([]SourcePost, error)
}
