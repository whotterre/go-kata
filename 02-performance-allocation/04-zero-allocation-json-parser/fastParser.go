package main

import (
	"encoding/json"
	"io"
	"log"
)

type Fast struct{}

func NewFastParser() Parser {
	return &Fast{}
}

func (f *Fast) Parse(src io.Reader) []SensorData {
	var res []SensorData
	decoder := json.NewDecoder(src)

	decoder.Token()

	for decoder.More() {
		var item SensorRecord
		if err := decoder.Decode(&item); err != nil {
			log.Println("Error parsing:", err)
			continue
		}

		if len(item.Readings) > 0 {
			res = append(res, SensorData{
				SensorId: item.SensorId,
				Value:    item.Readings[0],
			})
		}
	}

	return res
}
