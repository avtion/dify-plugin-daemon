package plugin_entities

import (
	"fmt"
	"sync"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/parser"
)

// ConcurrentTestError represents an error that occurred during concurrent testing
type ConcurrentTestError struct {
	Message string
	Panic   string
}

func (e *ConcurrentTestError) Error() string {
	return fmt.Sprintf("%s: %s", e.Message, e.Panic)
}

// TestJSONSchemaConcurrentAccess tests that the isJSONSchema function
// doesn't panic when accessed concurrently by multiple goroutines.
// This test specifically addresses the "concurrent map iteration and map write" panic.
func TestJSONSchemaConcurrentAccess(t *testing.T) {
	// Create a test struct with a ToolOutputSchema field that uses json_schema validation
	type TestSchema struct {
		Schema ToolOutputSchema `validate:"json_schema"`
	}

	validator := validator.New()
	validator.RegisterValidation("json_schema", isJSONSchema)

	// Number of concurrent goroutines
	const numGoroutines = 100
	const numIterations = 10

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Channel to collect any errors/panics
	errChan := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(routineID int) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					errChan <- &ConcurrentTestError{
						Message: "panic in goroutine",
						Panic:   parser.MarshalJson(r),
					}
				}
			}()

			for j := 0; j < numIterations; j++ {
				// Create a schema with some variation to make it more realistic
				schema := ToolOutputSchema{
					"type": "object",
					"properties": map[string]any{
						"message": map[string]any{
							"type": "string",
						},
						"status": map[string]any{
							"type": "number",
						},
						fmt.Sprintf("dynamic_%d_%d", routineID, j): map[string]any{
							"type": "string",
						},
					},
				}

				testData := TestSchema{Schema: schema}

				// Validate the schema (this triggers isJSONSchema function)
				if err := validator.Struct(&testData); err != nil {
					errChan <- err
					return
				}
			}
		}(i)
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(errChan)

	// Check if any errors occurred
	for err := range errChan {
		t.Errorf("Concurrent access test failed: %v", err)
	}
}

// TestDeepCopyValue tests that the deepCopyValue function correctly
// creates deep copies of complex nested structures
func TestDeepCopyValue(t *testing.T) {
	// Test basic types
	if result := deepCopyValue(nil); result != nil {
		t.Errorf("Expected nil, got %v", result)
	}

	if result := deepCopyValue("test"); result != "test" {
		t.Errorf("Expected 'test', got %v", result)
	}

	if result := deepCopyValue(42); result != 42 {
		t.Errorf("Expected 42, got %v", result)
	}

	// Test map[string]any
	originalMap := map[string]any{
		"string": "value",
		"number": 42,
		"nested": map[string]any{
			"inner": "value",
		},
		"array": []any{"a", "b", "c"},
	}

	copiedMap := deepCopyValue(originalMap).(map[string]any)

	// Verify it's a different instance
	if &originalMap == &copiedMap {
		t.Error("Deep copy should create a new map instance")
	}

	// Verify contents are equal
	if copiedMap["string"] != "value" {
		t.Errorf("Expected 'value', got %v", copiedMap["string"])
	}

	if copiedMap["number"] != 42 {
		t.Errorf("Expected 42, got %v", copiedMap["number"])
	}

	// Verify nested map is also copied
	nestedOriginal := originalMap["nested"].(map[string]any)
	nestedCopied := copiedMap["nested"].(map[string]any)

	if &nestedOriginal == &nestedCopied {
		t.Error("Deep copy should create new instances for nested maps")
	}

	if nestedCopied["inner"] != "value" {
		t.Errorf("Expected 'value', got %v", nestedCopied["inner"])
	}

	// Verify array is also copied
	arrayOriginal := originalMap["array"].([]any)
	arrayCopied := copiedMap["array"].([]any)

	if &arrayOriginal == &arrayCopied {
		t.Error("Deep copy should create new instances for arrays")
	}

	if len(arrayCopied) != 3 || arrayCopied[0] != "a" {
		t.Errorf("Array not copied correctly: %v", arrayCopied)
	}

	// Test that modifying the copy doesn't affect the original
	copiedMap["new_key"] = "new_value"
	if _, exists := originalMap["new_key"]; exists {
		t.Error("Modifying copy should not affect original")
	}

	nestedCopied["new_nested"] = "new_nested_value"
	if _, exists := nestedOriginal["new_nested"]; exists {
		t.Error("Modifying nested copy should not affect original")
	}
}