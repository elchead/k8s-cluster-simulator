package jobparser

import (
	"encoding/json"
	"time"
)

type JobData struct {
	Name string `json:"Name"`
	Memory []int64 `json:"Memory"`
	Time []int64 `json:"Time"`
}

func ParseJson(data []byte) ([]PodMemory, error) {
	var jobs []JobData
	err := json.Unmarshal(data, &jobs)
	if err != nil {
		return nil, err
	}
	
	pods := make([]PodMemory, 0,len(jobs))
	for _, job := range jobs {
		pods = append(pods,parseJobJson(job))
	}
	return pods, nil
}

func parseJobJson(job JobData) PodMemory {
	var podMemory PodMemory
	podMemory.Name = job.Name
	podMemory.Records = make([]Record, len(job.Memory))
	for i, mem := range job.Memory {
		podMemory.Records[i].Usage = float64(mem)
		t := job.Time[i] / 1e3
		podMemory.Records[i].Time = time.Unix(t, 0)
	}
	return podMemory
}
