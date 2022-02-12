package memory

import (
	"github.com/themisir/myfeed/pkg/adding"
	"github.com/themisir/myfeed/pkg/listing"
	"github.com/themisir/myfeed/pkg/updating"
	"log"
)

func NewFeedRepository() *feedRepository {
	return &feedRepository{
		feeds: []*feed{
			{FeedData{
				Id:        0,
				Name:      "Default",
				UserId:    "",
				SourceIds: []int{},
			}},
		},
	}
}

type FeedData struct {
	Id        int    `json:"id"`
	Name      string `json:"name"`
	UserId    string `json:"user_id"`
	SourceIds []int  `json:"source_ids"`
	IsPublic  bool   `json:"is_public"`
}

type feed struct {
	FeedData
}

func (f *feed) Id() int {
	return f.FeedData.Id
}

func (f *feed) Name() string {
	return f.FeedData.Name
}

func (f *feed) UserId() string {
	return f.FeedData.UserId
}

func (f *feed) IsPublic() bool {
	return f.FeedData.IsPublic
}

type feedRepository struct {
	feeds       []*feed
	persistence Persistence
}

func (r *feedRepository) Persist(p Persistence) {
	r.persistence = p
	r.load()
}

func (r *feedRepository) save() {
	if r.persistence != nil {
		if err := r.persistence.Save(&r.feeds); err != nil {
			log.Printf("Failed to save feeds: %s", err)
		}
	}
}

func (r *feedRepository) load() {
	if r.persistence != nil {
		if err := r.persistence.Load(&r.feeds); err != nil {
			log.Printf("Failed to load feeds: %s", err)
		}
	}
}

func (r *feedRepository) AddFeed(data adding.FeedData) (adding.Feed, error) {
	item := &feed{FeedData{
		Id:       len(r.feeds),
		Name:     data.Name,
		UserId:   data.UserId,
		IsPublic: data.IsPublic,
	}}

	r.feeds = append(r.feeds, item)
	r.save()
	return item, nil
}

func (r *feedRepository) GetUserFeeds(userId string) ([]listing.Feed, error) {
	var result []listing.Feed
	for _, feed := range r.feeds {
		if feed.UserId() == userId {
			result = append(result, feed)
		}
	}

	return result, nil
}

func (r *feedRepository) findFeed(feedId int) (*feed, int) {
	for i, feed := range r.feeds {
		if feed.Id() == feedId {
			return feed, i
		}
	}

	return nil, -1
}

func (r *feedRepository) GetFeed(feedId int) (listing.Feed, error) {
	feed, _ := r.findFeed(feedId)
	if feed != nil {
		return feed, nil
	}

	return nil, listing.ErrNotFound
}

func (r *feedRepository) RemoveFeed(feedId int) error {
	feed, i := r.findFeed(feedId)
	if feed != nil {
		newCount := len(r.feeds) - 1
		r.feeds[i] = r.feeds[newCount]
		r.feeds = r.feeds[:newCount]
		r.save()
		return nil
	}

	return listing.ErrNotFound
}

func (r *feedRepository) UpdateFeed(feedId int, data updating.Feed) error {
	feed, _ := r.findFeed(feedId)
	if feed != nil {
		feed.FeedData.Name = data.Name
		feed.FeedData.IsPublic = data.IsPublic
		r.save()
		return nil
	}

	return listing.ErrNotFound
}

func (r *feedRepository) UpdateFeedSources(feedId int, sourceIds ...int) error {
	feed, _ := r.findFeed(feedId)
	if feed != nil {
		feed.SourceIds = sourceIds
		r.save()
		return nil
	}

	return listing.ErrNotFound
}
