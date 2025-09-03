package plugin_entities

import (
	"testing"

	"github.com/go-playground/validator/v10"
)

// BenchmarkJSONSchemaValidation benchmarks the performance of isJSONSchema function
// to ensure our deep copy fix doesn't introduce significant performance regression
func BenchmarkJSONSchemaValidation(b *testing.B) {
	// Create a realistic complex JSON schema
	schema := ToolOutputSchema{
		"type": "object",
		"properties": map[string]any{
			"message": map[string]any{
				"type":        "string",
				"description": "A text message",
			},
			"status": map[string]any{
				"type":        "number",
				"description": "Status code",
				"minimum":     0,
				"maximum":     999,
			},
			"metadata": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"timestamp": map[string]any{
						"type":   "string",
						"format": "date-time",
					},
					"source": map[string]any{
						"type": "string",
					},
					"tags": map[string]any{
						"type": "array",
						"items": map[string]any{
							"type": "string",
						},
					},
				},
				"required": []any{"timestamp", "source"},
			},
			"results": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"id": map[string]any{
							"type": "string",
						},
						"value": map[string]any{
							"type": "number",
						},
						"enabled": map[string]any{
							"type": "boolean",
						},
					},
					"required": []any{"id", "value"},
				},
			},
		},
		"required": []any{"message", "status"},
	}

	type TestSchema struct {
		Schema ToolOutputSchema `validate:"json_schema"`
	}

	validator := validator.New()
	validator.RegisterValidation("json_schema", isJSONSchema)

	testData := TestSchema{Schema: schema}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := validator.Struct(&testData); err != nil {
			b.Fatalf("Validation failed: %v", err)
		}
	}
}

// BenchmarkDeepCopyValue benchmarks the performance of the deepCopyValue function
func BenchmarkDeepCopyValue(b *testing.B) {
	// Create a complex nested structure to copy
	complexValue := map[string]any{
		"string_val": "test_string",
		"number_val": 42.5,
		"bool_val":   true,
		"nested_map": map[string]any{
			"inner_string": "inner_value",
			"inner_number": 123,
			"deep_nested": map[string]any{
				"level3_string": "deep_value",
				"level3_array": []any{
					"array_item_1",
					"array_item_2",
					map[string]any{
						"array_object_key": "array_object_value",
					},
				},
			},
		},
		"top_level_array": []any{
			"array_string",
			456,
			map[string]any{
				"array_map_key": "array_map_value",
			},
			[]any{"nested_array_item_1", "nested_array_item_2"},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = deepCopyValue(complexValue)
	}
}

// BenchmarkJSONSchemaValidationComparison compares performance with and without deep copy
// This would show the performance impact of our fix
func BenchmarkJSONSchemaValidationComparison(b *testing.B) {
	// Test schema
	schema := ToolOutputSchema{
		"type": "object",
		"properties": map[string]any{
			"message": map[string]any{
				"type": "string",
			},
			"status": map[string]any{
				"type": "number",
			},
		},
	}

	b.Run("WithDeepCopy", func(b *testing.B) {
		type TestSchema struct {
			Schema ToolOutputSchema `validate:"json_schema"`
		}

		validator := validator.New()
		validator.RegisterValidation("json_schema", isJSONSchema)

		testData := TestSchema{Schema: schema}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if err := validator.Struct(&testData); err != nil {
				b.Fatalf("Validation failed: %v", err)
			}
		}
	})

	// Note: We can't easily create a "WithoutDeepCopy" benchmark here since we've fixed
	// the function, but this would show the overhead if we had both versions
	b.Run("DeepCopyOnly", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = deepCopyValue(schema)
		}
	})
}