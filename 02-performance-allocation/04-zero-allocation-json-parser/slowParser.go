package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
)

type Slow struct{}

func NewSlowParser() Parser {
	return &Slow{}
}

func (s *Slow) Parse(src io.Reader) []SensorData {
	var res []SensorData

	fileContents, err := io.ReadAll(src)
	if err != nil {
		log.Println("Failed to read input", err.Error())
		return res
	}

	var records []SensorRecord
	if err := json.Unmarshal(fileContents, &records); err != nil {
		log.Println("Failed to unmarshal data", err.Error())
		return res
	}

	for _, record := range records {
		if len(record.Readings) == 0 {
			continue
		}
		res = append(res, SensorData{
			SensorId: record.SensorId,
			Value:    record.Readings[0],
		})
	}

	return res
}

func main() {
	// Read the file
	file, err := os.Open("iot_stats.json")
	if err != nil {

	}

	defer file.Close()
	slowParser := NewSlowParser()
	res := slowParser.Parse(file)
	fmt.Println(res)
}
