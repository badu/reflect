package reflector

import (
	"sync"

	. "reflect"
)

func init() {
	cachedModels = &safeModelsMap{l: new(sync.RWMutex), m: make(map[Type]*Model)}
}
