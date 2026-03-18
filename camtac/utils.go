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

func CreateStrippedClassTable(records *CTRecords) []*SCT {
	maxNum := 0
	for _, ct := range records.CTs {
		if ct.Num > maxNum {
			maxNum = ct.Num
		}
	}
	ctByNum := make([]*SCT, maxNum+1)
	for i := range records.CTs {
		sr := records.CTs[i]
		ctByNum[records.CTs[i].Num] = &SCT{
			Domain:     sr.Domain,
			Class:      sr.Class,
			Type:       sr.Type,
			EntityType: sr.EntityType,
			EntityIdx:  sr.EntityIdx,
		}
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

func Contains[T comparable](a []T, x T) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}
