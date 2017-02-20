package reflector

import (
	"bytes"
	"strings"
)

// Stringer implementation
func (model Model) String() string {
	var result bytes.Buffer
	tabs := strings.Repeat("\t", model.printNesting)
	result.WriteString(tabs + model.Name + " struct { \n")
	for _, field := range model.Fields {
		field.printNesting = model.printNesting + 1
		result.WriteString(field.String())
	}
	result.WriteString(tabs + "}\n")
	return result.String()
}
