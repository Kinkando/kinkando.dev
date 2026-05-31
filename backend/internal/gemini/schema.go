package gemini

import (
	"reflect"
	"strings"

	"google.golang.org/genai"
)

// SchemaOf converts a Go struct value (or pointer) to a *genai.Schema.
// Property names come from json tags; descriptions from jsonschema tags.
// required explicitly lists which field names must appear in the model's response.
// Returns an empty object schema (never nil) — Gemini requires an explicit
// parameters field even for tools that take no arguments.
func SchemaOf(v any, required ...string) *genai.Schema {
	t := reflect.TypeOf(v)
	for t != nil && t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t == nil || t.Kind() != reflect.Struct || t.NumField() == 0 {
		return &genai.Schema{Type: genai.TypeObject, Properties: map[string]*genai.Schema{}}
	}
	s := objectSchema(t)
	if len(required) > 0 {
		s.Required = required
	}
	return s
}

func objectSchema(t reflect.Type) *genai.Schema {
	props := make(map[string]*genai.Schema, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		name := jsonFieldName(f)
		if name == "" || name == "-" {
			continue
		}
		s := fieldSchema(f.Type)
		if desc := f.Tag.Get("jsonschema"); desc != "" {
			s.Description = desc
		}
		props[name] = s
	}
	return &genai.Schema{Type: genai.TypeObject, Properties: props}
}

func fieldSchema(t reflect.Type) *genai.Schema {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	switch t.Kind() {
	case reflect.String:
		return &genai.Schema{Type: genai.TypeString}
	case reflect.Float32, reflect.Float64:
		return &genai.Schema{Type: genai.TypeNumber}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return &genai.Schema{Type: genai.TypeInteger}
	case reflect.Bool:
		return &genai.Schema{Type: genai.TypeBoolean}
	case reflect.Slice:
		return &genai.Schema{Type: genai.TypeArray, Items: fieldSchema(t.Elem())}
	case reflect.Struct:
		return objectSchema(t)
	default:
		return &genai.Schema{Type: genai.TypeString}
	}
}

func jsonFieldName(f reflect.StructField) string {
	tag := f.Tag.Get("json")
	if tag == "" {
		return f.Name
	}
	name, _, _ := strings.Cut(tag, ",")
	return name
}
