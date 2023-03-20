package conf

type cors struct {
	// Do not use this directly. Use the IsEnabled function instead
	Enabled *bool `yaml:"enabled,omitempty"`
}

func (c *cors) clone() cors {
	enabled := *c.Enabled
	out := cors{
		Enabled: &enabled,
	}
	return out
}

// Helper function that check for nil value on Enabled field
func (c *cors) IsEnabled() bool {
	return c.Enabled != nil && *c.Enabled
}
