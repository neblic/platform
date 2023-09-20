package digest

type ValueOptions struct {
	// Maximum number of fields to process when processing a sample
	MaxProcessedFields int
}

func DefaultValueOptions() ValueOptions {
	return ValueOptions{
		MaxProcessedFields: 100,
	}
}

func (o ValueOptions) WithMaxProcessedFields(num int) ValueOptions {
	o.MaxProcessedFields = num

	return o
}
