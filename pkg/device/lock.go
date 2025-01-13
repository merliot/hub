//go:build !tinygo

package device

//import "sync"

import (
	"github.com/ietxaniz/delock"
)

/*
type mutex struct {
	sync.Mutex
}

type rwMutex struct {
	sync.RWMutex
}
*/

type mutex struct {
	mu delock.Mutex
	id int
}

type rwMutex struct {
	mu delock.RWMutex
	id int
}

func (m *mutex) Lock() {
	var err error
	m.id, err = m.mu.Lock()
	if err != nil {
		panic(err)
	}
}

func (m *mutex) Unlock() {
	m.mu.Unlock(m.id)
}

func (m *rwMutex) Lock() {
	var err error
	m.id, err = m.mu.Lock()
	if err != nil {
		panic(err)
	}
}

func (m *rwMutex) Unlock() {
	m.mu.Unlock(m.id)
}

func (m *rwMutex) RLock() {
	var err error
	m.id, err = m.mu.RLock()
	if err != nil {
		panic(err)
	}
}

func (m *rwMutex) RUnlock() {
	m.mu.RUnlock(m.id)
}
