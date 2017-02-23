package reflector

import (
	"fmt"
	"reflect"
	"strings"
	"time"
)

func (m *Reflector) ComponentsScan(components ...interface{}) error {
	for _, component := range components {
		value := reflect.ValueOf(component)
		valueType := value.Type()

		if value.Kind() != reflect.Struct {
			return fmt.Errorf("We don't visit %v here!", value.Kind())
		}

		// set current model
		m.currentModel = &Model{ModelType: valueType, Name: valueType.Name(), printNesting: 0}

		m.visit(value)
	}
	return nil
}

func (m *Reflector) inspectMap(value reflect.Value) error {

	if err := m.InspectMap(value); err != nil {
		return err
	}

	for _, key := range value.MapKeys() {
		keyValue := value.MapIndex(key)

		if err := m.inspectMapKeyValue(value, key, keyValue); err != nil {
			return err
		}

		if err := m.inspect(key); err != nil {
			return err
		}

		if err := m.inspect(keyValue); err != nil {
			return err
		}

	}

	return nil
}

func (m *Reflector) inspectSlice(value reflect.Value) error {

	if err := m.InspectSlice(value); err != nil {
		return err
	}

	for i := 0; i < value.Len(); i++ {
		elem := value.Index(i)

		if err := m.inspectSliceElem(i, elem); err != nil {
			return err
		}

		if err := m.inspect(elem); err != nil {
			return err
		}

	}

	return nil
}

func (m *Reflector) inspectArray(value reflect.Value) error {

	if err := m.InspectArray(value); err != nil {
		return err
	}

	for i := 0; i < value.Len(); i++ {
		elem := value.Index(i)

		if err := m.InspectArrayElem(i, elem); err != nil {
			return err
		}

		if err := m.inspect(elem); err != nil {
			return err
		}

	}

	return nil
}

func (m *Reflector) InspectMap(theMap reflect.Value) error {
	//fmt.Printf("Map %v\n", theMap)
	return nil
}

func (m *Reflector) inspectMapKeyValue(theMap, key, value reflect.Value) error {
	//fmt.Printf("MapElem %v %v\n", key, value)
	return nil
}

func (m *Reflector) InspectSlice(value reflect.Value) error {
	if value.Len() == 0 {
		var pointedElement reflect.Type
		var pointedStruct reflect.Value
		if m.currentField.Type.Elem().Kind() == reflect.Ptr {
			pointedElement = m.currentField.Type.Elem().Elem()
		} else {
			pointedElement = m.currentField.Type.Elem()
		}
		pointedStruct = reflect.New(pointedElement).Elem()

		if printDebug {
			fmt.Printf("%s[Slice] field name %q of a struct %q\n", strings.Repeat("\t", m.currentModel.printNesting), m.currentField.Name, pointedStruct.Type().Name())
		}

		//newReflector := &Reflector{}
		// set current model
		//newReflector.currentModel =
		// before visiting, marking it as visiting, so circular references are avoided
		//visitingModels.set(pointedStruct.Type(), newReflector.currentModel)
		//newReflector.visit(pointedStruct)
		// set the relationship
		m.currentField.Relation = &Model{
			ModelType:    pointedStruct.Type(),
			Name:         pointedStruct.Type().Name(),
			Value:        pointedStruct,
			printNesting: m.currentModel.printNesting + 1}
		// set the flag
		m.currentField.flags = m.currentField.flags | (1 << ff_is_relation)

	}

	return nil

}
func (m *Reflector) inspectSliceElem(index int, value reflect.Value) error {
	fmt.Printf("SliceElem : #%d %v\n", index, value)
	return nil
}

func (m *Reflector) InspectArray(value reflect.Value) error {
	//fmt.Printf("Array %v\n", value)
	return nil
}

func (m *Reflector) InspectArrayElem(index int, value reflect.Value) error {
	//fmt.Printf("ArrayElem: #%d %v\n", index, value)
	return nil
}

func (m *Reflector) inspect(value reflect.Value) error {
	// Determine if we're receiving a pointer and if so notify the
	// The logic here is convoluted but very important (tests will fail if
	// almost any part is changed).
	//
	// First, we check if the value is an interface, if so, we really need
	// to check the interface's value to see whether it is a pointer.
	//
	// Check whether the value is then a pointer. If so, then set pointer
	// to true to notify the
	//
	// If we still have a pointer or an interface after the indirections, then
	// we unwrap another level
	//
	// At this time, we also set "value" to be the de-referenced value. This is
	// because once we've unwrapped the pointer we want to use that value.
	var err error
	valuePtr := value
	isPointer := false
	tabs := strings.Repeat("\t", m.currentModel.printNesting)

	for {
		switch valuePtr.Kind() {
		case reflect.Interface:
			valuePtr = valuePtr.Elem()
			// fallthrough, since it can be an interface
			fallthrough
		case reflect.Ptr:
			value = reflect.Indirect(valuePtr)
			valuePtr = value
			isPointer = true
		}

		// If we still have a pointer or interface we have to indirect another level.
		switch valuePtr.Kind() {
		case reflect.Ptr, reflect.Interface:
			continue
		}
		break
	}

	if value.Kind() == reflect.Interface {
		value = value.Elem()
	}

	kind := value.Kind()
	if kind >= reflect.Int && kind <= reflect.Complex128 {
		kind = reflect.Int
	}

	if printDebug {
		fmt.Printf("%s[level #%d] Current model type : %q\n", tabs, m.currentModel.printNesting, m.currentModel.ModelType)
	}

	switch kind {
	case reflect.Bool, reflect.Chan, reflect.Func, reflect.Int, reflect.String, reflect.Invalid:
		// Primitives
		return nil
	case reflect.Map:
		if printDebug {
			fmt.Printf("yes, it's a map : %v\n", value)
		}
		err = m.inspectMap(value)
		return err
	case reflect.Slice:
		if printDebug {
			err = m.inspectSlice(value)
		}
		return err
	case reflect.Array:
		if printDebug {
			fmt.Printf("yes, it's an array : %v\n", value)
		}
		err = m.inspectArray(value)
		return err
	case reflect.Struct:
		if isPointer {
			valueType := value.Type()
			if m.currentField != nil {
				m.currentField.Relation = &Model{
					ModelType:    valueType,
					Name:         valueType.Name(),
					Value:        value,
					printNesting: m.currentModel.printNesting + 1}
				m.currentField.flags = m.currentField.flags | (1 << ff_is_relation)
				if printDebug {
					fmt.Printf("%s[POINTER STRUCT] %q -> %q\n", tabs, m.currentField.Name, m.currentField.Relation.ModelType)
				}
			} else {
				if printDebug {
					fmt.Println("%s[POINTER STRUCT] Don't know how to visit %q\n", tabs, m.currentModel.ModelType)
				}
				/**
				if printDebug {
					fmt.Printf("%s[level #%d] Visiting NO CURRENT FIELD -> %q over %q\n", tabs, m.currentModel.printNesting, m.currentModel.ModelType, newReflector.currentModel.ModelType)
				}
				**/
			}

			// was a pointer, transfer model to the parent
			//if m.currentField == nil {
			//	m.currentModel = newReflector.currentModel
			//}

		} else {
			if m.currentField != nil {
				if printDebug {
					fmt.Printf("%s[STRUCT] CURRENT FIELD %q (%q)= `%v`\n", tabs, m.currentField.Name, m.currentField.Type, value)
				}
			}
			err = m.inspectStruct(value)
		}
		return err
	default:
		return fmt.Errorf("Inspector : unsupported type %q ", kind.String())
	}
	return err
}

func (m *Reflector) inspectStruct(value reflect.Value) error {

	valueType := value.Type()
	tabs := strings.Repeat("\t", m.currentModel.printNesting)
	for i := 0; i < valueType.NumField(); i++ {
		structField := valueType.Field(i)
		field := value.FieldByIndex([]int{i})
		fieldType := field.Type()
		// set current field as a new field
		m.currentField = &Field{
			Type: structField.Type,
			Name: structField.Name,
		}

		fmt.Printf("%s[Field] Current field %q : %q\n", tabs, structField.Name, structField.Type)

		if structField.Anonymous {

			m.currentField.flags = m.currentField.flags | (1 << ff_is_anonymous)

			if printDebug {
				fmt.Printf("%s[ANON] %d fields - %q (%q:%q)\n", tabs, field.NumField(), fieldType.Name(), structField.Name, structField.Type)
			}
			for j := 0; j < field.NumField(); j++ {
				subStructField := fieldType.Field(j)
				subField := field.FieldByIndex([]int{j})
				if printDebug {
					fmt.Printf("%s[ANON] %q %q = %v\n", tabs, subStructField.Name, subStructField.Type, subField)
				}
				switch subStructField.Type.Kind() {
				case reflect.Ptr:
					if printDebug {
						fmt.Println("POINTER")
					}
				case reflect.Struct:
					if printDebug {
						fmt.Println("STRUCT")
					}
					// set the relationship
					m.currentField.Relation = &Model{
						ModelType:    fieldType,
						Name:         fieldType.Name(),
						Value:        subField,
						printNesting: m.currentModel.printNesting + 1}
					// set the flag
					m.currentField.flags = m.currentField.flags | (1 << ff_is_relation)
				case reflect.Slice:
					if printDebug {
						fmt.Println("SLICE")
					}
				default:
					// field has values : visit it
					err := m.inspect(subField)
					if err != nil {
						return err
					}
				}

				// add the field to current model
				m.currentModel.Fields = append(m.currentModel.Fields, m.currentField)
			}

			if printDebug {
				fmt.Printf("%s[ANON] Finished.\n", tabs)
			}
			continue
		}

		if field.Kind() == reflect.Invalid {
			if printDebug {
				fmt.Printf("%sINVALID FIELD : %q = %q\n", tabs, structField.Name, structField.Type)
			}
			continue
		}

		switch structField.Type.Kind() {
		case reflect.Ptr:
			// set flag to pointer
			m.currentField.flags = m.currentField.flags | (1 << ff_is_pointer)

			fmt.Printf("%s[PTR] %q (valid = %t , nil = %t) `%v`\n", tabs, structField.Name, field.IsValid(), field.IsNil(), field)
			// fallthrough, since it can be a pointer to a struct, slice, whatever
			fallthrough
		case reflect.Struct:
			// set flag to struct
			m.currentField.flags = m.currentField.flags | (1 << ff_is_struct)
			if printDebug {
				fmt.Printf("%sSTRUCT %q : %q = `%v`\n", tabs, m.currentField.Name, m.currentModel.ModelType, field)
			}
			// if it's not a pointer
			if m.currentField.flags&(1<<ff_is_pointer) == 0 {
				// TODO : add Scanner, Valuer, Marshaller, Unmarshaller
				_, isTime := field.Interface().(time.Time)
				if isTime {
					// set flag to time
					m.currentField.flags = m.currentField.flags | (1 << ff_is_time)
				}

				// add the field to current model
				m.currentModel.Fields = append(m.currentModel.Fields, m.currentField)

				if printDebug {
					fmt.Printf("%s[ADD #1] %q (%q) = `%v` to %q\n", tabs, m.currentField.Name, m.currentField.Type, m.currentField.Value, m.currentModel.ModelType)
				}

				return nil

				if printDebug {
					fmt.Printf("%sINSPECT STRUCT %q : %q\n", tabs, m.currentField.Name, m.currentModel.ModelType)
				}
				// field has values : visit it
				err := m.inspect(field)
				if err != nil {
					return err
				}

			} else {
				if field.IsNil() {
					if printDebug {
						fmt.Printf("%sNIL STRUCT %q : %q\n", tabs, m.currentField.Name, m.currentModel.ModelType)
					}
				}
				// field is nil, we build one to visit it
				// it's a pointer. dereferencing it

				pointedElement := m.currentField.Type.Elem()
				pointedStruct := reflect.New(pointedElement).Elem()

				valueType := pointedStruct.Type()

				// add the field to current model
				m.currentModel.Fields = append(m.currentModel.Fields, m.currentField)

				if printDebug {
					fmt.Printf("%s[ADD #2] %q (%q) = `%v` to %q\n", tabs, m.currentField.Name, m.currentField.Type, m.currentField.Value, m.currentModel.ModelType)
				}

				return nil

				if printDebug {
					fmt.Printf("%s[Pointer] %q to a struct %q\n", tabs, m.currentField.Name, valueType.Name())
				}
				// set the relationship
				m.currentField.Relation = &Model{
					ModelType:    valueType,
					Name:         valueType.Name(),
					Value:        pointedStruct,
					printNesting: m.currentModel.printNesting + 1}
				// set the flag
				m.currentField.flags = m.currentField.flags | (1 << ff_is_relation)

			}
			//fmt.Printf("%s - %t, %t\n", tabs, field.IsNil(), field.IsValid())

		case reflect.Slice:
			// set the flag
			m.currentField.flags = m.currentField.flags | (1 << ff_is_relation)
			// set flag to slice
			m.currentField.flags = m.currentField.flags | (1 << ff_is_slice)
			// inspect it
			err := m.inspect(field)
			if err != nil {
				return err
			}
		default:
			m.currentField.Value = field
			// by default, we inspect
			err := m.inspect(field)
			if err != nil {
				return err
			}
		}
		if !structField.Anonymous {
			// add the field to current model
			m.currentModel.Fields = append(m.currentModel.Fields, m.currentField)

			if printDebug {
				fmt.Printf("%s[ADD #0] %q (%q) = `%v` to %q\n", tabs, m.currentField.Name, m.currentField.Type, m.currentField.Value, m.currentModel.ModelType)
			}
		}
	}

	return nil
}

func (m *Reflector) visit(value reflect.Value) error {
	valueType := value.Type()

	// Get cached Model
	if cachedValue := cachedModels.get(valueType); cachedValue != nil {
		return nil
	}

	// inspect the value
	err := m.inspect(value)

	switch value.Kind() {
	case
		reflect.Invalid,
		reflect.Bool,
		reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64,
		reflect.Uint,
		reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32,
		reflect.Uint64,
		reflect.Uintptr,
		reflect.Float32,
		reflect.Float64,
		reflect.Complex64,
		reflect.Complex128,
		reflect.Array,
		reflect.Chan,
		reflect.Func,
		reflect.Interface,
		reflect.Map,
		reflect.Ptr,
		reflect.Slice,
		reflect.String:
		// do nothing, it's primitive
	default:
		m.currentModel.visited = true
		// Set cached model
		cachedModels.set(valueType, m.currentModel)

	}

	for _, field := range m.currentModel.Fields {
		if field.HasRelation() && !field.Relation.visited {
			if printDebug{
				fmt.Printf("NOT VISITED : %q %q\n", field.Name, field.Type)
			}
			newReflector := &Reflector{}
			newReflector.currentModel = &Model{ModelType: field.Type, Name: field.Name, printNesting: field.Relation.printNesting}
			newReflector.visit(field.Relation.Value)
			field.Relation.visited = true
		}
	}

	return err

}
