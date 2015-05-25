package show

import (
	dlpb "diektronics.com/carter/dl/protos/dl"
)

type Show struct {
	Name string
	Eps  string
	Blob string
	Down *dlpb.Down
}

type ByAlpha []*Show

func (s ByAlpha) Len() int      { return len(s) }
func (s ByAlpha) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s ByAlpha) Less(i, j int) bool {
	return s[i].Name < s[j].Name || s[i].Name == s[j].Name && s[i].Eps < s[j].Eps
}
