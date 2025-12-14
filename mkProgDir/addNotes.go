package main

import (
	"fmt"
	"strings"

	"github.com/nickwells/param.mod/v6/param"
)

const (
	noteBaseName = "mkProgDir - "

	noteNameTemplateDir    = noteBaseName + "Template directories"
	noteNameGeneratedFiles = noteBaseName + "Template files - generated"
	noteNameCheckFiles     = noteBaseName + "Template files - checks"
)

// addNotes adds the notes, if any, for this program
func addNotes(prog *Prog) param.PSetOptFunc {
	var checkTypeNotes strings.Builder

	var ctiContains checkTypeInfo

	for _, cti := range checkTypes {
		checkTypeNotes.WriteString("- ")
		checkTypeNotes.WriteString(cti.suffix)
		checkTypeNotes.WriteString(" : ")
		checkTypeNotes.WriteString(cti.desc)
		checkTypeNotes.WriteString("\n")

		if cti.suffix == containsSuffix {
			ctiContains = cti
		}
	}

	if ctiContains.suffix == "" {
		panic(fmt.Errorf("%q is not in the list of check types",
			containsSuffix))
	}

	var genMacroNotes strings.Builder

	for _, mi := range availableMacros {
		genMacroNotes.WriteString("- ")
		genMacroNotes.WriteString(mi.name)
		genMacroNotes.WriteString(" : ")
		genMacroNotes.WriteString(mi.desc)
		genMacroNotes.WriteString("\n")
	}

	noteNames := []string{
		noteNameTemplateDir,
		noteNameGeneratedFiles,
		noteNameCheckFiles,
	}

	startMacro, endMacro := prog.macroCache.GetStartEndStrings()

	return func(ps *param.PSet) error {
		ps.AddNote(noteNameTemplateDir,
			"You can provide your own template directory in place"+
				" of the default one built in to this command. This"+
				" note will guide you on how to do so."+
				"\n\n"+
				"A minimal template directory simply contains the files"+
				" that you want to be copied into your target directory."+
				" You can also add files to the template directory that"+
				" will:"+
				"\n"+
				"- create checks to be performed on the content of the"+
				" files in the resulting directory.\n"+
				"- generate files that are not a straightforward copy but"+
				" instead have various values substituted at run time.",
			param.NoteSeeNote(noteNames...),
			param.NoteSeeParam(paramNameTemplateDir),
		)
		ps.AddNote(noteNameGeneratedFiles,
			"To generate a file that is not just a copy of the template"+
				" file, add the suffix '"+sfxGenerate+"' to the name of the"+
				" file. The contents of the file will then be passed"+
				" through macro substitution before being copied into the"+
				" target directory."+
				" The name of the resulting file will be the name of the"+
				" template file but with the suffix removed."+
				" The following macro substitutions are allowed:"+
				"\n"+
				genMacroNotes.String()+
				"\n"+
				"The macro name must be surrounded"+
				" by '"+startMacro+"' and '"+endMacro+"'.",
			param.NoteSeeNote(noteNames...),
			param.NoteSeeParam(paramNameTemplateDir),
		)
		ps.AddNote(noteNameCheckFiles,
			"To generate checks of the contents of a file, add another"+
				" entry to the template directory with the same name as the"+
				" file whose contents are to be checked but with an"+
				" additional suffix of '"+sfxCheck+"' added."+
				" The remainder of the name is then checked for an optional"+
				" ID number used to generate distinct names so that"+
				" multiple checks of the same type can be generated. Lastly"+
				" the remaining name, after the check suffix and any ID has"+
				" been removed is checked for a final suffix showing what"+
				" type of check should be performed. The following"+
				" check-type suffixes are allowed:"+
				"\n"+
				checkTypeNotes.String()+
				"\n"+
				"If the check to be performed on the contents of the file"+
				" relies on values that need to have macro substitution"+
				" performed on them then the filename in the template"+
				" directory should have the check suffixes followed by the"+
				" generate suffix."+
				"\n\n"+
				"For example, having a file in the template directory"+
				" called:"+
				"\n"+
				"   xxx"+ctiContains.suffix+sfxCheck+
				"\n"+
				"will generate a check function that will ensure"+
				" "+ctiContains.desc+
				"\n"+
				"The target file will be 'xxx'. No file will be generated"+
				" for this entry but there should be some corresponding"+
				" template directory entry which creates the file to be"+
				" checked by the generated function."+
				"\n\n"+
				"Similarly, having a file in the template directory"+
				" called:"+
				"\n"+
				"   xxx"+ctiContains.suffix+".1"+sfxCheck+sfxGenerate+
				"\n"+
				"will generate another check function against 'xxx' that"+
				" will ensure "+ctiContains.desc+
				", the difference being that the content of the check file"+
				" will have been subject to macro substitution before being"+
				" used to generate the check function."+
				"\n"+
				"The target file will again be 'xxx'. Again, no file will"+
				" be generated for this entry.",
			param.NoteSeeNote(noteNames...),
			param.NoteSeeParam(
				paramNameTemplateDir,
				paramNameCheck,
				paramNameAction),
		)

		return nil
	}
}
