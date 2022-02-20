package memory

import (
	"log"
	"sort"
	"time"

	"github.com/themisir/myfeed/pkg/adding"
	"github.com/themisir/myfeed/pkg/listing"
	"github.com/themisir/myfeed/pkg/updating"
)

func NewPostRepository(feedRepository *feedRepository, sourceRepository *sourceRepository) *postRepository {
	return &postRepository{
		posts:            []*post{},
		feedRepository:   feedRepository,
		sourceRepository: sourceRepository,
	}
}

type PostData struct {
	Id          int        `json:"id"`
	SourceId    int        `json:"source_id"`
	Url         string     `json:"url"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	PublishedAt *time.Time `json:"published_at"`
	UpdatedAt   *time.Time `json:"updated_at"`
}

type post struct {
	PostData
}

type sourcePost struct {
	post
	source listing.Source
}

func (p *sourcePost) Source() listing.Source {
	return p.source
}

func (p *post) Id() int {
	return p.PostData.Id
}

func (p *post) Title() string {
	return p.PostData.Title
}

func (p *post) Description() string {
	return p.PostData.Description
}

func (p *post) Url() string {
	return p.PostData.Url
}

func (p *post) PublishedAt() *time.Time {
	return p.PostData.PublishedAt
}

func (p *post) UpdatedAt() *time.Time {
	return p.PostData.UpdatedAt
}

type postRepository struct {
	posts            []*post
	feedRepository   *feedRepository
	sourceRepository *sourceRepository
	persistence      Persistence
}

func (r *postRepository) Persist(p Persistence) {
	r.persistence = p
	r.load()
}

func (r *postRepository) save() {
	if r.persistence != nil {
		if err := r.persistence.Save(&r.posts); err != nil {
			log.Printf("Failed to save posts: %s", err)
		}
	}
}

func (r *postRepository) load() {
	if r.persistence != nil {
		if err := r.persistence.Load(&r.posts); err != nil {
			log.Printf("Failed to load posts: %s", err)
		}
	}
}

func (r *postRepository) addPost(data adding.PostData) (adding.Post, error) {
	item := &post{PostData{
		Id:          len(r.posts),
		SourceId:    data.SourceId,
		Title:       data.Title,
		Url:         data.Url,
		Description: data.Description,
		PublishedAt: data.PublishedAt,
		UpdatedAt:   data.UpdatedAt,
	}}

	r.posts = append(r.posts, item)
	return item, nil
}

func (r *postRepository) AddPost(data adding.PostData) (adding.Post, error) {
	post, err := r.addPost(data)
	if err == nil {
		r.save()
	}
	return post, err
}

func (r *postRepository) AddManyPosts(items ...adding.PostData) error {
	for _, item := range items {
		_, err := r.addPost(item)
		if err != nil {
			return err
		}
	}

	r.save()
	return nil
}

func (r *postRepository) GetSourcePosts(sourceId int) ([]listing.Post, error) {
	var result []listing.Post
	for _, item := range r.posts {
		if item.SourceId == sourceId {
			result = append(result, item)
		}
	}

	sortPosts(result)
	return result, nil
}

func (r *postRepository) GetFeedPosts(feedId int) ([]listing.SourcePost, error) {
	feed, _ := r.feedRepository.findFeed(feedId)
	if feed == nil {
		return nil, listing.ErrNotFound
	}

	sources := make(map[int]listing.Source, len(feed.SourceIds))

	var result []listing.SourcePost
	for _, item := range r.posts {
		for _, id := range feed.SourceIds {
			if item.SourceId == id {
				var source listing.Source
				if s, ok := sources[id]; ok {
					source = s
				} else {
					s, err := r.sourceRepository.GetSource(id)
					if err != nil {
						return nil, err
					}

					sources[id] = s
					source = s
				}

				post := &sourcePost{*item, source}
				result = append(result, post)
			}
		}
	}

	sortPosts(result)
	return result, nil
}

func (r *postRepository) RemoveSourcePost(sourceId int, postId int) error {
	for i, item := range r.posts {
		if item.SourceId == sourceId && item.Id() == postId {
			newCount := len(r.posts) - 1
			r.posts[i] = r.posts[newCount]
			r.posts = r.posts[:newCount]
			r.save()
			return nil
		}
	}

	return listing.ErrNotFound
}

func (r *postRepository) RemoveAllSourcePosts(sourceId int) error {
	var posts []*post
	for _, post := range r.posts {
		if post.SourceId != sourceId {
			posts = append(posts, post)
		}
	}

	r.posts = posts
	r.save()
	return nil
}

func (r *postRepository) UpdateSourcePost(sourceId int, postId int, data updating.Post) error {
	for _, item := range r.posts {
		if item.SourceId == sourceId && item.Id() == postId {
			item.PostData.Title = data.Title
			item.PostData.Url = data.Url
			item.PostData.Description = data.Description
			item.PostData.PublishedAt = data.PublishedAt
			item.PostData.UpdatedAt = data.UpdatedAt
			r.save()
			return nil
		}
	}

	return listing.ErrNotFound
}

func sortPosts(posts interface{}) {
	var timeOf func(i int) *time.Time

	switch posts := posts.(type) {
	case []listing.SourcePost:
		timeOf = func(i int) *time.Time { return posts[i].PublishedAt() }
		break
	case []listing.Post:
		timeOf = func(i int) *time.Time { return posts[i].PublishedAt() }
		break
	default:
		panic("unsupported type")
	}

	sort.Slice(posts, func(i, j int) bool {
		time1 := timeOf(i)
		time2 := timeOf(j)

		if time1 == nil {
			return time2 == nil
		}
		if time2 == nil {
			return time1 == nil
		}

		return time1.After(*time2)
	})
}
