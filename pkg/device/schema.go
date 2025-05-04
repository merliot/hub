//go:build !tinygo

package device

import (
	"reflect"
	"strings"
)

// jsonSchema with detailed Properties
type jsonSchema struct {
	Type       string
	Properties map[string]*property
	Required   []string
}

type property struct {
	Type       string               `json:"Type,omitempty"`
	Format     string               `json:"Format,omitempty"`
	Desc       string               `json:"Description,omitempty"`
	Units      string               `json:"Units,omitempty"`
	Items      *jsonSchema          `json:"Items,omitempty"`      // For arrays
	Properties map[string]*property `json:"Properties,omitempty"` // For nested objects
	Required   []string             `json:"Required,omitempty"`   // For nested required fields
}

func structToSchema(t reflect.Type) jsonSchema {
	// Dereference pointer types until we get to the underlying type
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// Ensure we have a struct type after dereferencing
	if t.Kind() != reflect.Struct {
		return jsonSchema{
			Type:       "object",
			Properties: make(map[string]*property),
			Required:   []string{},
		}
	}

	schema := jsonSchema{
		Type:       "object",
		Properties: make(map[string]*property),
		Required:   []string{},
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldName := field.Name
		schemaTag := field.Tag.Get("schema")

		// Skip fields without schema tag
		if schemaTag == "" {
			continue
		}

		prop := &property{}
		fieldType := field.Type
		// Dereference pointer field types
		for fieldType.Kind() == reflect.Ptr {
			fieldType = fieldType.Elem()
		}

		switch fieldType.Kind() {
		case reflect.Int, reflect.Int64:
			prop.Type = "integer"
		case reflect.String:
			prop.Type = "string"
		case reflect.Struct:
			if fieldType.String() == "time.Time" {
				prop.Type = "string"
				prop.Format = "date-time"
			} else {
				// Handle nested struct
				prop.Type = "object"
				prop.Properties = structToSchema(fieldType).Properties
				prop.Required = structToSchema(fieldType).Required
			}
		case reflect.Slice, reflect.Array:
			// Handle arrays
			elemType := fieldType.Elem()
			// Dereference pointer element types
			for elemType.Kind() == reflect.Ptr {
				elemType = elemType.Elem()
			}
			if elemType.Kind() == reflect.Struct {
				prop.Type = "array"
				// Recursively generate schema for the array's element type
				prop.Items = &jsonSchema{
					Type:       "object",
					Properties: structToSchema(elemType).Properties,
					Required:   structToSchema(elemType).Required,
				}
			}
		}

		// Parse schema tags (e.g., required, min, max, format)
		for _, tag := range strings.Split(schemaTag, ",") {
			switch tag {
			case "required":
				schema.Required = append(schema.Required, fieldName)
			default:
				if strings.HasPrefix(tag, "desc=") {
					prop.Desc = strings.TrimPrefix(tag, "desc=")
				} else if strings.HasPrefix(tag, "units=") {
					prop.Units = strings.TrimPrefix(tag, "units=")
				} else if strings.HasPrefix(tag, "format=") {
					prop.Format = strings.TrimPrefix(tag, "format=")
				}
			}
		}

		schema.Properties[fieldName] = prop
	}

	return schema
}
