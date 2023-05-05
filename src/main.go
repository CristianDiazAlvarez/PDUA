package main

import (
	"flag"
	"fmt"
	"os"
)

const (
	subcommandError   = "expected either 'emulate' or 'compile' subcommands"
	missingInputError = "expected an input file"
)

var (
	compileFlag        = flag.NewFlagSet("compile", flag.ExitOnError)
	compileInputFile   = compileFlag.String("i", "", "The input assembly file")
	compileOutputFile  = compileFlag.String("o", "", "The output binary file from the assembly [optional]")
	compileDoEmulation = compileFlag.Bool("e", false, "Whether or not to do emulation [optional]")
	emulateFlag        = flag.NewFlagSet("emulate", flag.ExitOnError)
	emulateInputFile   = emulateFlag.String("i", "", "The input binary file for emulation")
)

func init() {
	flag.Parse()
}

func main() {
	if len(os.Args) < 2 {
		exit(subcommandError)
	}

	switch os.Args[1] {
	case "compile":
		if err := compile(); err != nil {
			exit(err)
		}
	case "emulate":
		if err := emulate(); err != nil {
			exit(err)
		}
	default:
		fmt.Println(subcommandError)
		os.Exit(1)
	}
}

func compile() error {
	if *compileInputFile == "" {
		return fmt.Errorf(missingInputError)
	}

	binary, err := CompileProgram(*compileInputFile)
	if err != nil {
		return err
	}

	if *compileOutputFile != "" {
		if err := os.WriteFile(*compileOutputFile, binary, 0644); err != nil {
			return err
		}
	}

	if *compileDoEmulation {
		if err := executeEmulation(binary); err != nil {
			return err
		}
	}

	return nil
}

func emulate() error {
	if *emulateInputFile == "" {
		return fmt.Errorf(missingInputError)
	}

	binary, err := os.ReadFile(*emulateInputFile)
	if err != nil {
		return err
	}

	if err := executeEmulation(binary); err != nil {
		return err
	}

	return nil
}

func executeEmulation(binary []byte) error {
	emulator := NewPDUAEmulator()
	if err := emulator.LoadProgram(binary); err != nil {
		return err
	}

	runner := NewRunner(emulator)
	if err := runner.Start(); err != nil {
		return err
	}

	return nil
}

func exit(msg any) {
	fmt.Println(msg)
	os.Exit(1)
}
