/*
 * Copyright 2009-2018 The Go Authors. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package old

import (
	"fmt"
	. "reflect"
	"time"
)

type (
	CompareResult struct {
		FieldName string
		OldValue  interface{}
		NewValue  interface{}
		IsSlice   bool
	}
)

func MakeMapFromDifferences(differences []*CompareResult) (map[string]interface{}, map[string]interface{}) {
	updatesMap, changesMap := make(map[string]interface{}), make(map[string]interface{})
	for _, difference := range differences {
		if !difference.IsSlice {
			changesMap["From"+difference.FieldName] = difference.OldValue
			changesMap["To"+difference.FieldName] = difference.NewValue
			updatesMap[difference.FieldName] = difference.NewValue
		}
	}
	return updatesMap, changesMap
}

func Compare(oldStruct interface{}, newStruct interface{}, onlyFields []string) ([]*CompareResult, error) {
	if oldStruct == nil || newStruct == nil {
		return nil, fmt.Errorf("Comparator : One of the inputs cannot be nil. struct1: %v, struct2 : %v ", oldStruct, newStruct)
	}

	//Get values of the structs
	oldValue, newValue := ValueOf(oldStruct), ValueOf(newStruct)

	//Handle pointers, if a non-pointer struct is passed in, Indirect will just return the element
	oldValue, newValue = Indirect(oldValue), Indirect(newValue)
	if !oldValue.IsValid() || !newValue.IsValid() {
		return nil, fmt.Errorf("Comparator : Types cannot be nil. SRC (%v) invalid = %v - DEST (%v) invalid %v", oldStruct, oldValue.IsValid(), newStruct, newValue.IsValid())
	}

	//Cache v1 struct type
	oldType := oldValue.Type()
	newType := newValue.Type()
	//Verify both v1 and v2 are the same type
	if oldType != newType {
		return nil, fmt.Errorf("Comparator : Structs must be the same type. Struct1 %v - Stuct2 -%v", oldType, newType)
	}

	//Verify v1 is a struct, if v1 is a struct then v2 is also a struct because we have already verified the types are equal
	if oldValue.Kind() != Struct || newValue.Kind() != Struct {
		return nil, fmt.Errorf("Comparator : Types must both be structs.  Kind1: %v, Kind2 :%v", oldValue.Kind(), newValue.Kind())
	}

	//Initialize differences to ensure length of 0 on return
	differences := make([]*CompareResult, 0)

	for i, numFields := 0, oldValue.NumField(); i < numFields; i++ {

		//Get values of the structure's fields
		oldField, newField := oldValue.Field(i), newValue.Field(i)

		//Get a reference to the field type
		fieldType := oldType.Field(i)
		fieldKind := fieldType.Type.Kind()
		currentFieldName := fieldType.Name

		if fieldType.PkgPath != "" {
			//If the field name is unexported, skip
			continue
		}

		//Handle nil pointers, if a non-pointer field is passed in, Indirect will just return the element
		oldField, newField = Indirect(oldField), Indirect(newField)

		switch valid1, valid2 := oldField.IsValid(), newField.IsValid(); {
		case valid1 && valid2:
		//If both are valid, do nothing
		case valid1:
			//If only field1 is valid, set field2 to Zero
			newField = Zero(oldField.Type())
		case valid2:
			//If only field1 is valid, set field2 to Zero
			oldField = Zero(newField.Type())
		default:
			//Both are invalid so skip loop body
			continue
		}

		if oldField.Kind() == Interface {
			return nil, fmt.Errorf("Update Compare : Type of field cannot be interface. field1: %v, field2: %v", oldField, newField)
		}

		var oldSubField, newSubField Value
		var subStructType Type

		// if we have embedded struct, we might need to check for the property with the same name inside
		isStruct := fieldKind == Struct
		if isStruct {
			oldSubField, newSubField = ValueOf(oldField.Interface()), ValueOf(newField.Interface())
			subStructType = oldSubField.Type()
		}

		currentSubfieldName := ""
		willAdd := false
		if len(onlyFields) > 0 {
			for _, fn := range onlyFields {
				if fn == currentFieldName {
					willAdd = true
					break
				}
			}
			// field was not found, and we have an embedded struct
			if !willAdd && isStruct {
				for j, subNumFields := 0, oldSubField.NumField(); j < subNumFields; j++ {
					subField := subStructType.Field(j)
					// searching for field in embedded structs
					for _, fn := range onlyFields {
						if fn == subField.Name {
							willAdd = true
							// store current subfield name to perform logic below
							currentSubfieldName = subField.Name
							break
						}
					}
				}
			}
		} else {
			// adding all fields - onlyFields is empty
			willAdd = true
		}

		// skipping non checked fields
		if !willAdd {
			continue
		}

		// convention - comparing slices by Id field
		if fieldKind == Slice {
			// what's in old and no more in new (was deleted)
			for index1 := 0; index1 < oldField.Len(); index1++ {
				oldElem := oldField.Index(index1)
				if oldElem.Kind() == Ptr {
					oldElem = oldElem.Elem()
				}
				foundInOld := false
				oldElemId := oldElem.FieldByName("Id").Interface()
				for index2 := 0; index2 < newField.Len(); index2++ {
					newElem := newField.Index(index2)
					if newElem.Kind() == Ptr {
						newElem = newElem.Elem()
					}
					if newElem.FieldByName("Id").Interface() == oldElemId {
						// item "might" be updated (check outside)
						result := &CompareResult{FieldName: currentFieldName, OldValue: oldElem.Interface(), NewValue: newElem.Interface(), IsSlice: true}
						differences = append(differences, result)
						foundInOld = true
						break
					}

				}
				if !foundInOld {
					// item was removed from slice
					result := &CompareResult{FieldName: currentFieldName, OldValue: oldElem.Interface(), NewValue: nil, IsSlice: true}
					differences = append(differences, result)
				}
			}
			// what's in new and not in old (was added)
			for index2 := 0; index2 < newField.Len(); index2++ {
				newElem := newField.Index(index2)
				if newElem.Kind() == Ptr {
					newElem = newElem.Elem()
				}
				foundInNew := false
				newElemId := newElem.FieldByName("Id").Interface()
				for index1 := 0; index1 < oldField.Len(); index1++ {
					oldElem := oldField.Index(index1)
					if oldElem.Kind() == Ptr {
						oldElem = oldElem.Elem()
					}
					if newElemId == oldElem.FieldByName("Id").Interface() {
						foundInNew = true
						break
					}
				}
				if !foundInNew {
					// item was added to slice
					result := &CompareResult{FieldName: currentFieldName, OldValue: nil, NewValue: newElem.Interface(), IsSlice: true}
					differences = append(differences, result)
				}

			}
			continue

		}

		oldFieldValue, newFieldValue := oldField.Interface(), newField.Interface()
		if isStruct {
			// if it's not an embedded struct
			if currentSubfieldName == "" {
				// substructs like nullstring, nulluint64, etc
				hasDifference := false
				// we're storing the values here, in case we find differences
				var storedOldValue, storedNewValue interface{}
				// the value might get nilled
				switchedToNil := false
				for j, subNumFields := 0, oldSubField.NumField(); j < subNumFields; j++ {
					subField := subStructType.Field(j)

					if subField.PkgPath != "" {
						//If the field name is unexported, skip
						continue
					}
					oldSubStructField, newSubStructField := oldSubField.Field(j), newSubField.Field(j)
					oldSubFieldValue, newSubFieldValue := oldSubStructField.Interface(), newSubStructField.Interface()

					switch oldSubFieldValue.(type) {

					case bool, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, string, time.Time:
						// boilerplate - Valid is always a bool
						if subField.Name == "Valid" {
							// if it was valid and now it's not, value will be nil
							if oldSubFieldValue.(bool) && !newSubFieldValue.(bool) {
								switchedToNil = true
							}
						} else if oldSubFieldValue != newSubFieldValue {
							storedOldValue = oldSubFieldValue
							storedNewValue = newSubFieldValue
							hasDifference = true
						}
					}
				}
				if hasDifference {
					result := &CompareResult{FieldName: currentFieldName, OldValue: storedOldValue, NewValue: storedNewValue}
					if switchedToNil {
						// was set to nil, we're doing it here
						result.NewValue = nil
					}
					differences = append(differences, result)
				}
			} else {
				// for embedded structs
				for j, subNumFields := 0, oldSubField.NumField(); j < subNumFields; j++ {
					subField := subStructType.Field(j)
					if subField.PkgPath != "" {
						//If the field name is unexported, skip
						continue
					}
					oldSubStructField, newSubStructField := oldSubField.Field(j), newSubField.Field(j)
					oldSubFieldValue, newSubFieldValue := oldSubStructField.Interface(), newSubStructField.Interface()

					switch oldSubFieldValue.(type) {
					case bool, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, string, time.Time:
						if oldSubFieldValue != newSubFieldValue {
							result := &CompareResult{FieldName: subField.Name, OldValue: oldSubFieldValue, NewValue: newSubFieldValue}
							differences = append(differences, result)
						}
					}
				}
			}
		} else {
			// it's a normal field or an alias for it
			switch oldFieldValue.(type) {
			case bool, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, string, time.Time:
				if oldFieldValue != newFieldValue {
					result := &CompareResult{FieldName: currentFieldName, OldValue: oldFieldValue, NewValue: newFieldValue}
					differences = append(differences, result)
				}
			default:
				if oldFieldValue != newFieldValue {
					result := &CompareResult{FieldName: currentFieldName, OldValue: oldFieldValue, NewValue: newFieldValue}
					differences = append(differences, result)
				}

			}
		}
	}

	return differences, nil
}
