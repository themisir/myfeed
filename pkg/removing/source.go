package removing

type SourceRepository interface {
	RemoveSource(sourceId int) error
}
