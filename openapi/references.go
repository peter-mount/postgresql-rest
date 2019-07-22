package openapi

import "strings"

type Reference struct {
	Ref string `yaml:"$ref,omitempty"`
}

func (r Reference) refSplit(c int) string {
	s := strings.Split(r.Ref, "/")
	if len(s) < c {
		return ""
	}
	return s[c-1]
}

func (r Reference) RefSection() string {
	return r.refSplit(2)
}

func (r Reference) RefType() string {
	return r.refSplit(3)
}

func (r Reference) RefName() string {
	return r.refSplit(4)
}
