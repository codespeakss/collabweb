package main

import (
	"encoding/json"
	"net/http"
)

type Device struct {
	ID         string `json:"id"` // d开头的12字节字符串
	Name       string `json:"name"`
	Type       string `json:"type"`
	LastOnline int64  `json:"lastOnline"` // 最近在线时间戳
}

func mockDevices() []Device {
	return []Device{
		{ID: "d000000000001", Name: "Device A", Type: "Sensor", LastOnline: 1695638400},
		{ID: "d000000000002", Name: "Device B", Type: "Actuator", LastOnline: 1695638500},
		{ID: "d000000000003", Name: "Device C", Type: "Gateway", LastOnline: 1695638600},
	}
}

func devicesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(mockDevices())
}

func main() {
	http.HandleFunc("/api/devices", devicesHandler)
	_ = http.ListenAndServe(":8080", nil)
}
