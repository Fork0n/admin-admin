package state

import "sync"

type Role int

const (
	RoleNone Role = iota
	RoleAdmin
	RoleWorker
)

// DeviceInfo contains all information about a connected device
type DeviceInfo struct {
	ID            string // Unique identifier for the worker
	Hostname      string
	OS            string
	Architecture  string
	IPAddress     string
	CPUUsage      float64
	RAMUsage      float64
	RAMTotal      uint64
	RAMUsed       uint64
	GPUName       string
	GPUUsage      float64
	InternetSpeed string
	Uptime        uint64
	SSHEnabled    bool
	SSHPort       int
}

// AdminInfo contains info about the connected admin (for worker)
type AdminInfo struct {
	Hostname string
}

type AppState struct {
	mu               sync.RWMutex
	currentRole      Role
	connectedDevices map[string]*DeviceInfo // Multiple workers by ID
	selectedWorkerID string                 // Currently selected worker
	connectedAdmin   *AdminInfo
}

func NewAppState() *AppState {
	return &AppState{
		currentRole:      RoleNone,
		connectedDevices: make(map[string]*DeviceInfo),
		connectedAdmin:   nil,
	}
}

func (s *AppState) SetRole(role Role) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.currentRole = role
}

func (s *AppState) GetRole() Role {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.currentRole
}

// AddConnectedDevice adds a worker to the connected devices
func (s *AppState) AddConnectedDevice(device *DeviceInfo) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if device.ID == "" {
		device.ID = device.IPAddress // Use IP as ID if not set
	}
	s.connectedDevices[device.ID] = device
	// Auto-select if it's the first device
	if s.selectedWorkerID == "" {
		s.selectedWorkerID = device.ID
	}
}

// RemoveConnectedDevice removes a worker from connected devices
func (s *AppState) RemoveConnectedDevice(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.connectedDevices, id)
	if s.selectedWorkerID == id {
		s.selectedWorkerID = ""
		// Select another worker if available
		for newID := range s.connectedDevices {
			s.selectedWorkerID = newID
			break
		}
	}
}

// GetConnectedDevices returns all connected workers
func (s *AppState) GetConnectedDevices() map[string]*DeviceInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()
	// Return a copy to avoid race conditions
	copy := make(map[string]*DeviceInfo)
	for k, v := range s.connectedDevices {
		copy[k] = v
	}
	return copy
}

// GetConnectedDevicesList returns connected workers as a slice
func (s *AppState) GetConnectedDevicesList() []*DeviceInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*DeviceInfo, 0, len(s.connectedDevices))
	for _, v := range s.connectedDevices {
		list = append(list, v)
	}
	return list
}

// SetSelectedWorker sets the currently selected worker
func (s *AppState) SetSelectedWorker(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.selectedWorkerID = id
}

// GetSelectedWorker returns the currently selected worker
func (s *AppState) GetSelectedWorker() *DeviceInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.selectedWorkerID == "" {
		return nil
	}
	return s.connectedDevices[s.selectedWorkerID]
}

// GetSelectedWorkerID returns the ID of the currently selected worker
func (s *AppState) GetSelectedWorkerID() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.selectedWorkerID
}

// GetConnectedDeviceByID returns a specific worker by ID
func (s *AppState) GetConnectedDeviceByID(id string) *DeviceInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.connectedDevices[id]
}

// Legacy methods for compatibility
func (s *AppState) SetConnectedDevice(device *DeviceInfo) {
	s.AddConnectedDevice(device)
}

func (s *AppState) GetConnectedDevice() *DeviceInfo {
	return s.GetSelectedWorker()
}

func (s *AppState) UpdateDeviceMetrics(cpuUsage, ramUsage, gpuUsage float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.selectedWorkerID != "" {
		if device, ok := s.connectedDevices[s.selectedWorkerID]; ok {
			device.CPUUsage = cpuUsage
			device.RAMUsage = ramUsage
			device.GPUUsage = gpuUsage
		}
	}
}

// UpdateDeviceMetricsByID updates metrics for a specific worker
func (s *AppState) UpdateDeviceMetricsByID(id string, cpuUsage, ramUsage, gpuUsage float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if device, ok := s.connectedDevices[id]; ok {
		device.CPUUsage = cpuUsage
		device.RAMUsage = ramUsage
		device.GPUUsage = gpuUsage
	}
}

func (s *AppState) SetConnectedAdmin(admin *AdminInfo) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.connectedAdmin = admin
}

func (s *AppState) GetConnectedAdmin() *AdminInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.connectedAdmin
}

func (s *AppState) ClearConnection() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.connectedDevices = make(map[string]*DeviceInfo)
	s.selectedWorkerID = ""
	s.connectedAdmin = nil
}

func (s *AppState) IsConnected() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.connectedDevices) > 0 || s.connectedAdmin != nil
}

func (s *AppState) IsWorkerConnected() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.connectedAdmin != nil
}

// GetWorkerCount returns the number of connected workers
func (s *AppState) GetWorkerCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.connectedDevices)
}
