package main

import (
	"fmt"
	"regexp"
	"strings"
)

const (
	beginsSuffix         = ".begins"
	endsSuffix           = ".ends"
	containsSuffix       = ".contains"
	doesNotContainSuffix = ".doesNotContain"
	matchesSuffix        = ".matches"
	doesNotMatchSuffix   = ".doesNotMatch"
)

type checkTypeInfo struct {
	suffix string
	desc   string
}

var checkTypes = []checkTypeInfo{
	{
		suffix: beginsSuffix,
		desc: "the contents of the target file" +
			" begins with the contents of this file",
	},
	{
		suffix: endsSuffix,
		desc: "the contents of the target file" +
			" ends with the contents of this file",
	},
	{
		suffix: containsSuffix,
		desc: "the contents of this file" +
			" appear somewhere in the target file",
	},
	{
		suffix: doesNotContainSuffix,
		desc: "the contents of this file do" +
			" not appear anywhere in the target file",
	},
	{
		suffix: matchesSuffix,
		desc: "the contents of this file (as a Regular Expression)" +
			" match the contents of the target file",
	},
	{
		suffix: doesNotMatchSuffix,
		desc: "the contents of this file (as a Regular Expression) do" +
			" not match the contents of the target file",
	},
}

type checkContentFunc func(string, string) int

type checkContentFuncMaker func(string) checkContentFunc

var checkTypeMap = map[string]checkContentFuncMaker{
	beginsSuffix:         checkContentBegins,
	endsSuffix:           checkContentEnds,
	containsSuffix:       checkContentContains,
	doesNotContainSuffix: checkContentDoesNotContain,
	matchesSuffix:        checkContentMatches,
	doesNotMatchSuffix:   checkContentDoesNotMatch,
}

// checkContentBegins returns a function that checks that the contents begin
// with the supplied value
func checkContentBegins(begins string) checkContentFunc {
	return func(path, contents string) int {
		if strings.HasPrefix(contents, begins) {
			return 0
		}

		const maxContentToShow = 40

		s := contents
		if len(s) > maxContentToShow {
			s = s[:maxContentToShow] + "..."
		}

		fmt.Printf("%q has unexpected content\n", path)
		fmt.Printf("\t   should start with:\n%s\n", begins)
		fmt.Printf("\tactually starts with:\n%s\n", s)

		return 1
	}
}

// checkContentEnds returns a function that checks that the contents begin
// with the supplied value
func checkContentEnds(ends string) checkContentFunc {
	return func(path, contents string) int {
		if strings.HasSuffix(contents, ends) {
			return 0
		}

		const maxContentToShow = 40

		e := contents
		if len(e) > maxContentToShow {
			e = "..." + e[len(e)-maxContentToShow-1:]
		}

		fmt.Printf("%q has unexpected content\n", path)
		fmt.Printf("\t   should end with:\n%s\n", ends)
		fmt.Printf("\tactually ends with:\n%s\n", e)

		return 1
	}
}

// checkContentContains returns a function that checks that the contents
// contain the supplied value
func checkContentContains(contains string) checkContentFunc {
	return func(path, contents string) int {
		if strings.Contains(contents, contains) {
			return 0
		}

		fmt.Printf("%q has unexpected content\n", path)
		fmt.Printf("\tdoes not contain:\n%s\n", contains)

		return 1
	}
}

// checkContentDoesNotContain returns a function that checks that the contents
// do not contain the supplied value
func checkContentDoesNotContain(contains string) checkContentFunc {
	return func(path, contents string) int {
		if strings.Contains(contents, contains) {
			fmt.Printf("%q has unexpected content\n", path)
			fmt.Printf("\tcontains:\n%s\n", contains)

			return 1
		}

		return 0
	}
}

// checkContentMatches returns a function that checks that the contents match
// the supplied value (which should compile to a valid Regular Expression)
func checkContentMatches(reStr string) checkContentFunc {
	re := regexp.MustCompile(reStr)

	return func(path, contents string) int {
		if len(re.FindStringSubmatch(contents)) > 0 {
			return 0
		}

		fmt.Printf("%q has unexpected content\n", path)
		fmt.Printf("\tdoes not match:\n%s\n", re)

		return 1
	}
}

// checkContentDoesNotMatch returns a function that checks that the contents
// do not match the supplied value (which should compile to a valid Regular
// Expression)
func checkContentDoesNotMatch(reStr string) checkContentFunc {
	re := regexp.MustCompile(reStr)

	return func(path, contents string) int {
		if len(re.FindStringSubmatch(contents)) > 0 {
			fmt.Printf("%q has unexpected content\n", path)
			fmt.Printf("\tmatches:\n%s\n", re)

			return 1
		}

		return 0
	}
}
