package device

import "slices"

type Maker func() Devicer

type Model struct {
	Package string
	Maker
	Config
}

type ModelMap map[string]Model // key: model name

func (s *server) childModels(d *device) ModelMap {
	var models = make(ModelMap)
	for name, model := range s.models {
		if slices.Contains(model.Config.Parents, d.Model) {
			models[name] = model
		}
	}
	return models
}
