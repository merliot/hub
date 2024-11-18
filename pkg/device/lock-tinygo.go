//go:build tinygo

package device

import "sync"

type mutex struct {
	sync.Mutex
}

type rwMutex struct {
	sync.RWMutex
}
