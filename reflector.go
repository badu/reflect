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

		// set current model
		m.currentModel = &Model{ModelType: valueType, Name: valueType.PkgPath(), printNesting: 0}

		// before visiting, marking it as visiting, so circular references are avoided
		visitingModels.set(valueType, m.currentModel)

		if value.Kind() == reflect.Interface {
			if printDebug {
				fmt.Printf("Yes, it's an interface from the very beginning\n")
			}
		}

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

		// check already visiting
		if cachedValue := visitingModels.get(pointedStruct.Type()); cachedValue == nil {
			if printDebug {
				fmt.Printf("%s[Slice] field name %q of a struct %q\n", strings.Repeat("\t", m.currentModel.printNesting), m.currentField.Name, pointedStruct.Type().Name())
			}

			newReflector := &Reflector{}
			// set current model
			newReflector.currentModel = &Model{ModelType: pointedStruct.Type(), Name: pointedStruct.Type().PkgPath()}
			newReflector.currentModel.printNesting = m.currentModel.printNesting + 1
			// before visiting, marking it as visiting, so circular references are avoided
			visitingModels.set(pointedStruct.Type(), newReflector.currentModel)
			newReflector.visit(pointedStruct)
			// set the relationship
			m.currentField.Relation = newReflector.currentModel
		}
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

			newReflector := &Reflector{}
			// set current model
			newReflector.currentModel = &Model{ModelType: valueType, Name: valueType.PkgPath()}
			newReflector.currentModel.printNesting = m.currentModel.printNesting + 1

			if m.currentField != nil {
				if printDebug {
					fmt.Printf("%s[Visiting] %q -> %q\n", tabs, m.currentField.Name, newReflector.currentModel.ModelType)
				}
				m.currentField.Relation = newReflector.currentModel
			} else {
				if printDebug {
					fmt.Printf("%s[level #%d] Visiting NULLFILED -> %q over %q\n", tabs, m.currentModel.printNesting, m.currentModel.ModelType, newReflector.currentModel.ModelType)
				}
			}

			newReflector.visit(value)
			// was a pointer, transfer model to the parent
			if m.currentField == nil {
				m.currentModel = newReflector.currentModel
			}

		} else {
			if m.currentField != nil {
				if printDebug {
					fmt.Printf("%s[Inspecting] field struct %q (%q)= `%v`\n", tabs, m.currentField.Name, m.currentField.Type, value)
				}
			} else {
				if printDebug {
					fmt.Printf("%s[Inspecting] NULLIFIED `%v`\n", tabs, value)
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
		if structField.Type.Kind() == reflect.Ptr {
			//fmt.Printf("%s[Field] %q (valid = %t , nil = %t) `%v`\n", tabs, structField.Name, field.IsValid(), field.IsNil(), field)
			//fmt.Printf("%s[StructField] %v\n", tabs, structField)
		}
		if structField.Anonymous {
			if printDebug {
				fmt.Printf("Skipping Anonymous field %q %s\n", structField.Name, structField.Type)
			}
			continue
		}

		// set current field as a new field
		m.currentField = &Field{
			Type:  structField.Type,
			Name:  structField.Name,
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
			fallthrough
		case reflect.Struct:
			// set flag to struct
			m.currentField.flags = m.currentField.flags | (1 << ff_is_struct)
			// if it's not a pointer
			if m.currentField.flags&(1<<ff_is_pointer) == 0 {
				// TODO : add Scanner, Valuer, Marshaller, Unmarshaller
				_, isTime := field.Interface().(time.Time)
				if isTime {
					// set flag to time
					m.currentField.flags = m.currentField.flags | (1 << ff_is_time)
				}
				valueType := field.Type()
				// check already visiting
				if cachedValue := visitingModels.get(valueType); cachedValue == nil {
					newReflector := &Reflector{}

					// set current model
					newReflector.currentModel = &Model{ModelType: valueType, Name: valueType.PkgPath()}
					newReflector.currentModel.printNesting = m.currentModel.printNesting + 1
					// before visiting, marking it as visiting, so circular references are avoided
					visitingModels.set(valueType, newReflector.currentModel)
					if printDebug {
						fmt.Printf("%s[Child struct] %q -> %q\n", tabs, m.currentField.Name, newReflector.currentModel.ModelType)

					}
					newReflector.visit(field)
					// set the relationship
					m.currentField.Relation = newReflector.currentModel
				}
			} else {
				// field is nil, we build one to visit it
				if field.IsNil() {
					// it's a pointer. dereferencing it
					pointedElement := m.currentField.Type.Elem()
					pointedStruct := reflect.New(pointedElement).Elem()

					valueType := pointedStruct.Type()
					// check already visiting
					if cachedValue := visitingModels.get(valueType); cachedValue == nil {
						if printDebug {
							fmt.Printf("%s[Pointer] %q to a struct %q\n", tabs, m.currentField.Name, valueType.Name())
						}

						newReflector := &Reflector{}
						// set current model
						newReflector.currentModel = &Model{ModelType: valueType, Name: valueType.PkgPath()}
						newReflector.currentModel.printNesting = m.currentModel.printNesting + 1
						// before visiting, marking it as visiting, so circular references are avoided
						visitingModels.set(valueType, newReflector.currentModel)
						newReflector.visit(pointedStruct)
						// set the relationship
						m.currentField.Relation = newReflector.currentModel
					}
				} else {
					// field has values : visit it
					err := m.inspect(field)
					if err != nil {
						return err
					}
				}
			}
		case reflect.Slice:
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

		// add the field to current model
		m.currentModel.Fields = append(m.currentModel.Fields, m.currentField)

		if printDebug {
			fmt.Printf("%s[ADD] field %q (%q) = `%v` to %q\n", tabs, m.currentField.Name, m.currentField.Type, m.currentField.Value, m.currentModel.ModelType)
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

	// Set cached model
	cachedModels.set(valueType, m.currentModel)

	return err
}
