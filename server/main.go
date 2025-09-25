package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

type Device struct {
	ID         string `json:"id"` // d开头的12字节字符串
	Name       string `json:"name"`
	Type       string `json:"type"`
	LastOnline int64  `json:"lastOnline"` // 最近在线时间戳
}

func mockDevices() []Device {
	devices := make([]Device, 300)
	types := []string{"Sensor", "Actuator", "Gateway", "Camera"}
	for i := 0; i < 300; i++ {
		devices[i] = Device{
			ID:         "d" + fmt.Sprintf("%012d", i+1),
			Name:       fmt.Sprintf("设备%03d", i+1),
			Type:       types[i%len(types)],
			LastOnline: 1695638400 + int64(i*60),
		}
	}
	return devices
}

func devicesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	devices := mockDevices()
	// 分页参数
	page := 1
	pageSize := 20
	q := r.URL.Query()
	if p := q.Get("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}
	if ps := q.Get("pageSize"); ps != "" {
		if v, err := strconv.Atoi(ps); err == nil && v > 0 {
			pageSize = v
		}
	}
	start := (page - 1) * pageSize
	end := start + pageSize
	if start > len(devices) {
		start = len(devices)
	}
	if end > len(devices) {
		end = len(devices)
	}
	paged := devices[start:end]
	// 返回分页数据和总数
	resp := map[string]interface{}{
		"devices":  paged,
		"total":    len(devices),
		"page":     page,
		"pageSize": pageSize,
	}
	json.NewEncoder(w).Encode(resp)
}

func main() {
	http.HandleFunc("/api/devices", devicesHandler)
	_ = http.ListenAndServe(":8080", nil)
}
