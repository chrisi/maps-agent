package camtac

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
