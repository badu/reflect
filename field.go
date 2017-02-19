package reflector

import (
	"bytes"
	"fmt"
	"strings"
)

func (field *Field) setFlag(value uint8) {
	field.flags = field.flags | (1 << value)
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

// Stringer implementation
func (field Field) String() string {
	var result bytes.Buffer
	tabs := strings.Repeat("\t", field.printNesting)
	result.WriteString(tabs + "Field Name : " + field.Name + "\n")
	result.WriteString(tabs + "Field Type : " + field.Type.Name() + "\n")
	result.WriteString(tabs + fmt.Sprintf("Field Value : `%v", field.Value) + "`\n")
	result.WriteString(tabs)
	if field.flags&(1<<ff_is_anonymous) != 0 {
		result.WriteString("Anonymous ")
	}
	if field.flags&(1<<ff_is_time) != 0 {
		result.WriteString("Time ")
	}
	if field.flags&(1<<ff_is_slice) != 0 {
		result.WriteString("Slice ")
	}
	if field.flags&(1<<ff_is_struct) != 0 {
		result.WriteString("Struct ")
	}
	if field.flags&(1<<ff_is_map) != 0 {
		result.WriteString("Map ")
	}
	if field.flags&(1<<ff_is_pointer) != 0 {
		result.WriteString("Pointer ")
	}
	if field.flags&(1<<ff_is_relation) != 0 {
		result.WriteString("Relation ")
	}
	if field.flags&(1<<ff_is_interface) != 0 {
		result.WriteString("Interface ")
	}
	result.WriteString("\n")
	if field.Relation != nil {
		field.Relation.printNesting = field.printNesting + 1
		result.WriteString(tabs + "Relation :\n")
		result.WriteString(field.Relation.String())
	}
	return result.String()
}
