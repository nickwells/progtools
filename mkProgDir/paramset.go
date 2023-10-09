package main

import (
	"github.com/nickwells/param.mod/v6/param"
	"github.com/nickwells/param.mod/v6/paramset"
	"github.com/nickwells/verbose.mod/verbose"
	"github.com/nickwells/versionparams.mod/versionparams"
)

// makeParamSet generates the param set ready for parsing
func makeParamSet(prog *Prog) *param.PSet {
	return paramset.NewOrPanic(
		verbose.AddParams,
		verbose.AddTimingParams(prog.stack),
		versionparams.AddParams,

		addParams(prog),
		addNotes(prog),

		param.SetProgramDescription(
			"This will populate a directory with files suitable to form a"+
				" skeleton program (Go by default). There will be TODO"+
				" comments at various points in the code describing changes"+
				" or additions that need to be made to flesh out the program."+
				"\n\n"+
				"Note that, by default, the generated program will use"+
				" packages from the param module rather than the standard"+
				" library flag package. This is available"+
				" from github.com/nickwells/param.mod. Other modules from"+
				" the same repository base are also used."+
				"\n\n"+
				"The template used to generate the program can be"+
				" changed or replaced with another template directory."),
	)
}
