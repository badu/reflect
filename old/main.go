/*
 * Copyright 2009-2018 The Go Authors. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package old

import (
	"sync"

	. "reflect"
)

func init() {
	cachedModels = &safeModelsMap{l: new(sync.RWMutex), m: make(map[Type]*Model)}
}
