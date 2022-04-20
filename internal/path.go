package internal

import (
	"fmt"
	"net/url"
	"path"
	"strings"
)

type PartType int

const (
	PartStatic PartType = iota
	PartPlaceholder
	PartWildcard
)

type Part struct {
	Name string
	Type PartType
}

type Path string

func (p Path) Split() []Part {
	var ans []Part
	for _, part := range strings.Split(string(p), "/") {
		if len(part) == 0 {
			continue
		}
		switch part[0] {
		case ':':
			ans = append(ans, Part{
				Name: part[1:],
				Type: PartPlaceholder,
			})
		case '*':
			ans = append(ans, Part{
				Name: part[1:],
				Type: PartWildcard,
			})
		default:
			ans = append(ans, Part{
				Name: part,
				Type: PartStatic,
			})
		}
	}
	return ans
}

func (p Path) Build(args ...string) (string, error) {
	var ans []string
LOOP:
	for _, arg := range p.Split() {
		switch arg.Type {
		case PartStatic:
			ans = append(ans, arg.Name)
		case PartPlaceholder:
			if len(args) == 0 {
				return "", fmt.Errorf("not enough arguments to build path")
			}
			param := args[0]
			args = args[1:]
			ans = append(ans, url.PathEscape(param))
		case PartWildcard:
			for _, param := range args {
				ans = append(ans, url.PathEscape(param))
			}
			break LOOP
		}
	}
	return "/" + path.Join(ans...), nil
}
