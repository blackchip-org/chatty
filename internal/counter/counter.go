package counter

import (
	"sync"
)

var (
	current = uint64(1)
	mutex   sync.Mutex
)

func Next() uint64 {
	mutex.Lock()
	defer mutex.Unlock()

	val := current
	current++
	return val
}

func Reset() {
	mutex.Lock()
	defer mutex.Unlock()
	current = 1
}
