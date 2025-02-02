package main

import (
	"fmt"
	"log"
	"math"
	"time"

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

		printChan := make(chan string)
		callbacks := GetCallbacks(Options{
			Frequency: 440,
			Threshold: args.Run.Threshold,
			Amplitude: args.Run.Amplitude,
			PrintChan: printChan,
		})

		go PrintData(printChan)

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

type Options struct {
	Threshold int16
	Amplitude int16
	Frequency uint32
	PrintChan chan string
}

func GetCallbacks(options Options) malgo.DeviceCallbacks {
	return malgo.DeviceCallbacks{
		Data: OnRecvFramesFactory(options),
	}
}

func PrintData(channel chan string) {
	for channel != nil {
		fmt.Println(<-channel)
	}
}

func OnRecvFramesFactory(options Options) malgo.DataProc {
	var deltaSampleSecond uint32 = 0
	var sampleT1 uint32 = 0

	var frameCount uint64 = 0
	var lastStutterFrame uint64 = 0

	var lastTime time.Time
	return func(pOutputSamples, pInputSamples []byte, sampleCount uint32) {
		// Print the time difference between the expected and actual time
		frameCount++
		now := time.Now()
		expectedTime := float64(sampleCount) / SampleRate
		actualTime := now.Sub(lastTime).Seconds()
		lastTime = now
		diff := expectedTime - actualTime
		if diff < 0 {
			options.PrintChan <- fmt.Sprintf("frame: %d/%d, expected: %f, actual: %f, diff: %f", frameCount, frameCount-lastStutterFrame, expectedTime, actualTime, -diff)
			lastStutterFrame = frameCount
		}

		for i := uint32(0); i < sampleCount; i++ {
			// calculate the time in seconds
			t := float64(sampleT1+i) / SampleRate
			inputSample := int16(pInputSamples[i*2]) | int16(pInputSamples[i*2+1])<<8

			// absolute value
			if inputSample < 0 {
				inputSample = -inputSample
			}

			if inputSample > options.Threshold {
				sampleT1 = 0
			}

			// calculate playback sound
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

		// count the samples
		sampleT1 += sampleCount
		deltaSampleSecond += sampleCount
		if deltaSampleSecond > SampleRate {
			deltaSampleSecond = 0
		}
	}
}
