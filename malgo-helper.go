package main

import (
	"github.com/gen2brain/malgo"
)

type MalgoDeviceInfo struct {
	Name string
	Id   string
}

type MalgoHelper struct {
	contextInitialized bool
	context            *malgo.AllocatedContext

	deviceInitialized bool
	deviceConfig      malgo.DeviceConfig
	device            *malgo.Device
}

func NewMalgoHelper() *MalgoHelper {
	return &MalgoHelper{}
}

func (mh *MalgoHelper) InitializeContext() error {
	if mh.contextInitialized {
		return nil
	}

	var err error
	mh.context, err = malgo.InitContext(nil, malgo.ContextConfig{}, nil)

	if err != nil {
		return err
	}

	mh.contextInitialized = true
	return nil
}

func (mh *MalgoHelper) ensureContext() error {
	if !mh.contextInitialized {
		return mh.InitializeContext()
	}
	return nil
}

func (mh *MalgoHelper) UninitializeContext() {
	if mh.contextInitialized {
		_ = mh.context.Uninit()
		mh.context.Free()
		mh.context = nil
		mh.contextInitialized = false
	}
}

func (mh *MalgoHelper) ListDevices() ([]MalgoDeviceInfo, error) {
	err := mh.ensureContext()
	if err != nil {
		return nil, err
	}

	devices, err := mh.context.Devices(malgo.Capture)
	if err != nil {
		return nil, err
	}

	var deviceInfos []MalgoDeviceInfo

	for i := range devices {
		info := &devices[i]
		deviceInfos = append(deviceInfos, MalgoDeviceInfo{
			Name: info.Name(),
			Id:   info.ID.String(),
		})
	}
	return deviceInfos, nil
}

func (mh *MalgoHelper) InitializeDevice(
	id string,
) error {
	if mh.deviceInitialized {
		return nil
	}

	err := mh.ensureContext()
	if err != nil {
		return err
	}

	devices, err := mh.context.Devices(malgo.Capture)
	if err != nil {
		return err
	}

	var deviceInfo *malgo.DeviceInfo
	for i := range devices {
		info := &devices[i]
		if info.ID.String() == id {
			deviceInfo = info
			break
		}
	}

	if deviceInfo == nil {
		return &MalgoHelperDeviceNotFoundError{DeviceId: id}
	}

	dev, err = mh.context.DeviceInfo(malgo.Capture, deviceInfo.ID, malgo.Shared)
	if err != nil {
		return err
	}
	// mh.deviceConfig.Capture.DeviceID = malgo.MustDeviceID(id)
	// deviceConfig.Capture.Format = malgo.FormatS16
	// deviceConfig.Capture.Channels = 1
	return nil
}
