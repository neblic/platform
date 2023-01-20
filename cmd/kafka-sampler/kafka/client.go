package kafka

type Client interface {
	// Topics returns the set of available topics as retrieved from cluster metadata.
	Topics() ([]string, error)

	// RefreshMetadata takes a list of topics and queries the cluster to refresh the
	// available metadata for those topics. If no topics are provided, it will refresh
	// metadata for all topics.
	RefreshMetadata(topics ...string) error

	Close() error
}
