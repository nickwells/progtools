package main

import "github.com/nickwells/verbose.mod/verbose"

// prog holds program parameters and status
type prog struct {
	exitStatus int
	stack      *verbose.Stack
	// parameters
	// TODO: add the program parameter values

	// program data
	// TODO: add the program data members (if any)
}

// newProg returns a new Prog instance with the default values set
func newProg() *prog {
	return &prog{
		stack: &verbose.Stack{},
		// TODO: set the initial values of the Prog members
	}
}

// run is the starting point for the program, it should be called from main()
// after the command-line parameters have been parsed. Use the setExitStatus
// method to record the exit status and then main can exit with that status.
func (prog *prog) run() {
	// TODO: add the program code
}
