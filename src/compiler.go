package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

const (
	LABLE_PLACEHOLDER = 0x00
)

func CompileProgram(filepath string) ([]byte, error) {
	buffer, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	return Compile(string(buffer))
}

func Compile(assembly string) ([]byte, error) {
	var (
		binary  = make([]byte, 0)
		labels  = make(map[string]uint8) // label -> origin
		queries = make(map[uint8]string) // offset -> label
	)

	// For each line
	for i, line := range strings.Split(assembly, "\n") {
		line = PurifyLine(line)
		line = UpdateLabels(line, binary, labels)

		// If the line is empty, skip it
		if line == "" {
			continue
		}

		// If the line is an instruction, add it to the binary
		inst, args := ParseInstruction(line)
		switch inst {
		case "NOP":
			binary = append(binary, 0x00)
		case "MOV":
			switch args {
			case "ACC,A":
				binary = append(binary, 0x01)
			case "A,ACC":
				binary = append(binary, 0x02)
			case "ACC,[DTPR]":
				binary = append(binary, 0x04)
			case "DTPR,ACC":
				binary = append(binary, 0x05)
			case "[DTPR],ACC":
				binary = append(binary, 0x06)
			default:
				if !strings.HasPrefix(args, "ACC,") {
					return nil, fmt.Errorf("invalid instruction at line: %d", i+1)
				}

				dst, err := ParseRaw(args[4:])
				if err != nil {
					UpdateQuery(&dst, args[4:], binary, queries)
				}

				binary = append(binary, 0x03, dst)
			}
		case "INV":
			binary = append(binary, 0x07)
		case "AND":
			binary = append(binary, 0x08)
		case "ADD":
			binary = append(binary, 0x09)
		case "JMP":
			dst, err := ParseRaw(args)
			if err != nil {
				UpdateQuery(&dst, args, binary, queries)
			}

			binary = append(binary, 0x0A, dst)
		case "JZ":
			dst, err := ParseRaw(args)
			if err != nil {
				UpdateQuery(&dst, args, binary, queries)
			}

			binary = append(binary, 0x0B, dst)
		case "JN":
			dst, err := ParseRaw(args)
			if err != nil {
				UpdateQuery(&dst, args, binary, queries)
			}

			binary = append(binary, 0x0C, dst)
		case "JC":
			dst, err := ParseRaw(args)
			if err != nil {
				UpdateQuery(&dst, args, binary, queries)
			}

			binary = append(binary, 0x0D, dst)
		case "CALL":
			dst, err := ParseRaw(args)
			if err != nil {
				UpdateQuery(&dst, args, binary, queries)
			}

			binary = append(binary, 0x0E, dst)
		case "RET":
			binary = append(binary, 0x0F)
		case "RSH":
			switch args {
			case "ACC":
				binary = append(binary, 0x10)
			case "ACC,A":
				binary = append(binary, 0x11)
			default:
				if !strings.HasPrefix(args, "ACC,") {
					return nil, fmt.Errorf("invalid instruction at line: %d", i+1)
				}

				dst, err := ParseRaw(args[4:])
				if err != nil {
					UpdateQuery(&dst, args[4:], binary, queries)
				}

				binary = append(binary, 0x12, dst)
			}
		case "LSH":
			switch args {
			case "ACC":
				binary = append(binary, 0x14)
			case "ACC,A":
				binary = append(binary, 0x15)
			default:
				if !strings.HasPrefix(args, "ACC,") {
					return nil, fmt.Errorf("invalid instruction at line: %d", i+1)
				}

				dst, err := ParseRaw(args[4:])
				if err != nil {
					UpdateQuery(&dst, args[4:], binary, queries)
				}

				binary = append(binary, 0x16, dst)
			}
		case "HLT":
			binary = append(binary, 0xFF)
		default:
			number, err := ParseRaw(inst)
			if err != nil {
				return nil, fmt.Errorf("invalid instruction at line: %d", i+1)
			}

			binary = append(binary, number)
		}
	}

	// Parse labels
	for offset, label := range queries {
		pos, ok := labels[label]
		if !ok {
			return nil, fmt.Errorf("invalid label \"%s\" position at offset: %d", label, offset)
		}

		binary[offset] = pos
	}

	return binary, nil
}

func PurifyLine(line string) string {
	// Remove comments
	if commentIndex := strings.Index(line, ";"); commentIndex != -1 {
		line = line[:commentIndex]
	}

	// Remove whitespace
	line = strings.TrimSpace(line)

	// Remove tabs
	line = strings.ReplaceAll(line, "\t", "")

	// Remove newlines
	line = strings.ReplaceAll(line, "\n", "")

	// Remove carriage returns
	line = strings.ReplaceAll(line, "\r", "")

	// Remove all spaces after the first one
	if spaceIndex := strings.Index(line, " "); spaceIndex != -1 {
		line = line[:spaceIndex+1] + strings.ReplaceAll(line[spaceIndex:], " ", "")
	}

	// Make all characters uppercase
	line = strings.ToUpper(line)

	return line
}

func ParseInstruction(line string) (string, string) {
	params := strings.Split(line, " ")
	if len(params) == 1 {
		return params[0], ""
	}

	return params[0], params[1]
}

func ParseRaw(line string) (uint8, error) {
	var (
		number uint64
		err    error
	)

	if strings.HasPrefix(line, "0X") {
		number, err = strconv.ParseUint(line[2:], 16, 8)
	} else if strings.HasPrefix(line, "0B") {
		number, err = strconv.ParseUint(line[2:], 2, 8)
	} else {
		number, err = strconv.ParseUint(line, 10, 8)
	}

	if err != nil {
		return 0, err
	}

	return uint8(number), nil
}

func UpdateLabels(line string, binary []byte, labels map[string]uint8) string {
	// Find label positions and remove them from the line
	if labelIndex := strings.Index(line, ":"); labelIndex != -1 {
		labels[line[:labelIndex]] = uint8(len(binary))
		line = line[labelIndex+1:]
	}

	// Re-purify the line
	return PurifyLine(line)
}

func UpdateQuery(dst *uint8, labelstr string, binary []byte, queries map[uint8]string) {
	*dst = LABLE_PLACEHOLDER
	queries[uint8(len(binary)+1)] = labelstr
}
