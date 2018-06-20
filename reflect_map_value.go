/*
 * Copyright 2009-2018 The Go Authors. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package reflect

import "unsafe"

// Len returns v's length.
func (v MapValue) Len() int {
	return maplen(v.pointer())
}

// MapIndex returns the value associated with key in the map v.
// It returns the zero Value if key is not found in the map or if v represents a nil map.
// As in Go, the key's value must be assignable to the map's key type.
func (v MapValue) MapIndex(key Value) Value {
	mapType := v.Type.ConvToMap()
	// Do not require key to be exported, so that DeepEqual and other programs can use all the keys returned by MapKeys as arguments to MapIndex. If either the map or the key is unexported, though, the result will be considered unexported.
	// This is consistent with the behavior for structs, which allow read but not write of unexported fields.
	key = key.assignTo(mapType.KeyType, nil)

	var keyPtr unsafe.Pointer
	if key.isPointer() {
		keyPtr = key.Ptr
	} else {
		keyPtr = unsafe.Pointer(&key.Ptr)
	}

	elemPtr := mapaccess(v.Type, v.pointer(), keyPtr)
	if elemPtr == nil {
		// we could return nil, but deep equal will panic
		return Value{}
	}

	mapElemType := mapType.ElemType
	fl := v.ro() | key.ro()
	fl |= Flag(mapElemType.Kind())
	if !mapElemType.isDirectIface() {
		return Value{mapElemType, convPtr(elemPtr), fl}
	}

	// Copy result so future changes to the map won't change the underlying value.
	mapElemValue := unsafeNew(mapElemType)
	typedmemmove(mapElemType, mapElemValue, elemPtr)
	return Value{mapElemType, mapElemValue, fl | pointerFlag}
}

// MapKeys returns a slice containing all the keys present in the map, in unspecified order.
// It returns an empty slice if v represents a nil map.
func (v MapValue) MapKeys() []Value {
	mapType := v.Type.ConvToMap()
	keyType := mapType.KeyType

	fl := v.ro() | Flag(keyType.Kind())

	mapPtr := v.pointer()
	mapLen := int(0)
	if mapPtr != nil {
		mapLen = maplen(mapPtr)
	}

	it := mapiterinit(v.Type, mapPtr)
	result := make([]Value, mapLen)
	var i int
	for i = 0; i < len(result); i++ {
		key := mapiterkey(it)
		if key == nil {
			// Someone deleted an entry from the map since we called maplen above. It's a data race, but nothing we can do about it.
			break
		}
		if keyType.isDirectIface() {
			// Copy result so future changes to the map won't change the underlying value.
			keyValue := unsafeNew(keyType)
			typedmemmove(keyType, keyValue, key)
			result[i] = Value{keyType, keyValue, fl | pointerFlag}
		} else {
			result[i] = Value{keyType, convPtr(key), fl}
		}
		mapiternext(it)
	}
	return result[:i]
}

// SetMapIndex sets the value associated with key in the map v to val.
// If val is the zero Value, SetMapIndex deletes the key from the map.
// Otherwise if v holds a nil map, SetMapIndex will panic.
// As in Go, key's value must be assignable to the map's key type, and val's value must be assignable to the map's value type.
func (v MapValue) SetMapIndex(key, value Value) {
	if !v.IsValid() || !v.isExported() {
		if willPrintDebug {
			println("reflect.MapValue.SetMapIndex: map must be exported")
		}
		return
	}
	if !key.IsValid() || !key.isExported() {
		if willPrintDebug {
			println("reflect.MapValue.SetMapIndex: key must be exported")
		}
		return
	}

	mapType := v.Type.ConvToMap()
	key = key.assignTo(mapType.KeyType, nil)
	var keyPtr unsafe.Pointer
	if key.isPointer() {
		keyPtr = key.Ptr
	} else {
		keyPtr = unsafe.Pointer(&key.Ptr)
	}

	if value.Type == nil {
		// this allows us to delete from map, when setting key to nil value
		mapdelete(v.Type, v.pointer(), keyPtr)
		return
	}

	// now we check the value too - we don't check this earlier, because delete operations
	if !value.IsValid() || !value.isExported() {
		return
	}

	value = value.assignTo(mapType.ElemType, nil)
	var elemPtr unsafe.Pointer
	if value.isPointer() {
		elemPtr = value.Ptr
	} else {
		elemPtr = unsafe.Pointer(&value.Ptr)
	}
	mapassign(v.Type, v.pointer(), keyPtr, elemPtr)
}
