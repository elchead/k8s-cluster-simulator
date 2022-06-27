package jobparser

import (
	"fmt"
	"sort"
	"time"
)

type PodMemory struct {
	Name    string
	Records []Record
	StartAt time.Time
	EndAt time.Time
	IsMigratingToNode string
	isMigrating bool
}

func (p PodMemory) IsMigrating() bool {
	return p.isMigrating
}

func (p *PodMemory) FinishedMigration() {
	p.isMigrating = false
}

func (p *PodMemory) StartMigration() {
	p.isMigrating = true
}



type Record struct {
	Time  time.Time
	Usage float64
}

func GetJob(name string,jobs []PodMemory) *PodMemory {  
	for i, _ := range jobs {
		if jobs[i].Name == name {
			return &jobs[i]
		}
	}
	return nil
}

func SortPodMemoriesByTime(podMemory []PodMemory) {
	sort.Slice(podMemory, func(i, j int) bool {
		return podMemory[i].StartAt.Before(podMemory[j].StartAt)
	})
}

func SetStartTime(pod *PodMemory) error {
	for _, record := range pod.Records {
		if record.Usage != 0. {
			pod.StartAt = record.Time
			return nil
		}
	}
	return fmt.Errorf("no start time found for pod %s", pod.Name)
}
