package conf

type cors struct {
	Enabled bool `yaml:"enabled,omitempty"`
}

func (c *cors) merge(target cors) {
	// c.Enabled = false
	// if target.Enabled == false
	// if target.Methods == nil {
	// 	c.Methods = ConfInst.Middlewares.Cache.Methods
	// } else if len(target.Methods) == 0 {
	// 	c.Methods = nil
	// }

	// if target.KeyHeaders == nil {
	// 	c.KeyHeaders = ConfInst.Middlewares.Cache.KeyHeaders
	// } else if len(target.KeyHeaders) == 0 {
	// 	c.KeyHeaders = nil
	// }
}
