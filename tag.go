package reflector

import (
	"bytes"
	"fmt"
	"strings"
)

func (t *Tags) Get(key string) (*Tag, error) {
	for _, tag := range t.tags {
		if tag.Key == key {
			return tag, nil
		}
	}

	return nil, errTagNotExist
}

func (t *Tags) Tags() []*Tag {
	return t.tags
}

func (t *Tags) Keys() []string {
	var keys []string
	for _, tag := range t.tags {
		keys = append(keys, tag.Key)
	}
	return keys
}

// Stringer implementation
func (t *Tags) String() string {
	tags := t.Tags()
	if len(tags) == 0 {
		return ""
	}

	var buf bytes.Buffer
	for i, tag := range t.Tags() {
		buf.WriteString(tag.String())
		if i != len(tags)-1 {
			buf.WriteString(" ")
		}
	}
	return buf.String()
}

func (t *Tag) HasOption(opt string) bool {
	for _, tagOpt := range t.Options {
		if tagOpt == opt {
			return true
		}
	}

	return false
}

// String reassembles the tag into a valid tag field representation
func (t *Tag) String() string {
	options := strings.Join(t.Options, ",")
	if options != "" {
		return fmt.Sprintf(`%s:"%s,%s"`, t.Key, t.Name, options)
	}
	return fmt.Sprintf(`%s:"%s"`, t.Key, t.Name)
}

// GoString implements the fmt.GoStringer interface
func (t *Tag) GoString() string {
	template := `{
		Key:    '%s',
		Name:   '%s',
		Option: '%s',
	}`

	if t.Options == nil {
		return fmt.Sprintf(template, t.Key, t.Name, "nil")
	}

	options := strings.Join(t.Options, ",")
	return fmt.Sprintf(template, t.Key, t.Name, options)
}

func (t *Tags) Len() int {
	return len(t.tags)
}

func (t *Tags) Less(i int, j int) bool {
	return t.tags[i].Key < t.tags[j].Key
}

func (t *Tags) Swap(i int, j int) {
	t.tags[i], t.tags[j] = t.tags[j], t.tags[i]
}
