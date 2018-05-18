/*
 * Copyright 2009-2018 The Go Authors. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package old

import (
	"bytes"
	"fmt"
	"strings"
)

func (model *Model) addField(field *Field) {
	/**
		_, file, line, ok := runtime.Caller(1)
		if ok {
			fmt.Printf("%s %d %s\n", model.tabs(), line, file)
		}
	**/
	if printDebug {
		fmt.Printf("%s[ADD] %q %q to %q (`%v`)\n", model.tabs(), field.Name, field.Type, model.ModelType.Name(), field.Value)
	}
	model.Fields = append(model.Fields, field)
}

func (model *Model) tabs() string {
	if model.printTabs == "" {
		model.printTabs = strings.Repeat("\t", model.printNesting)
	}
	return model.printTabs
}

// Stringer implementation
func (model Model) String() string {
	var result bytes.Buffer
	tabs := strings.Repeat("\t", model.printNesting)
	result.WriteString(tabs + "`" + model.ModelType.Name() + "` struct { \n")
	for _, field := range model.Fields {
		field.printNesting = model.printNesting + 1
		result.WriteString(field.String())
	}
	result.WriteString(tabs + "}\n")
	return result.String()
}
