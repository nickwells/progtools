package main

import (
	"embed"
	"fmt"
	"io/fs"
	"regexp"
	"strings"

	"github.com/nickwells/location.mod/location"
)

const (
	languageGo = "Go"
)

//go:embed _templates/Go/*
var goTemplate embed.FS

var templates = map[string]struct {
	name string
	fs   embed.FS
}{
	languageGo: {
		name: "_templates/Go",
		fs:   goTemplate,
	},
}

const (
	sfxGenerate = "--mkProgDir-Generate"
	sfxCheck    = "--mkProgDir-Check"
	sfxOptional = "--mkProgDir-Optional"
)

// TemplateFileInfo contains the information about a template file
type TemplateFileInfo struct {
	path   string
	target string

	contents string

	isADir           bool
	isTheTemplateDir bool
	isAGenFile       bool
	isACheckFile     bool
	isAnOptionalFile bool

	checkTypeSuffix string
}

// dot followed by one or more digits at the end of the string
var idRE = regexp.MustCompile(`\.\d+$`)

// trimNumSuffix strips the trailing numeric suffix if present and returns
// the stripped path
func trimNumSuffix(path string) string {
	// strip the .1, .2 etc. suffix (used to allow multiple checks of the
	// same type)
	if idSuffix := idRE.FindString(path); idSuffix != "" {
		return strings.TrimSuffix(path, idSuffix)
	}

	return path
}

// populateTFIDirInfo fills in the directory information
func (prog Prog) populateTFIDirInfo(tfi *TemplateFileInfo) {
	if tfi.path == prog.walkerBase {
		tfi.isTheTemplateDir = true
		return
	}

	tfi.target = prog.makeNewPath(tfi.path)
}

// getTFIContent gets the file contents, performing macro substitution if the
// file has the appropriate extension
func (prog Prog) getTFIContent(tfi *TemplateFileInfo) error {
	b, err := fs.ReadFile(prog.templateFS, tfi.path)
	if err != nil {
		return fmt.Errorf("can't read the template file %q: %w", tfi.path, err)
	}

	tfi.contents = string(b)

	if tfi.isAGenFile {
		loc := location.New(tfi.path)

		tfi.contents, err = prog.macroCache.Substitute(string(b), loc)
		if err != nil {
			return fmt.Errorf("can't replace macros from %q: %s", tfi.path, err)
		}
	}

	return nil
}

// getCheckTypeSuffix sets the tfi.checkTypeSuffix or else returns an error
// indicating that no valid suffix could be found.
func getCheckTypeSuffix(tfi *TemplateFileInfo, path string) error {
	for suffix := range checkTypeMap {
		if strings.HasSuffix(path, suffix) {
			tfi.checkTypeSuffix = suffix
			return nil
		}
	}

	return fmt.Errorf("%q : has no valid check-type suffix", tfi.path)
}

// getTemplateFileInfo returns the information about the template file
func (prog Prog) getTemplateFileInfo(path string, d fs.DirEntry,
) (TemplateFileInfo, error) {
	tfi := TemplateFileInfo{
		path:       path,
		isADir:     d.IsDir(),
		isAGenFile: strings.HasSuffix(path, sfxGenerate),
	}
	if tfi.isADir {
		prog.populateTFIDirInfo(&tfi)
		return tfi, nil
	}

	err := prog.getTFIContent(&tfi)
	if err != nil {
		return TemplateFileInfo{}, err
	}

	if tfi.isAGenFile {
		path = strings.TrimSuffix(path, sfxGenerate)
	}

	if strings.HasSuffix(path, sfxOptional) {
		tfi.isAnOptionalFile = true
		path = strings.TrimSuffix(path, sfxOptional)
	}

	if strings.HasSuffix(path, sfxCheck) {
		tfi.isACheckFile = true
		path = strings.TrimSuffix(path, sfxCheck)
		path = trimNumSuffix(path)

		err := getCheckTypeSuffix(&tfi, path)
		if err != nil {
			return TemplateFileInfo{}, err
		}

		path = strings.TrimSuffix(path, tfi.checkTypeSuffix)
	}

	tfi.target = prog.makeNewPath(path)

	return tfi, nil
}
