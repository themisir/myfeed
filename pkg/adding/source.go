package adding

type (
	SourceData struct {
		Title string
		Url   string
	}
	Source interface {
		Id() int
		Title() string
		Url() string
	}
	SourceRepository interface {
		AddSource(data SourceData) (Source, error)
	}
)
