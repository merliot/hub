//go:build !tinygo

package device

import (
	"slices"
	"sync"
)

// Maker is a function type that creates new instances of a device model.
// It is used by the device system to instantiate devices of a particular model.
// Each Maker function returns a new Devicer implementation that represents
// a specific device instance's behavior and state.
type Maker func() Devicer

// Model represents a device model in the system. It contains:
// - Package: The Go package name where the model is implemented
// - Maker: A function to create new instances of this model
// - Config: The default configuration for devices of this model
type Model struct {
	Package string
	Maker
	Config
}

// Models is a map of device models keyed by model name.
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

func (mm *modelMap) unload() Models {
	var models = make(Models)
	mm.drange(func(name string, model *Model) bool {
		models[name] = model
		return true
	})
	return models
}

func (s *server) childModels(d *device) *modelMap {
	var models modelMap
	s.models.drange(func(name string, model *Model) bool {
		if slices.Contains(model.Config.Parents, d.Model) {
			models.Store(name, model)
		}
		return true
	})
	return &models
}
