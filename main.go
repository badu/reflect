package reflector

import (
	"reflect"
	"sync"
)

func init() {
	cachedModels = &safeModelsMap{l: new(sync.RWMutex), m: make(map[reflect.Type]*Model)}
}