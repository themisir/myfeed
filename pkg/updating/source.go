package updating

type Source struct {
	Title string
}

type SourceRepository interface {
	UpdateSource(sourceId int, data Source) error
}
