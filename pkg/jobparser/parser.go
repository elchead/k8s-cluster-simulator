package jobparser

import (
	"encoding/csv"
	"io"
	"log"
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
	data, err := csvReader.ReadAll()
	if err != nil {
		log.Fatal(err)
	}
	return parse(data)
}

func parse(records [][]string) []PodMemory {
	header := records[0]
	res := make([]PodMemory, 0, len(header))
	for i, title := range header {
		if i > 0 { // omit time column
			name := extractPodName(title)
			res = append(res, PodMemory{Name: name})
		}
	}
	// for i, line := range records {
	//     if i > 0 { // omit header line
	// 	}
	// }
	return res

}

func extractPodName(header string) string {
	return strings.Split(header, " ")[0]
}
