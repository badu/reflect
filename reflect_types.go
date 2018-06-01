/*
 * Copyright 2009-2018 The Go Authors. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package reflect

import (
	"errors"
	"runtime"
	"unsafe"

	systemReflect "reflect"
)

const (
	maxUint = ^uint(0)
	maxInt  = int(maxUint >> 1)
)

const (
	kindWidthFlag = 5 // there are 27 kinds

	// The lowest bits are flag bits:
	stickyROFlag    Flag = 1 << 5 // obtained via unexported not embedded field, so read-only
	embedROFlag     Flag = 1 << 6 // obtained via unexported embedded field, so read-only
	pointerFlag     Flag = 1 << 7 // value holds a pointer to the data
	addressableFlag Flag = 1 << 8 // v.CanAddr is true (implies pointerFlag)
	methodFlag      Flag = 1 << 9 // v is a method value
	// The next five bits give the Kind of the value.
	// This repeats Type.Kind() except for method values.
	kindMaskFlag Flag = 1<<kindWidthFlag - 1
	exportFlag   Flag = stickyROFlag | embedROFlag

	// The remaining 23+ bits give a method number for method values.
	// If flag.kind() != Func, code can assume that methodFlag is unset.
	// If isDirectIface(Type), code can assume that pointerFlag is set.
	methodShiftFlag = 10

	willPrintDebug = true

	kindDirectIface = 1 << 5
	kindGCProg      = 1 << 6 // Type.gc points to GC program
	kindNoPointers  = 1 << 7
	kindMask        = (1 << 5) - 1
	// hasExtraInfoFlag means that there is a pointer, *info, just beyond the outer type structure.
	//
	// For example, if t.Kind() == Struct and t.extraTypeFlag&hasExtraInfoFlag != 0,
	// then t has info data and it can be accessed as:
	//
	// 	type tUncommon struct {
	// 		structType
	// 		u info
	// 	}
	// 	u := &(*tUncommon)(unsafe.Pointer(t)).u
	hasExtraInfoFlag extraFlag = 1 << 0

	// hasExtraStarFlag means the name in the str field has an
	// extraneous '*' prefix. This is because for most types T in
	// a program, the type *T also exists and reusing the str data
	// saves binary size.
	hasExtraStarFlag extraFlag = 1 << 1

	// hasNameFlag means the type has a name.
	hasNameFlag extraFlag = 1 << 2

	maxPtrMaskBytes  = 2048 // See cmd/compile/x/gc/reflect.go for derivation of constant.
	additionalOffset = unsafe.Sizeof(uncommonType{})

	PtrSize    = 4 << (^uintptr(0) >> 63) // unsafe.Sizeof(uintptr(0)) but an ideal const
	IsAMD64p32 = runtime.GOARCH == "amd64p32"

	// Make sure these routines stay in sync with ../../runtime/hashmap.go!
	// These types exist only for GC, so we only fill out GC relevant info.
	// Currently, that's just size and the GC program. We also fill in string
	// for possible debugging use.
	bucketSize uintptr = 8
	maxKeySize uintptr = 128
	maxValSize uintptr = 128

	comma     byte = ','
	openPar   byte = '('
	closePar  byte = ')'
	sqOpenPar byte = '['
	sqClosPar byte = ']'
	star      byte = '*'

	mapStr    = "map"
	bucketStr = "bucket"
	methStr   = "methodargs"
	fnStr     = "funcargs"
)

const (
	Invalid       Kind = iota // 0
	Bool                      // 1
	Int                       // 2
	Int8                      // 3
	Int16                     // 4
	Int32                     // 5
	Int64                     // 6
	Uint                      // 7
	Uint8                     // 8
	Uint16                    // 9
	Uint32                    // 10
	Uint64                    // 11
	UintPtr                   // 12
	Float32                   // 13
	Float64                   // 14
	Complex64                 // 15
	Complex128                // 16
	Array                     // 17
	Chan                      // 18
	Func                      // 19
	Interface                 // 20
	Map                       // 21
	Ptr                       // 22
	Slice                     // 23
	String                    // 24
	Struct                    // 25
	UnsafePointer             // 26
)

var (
	uint8Type *RType
	kindNames = []string{
		Invalid:       "invalid",
		Bool:          "bool",
		Int:           "int",
		Int8:          "int8",
		Int16:         "int16",
		Int32:         "int32",
		Int64:         "int64",
		Uint:          "uint",
		Uint8:         "uint8",
		Uint16:        "uint16",
		Uint32:        "uint32",
		Uint64:        "uint64",
		UintPtr:       "uintptr",
		Float32:       "float32",
		Float64:       "float64",
		Complex64:     "complex64",
		Complex128:    "complex128",
		Array:         "array",
		Chan:          "chan",
		Func:          "func",
		Interface:     "interface",
		Map:           "map",
		Ptr:           "ptr",
		Slice:         "slice",
		String:        "string",
		Struct:        "struct",
		UnsafePointer: "unsafe.Pointer",
	}
	ErrSyntax = errors.New("invalid syntax")
)

type (
	// Aliases
	// -------
	ptr     = unsafe.Pointer // alias, for readability
	nameOff = int32          // offset to a name
	typeOff = int32          // offset to an *Type
	textOff = int32          // offset from top of text section
	Flag    = uintptr
	// extraTypeFlag is used by an Type to signal what extra type information is available in the memory directly following the Type value.
	//
	// extraTypeFlag values must be kept in sync with copies in:
	// 	cmd/compile/x/gc/reflect.go
	// 	cmd/link/x/ld/decodesym.go
	// 	runtime/type.go
	extraFlag = uint8
	// A Kind represents the specific kind of type that a Type represents. The zero Kind is not a valid kind.
	Kind = uint

	// Types
	// -----

	// a copy of runtime.typeAlg
	// (COMPILER)
	algo struct {
		hash  func(unsafe.Pointer, uintptr) uintptr     // function for hashing objects of this type (ptr to object, seed) -> hash
		equal func(unsafe.Pointer, unsafe.Pointer) bool // function for comparing objects of this type (ptr to object A, ptr to object B) -> ==?
	}

	// Method on non-interface type
	// (COMPILER)
	method struct {
		nameOffset nameOff // name of method
		typeOffset typeOff // method type (without receiver)
		ifaceCall  textOff // fn used in interface call (one-word receiver)
		normCall   textOff // fn used for normal method call
	}
	// uncommonType is present only for types with names or methods
	// (if T is a named type, the uncommonTypes for T and *T have methods).
	// Using a pointer to this struct reduces the overall size required
	// to describe an unnamed type with no methods.
	// (COMPILER)
	uncommonType struct {
		pkgPath nameOff // import path; empty for built-in types like int, string
		mCount  uint16  // number of methods
		_       uint16  // unused (future exported methods)
		mOffset uint32  // offset from this uncommontypeto [mCount]method
		_       uint32  // unused
	}
	// (COMPILER)
	uncommonStruct struct {
		structType
		u uncommonType
	}
	// (COMPILER)
	uncommonPtr struct {
		ptrType
		u uncommonType
	}
	// (COMPILER)
	uncommonFunc struct {
		funcType
		u uncommonType
	}
	// (COMPILER)
	uncommonSlice struct {
		sliceType
		u uncommonType
	}
	// (COMPILER)
	uncommonArray struct {
		arrayType
		u uncommonType
	}
	// (COMPILER)
	uncommonInterface struct {
		ifaceType
		u uncommonType
	}
	// (COMPILER)
	uncommonConcrete struct {
		RType
		u uncommonType
	}
	// arrayType represents a fixed array type.
	// (COMPILER)
	arrayType struct {
		RType     `reflect:"array"`
		ElemType  *RType // array element type
		SliceType *RType // slice type
		Len       uintptr
	}

	// funcType represents a function type.
	//
	// A *Type for each in and out parameter is stored in an array that
	// directly follows the funcType (and possibly its info). So
	// a function type with one method, one input, and one output is:
	//
	// 	struct {
	// 		funcType
	// 		info
	// 		[2]*Type    // [0] is in, [1] is out
	// 	}
	// (COMPILER)
	funcType struct {
		RType  `reflect:"func"`
		InLen  uint16
		OutLen uint16 // top bit is set if last input parameter is ...
	}

	// ifaceMethod represents a method on an interface type
	// (COMPILER)
	ifaceMethod struct {
		nameOffset nameOff // name of method
		typeOffset typeOff // .(*MethodType) underneath
	}

	// ifaceType represents an interface type.
	// (COMPILER)
	ifaceType struct {
		RType   `reflect:"interface"`
		pkgPath name          // import path
		methods []ifaceMethod // sorted by hash
	}

	// mapType represents a map type.
	// (COMPILER)
	mapType struct {
		RType          `reflect:"map"`
		KeyType        *RType // map key type
		ElemType       *RType // map element (value) type
		bucket         *RType // x bucket structure
		header         *RType // x map header
		keySize        uint8  // size of key slot
		indirectKey    uint8  // store ptr to key instead of key itself
		valueSize      uint8  // size of value slot
		indirectValue  uint8  // store ptr to value instead of value itself
		bucketSize     uint16 // size of bucket
		reflexiveKey   bool   // true if k==k for all keys
		needsKeyUpdate bool   // true if we need to update key on an overwrite
	}

	// ptrType represents a pointer type.
	// (COMPILER)
	ptrType struct {
		RType `reflect:"ptr"`
		Type  *RType // pointer element (pointed at) type
	}

	// sliceType represents a slice type.
	// (COMPILER)
	sliceType struct {
		RType    `reflect:"slice"`
		ElemType *RType // slice element type
	}

	// Struct field (CORE)
	// (COMPILER)
	structField struct {
		name        name    // name is always non-empty
		Type        *RType  // type of field
		offsetEmbed uintptr // byte offset of field<<1 | isEmbed
	}

	// structType represents a struct type.
	// (COMPILER)
	structType struct {
		RType   `reflect:"struct"`
		pkgPath name
		fields  []structField // sorted by offset
	}

	// name is an encoded type name with optional extra data.
	//
	// The first byte is a bit field containing:
	//
	// 	1<<0 the name is exported
	// 	1<<1 tag data follows the name
	// 	1<<2 pkgPath nameOff follows the name and tag
	//
	// The next two bytes are the data length:
	//
	// 	 l := uint16(data[1])<<8 | uint16(data[2])
	//
	// Bytes [3:3+l] are the string data.
	//
	// If tag data follows then bytes 3+l and 3+l+1 are the tag length,
	// with the data following.
	//
	// If the import path follows, then 4 bytes at the end of
	// the data form a nameOff. The import path is only set for concrete
	// methods that are defined in a different package than their type.
	//
	// If a name starts with "*", then the exported bit represents
	// whether the pointed to type is exported.
	name struct {
		bytes *byte
	}

	// Layout matches runtime.gobitvector (well enough).
	bitVector struct {
		num  uint32 // number of bits
		data []byte
	}

	// ifaceRtype is the header for an interface{} value.
	// (COMPILER)
	ifaceRtype struct {
		Type *RType
		word unsafe.Pointer
	}

	// nonEmptyInterface is the header for an interface value with methods.
	// see ../runtime/iface.go:/Itab
	// (COMPILER)
	concreteRtype struct {
		iTab *struct {
			IfaceType *RType                 // static interface type
			Type      *RType                 // dynamic concrete type
			hash      uint32                 // copy of Type.hash
			_         [4]byte                // ignored
			fun       [100000]unsafe.Pointer // method table
		}
		word unsafe.Pointer
	}

	// stringHeader is a safe version of StringHeader used within this package.
	// (COMPILER)
	stringHeader struct {
		Data unsafe.Pointer
		Len  int
	}

	// SliceHeader is the runtime representation of a slice.
	// It cannot be used safely or portable and its representation may
	// change in a later release.
	// Moreover, the Data field is not sufficient to guarantee the data
	// it references will not be garbage collected, so programs must keep
	// a separate, correctly typed pointer to the underlying data.
	SliceHeader struct {
		Data uintptr
		Len  int
		Cap  int
	}

	// sliceHeader is a safe version of SliceHeader used within this package.
	// (COMPILER)
	sliceHeader struct {
		Data unsafe.Pointer
		Len  int
		Cap  int
	}

	// Type is the representation of a Go type.
	//
	// Not all methods apply to all kinds of types. Restrictions,
	// if any, are noted in the documentation for each method.
	// Use the Kind method to find out the kind of type before
	// calling kind-specific methods. Calling a method
	// inappropriate to the kind of type causes a run-time panic.
	//
	// Type values are comparable, such as with the == operator,
	// so they can be used as map keys.
	// Two Type values are equal if they represent identical types.

	// Type is the common implementation of most values.
	// It is embedded in other, public struct types, but always with a unique tag like `reflect:"array"` or `reflect:"ptr"` so that code cannot convert from, say, *arrayType to *ptrType.
	//
	// Type must be kept in sync with ../runtime/type.go:/^type._type.
	// (COMPILER)
	RType struct {
		size          uintptr
		ptrData       uintptr   // number of bytes in the type that can contain pointers
		hash          uint32    // hash of type; avoids computation in hash tables
		extraTypeFlag extraFlag // extra type information flags
		align         uint8     // alignment of variable with this type
		fieldAlign    uint8     // alignment of struct field with this type
		kind          uint8     // enumeration for C
		alg           *algo     // algorithm table
		gcData        *byte     // garbage collection data
		str           nameOff   // string form
		ptrToThis     typeOff   // type for pointer to this type, may be zero
	}

	// The first two words of this type must be kept in sync with makeFuncImpl and runtime.reflectMethodValue.
	// Any changes should be reflected in all three.
	// (COMPILER ???)
	methodValue struct {
		fnUintPtr uintptr
		stack     *bitVector
		method    int
		rcvrVal   Value
	}

	// Value is the reflection interface to a Go value.
	//
	// Not all methods apply to all kinds of values. Restrictions,
	// if any, are noted in the documentation for each method.
	// Use the Kind method to find out the kind of value before
	// calling kind-specific methods. Calling a method
	// inappropriate to the kind of type causes a run time panic.
	//
	// The zero Value represents no value.
	// Its IsValid method returns false, its Kind method returns Invalid,
	// its String method returns "<invalid Value>", and all other methods panic.
	// Most functions and methods never return an invalid value.
	// If one does, its documentation states the conditions explicitly.
	//
	// A Value can be used concurrently by multiple goroutines provided that
	// the underlying Go value can be used concurrently for the equivalent
	// direct operations.
	//
	// To compare two Values, compare the results of the Interface method.
	// Using == on two Values does not compare the underlying values
	// they represent.
	Value struct {
		Type *RType         // Type holds the type of the value represented by a Value.
		Ptr  unsafe.Pointer // Pointer-valued data or, if pointerFlag is set, pointer to data. Valid when either pointerFlag is set or Type.pointers() is true.
		Flag
		// A method value represents a curried method invocation
		// like r.Read for some receiver r. The Type+val+flag bits describe
		// the receiver r, but the flag's Kind bits say Func (methods are
		// functions), and the top bits of the flag give the method number
		// in r's type's method table.
	}

	InspectTypeFn   func(typ *RType, name []byte, tag []byte, pack []byte, embedded, exported bool, offset uintptr, index int)
	InspectValueFn  func(typ *RType, name []byte, tag []byte, pack []byte, embedded, exported bool, offset uintptr, index int, valPtr unsafe.Pointer)
	MethodInspectFn func(name []byte, index int, flag Flag, inParams, outParams []*RType)

	BasicValue struct {
		ptr  unsafe.Pointer
		size uintptr
		flag Flag
	}

	UintValue struct {
		BasicValue
	}

	IntValue struct {
		BasicValue
	}

	FloatValue struct {
		BasicValue
	}

	ComplexValue struct {
		BasicValue
	}

	StringValue struct {
		BasicValue
		Debug string
	}

	BoolValue struct {
		BasicValue
	}

	PointerValue struct {
		BasicValue
	}

	StructValue struct {
		Value
	}

	MapValue struct {
		Value
	}

	SliceValue struct {
		Value
	}
)
type З struct{}

func (d З) З() {}

// fool linker that it has seen reflect, so it don't skip the method type filling (method types are zero otherwise)
// see cmd\compile\internal\gc\syntax.go -> funcReflectMethod
// see cmd\internal\obj\link.go -> AttrReflectMethod
// also `usemethod` in cmd\compile\internal\gc\walk.go function, which checks interface method calls for uses of reflect.Type.Method.
var _ = systemReflect.TypeOf(З{}).Method(0)

func init() {
	// info is present only for types with names or methods
	// (if T is a named type, the uncommonTypes for T and *T have methods).
	// Using a pointer to this struct reduces the overall size required
	// to describe an unnamed type with no methods.
	uint8Type = TypeOf(uint8(0))
}
