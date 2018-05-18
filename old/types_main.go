/*
 * Copyright 2009-2018 The Go Authors. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package old

import (
	"errors"
	"sync"

	. "reflect"
)

type (
	InspectType uint

	// Field's Tag
	Tag struct {
		Key     string
		Name    string
		Options []string
	}

	// Tags represent a set of tags from a single struct field
	Tags struct {
		tags []*Tag
	}

	// Field - flags are not uint16 because of padding
	Field struct {
		flags        uint64 // flags that indicates pointer, struct, slice, etc - see below
		Name         string
		Type         Type
		Value        Value
		Tags         *Tags
		Relation     *Model // Relation represents a relation between a Field and a Model
		printNesting int    // internal for printing tabs
	}

	// Model
	Model struct {
		ModelType    Type
		Value        Value // value is needed to pass over non-initialized structs
		Fields       []*Field
		Methods      map[string]bool // has the method from method lookup
		printNesting int             // internal, for printing tabs
		printTabs    string          // internal, for printing tabs
		visited      bool            // internal, for keeping track of visited models
	}

	// Reflector
	Reflector struct {
		currentModel  *Model   // keeps track of current visiting model
		MethodsLookup []string // white list of methods that struct have declared
	}

	// a safe map of models that Reflector keeps as cached
	safeModelsMap struct {
		m map[Type]*Model
		l *sync.RWMutex
	}

	Function struct {
		Caller caller
	}
	caller interface {
		Call(args ...interface{}) error
	}
	callback func(args ...interface{}) error
)

const (
	// flags
	ff_is_anonymous      uint8 = 0
	ff_is_time           uint8 = 1
	ff_is_slice          uint8 = 2
	ff_is_struct         uint8 = 3
	ff_is_map            uint8 = 4
	ff_is_pointer        uint8 = 5
	ff_is_relation       uint8 = 6
	ff_is_interface      uint8 = 7
	ff_is_self_reference uint8 = 8

	// relations kinds
	ff_rk_one_to_many  uint8 = 9
	ff_rk_belongs_to   uint8 = 10
	ff_rk_many_to_many uint8 = 11

	// reserved (unused) flags
	ff_reserved_1 uint8 = 12
	ff_reserved_2 uint8 = 13
	ff_reserved_3 uint8 = 14
	ff_reserved_4 uint8 = 15
)

var (
	printDebug bool = false

	// keeps known models (already visited)
	cachedModels *safeModelsMap

	errTagSyntax      = errors.New("bad syntax for struct tag pair")
	errTagKeySyntax   = errors.New("bad syntax for struct tag key")
	errTagValueSyntax = errors.New("bad syntax for struct tag value")
	errTagNotExist    = errors.New("tag does not exist")
)
