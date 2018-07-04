/*
 * Copyright 2009-2018 The Go Authors. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package reflect

import (
	"errors"
	"testing"
	"unsafe"
)

func TestProduceErrorInsideMakeFunc(t *testing.T) {
	type User struct {
		Id   uint64
		Name string
	}
	type UsersCollbection []User
	type GetByUsername func(username string) (UsersCollbection, error)
	type UserRepositoryImplementation struct {
		//UserRepository
		//Extender
		GetByUsername GetByUsername
		GetPaged      func(page int) ([10]User, error)
	}
	ur := &UserRepositoryImplementation{}

	repositoryValue := ReflectOnPtr(ur)
	repositoryStruct := repositoryValue.Type.convToStruct()

	for idx := range repositoryStruct.fields {
		repositoryField := &repositoryStruct.fields[idx]

		funcType := repositoryField.Type.convToFn()
		outputs := funcType.outParams()

		newFn := MakeFunc(repositoryField.Type, func(in []Value) []Value {
			println("Yes, we've replaced " + string(repositoryField.name.name()) + " function.")
			var result []Value
			for idx, output := range outputs {
				if idx == 1 {
					// TODO : pattern, if you want to return an error
					/**
					Because ReflectOnPtr takes an interface argument, it effectively always uses the dynamic type.
					You have to work a bit to get an interface type into a Value, like this:
					*/
					testError := errors.New("Test error")
					result = append(result, ReflectOnPtr(&testError))
					/**
					We're using ReflectOnPtr so we don't need to call Deref on the Value.
					*/
				} else {
					result = append(result, internalNew(output))
				}
			}
			return result
		})

		destFieldPtr := add(repositoryValue.Ptr, structFieldOffset(repositoryField))
		*(*unsafe.Pointer)(destFieldPtr) = newFn.Ptr
	}

	_, err := ur.GetPaged(3)
	if err == nil {
		t.Fail()
	}
	//we're expecting an error, since this test is for producing an error inside MakeFunc
	t.Logf("Test error ? -> %v", err)
}
