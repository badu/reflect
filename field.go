package reflector

import (
	"bytes"
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
	result.WriteString(tabs + field.Name + "\t")

	if field.flags&(1<<ff_is_slice) != 0 {
		result.WriteString("[]")
	}
	if field.flags&(1<<ff_is_anonymous) != 0 {
		result.WriteString("Anonymous ")
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
		result.WriteString(tabs + "\t" + field.Type.Name() + "\n")
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
