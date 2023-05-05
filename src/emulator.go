package main

import "fmt"

type PDUAEmulator struct {
	ProgramCounter uint8
	Pointer        uint8
	Accumulator    int8
	XRegister      int8

	// Flags
	ZeroFlag     bool
	NegativeFlag bool
	OverflowFlag bool

	// Memory
	Memory      [256]uint8
	Stack       []uint8
	ProgramSize uint8
}

func NewPDUAEmulator() *PDUAEmulator {
	emulator := &PDUAEmulator{}
	emulator.Reset()
	return emulator
}

func (e *PDUAEmulator) Reset() {
	e.ProgramCounter = 0
	e.Accumulator = 0
	e.Pointer = 0
	e.XRegister = 0

	e.ZeroFlag = true
	e.NegativeFlag = false
	e.OverflowFlag = false

	for i := range e.Memory {
		e.Memory[i] = 0
	}

	e.Stack = make([]uint8, 0)
}

func (e *PDUAEmulator) LoadProgram(program []byte) error {
	if len(program) > len(e.Memory) {
		return fmt.Errorf("binary too large to fit in memory")
	}

	e.ProgramSize = uint8(len(program))
	for i, v := range program {
		e.Memory[i] = uint8(v)
	}

	return nil
}

/*
 * Instruction Set:
 * 0x00: NOP
 * 0x01: MOV ACC, A
 * 0x02: MOV A, ACC
 * 0x03: MOV ACC, CTE (2 bytes)
 * 0x04: MOV ACC, [DPTR]
 * 0x05: MOV DPTR, ACC
 * 0x06: MOV [DPTR], ACC
 * 0x07: INV ACC
 * 0x08: AND ACC, A
 * 0x09: ADD ACC, A
 * 0x0A: JMP CTE (2 bytes)
 * 0x0B: JZ CTE (2 bytes)
 * 0x0C: JN CTE (2 bytes)
 * 0x0D: JC CTE (2 bytes)
 * 0x0E: CALL CTE (2 bytes)
 * 0x0F: RET
 */
func (e *PDUAEmulator) Step() error {
	instruction := e.Memory[e.ProgramCounter]
	switch instruction {
	case 0x00:
		// NOP
		e.ProgramCounter++
	case 0x01:
		// MOV ACC, A
		e.Accumulator = e.XRegister
		e.ProgramCounter++
	case 0x02:
		// MOV A, ACC
		e.XRegister = e.Accumulator
		e.ProgramCounter++
	case 0x03:
		// MOV ACC, CTE (2 bytes)
		e.Accumulator = int8(e.Memory[e.ProgramCounter+1])
		e.ProgramCounter += 2
	case 0x04:
		// MOV ACC, [DPTR]
		e.Accumulator = int8(e.Memory[e.Pointer])
		e.ProgramCounter++
	case 0x05:
		// MOV DPTR, ACC
		e.Pointer = uint8(e.Accumulator)
		e.ProgramCounter++
	case 0x06:
		// MOV [DPTR], ACC
		e.Memory[e.Pointer] = uint8(e.Accumulator)
		e.ProgramCounter++
	case 0x07:
		// INV ACC
		e.Accumulator = ^e.Accumulator
		e.ProgramCounter++
	case 0x08:
		// AND ACC, A
		e.Accumulator &= e.XRegister
		e.ProgramCounter++
	case 0x09:
		// ADD ACC, A

		// Check for overflow
		if e.Accumulator > 0 && e.XRegister > 0 && e.Accumulator+e.XRegister < 0 {
			e.OverflowFlag = true
		} else if e.Accumulator < 0 && e.XRegister < 0 && e.Accumulator+e.XRegister > 0 {
			e.OverflowFlag = true
		} else {
			e.OverflowFlag = false
		}

		e.Accumulator += e.XRegister
		e.ProgramCounter++
	case 0x0A:
		// JMP CTE (2 bytes)
		e.ProgramCounter = e.Memory[e.ProgramCounter+1]
	case 0x0B:
		// JZ CTE (2 bytes)
		if e.ZeroFlag {
			e.ProgramCounter = e.Memory[e.ProgramCounter+1]
		} else {
			e.ProgramCounter += 2
		}
	case 0x0C:
		// JN CTE (2 bytes)
		if e.NegativeFlag {
			e.ProgramCounter = e.Memory[e.ProgramCounter+1]
		} else {
			e.ProgramCounter += 2
		}
	case 0x0D:
		// JC CTE (2 bytes)
		if e.OverflowFlag {
			e.ProgramCounter = e.Memory[e.ProgramCounter+1]
		} else {
			e.ProgramCounter += 2
		}
	case 0x0E:
		// CALL CTE (2 bytes)
		address := e.Memory[e.ProgramCounter+1]
		e.Stack = append(e.Stack, e.ProgramCounter+2)
		e.ProgramCounter = address
	case 0x0F:
		// RET
		if len(e.Stack) == 0 {
			return fmt.Errorf("stack underflow")
		}

		e.ProgramCounter = e.Stack[len(e.Stack)-1]
		e.Stack = e.Stack[:len(e.Stack)-1]
	case 0x10:
		// RSH ACC
		e.Accumulator >>= 1
		e.Accumulator &= 0x7F
		e.ProgramCounter++
	case 0x11:
		// LSH ACC
		e.Accumulator <<= 1
		e.ProgramCounter++
	default:
		return fmt.Errorf("unknown instruction: %x", instruction)
	}

	e.UpdateFlags()
	return nil
}

func (e *PDUAEmulator) UpdateFlags() {
	e.ZeroFlag = e.Accumulator == 0
	e.NegativeFlag = e.Accumulator < 0
}
