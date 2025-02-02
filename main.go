package main

import (
	"fmt"
	"log"
	"math"

	"github.com/gen2brain/malgo"
)

func GetContextConfig() malgo.ContextConfig {
	return malgo.ContextConfig{}
}

func GetDefaultDevice(devices []malgo.DeviceInfo) (malgo.DeviceInfo, error) {
	for _, device := range devices {
		if device.IsDefault == 1 {
			return device, nil
		}
	}
	return malgo.DeviceInfo{}, fmt.Errorf("no default device found")
}

func GetDeviceById(devices []malgo.DeviceInfo, id string) malgo.DeviceInfo {
	for _, device := range devices {
		if device.ID.String() == id {
			return device
		}
	}
	return malgo.DeviceInfo{}
}

const SampleRate = 44100

func main() {
	args := ParseArgs()

	context, err := malgo.InitContext(nil, GetContextConfig(), nil)
	if err != nil {
		log.Fatalf("Failed to initialize context: %v", err)
	}
	defer context.Uninit()
	defer context.Free()

	captureDevices, err := context.Devices(malgo.Capture)
	if err != nil {
		log.Fatalf("Failed to get capture devices: %v", err)
	}

	playbackDevices, err := context.Devices(malgo.Playback)
	if err != nil {
		log.Fatalf("Failed to get playback devices: %v", err)
	}

	if args.Devices != nil {
		fmt.Println("Capture devices:")
		for _, device := range captureDevices {
			fmt.Printf("- %s %s\n", device.Name(), device.ID.String())
		}

		fmt.Println("Playback devices:")
		for _, device := range playbackDevices {
			fmt.Printf("- %s %s\n", device.Name(), device.ID.String())
		}
	}

	if args.Run != nil {
		context, err := malgo.InitContext(nil, GetContextConfig(), nil)
		if err != nil {
			log.Fatalf("Failed to initialize context: %v", err)
		}
		defer context.Uninit()
		defer context.Free()

		var playbackDeviceInfo malgo.DeviceInfo
		if args.Run.PlaybackDeviceId != "" {
			playbackDeviceInfo = GetDeviceById(playbackDevices, args.Run.PlaybackDeviceId)
		} else {
			playbackDeviceInfo, err = GetDefaultDevice(playbackDevices)
			if err != nil {
				log.Fatalf("Failed to get default playback device: %v", err)
			}
		}

		var captureDeviceInfo malgo.DeviceInfo
		if args.Run.CaptureDeviceId != "" {
			captureDeviceInfo = GetDeviceById(captureDevices, args.Run.CaptureDeviceId)
		} else {
			captureDeviceInfo, err = GetDefaultDevice(captureDevices)
			if err != nil {
				log.Fatalf("Failed to get default capture device: %v", err)
			}
		}

		deviceConfig := malgo.DefaultDeviceConfig(malgo.Duplex)
		deviceConfig.Capture.DeviceID = captureDeviceInfo.ID.Pointer()
		deviceConfig.Capture.Format = malgo.FormatS16
		deviceConfig.Capture.Channels = 1
		deviceConfig.Playback.DeviceID = playbackDeviceInfo.ID.Pointer()
		deviceConfig.Playback.Format = malgo.FormatS16
		deviceConfig.Playback.Channels = 1
		deviceConfig.SampleRate = SampleRate
		deviceConfig.Alsa.NoMMap = 1

		callbacks := GetCallbacks(Optionns{
			Frequency: 440,
			Threshold: args.Run.Threshold,
			Amplitude: args.Run.Amplitude,
		})

		device, err := malgo.InitDevice(context.Context, deviceConfig, callbacks)
		if err != nil {
			log.Fatalf("Failed to initialize device: %v", err)
		}

		err = device.Start()
		if err != nil {
			log.Fatalf("Failed to start device: %v", err)
		}

		select {}
	}
}

type Optionns struct {
	Threshold uint16
	Amplitude uint16
	Frequency uint32
}

func GetCallbacks(options Optionns) malgo.DeviceCallbacks {
	return malgo.DeviceCallbacks{
		Data: OnRecvFramesFactory(options),
	}
}

func OnRecvFramesFactory(options Optionns) malgo.DataProc {
	var deltaSampleSecond uint32 = 0
	var sampleT1 uint32 = 0

	return func(pOutputSamples, pInputSamples []byte, sampleCount uint32) {
		values := make([]int16, len(pInputSamples)/2)
		for i := 0; i < len(pInputSamples); i += 2 {
			values[i/2] = int16(pInputSamples[i]) | int16(pInputSamples[i+1])<<8
		}

		absolutes := make([]int16, len(values))
		for i, value := range values {
			if value < 0 {
				absolutes[i] = -value
			} else {
				absolutes[i] = value
			}
		}

		max := int16(0)
		for _, value := range absolutes {
			if value > max {
				max = value
			}
		}

		if max > int16(options.Threshold) {
			sampleT1 = 0
		}

		for i := uint32(0); i < sampleCount; i++ {
			t := float64(sampleT1+i) / SampleRate

			const T1 = 0.5
			var e float64 = 1
			if t > T1 {
				e = math.Exp(-5 * (t - T1))
			}

			deltaSecondT := float64((deltaSampleSecond+i)%SampleRate) / SampleRate

			y := int16(float64(options.Amplitude) * e * math.Sin(2*math.Pi*float64(options.Frequency)*deltaSecondT))
			pOutputSamples[i*2] = byte(y & 0xff)
			pOutputSamples[i*2+1] = byte(y >> 8)
		}

		sampleT1 += sampleCount
		deltaSampleSecond += sampleCount
		if deltaSampleSecond > SampleRate {
			deltaSampleSecond = 0
		}
	}
}
