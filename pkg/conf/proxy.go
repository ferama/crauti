package conf

type Proxy struct {
	// default false. If true, the ReverseProxy will not set
	// the Host header to the real upstream host while forwarding the
	// request to the upstream
	PreserveHostHeader *bool `yaml:"preserveHostHeader,omitempty"`
}

// Helper function that check for nil value on Enabled field
func (p *Proxy) IsHostHeaderPreerved() bool {
	return p.PreserveHostHeader != nil && *p.PreserveHostHeader
}

func (p *Proxy) clone() Proxy {
	preserveHostHeader := *p.PreserveHostHeader
	out := Proxy{
		PreserveHostHeader: &preserveHostHeader,
	}
	return out
}
