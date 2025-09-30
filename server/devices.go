package main

import (
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "strconv"
    "strings"
    "sort"
    "sync"
    "time"
)

type Device struct {
	ID         string `json:"id"` // d开头的12字节字符串
	Name       string `json:"name"`
	Type       string `json:"type"`
	LastOnline int64  `json:"lastOnline"` // 最近在线时间戳
	CreatedAt  int64  `json:"createdAt"`  // 创建时间戳
	UpdatedAt  int64  `json:"updatedAt"`  // 更新时间戳
}

type CreateDeviceRequest struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type UpdateDeviceRequest struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// 内存存储设备数据
var (
	devicesMu    sync.RWMutex
	devicesStore = make(map[string]Device)
	deviceSeq    = 0
)

func init() {
	// 初始化一些示例设备
	types := []string{"Sensor", "Actuator", "Gateway", "Camera"}
	now := time.Now().Unix()
	for i := 0; i < 50; i++ {
		deviceSeq++
		id := fmt.Sprintf("d%012d", deviceSeq)
		device := Device{
			ID:         id,
			Name:       fmt.Sprintf("设备%03d", i+1),
			Type:       types[i%len(types)],
			LastOnline: now - int64(i*60),
			CreatedAt:  now - int64(i*3600),
			UpdatedAt:  now - int64(i*60),
		}
		devicesStore[id] = device
	}
}

// GET /api/v1/devices (list), POST /api/v1/devices (create)
func devicesCollectionHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getDevicesList(w, r)
	case http.MethodPost:
		createDevice(w, r)
	default:
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"})
	}
}

// GET /api/v1/devices/{id}, PUT /api/v1/devices/{id}, DELETE /api/v1/devices/{id}
func deviceResourceHandler(w http.ResponseWriter, r *http.Request) {
	// 提取设备 ID
	path := r.URL.Path
	prefix := "/api/v1/devices/"
	if !strings.HasPrefix(path, prefix) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Not found"})
		return
	}
	id := strings.TrimPrefix(path, prefix)
	if id == "" || strings.Contains(id, "/") {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid device ID"})
		return
	}

	switch r.Method {
	case http.MethodGet:
		getDevice(w, r, id)
	case http.MethodPut:
		updateDevice(w, r, id)
	case http.MethodDelete:
		deleteDevice(w, r, id)
	default:
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"})
	}
}

func getDevicesList(w http.ResponseWriter, r *http.Request) {
    devicesMu.RLock()
    defer devicesMu.RUnlock()

    // 转换为切片
    devices := make([]Device, 0, len(devicesStore))
    for _, device := range devicesStore {
        devices = append(devices, device)
    }

    // 读取排序与分页参数（REST 风格：下划线命名）
    page := 1
    pageSize := 20
    q := r.URL.Query()
    sortBy := strings.ToLower(strings.TrimSpace(q.Get("sort_by")))
    order := strings.ToLower(strings.TrimSpace(q.Get("order")))
    if sortBy == "" { sortBy = "lastonline" }
    if order != "asc" { order = "desc" }
    if p := q.Get("page"); p != "" {
        if v, err := strconv.Atoi(p); err == nil && v > 0 {
            page = v
        }
    }
    if ps := q.Get("page_size"); ps != "" {
        if v, err := strconv.Atoi(ps); err == nil && v > 0 {
            pageSize = v
        }
    }

    // 排序（默认 lastOnline desc）
    less := func(i, j int) bool { return false }
    switch sortBy {
    case "lastonline":
        less = func(i, j int) bool { if devices[i].LastOnline == devices[j].LastOnline { return devices[i].ID < devices[j].ID } ; return devices[i].LastOnline < devices[j].LastOnline }
    case "createdat":
        less = func(i, j int) bool { if devices[i].CreatedAt == devices[j].CreatedAt { return devices[i].ID < devices[j].ID } ; return devices[i].CreatedAt < devices[j].CreatedAt }
    case "updatedat":
        less = func(i, j int) bool { if devices[i].UpdatedAt == devices[j].UpdatedAt { return devices[i].ID < devices[j].ID } ; return devices[i].UpdatedAt < devices[j].UpdatedAt }
    case "name":
        less = func(i, j int) bool { if devices[i].Name == devices[j].Name { return devices[i].ID < devices[j].ID } ; return devices[i].Name < devices[j].Name }
    case "id":
        less = func(i, j int) bool { return devices[i].ID < devices[j].ID }
    case "type":
        less = func(i, j int) bool { if devices[i].Type == devices[j].Type { return devices[i].ID < devices[j].ID } ; return devices[i].Type < devices[j].Type }
    default:
        // fallback
        less = func(i, j int) bool { if devices[i].LastOnline == devices[j].LastOnline { return devices[i].ID < devices[j].ID } ; return devices[i].LastOnline < devices[j].LastOnline }
    }
    if order == "asc" {
        sort.Slice(devices, less)
    } else {
        sort.Slice(devices, func(i, j int) bool { return less(j, i) })
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

    resp := map[string]interface{}{
        "devices":   paged,
        "total":     len(devices),
        "page":      page,
        "page_size": pageSize,
    }
    writeJSON(w, http.StatusOK, resp)
}

func createDevice(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}

	var req CreateDeviceRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON"})
		return
	}

	// 验证必填字段
	if strings.TrimSpace(req.Name) == "" {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": "Device name is required"})
		return
	}
	if strings.TrimSpace(req.Type) == "" {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": "Device type is required"})
		return
	}

	devicesMu.Lock()
	defer devicesMu.Unlock()

	// 生成新设备 ID
	deviceSeq++
	id := fmt.Sprintf("d%012d", deviceSeq)
	now := time.Now().Unix()

	device := Device{
		ID:         id,
		Name:       strings.TrimSpace(req.Name),
		Type:       strings.TrimSpace(req.Type),
		LastOnline: now,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	devicesStore[id] = device
	writeJSON(w, http.StatusCreated, device)
}

func getDevice(w http.ResponseWriter, r *http.Request, id string) {
	devicesMu.RLock()
	defer devicesMu.RUnlock()

	device, exists := devicesStore[id]
	if !exists {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Device not found"})
		return
	}

	writeJSON(w, http.StatusOK, device)
}

func updateDevice(w http.ResponseWriter, r *http.Request, id string) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}

	var req UpdateDeviceRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON"})
		return
	}

	devicesMu.Lock()
	defer devicesMu.Unlock()

	device, exists := devicesStore[id]
	if !exists {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Device not found"})
		return
	}

	// 更新字段
	if strings.TrimSpace(req.Name) != "" {
		device.Name = strings.TrimSpace(req.Name)
	}
	if strings.TrimSpace(req.Type) != "" {
		device.Type = strings.TrimSpace(req.Type)
	}
	device.UpdatedAt = time.Now().Unix()

	devicesStore[id] = device
	writeJSON(w, http.StatusOK, device)
}

func deleteDevice(w http.ResponseWriter, r *http.Request, id string) {
	devicesMu.Lock()
	defer devicesMu.Unlock()

	_, exists := devicesStore[id]
	if !exists {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Device not found"})
		return
	}

	delete(devicesStore, id)
	writeJSON(w, http.StatusNoContent, nil)
}
