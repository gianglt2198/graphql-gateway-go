package annotation

// SkipRepository is an annotation to skip repository generation
type SkipRepository struct{}

func (SkipRepository) Name() string {
	return "SkipRepository"
}
