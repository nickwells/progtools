package main

import "github.com/nickwells/verbose.mod/verbose"

// Prog holds program parameters and status
type Prog struct {
	exitStatus int
	stack      *verbose.Stack
	// TODO: add the program parameter values
}

// NewProg returns a new Prog instance with the default values set
func NewProg() *Prog {
	return &Prog{
		stack: &verbose.Stack{},
		// TODO: set the initial values of the Prog members
	}
}

// SetExitStatus sets the exit status to the new value. It will not do this
// if the exit status has already been set to a non-zero value.
func (prog *Prog) SetExitStatus(es int) {
	if prog.exitStatus == 0 {
		prog.exitStatus = es
	}
}

// ForceExitStatus sets the exit status to the new value. It will do this
// regardless of the existing exit status value.
func (prog *Prog) ForceExitStatus(es int) {
	prog.exitStatus = es
}

// Run is the starting point for the program, it should be called from main()
// after the command-line parameters have been parsed. Use the setExitStatus
// method to record the exit status and then main can exit with that status.
func (prog *Prog) Run() {
	// TODO: add the program code
}
