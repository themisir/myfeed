package sources

import (
	"fmt"
	"github.com/themisir/myfeed/pkg/listing"
	"time"

	"github.com/themisir/myfeed/pkg/adding"
	"github.com/themisir/myfeed/pkg/models"
	"github.com/themisir/myfeed/pkg/updating"
)

func NewManager(sourceRepository models.SourceRepository, postRepository models.PostRepository, feedUpdating updating.FeedRepository) *Manager {
	return &Manager{
		sourceRepository: sourceRepository,
		postRepository:   postRepository,
		feedUpdating:     feedUpdating,
		resolver:         &resolver{},
		immediateQueue:   make(chan sourceQueueEntry),
		delayedQueue:     make(chan sourceQueueEntry),
	}
}

type Manager struct {
	sourceRepository models.SourceRepository
	postRepository   models.PostRepository
	feedUpdating     updating.FeedRepository

	resolver Resolver

	immediateQueue chan sourceQueueEntry
	delayedQueue   chan sourceQueueEntry
}

type sourceQueueEntry struct {
	id  int
	url string
}

func (m *Manager) UpdateFeedSources(feedId int, sourceUrls ...string) error {
	sourceIds := make([]int, len(sourceUrls))

	for i, url := range sourceUrls {
		source, _ := m.sourceRepository.FindSourceByUrl(url)
		if source != nil {
			sourceIds[i] = source.Id()
		} else {
			// Create a new source
			source, err := m.sourceRepository.AddSource(adding.SourceData{
				Title: "Processing...",
				Url:   url,
			})
			if err != nil {
				return fmt.Errorf("failed to add source: %s", err)
			}

			// Enqueue source for processing
			go m.enqueue(source)

			sourceIds[i] = source.Id()
		}
	}

	return m.feedUpdating.UpdateFeedSources(feedId, sourceIds...)
}

func (m *Manager) Start() error {
	go m.processSources()
	go m.processDelayedQueue()

	return m.enqueueExistingSources()
}

func (m *Manager) enqueueExistingSources() error {
	sources, err := m.sourceRepository.GetSources()
	if err != nil {
		return err
	}

	for _, source := range sources {
		go m.enqueue(source)
	}
	return nil
}

func (m *Manager) enqueue(source listing.Source) {
	m.immediateQueue <- sourceQueueEntry{
		id:  source.Id(),
		url: source.Url(),
	}
}

func (m *Manager) processSources() {
	for {
		source := <-m.immediateQueue

		resolved, err := m.resolver.Resolve(source.url)
		if err != nil {
			fmt.Printf("%s", err)
			continue
			// TODO: handle error
		}

		// Update source details
		_ = m.sourceRepository.UpdateSource(source.id, updating.Source{
			Title: resolved.Title,
		})

		// Map resolved items into posts
		posts := make([]adding.PostData, len(resolved.Items))
		for i, item := range resolved.Items {
			posts[i] = adding.PostData{
				SourceId:    source.id,
				Title:       item.Title,
				Description: item.Description,
				Url:         item.Url,
				PublishedAt: item.PublishedAt,
				UpdatedAt:   item.UpdatedAt,
			}
		}

		// Update cached posts
		_ = m.postRepository.RemoveAllSourcePosts(source.id)
		_ = m.postRepository.AddManyPosts(posts...)

		go func() {
			m.delayedQueue <- source
		}()
	}
}

func (m *Manager) processDelayedQueue() {
	for {
		time.Sleep(3 * time.Minute)
		go func() {
			m.immediateQueue <- <-m.delayedQueue
		}()
	}
}
