package main

import "fmt"

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
 * 0x10: RSH ACC
 * 0x11: LSH ACC
 */

type Instruction struct {
	Offset      uint8
	Opcode      uint8
	Argument    uint8
	HasArgument bool
}

func (i Instruction) String() string {
	// format: 0x(offset in hex) 0x(opcode in hex)(argument in hex) (opcode in string), (argument in string)
	if i.HasArgument {
		inst := uint16(i.Opcode)<<8 | uint16(i.Argument)
		return fmt.Sprintf("0x%02X  0x%04X │ %s,0x%02X", i.Offset, inst, OpcodeToString(i.Opcode), i.Argument)
	}

	return fmt.Sprintf("0x%02X    0x%02X │ %s", i.Offset, i.Opcode, OpcodeToString(i.Opcode))
}

func GetInstructionsFromBinary(binary [256]uint8) ([]Instruction, error) {
	instructions := make([]Instruction, 0)

	for i := 0; i < len(binary); i++ {
		instruction := Instruction{
			Offset:      uint8(i),
			Opcode:      binary[i],
			HasArgument: false,
		}

		switch instruction.Opcode {
		case 0x03, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E:
			instruction.HasArgument = true
			instruction.Argument = binary[i+1]
			i++
		}

		instructions = append(instructions, instruction)
	}

	return instructions, nil
}

func OpcodeToString(opcode uint8) string {
	switch opcode {
	case 0x00:
		return "NOP"
	case 0x01:
		return "MOV ACC, A"
	case 0x02:
		return "MOV A, ACC"
	case 0x03:
		return "MOV ACC"
	case 0x04:
		return "MOV ACC, [DPTR]"
	case 0x05:
		return "MOV DPTR, ACC"
	case 0x06:
		return "MOV [DPTR], ACC"
	case 0x07:
		return "INV ACC"
	case 0x08:
		return "AND ACC, A"
	case 0x09:
		return "ADD ACC, A"
	case 0x0A:
		return "JMP"
	case 0x0B:
		return "JZ"
	case 0x0C:
		return "JN"
	case 0x0D:
		return "JC"
	case 0x0E:
		return "CALL"
	case 0x0F:
		return "RET"
	case 0x10:
		return "RSH ACC"
	case 0x11:
		return "LSH ACC"
	default:
		return "???"
	}
}
