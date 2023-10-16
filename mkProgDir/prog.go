package main

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"github.com/nickwells/macros.mod/macros"
	"github.com/nickwells/verbose.mod/verbose"
)

type action string

const (
	aCreate = action("create")
	aCheck  = action("check")
)

// Prog holds program parameters and status
type Prog struct {
	exitStatus int

	stack *verbose.Stack

	dir    string
	name   string
	action action

	walkerBase      string
	templateDirName string
	templateFS      fs.FS

	reportAllFiles bool

	checkPerms bool
	filePerms  fs.FileMode
	dirPerms   fs.FileMode

	fileChecks map[string][]checkContentFunc

	macroCache *macros.Cache
}

// NewProg returns a new Prog instance with the default values set
func NewProg() *Prog {
	mc, err := macros.NewCache()
	if err != nil {
		panic(fmt.Errorf("cannot build the macro cache: %w", err))
	}

	tmpl, ok := templates[languageGo]
	if !ok {
		panic(fmt.Errorf(
			"there is no template for the default language (%s)", languageGo))
	}

	return &Prog{
		filePerms:       0o664,
		dirPerms:        0o775,
		action:          aCreate,
		walkerBase:      tmpl.name,
		templateDirName: tmpl.name,
		templateFS:      tmpl.fs,
		fileChecks:      map[string][]checkContentFunc{},
		macroCache:      mc,
		stack:           &verbose.Stack{},
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

// setFileChecks populates the content check functions
func (prog *Prog) setFileChecks() {
	defer prog.stack.Start("setFileChecks", "Start")()

	err := fs.WalkDir(prog.templateFS, prog.walkerBase, prog.makeFileCheck())
	if err != nil {
		fmt.Printf("Problem found walking the template directory: %s\n", err)
		prog.SetExitStatus(1)
	}
}

// makeFileCheck returns a function used to walk the file system
// directory. For each file ending in the checkSuffix it will strip off the
// suffixes in order:
//   - the check suffix
//   - any numeric id
//   - the check-type suffix
//
// The remainder is the name of the file to be checked. A check function will
// be generated from the contents of the file and the value of the check-type
// suffix and will be added to the list of check funcs for that file.
func (prog *Prog) makeFileCheck() fs.WalkDirFunc {
	return func(path string, d fs.DirEntry, err error) error {
		defer prog.stack.Start("makeFileCheck",
			fmt.Sprintf("Start%25s: %q", "template file", path))()
		intro := prog.stack.Tag()
		defer (func() {
			// checkContentMatches can panic if the regexp doesn't compile
			if panicVal := recover(); panicVal != nil {
				verbose.Printf("%s PANIC: %v\n", intro, panicVal)
				err = fmt.Errorf("file checks for %q could not be made: %s",
					path, panicVal)
			}
		})()

		if err != nil {
			fmt.Println(err)
			return nil
		}

		tfi, err := prog.getTemplateFileInfo(path, d)
		if err != nil {
			fmt.Println(err)
			return err
		}

		if tfi.isAGenFile {
			verbose.Printf("%s %30s: %s\n", intro, "", "a generated file")
		}
		if tfi.isAnOptionalFile {
			verbose.Printf("%s %30s: %s\n", intro, "", "an optional file")
		}
		if !tfi.isACheckFile {
			verboseSkipMsg(intro, "not a check file")
			return nil
		}
		verbose.Printf("%s %30s: %s\n", intro, "check-type",
			tfi.checkTypeSuffix)

		f, ok := checkTypeMap[tfi.checkTypeSuffix]
		if !ok {
			verbose.Printf("%s %30s: %s\n", intro, "", "bad check-type")
			return fmt.Errorf("bad check-type suffix: %q", tfi.checkTypeSuffix)
		}

		prog.fileChecks[tfi.target] = append(prog.fileChecks[tfi.target],
			f(tfi.contents))

		return nil
	}
}

// Run is the starting point for the program, it should be called from main()
// after the command-line parameters have been parsed.
func (prog *Prog) Run() {
	defer prog.stack.Start("Run", os.Args[0])()

	prog.addAllMacros()

	switch prog.action {
	case aCreate:
		prog.CreateAllFiles()
		return
	case aCheck:
		prog.setFileChecks()
		prog.CheckAllFiles()
		return
	}
	fmt.Printf(
		"Unexpected action: %q - there is no code to handle this action\n",
		prog.action)
}

// CreateAllFiles creates the directory and all the files that it should
// contain.
func (prog *Prog) CreateAllFiles() {
	defer prog.stack.Start("CreateAllFiles", "Start")()
	intro := prog.stack.Tag()

	verbose.Println(intro, " make program directory")
	err := os.MkdirAll(prog.dir, prog.dirPerms)
	if err != nil {
		fmt.Printf("Cannot create the program directory (%q): %s\n",
			prog.dir, err)
		prog.SetExitStatus(1)
		return
	}

	verbose.Println(intro, " walking the template directory")
	err = fs.WalkDir(prog.templateFS, prog.walkerBase, prog.createFileFunc())

	if err != nil {
		fmt.Printf("Problem found walking the template directory: %s\n", err)
		prog.SetExitStatus(1)
	}
}

// CheckDir checks that the named directory exists and has the expected
// permissions. It reports any errors and returns false if the path does not
// refer to a Stat-able directory
func (prog *Prog) CheckDir(path string) bool {
	defer prog.stack.Start("CheckDir",
		fmt.Sprintf("Start%25s: %q", "directory to check", path))()
	intro := prog.stack.Tag()

	fi, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("directory %q does not exist\n", path)
			prog.SetExitStatus(1)
			return false
		}
		fmt.Printf("Cannot check the directory (%q):\n\t%s\n", path, err)
		prog.SetExitStatus(1)
		return false
	}
	verbose.Printf("%s %30s: %s\n", intro, "", "directory exists")
	if !fi.Mode().IsDir() {
		fmt.Printf("%q is not a directory\n", path)
		prog.SetExitStatus(1)
		return false
	}
	verbose.Printf("%s %30s: %s\n", intro, "", "is a directory")
	if prog.checkPerms &&
		fi.Mode()&fs.ModePerm != prog.dirPerms {
		fmt.Printf("Directory: %q has unexpected permissions\n", path)
		fmt.Printf("\texpected permissions %04o\n", prog.dirPerms)
		fmt.Printf("\t  actual permissions %04o\n", fi.Mode()&fs.ModePerm)
	}
	return true
}

// CheckFile checks that the given file exists, is a file, has the correct
// permissions and passes the supplied check functions.
func (prog *Prog) CheckFile(tfi TemplateFileInfo) {
	path := tfi.target
	defer prog.stack.Start("CheckFile",
		fmt.Sprintf("Start%25s: %q", "file to check", path))()
	intro := prog.stack.Tag()

	fi, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			if tfi.isAnOptionalFile && !prog.reportAllFiles {
				verbose.Printf("%s %30s: %s\n",
					intro, "", "file does not exist but is optional")
				return
			}

			fmt.Printf("%q does not exist\n", path)
			prog.SetExitStatus(1)
			return
		}
		fmt.Printf("Cannot check the file (%q):\n\t%s\n", path, err)
		prog.SetExitStatus(1)
		return
	}
	verbose.Printf("%s %30s: %s\n", intro, "", "file exists")
	if !fi.Mode().IsRegular() {
		fmt.Printf("%q is not a regular file\n", path)
		prog.SetExitStatus(1)
		return
	}
	verbose.Printf("%s %30s: %s\n", intro, "", "is a regular file")
	if prog.checkPerms &&
		fi.Mode()&fs.ModePerm != prog.filePerms {
		fmt.Printf("File: %q has unexpected permissions\n", path)
		fmt.Printf("\texpected permissions %04o\n", prog.filePerms)
		fmt.Printf("\t  actual permissions %04o\n", fi.Mode()&fs.ModePerm)
	}

	contents, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("File: %q can't be read: %s", path, err)
		prog.SetExitStatus(1)
		return
	}
	verbose.Printf("%s %30s: %s\n", intro, "", "can be read")

	chkFuncs := prog.fileChecks[tfi.target]
	if len(chkFuncs) == 0 {
		return
	}

	for _, cf := range chkFuncs {
		s := cf(path, string(contents))
		if s != 0 {
			prog.SetExitStatus(1)
			return
		}
	}
	verbose.Printf("%s %30s: %s\n", intro, "", "contents OK")
}

// CheckAllFiles checks the directory and all the files that it should
// contain.
func (prog *Prog) CheckAllFiles() {
	defer prog.stack.Start("CheckAllFiles", "Start")()
	intro := prog.stack.Tag()

	if !prog.CheckDir(prog.dir) {
		return
	}

	verbose.Println(intro, " walking the template directory")
	err := fs.WalkDir(prog.templateFS, prog.walkerBase, prog.checkFileFunc())
	if err != nil {
		fmt.Printf("Problem found walking the template directory: %s\n", err)
		prog.SetExitStatus(1)
	}
}

// checkFileFunc returns a function that will check that a file in the
// template directory is present in the target directory
func (prog *Prog) checkFileFunc() fs.WalkDirFunc {
	return func(path string, d fs.DirEntry, err error) error {
		defer prog.stack.Start("checkFileFunc",
			fmt.Sprintf("Start%25s: %q", "template file", path))()
		intro := prog.stack.Tag()

		if err != nil {
			fmt.Println(err)
			prog.SetExitStatus(1)
			return nil
		}

		tfi, err := prog.getTemplateFileInfo(path, d)
		if err != nil {
			fmt.Println(err)
			return err
		}

		if tfi.isTheTemplateDir {
			verboseSkipMsg(intro, "is the template dir")
			return nil
		}
		if tfi.isACheckFile {
			verboseSkipMsg(intro, "is a check file")
			return nil
		}

		if tfi.isAGenFile {
			verbose.Printf("%s %30s: %s\n", intro, "", "a generated file")
		}

		if tfi.isADir {
			prog.CheckDir(tfi.target)
			return nil
		}

		prog.CheckFile(tfi)
		return nil
	}
}

// createFileFunc returns a function that will copy a file from the template
// directory into the target directory or generate it from the template file
func (prog *Prog) createFileFunc() fs.WalkDirFunc {
	return func(path string, d fs.DirEntry, err error) error {
		defer prog.stack.Start("createFileFunc",
			fmt.Sprintf("Start%25s: %q", "template file", path))()
		intro := prog.stack.Tag()

		if err != nil {
			fmt.Println("Error walking the template directory: ", err)
			return nil
		}

		tfi, err := prog.getTemplateFileInfo(path, d)
		if err != nil {
			fmt.Println("Error getting the template file info: ", err)
			return err
		}

		if tfi.isTheTemplateDir {
			verboseSkipMsg(intro, "is the template dir")
			return nil
		}
		if tfi.isACheckFile {
			verboseSkipMsg(intro, "is a check file")
			return nil
		}

		if tfi.isAGenFile {
			verbose.Printf("%s %30s: %s\n", intro, "", "a generated file")
		}

		verbose.Printf("%s %30s: %q\n", intro, "file to create", tfi.target)

		if tfi.isADir {
			err = os.Mkdir(tfi.target, prog.dirPerms)
			if err != nil {
				fmt.Printf("Can't create directory %q: %s\n", tfi.target, err)
				prog.SetExitStatus(1)
			}
			return err
		}

		err = os.WriteFile(tfi.target, []byte(tfi.contents), prog.filePerms)
		if err != nil {
			fmt.Printf("Can't create %q: %s\n", tfi.target, err)
			prog.SetExitStatus(1)
			return err
		}

		verbose.Printf("%s %30s: %s\n", intro, "", "file created")
		return nil
	}
}

// makeNewPath constructs a path in the target directory from the template
// directory. Note that the template directory has files in an embedded File
// System which uses '/' as a file separator regardless of the separator used
// in the target File System.
//
// - The supplied path is first split into its component parts
// - then the first part (the name of the template directory) is removed
// - then the target directory name is added at the beginning
// - finally the new path name is built using the appropriate separator
func (prog Prog) makeNewPath(path string) string {
	path = strings.TrimPrefix(path, prog.walkerBase)
	pathParts := splitPath(path)

	pathParts = slices.Insert[[]string, string](pathParts, 0, prog.dir)
	return filepath.Join(pathParts...)
}

// splitPath splits the path (with a file separator of '/' from the embedded
// fs) into a slice of parts
func splitPath(fsPath string) []string {
	results := []string{}
	for ; fsPath != "." && fsPath != "/"; fsPath = path.Dir(fsPath) {
		file := path.Base(fsPath)
		results = append(results, file)
	}
	slices.Reverse[[]string, string](results)
	return results
}

// verboseSkipMsg prints a skipping message using the verbose print calls
func verboseSkipMsg(intro, reason string) {
	verbose.Printf("%s %30s: ** Skipping ** - %s\n", intro, "", reason)
}
