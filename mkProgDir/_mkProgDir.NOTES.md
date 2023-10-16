<!-- Created by mkdoc DO NOT EDIT. -->

# Notes

## mkProgDir \- Template directories
You can provide your own template directory in place of the default one built in
to this command\. This note will guide you on how to do so\.



A minimal template directory simply contains the files that you want to be
copied into your target directory\. You can also add files to the template
directory that will:

\- create checks to be performed on the content of the files in the resulting
directory\.

\- generate files that are not a straightforward copy but instead have various
values substituted at run time\.
### See Parameter
* template\-directory

### See Notes
* mkProgDir \- Template files \- checks
* mkProgDir \- Template files \- generated



## mkProgDir \- Template files \- checks
To generate checks of the contents of a file, add another entry to the template
directory with the same name as the file whose contents are to be checked but
with an additional suffix of &apos;\-\-mkProgDir\-Check&apos; added\. The
remainder of the name is then checked for an optional ID number used to generate
distinct names so that multiple checks of the same type can be generated\.
Lastly the remaining name, after the check suffix and any ID has been removed is
checked for a final suffix showing what type of check should be performed\. The
following check\-type suffixes are allowed:

\- \.begins : the contents of the target file begins with the contents of this
file

\- \.ends : the contents of the target file ends with the contents of this file

\- \.contains : the contents of this file appear somewhere in the target file

\- \.doesNotContain : the contents of this file do not appear anywhere in the
target file

\- \.matches : the contents of this file \(as a Regular Expression\) match the
contents of the target file

\- \.doesNotMatch : the contents of this file \(as a Regular Expression\) do not
match the contents of the target file



If the check to be performed on the contents of the file relies on values that
need to have macro substitution performed on them then the filename in the
template directory should have the check suffixes followed by the generate
suffix\.



For example, having a file in the template directory called:

   xxx\.contains\-\-mkProgDir\-Check

will generate a check function that will ensure the contents of this file appear
somewhere in the target file

The target file will be &apos;xxx&apos;\. No file will be generated for this
entry but there should be some corresponding template directory entry which
creates the file to be checked by the generated function\.



Similarly, having a file in the template directory called:

   xxx\.contains\.1\-\-mkProgDir\-Check\-\-mkProgDir\-Generate

will generate another check function against &apos;xxx&apos; that will ensure
the contents of this file appear somewhere in the target file, the difference
being that the content of the check file will have been subject to macro
substitution before being used to generate the check function\.

The target file will again be &apos;xxx&apos;\. Again, no file will be generated
for this entry\.
### See Parameters
* action
* check
* template\-directory

### See Notes
* mkProgDir \- Template directories
* mkProgDir \- Template files \- generated



## mkProgDir \- Template files \- generated
To generate a file that is not just a copy of the template file, add the suffix
&apos;\-\-mkProgDir\-Generate&apos; to the name of the file\. The contents of
the file will then be passed through macro substitution before being copied into
the target directory\. The name of the resulting file will be the name of the
template file but with the suffix removed\. The following macro substitutions
are allowed:

\- ProgName : this translates to the program name



The macro name must be surrounded by &apos;$\{&apos; and &apos;\}&apos;\.
### See Parameter
* template\-directory

### See Notes
* mkProgDir \- Template directories
* mkProgDir \- Template files \- checks



