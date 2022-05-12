package jobparser

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"time"
)

type PodMemory struct {
	Name    string
	Records []Record
	StartAt time.Time
}

type Record struct {
	Time  time.Time
	Usage float64
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
func ParsePodMemories(f io.Reader) []PodMemory {
	csvReader := csv.NewReader(f)
	csvReader.LazyQuotes = true

	data, err := csvReader.ReadAll()
	if err != nil {
		log.Fatal(err)
	}
	return parse(data)
}

func parse(records [][]string) []PodMemory {
	header := records[0]
	res := initNamesForPodMemories(header)
	for timeIdx, line := range records {
		time, err := parseTime(line[0])
		if timeIdx > 0 { // omit header line
			if err != nil {
				log.Fatal(err)
			}
			for podIdx, strmem := range line[1:] {
				mem, _ := strconv.ParseFloat(strmem, 64)
				res[podIdx].Records = append(res[podIdx].Records, Record{Time: time, Usage: mem})
			}
		}

	}
	return res

}

// assume UTC
func parseTime(timestr string) (time.Time, error) {
	convertedTimeFormat := strings.Replace(timestr, " ", "T", 1) + "Z"
	time, err := time.Parse(time.RFC3339, convertedTimeFormat)
	return time, err
}

func initNamesForPodMemories(header []string) []PodMemory {
	res := make([]PodMemory, 0, len(header)-1)
	for i, title := range header {
		if i > 0 {
			name := extractPodName(title)
			res = append(res, PodMemory{Name: name, Records: make([]Record, 0, 100)})
		}
	}
	return res
}

func extractPodName(header string) string {
	return strings.Split(header, " ")[0]
}
