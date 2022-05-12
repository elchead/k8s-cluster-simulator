package jobparser

import (
	"encoding/csv"
	"io"
	"log"
	"strconv"
	"strings"
)

type PodMemory struct {
	Name    string
	Records []Record
}

type Record struct {
	Time  string
	Usage float64
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
	// omit time column
	res := initNamesForPodMemories(header)
	for timeIdx, line := range records {
		if timeIdx > 0 { // omit header line
			time := line[0]
			for podIdx, strmem := range line[1:] {
				mem, _ := strconv.ParseFloat(strmem, 64)
				res[podIdx].Records = append(res[podIdx].Records, Record{Time: time, Usage: mem})
			}
		}

	}
	return res

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
