package reflector

import (
	"bytes"
	"strings"
)

// Stringer implementation
func (model Model) String() string {
	var result bytes.Buffer
	tabs := strings.Repeat("\t", model.printNesting)
	result.WriteString(tabs + "Model Name : " + model.Name + " \n")
	result.WriteString(tabs + "Model Type : " + model.ModelType.String() + "\n")
	for _, field := range model.Fields {
		field.printNesting = model.printNesting + 1
		result.WriteString(field.String())
	}
	return result.String()
}
