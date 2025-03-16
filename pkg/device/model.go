//go:build !tinygo

package device

import (
	"slices"
	"sync"
)

type Maker func() Devicer

type Model struct {
	Package string
	Maker
	Config
}

type Models map[string]*Model

type modelMap struct {
	sync.Map // key: model name, value: *Model
}

func (mm *modelMap) drange(f func(string, *Model) bool) {
	mm.Range(func(key, value any) bool {
		name := key.(string)
		m := value.(*Model)
		return f(name, m)
	})
}

func (mm *modelMap) get(name string) (*Model, bool) {
	value, ok := mm.Load(name)
	if !ok {
		return nil, false
	}
	return value.(*Model), true
}

func (mm *modelMap) length() int {
	l := 0
	mm.Range(func(key, value any) bool {
		l++
		return true
	})
	return l
}

func (mm *modelMap) load(models Models) {
	mm.Clear()
	for name, model := range models {
		mm.Store(name, model)
	}
}

func (mm modelMap) unload() Models {
	var models = make(Models)
	mm.drange(func(name string, model *Model) bool {
		models[name] = model
		return true
	})
	return models
}

func (s *server) childModels(d *device) modelMap {
	var models modelMap
	s.models.drange(func(name string, model *Model) bool {
		if slices.Contains(model.Config.Parents, d.Model) {
			models.Store(name, model)
		}
		return true
	})
	return models
}
