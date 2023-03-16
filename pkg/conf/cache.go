package conf

import "time"

type redis struct {
	Host     string `yaml:"host,omitempty"`
	Port     int    `yaml:"port,omitempty"`
	Password string `yaml:"password,omitempty"`
}

type Cache struct {
	Enabled    bool          `yaml:"enabled"`
	Redis      redis         `yaml:"redis,omitempty"`
	TTL        time.Duration `yaml:"cacheTTL,omitempty"`
	Methods    []string      `yaml:"methods,omitempty"`
	KeyHeaders []string      `yaml:"keyHeaders,omitempty"`
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
}
