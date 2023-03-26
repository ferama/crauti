package conf

type proxy struct {
	// default false. If true, the ReverseProxy will not set
	// the Host header to the real upstream host while forwarding the
	// request to the upstream
	PreserveHostHeader *bool `yaml:"preserveHostHeader,omitempty"`
}

// Helper function that check for nil value on Enabled field
func (p *proxy) IsHostHeaderPreserved() bool {
	return p.PreserveHostHeader != nil && *p.PreserveHostHeader
}

func (p *proxy) clone() proxy {
	preserveHostHeader := *p.PreserveHostHeader
	out := proxy{
		PreserveHostHeader: &preserveHostHeader,
	}
	return out
}
