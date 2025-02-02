package main

import (
	"os"

	"github.com/alexflint/go-arg"
)

type DevicesArgs struct {
}

type RunArgs struct {
	PlaybackDeviceId string `arg:"--playback-device-id" help:"ID of the playback device to use"`
	CaptureDeviceId  string `arg:"--capture-device-id" help:"ID of the capture device to use"`
	Threshold        uint16 `arg:"--threshold" help:"Threshold for the volume level"`
	Amplitude        uint16 `arg:"--amplitude" help:"Amplitude of the volume level"`
}

type Args struct {
	Devices *DevicesArgs `arg:"subcommand:devices"`
	Run     *RunArgs     `arg:"subcommand:run"`
}

func (Args) Description() string {
	return "Notifies you when talking too loud"
}

var Version string

func (Args) Version() string {
	return Version
}

func ParseArgs() *Args {
	args := &Args{}
	program := arg.MustParse(args)
	if program.Subcommand() == nil {
		program.WriteHelp(os.Stderr)
		program.Fail("No subcommand given")
	}
	return args
}
