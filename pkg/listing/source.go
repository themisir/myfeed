package listing

type (
	Source interface {
		Id() int
		Title() string
		Url() string
	}
	SourceRepository interface {
		GetSource(sourceId int) (Source, error)
		GetSources() ([]Source, error)
		GetFeedSources(feedId int) ([]Source, error)
		FindSourceByUrl(url string) (Source, error)
	}
)
