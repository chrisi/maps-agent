package camtac

import (
	"encoding/xml"
	"fmt"
	"os"
)

func LoadCTRecords(filePath string) (*CTRecords, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("read xml file %q: %w", filePath, err)
	}

	var records CTRecords
	if err := xml.Unmarshal(data, &records); err != nil {
		return nil, fmt.Errorf("unmarshal ct records xml: %w", err)
	}

	return &records, nil
}
