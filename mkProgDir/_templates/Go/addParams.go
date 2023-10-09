package main

import (
	"github.com/nickwells/param.mod/v6/param"
)

const (
// paramNameXxx = "xxx-xxx"
)

// addParams adds the parameters for this program
func addParams(prog *Prog) param.PSetOptFunc {
	return func(ps *param.PSet) error {
		// TODO: add parameters, final-checks, parameter groups etc
		// ps.Add(paramNameXxx, psetter.Xxx{Value: &prog.xxx},"Xxx param desc")
		// ps.AddFinalCheck(func() error {return nil})
		// ps.AddGroup("group-name", "description")
		// ps.AddExample("example", "description")
		// ps.AddNote("headline", "text")
		// ps.AddReference("name", "description")
		return nil
	}
}
