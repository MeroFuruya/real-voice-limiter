package main

import (
	"log"
)

func main() {
	args := ParseArgs()

	if args.Devices != nil {
		mh := NewMalgoHelper()
		err := mh.InitializeContext()
		if err != nil {
			log.Fatalf("Failed to initialize context: %v", err)
		}
		defer mh.UninitializeContext()

		devices, err := mh.ListDevices()
		if err != nil {
			log.Fatalf("Failed to list devices: %v", err)
		}

		for _, device := range devices {
			log.Printf("%s %s", device.Name, device.Id)
		}
	}

	if args.Run != nil {
		mh := NewMalgoHelper()
		err := mh.InitializeContext()
		if err != nil {
			log.Fatalf("Failed to initialize context: %v", err)
		}
		defer mh.UninitializeContext()

		err = mh.InitializeDevice(args.Run.DeviceId)
		if err != nil {
			log.Fatalf("Failed to initialize device: %v", err)
		}
	}
}
