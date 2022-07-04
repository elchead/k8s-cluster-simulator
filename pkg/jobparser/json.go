package jobparser

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"time"
)

type JobData struct {
	Name string `json:"Name"`
	Memory []int64 `json:"Memory"`
	Time []int64 `json:"Time"`
}

func FindJob(name string, jobs []PodMemory) *PodMemory {
	for _, job := range jobs {
		if job.Name == name {
			return &job
		}
	}
	return nil
}

func ParsePodMemoriesFromJson(reader io.Reader) ([]PodMemory, error) {
	var jobs []JobData
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %v",err)
	}
	err = json.Unmarshal(data, &jobs)
	if err != nil {
		return nil, err
	}
	
	pods := make([]PodMemory, 0,len(jobs))
	for _, job := range jobs {
		if len(job.Memory) > 2 { // prevent that job starts and ends at the same time
			pod, err := parseJobJson(job)
			if err != nil {
				log.Println("Removing pod because:", err)
			}
			pods = append(pods,pod)
		}
	}
	SortPodMemoriesByTime(pods)
	return pods, nil
}

func parseJobJson(job JobData) (PodMemory,error) {
	var podMemory PodMemory
	podMemory.Name = job.Name
	podMemory.Records = make([]Record, len(job.Memory))
	for i, mem := range job.Memory {
		podMemory.Records[i].Usage = mem
		t := job.Time[i] / 1e3 // timestamp in seconds (not milliseconds)
		podMemory.Records[i].Time = time.Unix(t, 0)
	}
	podMemory.StartAt = podMemory.Records[0].Time
	podMemory.EndAt = podMemory.Records[len(podMemory.Records)-1].Time
	return podMemory,nil
}
