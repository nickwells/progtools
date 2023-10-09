package main

const (
	macroProgName = "ProgName"
)

// macroInfo records information about the available macros
type macroInfo struct {
	name string
	desc string
}

var availableMacros = []macroInfo{
	{
		name: macroProgName,
		desc: "this translates to the program name",
	},
}

// addAllMacros populates the macroCache
func (prog *Prog) addAllMacros() {
	prog.macroCache.AddMacro(macroProgName, prog.name)
}
