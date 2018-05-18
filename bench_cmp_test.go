/*
 * Copyright 2009-2018 The Go Authors. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package reflect_test

import (
	"github.com/badu/reflect"
	systemReflect "reflect"
	"testing"
)

func BenchmarkReflect(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		value := reflect.ReflectOn(Invoice{})
		value.Type.Fields(func(Type *reflect.RType, name []byte, tag []byte, pack []byte, embedded, exported bool, offset uintptr, index int) {

		})
	}
}

func BenchmarkOldReflect(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		value := systemReflect.ValueOf(Invoice{})
		for j := 0; j < value.Type().NumField(); j++ {
			value.Type().Field(j)
		}
	}
}
