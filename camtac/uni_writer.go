package camtac

import (
	"encoding/json"
	"os"
)

// WriteUnitsToJSON serializes a slice of Unit structures to a JSON file.
func WriteUnitsToJSON(units []interface{}, filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "  ")
	return encoder.Encode(units)
}
