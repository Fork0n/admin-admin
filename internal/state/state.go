package state

type Role int

const (
	RoleNone Role = iota
	RoleAdmin
	RoleWorker
)

type DeviceInfo struct {
	Hostname     string
	OS           string
	Architecture string
	IPAddress    string
}

type AppState struct {
	currentRole     Role
	connectedDevice *DeviceInfo
}

func NewAppState() *AppState {
	return &AppState{
		currentRole:     RoleNone,
		connectedDevice: nil,
	}
}

func (s *AppState) SetRole(role Role) {
	s.currentRole = role
}

func (s *AppState) GetRole() Role {
	return s.currentRole
}

func (s *AppState) SetConnectedDevice(device *DeviceInfo) {
	s.connectedDevice = device
}

func (s *AppState) GetConnectedDevice() *DeviceInfo {
	return s.connectedDevice
}

func (s *AppState) ClearConnection() {
	s.connectedDevice = nil
}

func (s *AppState) IsConnected() bool {
	return s.connectedDevice != nil
}
