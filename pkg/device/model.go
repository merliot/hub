package device

import "slices"

type Maker func() Devicer

type Model struct {
	Package string
	Maker
	Config
}

type ModelMap map[string]Model // key: model name

var Models = ModelMap{}
