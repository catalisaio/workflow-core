package harness

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// CompareOutputs compares expected and actual outputs after normalization.
func CompareOutputs(expected, actual []map[string]interface{}) error {
	normExpected := NormalizeMapSlice(expected)
	normActual := NormalizeMapSlice(actual)

	if reflect.DeepEqual(normExpected, normActual) {
		return nil
	}

	eb, _ := json.MarshalIndent(normExpected, "", "  ")
	ab, _ := json.MarshalIndent(normActual, "", "  ")
	return fmt.Errorf("output mismatch\nexpected:\n%s\nactual:\n%s", string(eb), string(ab))
}
