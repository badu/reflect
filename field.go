package reflector

import (
	"bytes"
	"database/sql"
	"fmt"
	"reflect"
	"strings"
)
// should remain unused
func (field *Field) setFlag(value uint8) {
	field.flags = field.flags | (1 << value)
}
// should remain unused
func (field *Field) unsetFlag(value uint8) {
	field.flags = field.flags & ^(1 << value)
}

func (field *Field) IsAnonymous() bool {
	return field.flags&(1<<ff_is_anonymous) != 0
}

func (field *Field) IsTime() bool {
	return field.flags&(1<<ff_is_time) != 0
}

func (field *Field) IsSlice() bool {
	return field.flags&(1<<ff_is_slice) != 0
}

func (field *Field) IsStruct() bool {
	return field.flags&(1<<ff_is_struct) != 0
}

func (field *Field) IsMap() bool {
	return field.flags&(1<<ff_is_map) != 0
}

func (field *Field) IsPointer() bool {
	return field.flags&(1<<ff_is_pointer) != 0
}

func (field *Field) HasRelation() bool {
	return field.flags&(1<<ff_is_relation) != 0
}

func (field *Field) IsInterface() bool {
	return field.flags&(1<<ff_is_interface) != 0
}

func (field *Field) ptrToLoad() reflect.Value {
	return reflect.New(reflect.PtrTo(field.Value.Type()))
}

func (field *Field) IndirectValue(value interface{}) reflect.Value {
	result := reflect.ValueOf(value)
	for result.Kind() == reflect.Ptr {
		result = result.Elem()
	}
	return result
}

func (field *Field) makeSlice() (interface{}, reflect.Value) {
	basicType := field.Type
	if field.IsPointer() {
		basicType = reflect.PtrTo(field.Type)
	}
	sliceType := reflect.SliceOf(basicType)
	slice := reflect.New(sliceType)
	slice.Elem().Set(reflect.MakeSlice(sliceType, 0, 0))
	return slice.Interface(), field.IndirectValue(slice.Interface())
}

func (field *Field) Set(value interface{}) error {
	var (
		err        error
		fieldValue = field.Value
	)

	if !fieldValue.IsValid() {
		//TODO : @Badu - make errors more explicit : which field...
		return fmt.Errorf("Field not valid")
	}

	if !fieldValue.CanAddr() {
		return fmt.Errorf("Field not addressable")
	}

	reflectValue, ok := value.(reflect.Value)
	if !ok {
		//couldn't cast - reflecting it
		reflectValue = reflect.ValueOf(value)
	}

	if reflectValue.IsValid() {
		if reflectValue.Type().ConvertibleTo(fieldValue.Type()) {
			//we set it
			fieldValue.Set(reflectValue.Convert(fieldValue.Type()))
		} else {
			//we're working with a pointer?
			if fieldValue.Kind() == reflect.Ptr {
				//it's a pointer
				if fieldValue.IsNil() {
					//and it's NIL : we have to build it
					fieldValue.Set(reflect.New(field.Type))
				}
				//we dereference it
				fieldValue = fieldValue.Elem()
			}

			scanner, isScanner := fieldValue.Addr().Interface().(sql.Scanner)
			if isScanner {
				//implements Scanner - we pass it over
				err = scanner.Scan(reflectValue.Interface())

			} else if reflectValue.Type().ConvertibleTo(fieldValue.Type()) {
				//last attempt to set it
				fieldValue.Set(reflectValue.Convert(fieldValue.Type()))
			} else {
				//Oops
				//TODO : @Badu - make errors more explicit
				err = fmt.Errorf("Cannot convert %q %v %v", field.Name, reflectValue.Type(), fieldValue.Type())
			}
		}
		//then we check if the value is blank
		//field.checkIsBlank()
	} else {
		//set is blank
		//field.setFlag(ff_is_blank)
		//it's not valid : set empty
		//field.setZeroValue()
	}

	return err
}

// Stringer implementation
func (field Field) String() string {
	var result bytes.Buffer
	tabs := strings.Repeat("\t", field.printNesting)
	result.WriteString(tabs + field.Name + "\t")

	if field.flags&(1<<ff_is_slice) != 0 {
		result.WriteString("[]")
	}
	if field.flags&(1<<ff_is_anonymous) != 0 {
		result.WriteString("(embedded) ")
	}
	if field.flags&(1<<ff_is_map) != 0 {
		result.WriteString("Map ")
	}
	if field.flags&(1<<ff_is_pointer) != 0 {
		result.WriteString("*")
	}
	if field.flags&(1<<ff_is_interface) != 0 {
		result.WriteString("Interface ")
	}
	if field.flags&(1<<ff_is_struct) == 0 {
		// not a struct
		result.WriteString(tabs + "\t having = " + field.Type.Name() + "\n")
	}
	// if it's a relation, but not Time
	if field.flags&(1<<ff_is_relation) != 0 && field.flags&(1<<ff_is_time) == 0 {
		//field.Relation.printNesting = field.printNesting + 1
		//result.WriteString(tabs + "Relation :\n")
		result.WriteString(field.Relation.String())
	}
	if field.flags&(1<<ff_is_time) != 0 {
		result.WriteString("time.Time\n")
	}
	/**
	if field.Value.IsValid() {
		result.WriteString(tabs + fmt.Sprintf("Field Value : `%v", field.Value) + "`\n")
	} else {
		result.WriteString(tabs + fmt.Sprintf("Invalid Field Value\n"))
	}
	**/
	return result.String()
}
