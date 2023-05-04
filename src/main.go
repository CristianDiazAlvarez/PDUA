package main

import (
	"fmt"
	"os"
)

func main() {
	filepath := "test.asm"

	// Compile the program
	binary, err := CompileProgram(filepath)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Write binary to file
	if err := os.WriteFile("out.bin", binary, 0644); err != nil {
		fmt.Println(err)
		return
	}

	emulator := NewPDUAEmulator()
	if err := emulator.LoadProgram(binary); err != nil {
		panic(err)
	}

	runner := NewRunner(emulator)
	if err := runner.Start(); err != nil {
		panic(err)
	}
}
