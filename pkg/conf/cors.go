package conf

type Cors struct {
	// Do not use this directly. Use the IsEnabled function instead
	Enabled *bool `yaml:"enabled,omitempty"`
}

func (c *Cors) clone() Cors {
	enabled := *c.Enabled
	out := Cors{
		Enabled: &enabled,
	}
	return out
}

// Helper function that check for nil value on Enabled field
func (c *Cors) IsEnabled() bool {
	return c.Enabled != nil && *c.Enabled
}
