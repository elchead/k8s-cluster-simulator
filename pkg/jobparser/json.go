package jobparser

import "encoding/json"

type JobData struct {
	Name string `json:"Name"`
	Memory []int64 `json:"Memory"`
}

func ParseJson(data []byte) ([]JobData, error) {
	var jobs []JobData
	err := json.Unmarshal(data, &jobs)
	if err != nil {
		return nil, err
	}
	return jobs, nil
}
