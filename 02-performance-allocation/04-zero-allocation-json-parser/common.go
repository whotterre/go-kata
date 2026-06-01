package main

import "io"

type SensorRecord struct {
	SensorId string    `json:"sensor_id"`
	Readings []float64 `json:"readings"`
}

type SensorData struct {
	SensorId string  `json:"sensor_id"`
	Value    float64 `json:"value"`
}

type Parser interface {
	Parse(src io.Reader) []SensorData
}
