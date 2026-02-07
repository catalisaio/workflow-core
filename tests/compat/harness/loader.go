package harness

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// LoadCases loads all JSON case files from a directory.
func LoadCases(dir string) ([]Case, error) {
	pattern := filepath.Join(dir, "*.json")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("glob cases: %w", err)
	}
	sort.Strings(files)

	cases := make([]Case, 0, len(files))
	for _, file := range files {
		b, readErr := os.ReadFile(file)
		if readErr != nil {
			return nil, fmt.Errorf("read case %s: %w", file, readErr)
		}

		var c Case
		if unmarshalErr := json.Unmarshal(b, &c); unmarshalErr != nil {
			return nil, fmt.Errorf("unmarshal case %s: %w", file, unmarshalErr)
		}
		if c.ID == "" {
			return nil, fmt.Errorf("case %s missing id", file)
		}
		cases = append(cases, c)
	}

	return cases, nil
}
