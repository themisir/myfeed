package removing

type PostRepository interface {
	RemoveSourcePost(sourceId int, postId int) error
	RemoveAllSourcePosts(sourceId int) error
}
