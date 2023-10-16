package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"

	"github.com/nickwells/check.mod/v2/check"
	"github.com/nickwells/filecheck.mod/filecheck"
	"github.com/nickwells/location.mod/location"
	"github.com/nickwells/param.mod/v6/param"
	"github.com/nickwells/param.mod/v6/psetter"
)

const (
	paramNameProgName              = "program-name"
	paramNameAction                = "action"
	paramNameCheck                 = "check"
	paramNamePerms                 = "permissions"
	paramNameCheckPerms            = "check-permissions"
	paramNameTemplateDir           = "template-directory"
	paramNameReportMissingOptFiles = "report-missing-optional-files"
)

var progNameRE = regexp.MustCompile("[a-zA-Z][-_.a-zA-Z0-9]*")

// addParams adds the parameters for this program
func addParams(prog *Prog) param.PSetOptFunc {
	return func(ps *param.PSet) error {
		ps.Add(paramNameAction,
			psetter.Enum[action]{
				Value: &prog.action,
				AllowedVals: psetter.AllowedVals[action]{
					aCreate: "create the program directory and" +
						" populate it with the standard files",
					aCheck: "check that the directory exists and" +
						" that the standard files are all present.",
				},
			},
			"The action to perform.",
			param.AltNames("a"),
		)

		ps.Add(paramNameCheck,
			psetter.Nil{},
			"Check that all the standard files are present.",
			param.PostAction(
				func(_ location.L, _ *param.ByName, _ []string) error {
					prog.action = aCheck
					return nil
				}),
		)

		ps.Add(paramNameProgName,
			psetter.String[string]{
				Value: &prog.dir,
			},
			"The name of the program to generate or check."+
				" Note that the name given forms the last part of"+
				" a directory path, the first part being"+
				" the current directory."+
				" If you are creating the program directory"+
				" the directory must not exist.",
			param.AltNames("prog-name", "name"),
			param.Attrs(param.MustBeSet),
			param.PostAction(
				func(_ location.L, _ *param.ByName, _ []string) error {
					dir := filepath.Clean(prog.dir)
					prog.name = filepath.Base(dir)
					if len(prog.name) == 0 {
						return fmt.Errorf(
							"bad name - last part of the path (%q) is empty",
							prog.dir)
					}
					return check.StringMatchesPattern[string](
						progNameRE,
						"a string starting with a letter and"+
							" followed by zero or more"+
							" letters, digits,"+
							" '.', '-' or '_'")(prog.name)
				}),
		)

		ps.Add(paramNameTemplateDir,
			psetter.Pathname{
				Value: &prog.templateDirName,
				Expectation: filecheck.Provisos{
					Existence: filecheck.MustExist,
					Checks:    []check.FileInfo{check.FileInfoIsDir},
				},
			},
			"The name of the template directory"+
				" from which to generate the program.",
			param.AltNames("template-dir", "template"),
			param.PostAction(
				func(_ location.L, _ *param.ByName, _ []string) error {
					prog.templateFS = os.DirFS(prog.templateDirName)
					prog.walkerBase = "."
					return nil
				}),
		)

		ps.Add(paramNamePerms,
			psetter.Uint[fs.FileMode]{
				Value: &prog.filePerms,
				Checks: []check.ValCk[fs.FileMode]{
					check.ValLE[fs.FileMode](0o777),
				},
			},
			"The permissions to create files with."+
				" Note that this will be subject to the application"+
				" of the umask and so may be different from the"+
				" given value."+
				" Note also that directories are created with"+
				" execute (search) permission set.",
			param.PostAction(
				func(_ location.L, _ *param.ByName, _ []string) error {
					prog.dirPerms = prog.filePerms | 0o111
					return nil
				}),
		)

		ps.Add(paramNameCheckPerms,
			psetter.Bool{
				Value: &prog.checkPerms,
			},
			"Check that the permissions of the files and directories"+
				" match the given values. Note that objects are"+
				" created with the umask applied and so may be"+
				" different from the given value. This can mean that"+
				" the created files etc do not have the given"+
				" permissions (they may have fewer).",
		)

		ps.Add(paramNameReportMissingOptFiles,
			psetter.Bool{
				Value: &prog.reportAllFiles,
			},
			"Optional files that are missing from the target"+
				" directory are reported rather than being"+
				" silently ignored.",
			param.AltNames("report-all-files"),
		)

		ps.AddFinalCheck(func() error {
			if prog.dir == "" {
				return nil
			}
			var provisos filecheck.Provisos
			switch prog.action {
			case aCreate:
				provisos = filecheck.Provisos{
					Existence: filecheck.MustNotExist,
				}
			case aCheck:
				provisos = filecheck.Provisos{
					Existence: filecheck.MustExist,
					Checks:    []check.FileInfo{check.FileInfoIsDir},
				}
			}
			return provisos.StatusCheck(prog.dir)
		})
		// ps.AddGroup("group-name", "description")
		// ps.AddExample("example", "description")
		// ps.AddNote("headline", "text")
		// ps.AddReference("name", "description")

		return nil
	}
}
