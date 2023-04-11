package conf

type rewrite struct {
	Pattern string `yaml:"pattern"`
	Target  string `yaml:"target"`
}

func (r *rewrite) clone() rewrite {
	out := rewrite{
		Pattern: r.Pattern,
		Target:  r.Target,
	}
	return out
}
