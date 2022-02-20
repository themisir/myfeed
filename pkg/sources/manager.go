package sources

import (
	"fmt"
	"github.com/themisir/myfeed/pkg/listing"
	"github.com/themisir/myfeed/pkg/log"
	"time"

	"github.com/themisir/myfeed/pkg/adding"
	"github.com/themisir/myfeed/pkg/models"
	"github.com/themisir/myfeed/pkg/updating"
)

func NewManager(sourceRepository models.SourceRepository, postRepository models.PostRepository, feedUpdating updating.FeedRepository, logger log.Logger) *Manager {
	return &Manager{
		sourceRepository: sourceRepository,
		postRepository:   postRepository,
		feedUpdating:     feedUpdating,
		resolver:         &resolver{},
		queue:            make(chan sourceQueueEntry, 32),
		logger:           logger,
	}
}

type Manager struct {
	sourceRepository models.SourceRepository
	postRepository   models.PostRepository
	feedUpdating     updating.FeedRepository
	logger           log.Logger

	resolver Resolver

	queue chan sourceQueueEntry
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
	// Start periodic timer
	go m.runTimer()

	// Create 4 worker goroutine for processing sources
	for i := 0; i < 4; i++ {
		go m.processSources()
	}

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
	m.queue <- sourceQueueEntry{
		id:  source.Id(),
		url: source.Url(),
	}
}

func (m *Manager) processSources() {
	for {
		source := <-m.queue

		// Resolve source
		resolved, err := m.resolver.Resolve(source.url)
		if err != nil {
			m.logger.Errorf("failed to process source %v on '%s': %s", source.id, source.url, err)
			continue
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
	}
}

func (m *Manager) runTimer() {
	t := time.NewTimer(10 * time.Minute)
	defer t.Stop()

	for {
		select {
		case <-t.C:
			// Clean up
			if err := m.sourceRepository.RemoveEmptySources(); err != nil {
				m.logger.Errorf("failed to clean up unused sources: %s", err)
			}

			// Update
			if err := m.enqueueExistingSources(); err != nil {
				m.logger.Errorf("failed to enqueue existing sources: %s", err)
			}
		}
	}
}
