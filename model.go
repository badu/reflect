package reflector

import (
	"bytes"
	"fmt"
	"strings"
)

func (model *Model) addField(field *Field) {

	tabs := strings.Repeat("\t", model.printNesting)

	if printDebug {
		fmt.Printf("%s[ADD] %s %s to %s (`%v`)\n", tabs, field.Name, field.Type, model.Name, field.Value)
	}

	model.Fields = append(model.Fields, field)
}

// Stringer implementation
func (model Model) String() string {
	var result bytes.Buffer
	tabs := strings.Repeat("\t", model.printNesting)
	result.WriteString(tabs +"`" + model.Name + "` struct { \n")
	for _, field := range model.Fields {
		field.printNesting = model.printNesting + 1
		result.WriteString(field.String())
	}
	result.WriteString(tabs + "}\n")
	return result.String()
}
