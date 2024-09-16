package main

import (
	"os"

	"github.com/alexflint/go-arg"
)

type DevicesArgs struct {
}

type RunArgs struct {
	DeviceId string `arg:"required,positional"`
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
