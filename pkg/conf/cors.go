package conf

type cors struct {
	// Do not use this directly. Use the IsEnabled function instead
	Enabled *bool `yaml:"enabled,omitempty"`
}

// Helper function that check for nil value on Enabled field
func (c *cors) IsEnabled() bool {
	return c.Enabled != nil && *c.Enabled
}
