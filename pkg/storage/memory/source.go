package memory

import (
	"github.com/themisir/myfeed/pkg/adding"
	"github.com/themisir/myfeed/pkg/listing"
	"github.com/themisir/myfeed/pkg/updating"
	"log"
)

func NewSourceRepository(feedRepository *feedRepository) *sourceRepository {
	return &sourceRepository{
		sources:        []*source{},
		feedRepository: feedRepository,
	}
}

type SourceData struct {
	Id    int    `json:"id"`
	Title string `json:"title"`
	Url   string `json:"url"`
}

type source struct {
	SourceData
}

func (s *source) Id() int {
	return s.SourceData.Id
}

func (s *source) Title() string {
	return s.SourceData.Title
}

func (s *source) Url() string {
	return s.SourceData.Url
}

type sourceRepository struct {
	sources        []*source
	feedRepository *feedRepository
	persistence    Persistence
}

func (r *sourceRepository) Persist(p Persistence) {
	r.persistence = p
	r.load()
}

func (r *sourceRepository) save() {
	if r.persistence != nil {
		if err := r.persistence.Save(&r.sources); err != nil {
			log.Printf("Failed to save sources: %s", err)
		}
	}
}

func (r *sourceRepository) load() {
	if r.persistence != nil {
		if err := r.persistence.Load(&r.sources); err != nil {
			log.Printf("Failed to load sources: %s", err)
		}
	}
}

func (r *sourceRepository) GetSource(sourceId int) (listing.Source, error) {
	for _, source := range r.sources {
		if source.Id() == sourceId {
			return source, nil
		}
	}

	return nil, listing.ErrNotFound
}

func (r *sourceRepository) GetSources() ([]listing.Source, error) {
	result := make([]listing.Source, len(r.sources))
	for i, source := range r.sources {
		result[i] = source
	}

	return result, nil
}

func (r *sourceRepository) AddSource(data adding.SourceData) (adding.Source, error) {
	item := &source{SourceData{
		Id:    len(r.sources),
		Title: data.Title,
		Url:   data.Url,
	}}
	r.sources = append(r.sources, item)
	r.save()
	return item, nil
}

func (r *sourceRepository) GetFeedSources(feedId int) ([]listing.Source, error) {
	feed, _ := r.feedRepository.findFeed(feedId)
	if feed == nil {
		return nil, listing.ErrNotFound
	}

	var result []listing.Source
	for _, item := range r.sources {
		for _, id := range feed.SourceIds {
			if item.Id() == id {
				result = append(result, item)
			}
		}
	}

	return result, nil
}

func (r *sourceRepository) FindSourceByUrl(url string) (listing.Source, error) {
	for _, source := range r.sources {
		if source.Url() == url {
			return source, nil
		}
	}

	return nil, listing.ErrNotFound
}

func (r *sourceRepository) RemoveSource(sourceId int) error {
	for i, source := range r.sources {
		if source.Id() == sourceId {
			newCount := len(r.sources) - 1
			r.sources[i] = r.sources[newCount]
			r.sources = r.sources[:newCount]
			r.save()
			return nil
		}
	}

	return listing.ErrNotFound
}

func (r *sourceRepository) UpdateSource(sourceId int, data updating.Source) error {
	for _, source := range r.sources {
		if source.Id() == sourceId {
			source.SourceData.Title = data.Title
			r.save()
			return nil
		}
	}

	return listing.ErrNotFound
}
