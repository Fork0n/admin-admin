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
}

// AdminInfo contains info about the connected admin (for worker)
type AdminInfo struct {
	Hostname string
}

type AppState struct {
	mu              sync.RWMutex
	currentRole     Role
	connectedDevice *DeviceInfo
	connectedAdmin  *AdminInfo
}

func NewAppState() *AppState {
	return &AppState{
		currentRole:     RoleNone,
		connectedDevice: nil,
		connectedAdmin:  nil,
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

func (s *AppState) SetConnectedDevice(device *DeviceInfo) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.connectedDevice = device
}

func (s *AppState) GetConnectedDevice() *DeviceInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.connectedDevice
}

func (s *AppState) UpdateDeviceMetrics(cpuUsage, ramUsage, gpuUsage float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.connectedDevice != nil {
		s.connectedDevice.CPUUsage = cpuUsage
		s.connectedDevice.RAMUsage = ramUsage
		s.connectedDevice.GPUUsage = gpuUsage
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
	s.connectedDevice = nil
	s.connectedAdmin = nil
}

func (s *AppState) IsConnected() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.connectedDevice != nil || s.connectedAdmin != nil
}

func (s *AppState) IsWorkerConnected() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.connectedAdmin != nil
}
