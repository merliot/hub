//go:build tinygo

package hub

import "sync"

type mutex struct {
	sync.Mutex
}

type rwMutex struct {
	sync.RWMutex
}
