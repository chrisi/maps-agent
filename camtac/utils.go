package camtac

import (
	"encoding/json"
	"os"
)

func CreateClassTable(records *CTRecords) []*CT {
	maxNum := 0
	for _, ct := range records.CTs {
		if ct.Num > maxNum {
			maxNum = ct.Num
		}
	}

	ctByNum := make([]*CT, maxNum+1)
	for i := range records.CTs {
		ctByNum[records.CTs[i].Num] = &records.CTs[i]
	}

	return ctByNum
}

func WriteToJSON(data any, filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(f)
	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}
