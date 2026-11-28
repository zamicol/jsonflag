package jsonflag

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
)

// configIndent is the indent used in written configs.
//   - Four spaces, matching config.json5.
const configIndent = "    "

// WriteDefault writes a JSON5 config of c's defaults to path.
//   - path is literal. `$` is not expanded. An empty path is an error.
//   - c must be a struct or a non-nil pointer to a struct.
//   - Existing files are not overwritten. That error wraps os.ErrExist.
//   - Exported fields are written. A `json:"-"` field is skipped.
//   - The key is the json tag before any comma, or the Go field name.
//   - A non-empty `default` tag is parsed like a struct-tag flag.
//   - Otherwise the field's current value is written.
//   - time.Duration is written as integer nanoseconds.
//   - A non-empty `desc` tag is a // comment on the lines above that field.
//   - flag.CommandLine is not changed.
func WriteDefault(path string, c interface{}) error {
	if path == "" {
		return fmt.Errorf("jsonflag: config path is empty")
	}
	v, err := configStruct(c)
	if err != nil {
		return err
	}
	data, err := marshalDefaultConfig(v)
	if err != nil {
		return err
	}

	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		return fmt.Errorf("jsonflag: write config %s: %w", path, err)
	}
	_, werr := f.Write(data)
	cerr := f.Close()
	if werr != nil {
		return fmt.Errorf("jsonflag: write config %s: %w", path, werr)
	}
	if cerr != nil {
		return fmt.Errorf("jsonflag: write config %s: %w", path, cerr)
	}
	return nil
}

// configStruct returns the struct held by c.
//   - c must be a struct or a non-nil pointer to a struct.
func configStruct(c interface{}) (reflect.Value, error) {
	if c == nil {
		return reflect.Value{}, fmt.Errorf("jsonflag: config is nil")
	}
	v := reflect.ValueOf(c)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return reflect.Value{}, fmt.Errorf("jsonflag: config is nil")
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return reflect.Value{}, fmt.Errorf("jsonflag: config must be a struct or pointer to struct, got %s", v.Kind())
	}
	return v, nil
}

// marshalDefaultConfig renders v as an indented JSON5 object.
func marshalDefaultConfig(v reflect.Value) ([]byte, error) {
	t := v.Type()
	var b bytes.Buffer
	b.WriteString("{\n")
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}
		name, ok := jsonFieldName(field)
		if !ok {
			continue
		}
		val, err := fieldJSONValue(field, v.Field(i))
		if err != nil {
			return nil, err
		}
		encoded, err := json.Marshal(val)
		if err != nil {
			return nil, fmt.Errorf("jsonflag: marshal %s: %w", field.Name, err)
		}
		if desc := field.Tag.Get("desc"); desc != "" {
			for _, line := range strings.Split(desc, "\n") {
				b.WriteString(configIndent)
				b.WriteString("// ")
				b.WriteString(line)
				b.WriteByte('\n')
			}
		}
		key, err := json.Marshal(name)
		if err != nil {
			return nil, fmt.Errorf("jsonflag: marshal key %s: %w", field.Name, err)
		}
		b.WriteString(configIndent)
		b.Write(key)
		b.WriteString(": ")
		b.Write(encoded)
		b.WriteString(",\n")
	}
	b.WriteString("}\n")
	return b.Bytes(), nil
}

// jsonFieldName returns the object key for field.
//   - `json:"-"` returns ok false.
//   - Otherwise the name is the json tag before a comma, or the Go field name.
func jsonFieldName(field reflect.StructField) (string, bool) {
	tag := field.Tag.Get("json")
	if tag == "-" {
		return "", false
	}
	name, _, _ := strings.Cut(tag, ",")
	if name == "" {
		name = field.Name
	}
	return name, true
}

// fieldJSONValue returns the value to write for field.
//   - A non-empty default tag is parsed with parseTagDefault.
//   - Otherwise the current value is used.
func fieldJSONValue(field reflect.StructField, fv reflect.Value) (interface{}, error) {
	if tag := field.Tag.Get("default"); tag != "" {
		return parseTagDefault(field.Type, field.Name, tag)
	}
	return fv.Interface(), nil
}
