package conf

import "time"

type timeout struct {
	// Do not use this directly. Use the IsEnabled function instead
	Duration time.Duration `yaml:"duration,omitempty"`
}
