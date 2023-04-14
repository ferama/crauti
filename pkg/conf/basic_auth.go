package conf

type credentials struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type BasiAuth struct {
	Enabled     *bool         `yaml:"enabled,omitempty"`
	Realm       string        `yaml:"realm,omitempty"`
	Credentials []credentials `yaml:"credentials,omitempty"`
}

// Helper function that check for nil value on Enabled field
func (c *BasiAuth) IsEnabled() bool {
	return c.Enabled != nil && *c.Enabled
}

func (c *BasiAuth) clone() BasiAuth {
	enabled := *c.Enabled
	out := BasiAuth{
		Enabled: &enabled,
	}
	out.Credentials = append(out.Credentials, c.Credentials...)
	return out
}

// slice types needs manually merging logic
// When not defined (nil case) we should use the global values
// If defined but empty ([] case), we should use a nil value
func (c *BasiAuth) merge(target BasiAuth) {
	if target.Credentials == nil {
		c.Credentials = ConfInst.Middlewares.BasicAuth.Credentials
	} else if len(target.Credentials) == 0 {
		c.Credentials = nil
	}
}
