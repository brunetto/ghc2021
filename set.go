package main

type Set map[string]nothing
type nothing struct{}

func NewSet() Set {
	return Set{}
}

func (s Set) Add(items ...string) Set {
	if s == nil {
		s = NewSet()
	}

	for _, i := range items {
		s[i] = nothing{}
	}

	return s
}

func (s Set) Contains(item string) bool {
	_, exists := s[item]
	return exists
}

func (s Set) Intersect(other Set) Set {
	var smaller, larger Set

	if len(s) < len(other) {
		smaller = s
		larger = other
	} else {
		smaller = other
		larger = s
	}

	res := make(Set, len(smaller))

	for item := range smaller {
		if larger.Contains(item) {
			res.Add(item)
		}
	}

	return res
}
