package reflector

import (
	"bytes"
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	. "reflect"
)

//TODO : @badu - benchmark inlining version
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

func (field *Field) ptrToLoad() Value {
	return New(PtrTo(field.Value.Type()))
}

func (field *Field) IndirectValue(value interface{}) Value {
	result := ValueOf(value)
	for result.Kind() == Ptr {
		result = result.Elem()
	}
	return result
}

func (field *Field) makeSlice() (interface{}, Value) {
	basicType := field.Type
	if field.IsPointer() {
		basicType = PtrTo(field.Type)
	}
	sliceType := SliceOf(basicType)
	slice := New(sliceType)
	slice.Elem().Set(MakeSlice(sliceType, 0, 0))
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

	reflectValue, ok := value.(Value)
	if !ok {
		//couldn't cast - reflecting it
		reflectValue = ValueOf(value)
	}

	if reflectValue.IsValid() {
		if reflectValue.Type().ConvertibleTo(fieldValue.Type()) {
			//we set it
			fieldValue.Set(reflectValue.Convert(fieldValue.Type()))
		} else {
			//we're working with a pointer?
			if fieldValue.Kind() == Ptr {
				//it's a pointer
				if fieldValue.IsNil() {
					//and it's NIL : we have to build it
					fieldValue.Set(New(field.Type))
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

// Parse parses a single struct field tag and returns the set of tags.
func (field Field) ParseStructTag(tag string) (*Tags, error) {
	var tags []*Tag

	for tag != "" {
		// Skip leading space.
		i := 0
		for i < len(tag) && tag[i] == ' ' {
			i++
		}
		tag = tag[i:]
		if tag == "" {
			return nil, nil
		}

		// Scan to colon. A space, a quote or a control character is a syntax
		// error. Strictly speaking, control chars include the range [0x7f,
		// 0x9f], not just [0x00, 0x1f], but in practice, we ignore the
		// multi-byte control characters as it is simpler to inspect the tag's
		// bytes than the tag's runes.
		i = 0
		for i < len(tag) && tag[i] > ' ' && tag[i] != ':' && tag[i] != '"' && tag[i] != 0x7f {
			i++
		}

		if i == 0 {
			return nil, errTagKeySyntax
		}
		if i+1 >= len(tag) || tag[i] != ':' {
			return nil, errTagSyntax
		}
		if tag[i+1] != '"' {
			return nil, errTagValueSyntax
		}

		key := string(tag[:i])
		tag = tag[i+1:]

		// Scan quoted string to find value.
		i = 1
		for i < len(tag) && tag[i] != '"' {
			if tag[i] == '\\' {
				i++
			}
			i++
		}
		if i >= len(tag) {
			return nil, errTagValueSyntax
		}

		qvalue := string(tag[:i+1])
		tag = tag[i+1:]

		value, err := strconv.Unquote(qvalue)
		if err != nil {
			return nil, errTagValueSyntax
		}

		res := strings.Split(value, ",")
		name := res[0]
		options := res[1:]
		if len(options) == 0 {
			options = nil
		}

		tags = append(tags, &Tag{
			Key:     key,
			Name:    name,
			Options: options,
		})
	}

	return &Tags{
		tags: tags,
	}, nil
}

// Stringer implementation
func (field Field) String() string {
	var result bytes.Buffer
	tabs := strings.Repeat("\t", field.printNesting)
	result.WriteString(tabs + field.Name + "\t")

	if field.printNesting > 10 {
		result.WriteString(tabs + "Nesting exceeded 20 levels.\t")
		return result.String()
	}

	if field.flags&(1<<ff_is_slice) != 0 {
		if field.flags&(1<<ff_is_pointer) != 0 {
			result.WriteString("[]*" + field.Relation.ModelType.Name())
		} else {
			result.WriteString("[]" + field.Relation.ModelType.Name())
		}

	} else {
		if field.flags&(1<<ff_is_pointer) != 0 {
			result.WriteString("*")
		}
	}

	if field.flags&(1<<ff_is_anonymous) != 0 {
		result.WriteString("(embedded) ")
	}

	if field.flags&(1<<ff_is_map) != 0 {
		result.WriteString("Map ")
	}

	if field.flags&(1<<ff_is_interface) != 0 {
		result.WriteString("Interface ")
	}

	if field.flags&(1<<ff_is_time) != 0 {
		result.WriteString("time.Time\n")
	}

	if field.flags&(1<<ff_is_struct) == 0 {
		// not a struct
		result.WriteString(tabs + "\t " + field.Type.Name() + "\n")
		for _, t := range field.Tags.Tags() {
			result.WriteString(tabs + "\t\ttag: " + t.String() + "\n")
		}

	} else {
		//it's a struct, but not time
		if field.flags&(1<<ff_is_time) == 0 {
			result.WriteString(field.Relation.ModelType.Name() + "\n")
		}
	}

	// if it's a relation, but not Time and it's not self reference
	if field.flags&(1<<ff_is_relation) != 0 && field.flags&(1<<ff_is_time) == 0 && field.flags&(1<<ff_is_self_reference) == 0 {
		result.WriteString(field.Relation.String())
	}

	return result.String()
}
