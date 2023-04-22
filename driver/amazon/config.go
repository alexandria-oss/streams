package amazon

// Config is the basic configuration schema for Amazon messaging services.
type Config struct {
	AccountID string // AWS Account identifier streams belongs to.
	Region    string // AWS Region a where streams are located.
}
