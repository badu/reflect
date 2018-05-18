/*
 * Copyright 2009-2018 The Go Authors. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package reflect

func newName(newName []byte) name {
	if len(newName) > 1<<16-1 {
		panic("reflect.x.error : name too long: " + string(newName))
	}
	var bits byte
	l := 1 + 2 + len(newName)
	b := make([]byte, l)
	b[0] = bits
	b[1] = uint8(len(newName) >> 8)
	b[2] = uint8(len(newName))
	copy(b[3:], newName)
	return name{bytes: &b[0]}
}

func (n name) data(offset int) *byte {
	return (*byte)(ptr(uintptr(ptr(n.bytes)) + uintptr(offset)))
}

func (n name) isExported() bool {
	return (*n.bytes)&(1<<0) != 0
}

func (n name) nameLen() int {
	if n.bytes == nil {
		return 0
	}
	return int(uint16(*n.data(1))<<8 | uint16(*n.data(2)))
}

func (n name) tagLen() int {
	if *n.data(0)&(1<<1) == 0 {
		return 0
	}
	offset := 3 + n.nameLen()
	return int(uint16(*n.data(offset))<<8 | uint16(*n.data(offset + 1)))
}

func (n name) name() []byte {
	if n.bytes == nil {
		return nil
	}
	info := (*[4]byte)(ptr(n.bytes))
	nameLen := int(info[1])<<8 | int(info[2])
	result := make([]byte, nameLen)
	header := (*stringHeader)(ptr(&result))
	header.Data = ptr(&info[3])
	header.Len = nameLen
	return result
}

func (n name) tag() []byte {
	tagLen := n.tagLen()
	if tagLen == 0 {
		return nil
	}
	nameLen := n.nameLen()
	result := make([]byte, tagLen)
	header := (*stringHeader)(ptr(&result))
	header.Data = ptr(n.data(5 + nameLen))
	header.Len = tagLen
	return result
}

func (n name) pkgPath() []byte {
	if n.bytes == nil || *n.data(0)&(1<<2) == 0 {
		return nil
	}
	offset := 3 + n.nameLen()
	if tagLen := n.tagLen(); tagLen > 0 {
		offset += 2 + tagLen
	}
	var nameOffset int32
	// Note that this field may not be aligned in memory, so we cannot use a direct int32 assignment here.
	copy((*[4]byte)(ptr(&nameOffset))[:], (*[4]byte)(ptr(n.data(offset)))[:])
	pkgPathName := name{(*byte)(resolveTypeOff(ptr(n.bytes), nameOffset))}
	return pkgPathName.name()
}
