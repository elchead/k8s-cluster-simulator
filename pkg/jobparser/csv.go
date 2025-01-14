package jobparser

import (
	"encoding/csv"
	"io"
	"log"
	"strconv"
	"strings"
	"time"
)

func ParsePodMemories(f io.Reader) []PodMemory {
	csvReader := csv.NewReader(f)
	csvReader.LazyQuotes = true

	data, err := csvReader.ReadAll()
	if err != nil {
		log.Fatal("Failed to read csv:", err)
	}
	res := parse(data)
	valid := make([]PodMemory, 0, len(res))
	for i, _ := range res {
		err := SetStartTime(&res[i])
		if err != nil {
			log.Println("Removing pod because:", err)
		} else {
			valid = append(valid, res[i])
		}
	}
	SortPodMemoriesByTime(valid)
	return valid
}

func parse(records [][]string) []PodMemory {
	header := records[0]
	res := initNamesForPodMemories(header)
	for timeIdx, line := range records {
		if timeIdx > 0 { // omit header line
			time, err := parseTime(line[0])
			if err != nil {
				log.Fatal("Failed to parse time:", line[0], err)
			}
			for podIdx, strmem := range line[1:] {
				mem, _ := strconv.ParseInt(strmem,10, 64)
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
