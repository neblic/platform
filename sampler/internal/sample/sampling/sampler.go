package sampling

type Sampler interface {
	Sample(determinant string) bool
}
