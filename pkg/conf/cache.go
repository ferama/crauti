package conf

import "time"

type Cache struct {
	// Do not use this directly. Use the IsEnabled function instead
	// See the conf.Update() function for more details
	Enabled    *bool         `yaml:"enabled,omitempty"`
	TTL        time.Duration `yaml:"TTL,omitempty"`
	Methods    []string      `yaml:"methods,omitempty"`
	KeyHeaders []string      `yaml:"keyHeaders,omitempty"`
	KeyClaims  []string      `yaml:"keyClaims,omitempty"`
}

func (c *Cache) clone() Cache {
	enabled := *c.Enabled
	out := Cache{
		Enabled: &enabled,
		TTL:     c.TTL,
	}
	out.Methods = append(out.Methods, c.Methods...)
	out.KeyHeaders = append(out.KeyHeaders, c.KeyHeaders...)
	out.KeyClaims = append(out.KeyClaims, c.KeyClaims...)
	return out
}

// Helper function that check for nil value on Enabled field
func (c *Cache) IsEnabled() bool {
	return c.Enabled != nil && *c.Enabled
}

// slice types needs manually merging logic
// When not defined (nil case) we should use the global values
// If defined but empty ([] case), we should use a nil value
func (c *Cache) merge(target Cache) {
	if target.Methods == nil {
		c.Methods = ConfInst.Middlewares.Cache.Methods
	} else if len(target.Methods) == 0 {
		c.Methods = nil
	}

	if target.KeyHeaders == nil {
		c.KeyHeaders = ConfInst.Middlewares.Cache.KeyHeaders
	} else if len(target.KeyHeaders) == 0 {
		c.KeyHeaders = nil
	}

	if target.KeyClaims == nil {
		c.KeyClaims = ConfInst.Middlewares.Cache.KeyClaims
	} else if len(target.KeyClaims) == 0 {
		c.KeyClaims = nil
	}
}
