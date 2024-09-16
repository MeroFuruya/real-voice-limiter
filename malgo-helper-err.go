package main

type MalgoHelperDeviceNotFoundError struct {
	DeviceId string
}

func (e *MalgoHelperDeviceNotFoundError) Error() string {
	return "device not found: " + e.DeviceId
}
