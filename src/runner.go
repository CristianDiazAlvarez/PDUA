package main

import (
	"fmt"

	"github.com/jroimartin/gocui"
)

const (
	PaddingX     = 1
	PaddingY     = 1
	AutoscrollAt = 1.0 / 4.0
)

type Runner struct {
	gui      *gocui.Gui
	emulator *PDUAEmulator
}

func NewRunner(emulator *PDUAEmulator) *Runner {
	return &Runner{
		emulator: emulator,
	}
}

func (r *Runner) Start() error {
	gui, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		return err
	}
	defer gui.Close()

	if err := r.SetupGUI(gui); err != nil {
		return err
	}

	if err := gui.MainLoop(); err != nil && err != gocui.ErrQuit {
		return err
	}
	return nil
}

func (r *Runner) SetupGUI(gui *gocui.Gui) error {
	r.gui = gui
	r.gui.Cursor = false
	r.gui.SetManagerFunc(r.Layout)

	// Step
	if err := r.gui.SetKeybinding("", gocui.KeyEnter, gocui.ModNone, r.EmulatorStep); err != nil {
		return err
	}

	// Step (alternative)
	if err := r.gui.SetKeybinding("", gocui.KeyArrowRight, gocui.ModNone, r.EmulatorStep); err != nil {
		return err
	}

	// Reset
	if err := r.gui.SetKeybinding("", 'r', gocui.ModNone, r.EmulatorReset); err != nil {
		return err
	}

	// Quit
	if err := r.gui.SetKeybinding("", 'q', gocui.ModNone, r.Quit); err != nil {
		return err
	}

	// Quit (alternative)
	if err := r.gui.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, r.Quit); err != nil {
		return err
	}

	return nil
}

func (r *Runner) Quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func (r *Runner) EmulatorReset(g *gocui.Gui, v *gocui.View) error {
	r.emulator.Reset(true)
	r.UpdateViews(g)
	return nil
}

func (r *Runner) EmulatorStep(g *gocui.Gui, v *gocui.View) error {
	if err := r.emulator.Step(); err != nil {
		return err
	}

	r.UpdateViews(g)
	return nil
}

func (r *Runner) UpdateViews(g *gocui.Gui) {
	for _, v := range g.Views() {
		v.Clear()

		switch v.Name() {
		case "assembler":
			r.UpdateAssemblerView(v)
		case "state":
			r.UpdateStateView(v)
		case "stack":
			r.UpdateStackView(v)
		}
	}
}

func (r *Runner) Layout(g *gocui.Gui) error {
	if err := r.AssemblerView(g); err != nil {
		return err
	}

	if err := r.StateViewer(g); err != nil {
		return err
	}

	if err := r.StackViewer(g); err != nil {
		return err
	}

	return nil
}

func (r *Runner) AssemblerView(g *gocui.Gui) error {
	var (
		maxX, maxY = g.Size()
		width      = int(float32(maxX-1)*(2.0/3.0)) - PaddingX
		height     = maxY - 1
	)

	if v, err := g.SetView("assembler", 0, 0, width, height); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		r.UpdateAssemblerView(v)
	}
	return nil
}

func (r *Runner) UpdateAssemblerView(v *gocui.View) {
	v.Title = "Assembler"
	v.Wrap = true
	v.Autoscroll = false
	v.Editable = false
	v.Frame = true

	instructions, err := GetInstructionsFromBinary(r.emulator.Memory)
	if err != nil {
		fmt.Fprintln(v, err)
		return
	}

	var currentLine int
	for i, instruction := range instructions {
		if instruction.Offset == r.emulator.ProgramCounter {
			currentLine = i
			break
		}
	}

	_, screenHeight := v.Size()
	for i, instruction := range instructions {
		var (
			autoscroll = float32(i) < float32(currentLine)-float32(screenHeight)*AutoscrollAt
			overflow   = screenHeight >= len(instructions)-i
		)

		if autoscroll && !overflow {
			continue
		}

		if instruction.Offset == r.emulator.ProgramCounter {
			fmt.Fprintf(v, "\033[30;47m")
		}
		fmt.Fprintf(v, "%s\n", instruction)
		fmt.Fprintf(v, "\033[0m")
	}

}

func (r *Runner) StateViewer(g *gocui.Gui) error {
	var (
		maxX, maxY = g.Size()
		offsetX    = int(float32(maxX-1)*(2.0/3.0)) + PaddingX
		width      = maxX - 1
		height     = int(float32(maxY-1)*(2.0/3.0)) - PaddingY
	)

	if v, err := g.SetView("state", offsetX, 0, width, height); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		r.UpdateStateView(v)
	}

	return nil
}

func (r *Runner) UpdateStateView(v *gocui.View) {
	v.Title = "State Viewer"
	v.Wrap = true
	v.Autoscroll = false
	v.Editable = false
	v.Frame = true

	// Registers
	fmt.Fprintln(v, "\tRegisters:")
	fmt.Fprintf(v, "\t\tPC:\t0x%02X\n", uint8(r.emulator.ProgramCounter))
	fmt.Fprintf(v, "\t\tA:\t0x%02X (%d)\n", uint8(r.emulator.XRegister), r.emulator.XRegister)
	fmt.Fprintf(v, "\t\tACC:\t0x%02X (%d)\n", uint8(r.emulator.Accumulator), r.emulator.Accumulator)
	fmt.Fprintf(v, "\t\tDTPR:\t0x%02X (%d)\n", uint8(r.emulator.Pointer), r.emulator.Pointer)
	fmt.Fprintf(v, "\t\t[DTPR]:\t0x%02X (%d)\n", r.emulator.Memory[r.emulator.Pointer], r.emulator.Memory[r.emulator.Pointer])

	fmt.Fprintln(v, "") // Newline

	// Flags
	fmt.Fprintln(v, "\tFlags:")
	fmt.Fprintf(v, "\t\t(Z)ero:\t%t\n", r.emulator.ZeroFlag)
	fmt.Fprintf(v, "\t\t(N)egative:\t%t\n", r.emulator.NegativeFlag)
	fmt.Fprintf(v, "\t\t(C)arry:\t%t\n", r.emulator.OverflowFlag)

	fmt.Fprintln(v, "") // Newline

	// Binary Info
	fmt.Fprintln(v, "\tBinary Info:")
	fmt.Fprintf(v, "\t\tBinary size:\t%d bytes\n", uint8(r.emulator.ProgramSize))
	fmt.Fprintf(v, "\t\tTotal memory:\t%d bytes\n", len(r.emulator.Memory))
}

func (r *Runner) StackViewer(g *gocui.Gui) error {
	var (
		maxX, maxY = g.Size()
		offsetX    = int(float32(maxX-1)*(2.0/3.0)) + PaddingX
		width      = maxX - 1
		offsetY    = int(float32(maxY-1) * (2.0 / 3.0))
		height     = maxY - 1
	)

	if v, err := g.SetView("stack", offsetX, offsetY, width, height); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		r.UpdateStackView(v)
	}

	return nil
}

func (r *Runner) UpdateStackView(v *gocui.View) {
	v.Title = "Stack Viewer"
	v.Wrap = true
	v.Autoscroll = true
	v.Editable = false
	v.Frame = true

	if len(r.emulator.Stack) > 0 {
		instructions, err := GetInstructionsFromBinary(r.emulator.Memory)
		if err != nil {
			fmt.Fprintln(v, err)
			return
		}

		for i, value := range r.emulator.Stack {
			var instruction Instruction
			for _, ins := range instructions {
				if ins.Offset == uint8(value) {
					instruction = ins
					break
				}
			}

			fmt.Fprintf(v, "S-%02x %s\n", i, instruction)
		}

	} else {
		fmt.Fprintln(v, "The stack is empty!")
	}
}
