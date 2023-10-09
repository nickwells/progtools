package main

import "os"

func main() {
	prog := NewProg()
	ps := makeParamSet(prog)
	ps.Parse()

	prog.Run()
	os.Exit(prog.exitStatus)
}
