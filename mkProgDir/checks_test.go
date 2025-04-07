package main

import (
	"testing"

	"github.com/nickwells/testhelper.mod/v2/testhelper"
)

func TestCheckTypeMap(t *testing.T) {
	if len(checkTypeMap) != len(checkTypes) {
		t.Log("Bad check type info")
		t.Errorf("\tthe checkTypeMap has %d entries, checkTypes has %d\n",
			len(checkTypeMap), len(checkTypes))
	}

	for _, ct := range checkTypes {
		if _, ok := checkTypeMap[ct.suffix]; !ok {
			t.Log("Bad check type info")
			t.Errorf("\t%q is not in checkTypeMap", ct.suffix)
		}
	}
}

func TestCheckTypeFuncs(t *testing.T) {
	const (
		testPath = "test/file/path"
		testText = `start
line 1
line 2
end
`
	)

	testCases := []struct {
		testhelper.ID
		testhelper.ExpPanic
		suffix    string
		paramText string
		expRval   int
		expOutput string
	}{
		{
			ID:        testhelper.MkID("begin - successful match"),
			suffix:    beginsSuffix,
			paramText: "start\nline 1",
			expRval:   0,
		},
		{
			ID:        testhelper.MkID("begin - bad match"),
			suffix:    beginsSuffix,
			paramText: "start\nbad line",
			expRval:   1,
			expOutput: `"` + testPath + `" has unexpected content` +
				"\n" +
				"\tit should start with:\n" +
				"start\n" +
				"bad line\n" +
				"\tactually starts with:\n" +
				testText + "\n",
		},
		{
			ID:        testhelper.MkID("end - successful match"),
			suffix:    endsSuffix,
			paramText: "line 2\nend\n",
			expRval:   0,
		},
		{
			ID:        testhelper.MkID("end - bad match"),
			suffix:    endsSuffix,
			paramText: "line 2\nbad ending\n",
			expRval:   1,
			expOutput: `"` + testPath + `" has unexpected content` +
				"\n" +
				"\tit should end with:\n" +
				"line 2\n" +
				"bad ending\n" +
				"\n" +
				"\tactually ends with:\n" +
				testText + "\n",
		},
		{
			ID:        testhelper.MkID("contain - successful match"),
			suffix:    containsSuffix,
			paramText: "line 2\n",
			expRval:   0,
		},
		{
			ID:        testhelper.MkID("contain - bad match"),
			suffix:    containsSuffix,
			paramText: "line 3\n",
			expRval:   1,
			expOutput: `"` + testPath + `" has unexpected content` +
				"\n" +
				"\tdoes not contain:\n" +
				"line 3\n" +
				"\n",
		},
		{
			ID:        testhelper.MkID("doesNotContain - successful match"),
			suffix:    doesNotContainSuffix,
			paramText: "line 3\n",
			expRval:   0,
		},
		{
			ID:        testhelper.MkID("doesNotContain - bad match"),
			suffix:    doesNotContainSuffix,
			paramText: "line 2\n",
			expRval:   1,
			expOutput: `"` + testPath + `" has unexpected content` +
				"\n" +
				"\tcontains:\n" +
				"line 2\n" +
				"\n",
		},
		{
			ID:        testhelper.MkID("match - successful match"),
			suffix:    matchesSuffix,
			paramText: "(?m)^line 2$",
			expRval:   0,
		},
		{
			ID:        testhelper.MkID("match - bad match"),
			suffix:    matchesSuffix,
			paramText: "(?m)^line 3$",
			expRval:   1,
			expOutput: `"` + testPath + `" has unexpected content` +
				"\n" +
				"\tdoes not match:\n" +
				"(?m)^line 3$" +
				"\n",
		},
		{
			ID:        testhelper.MkID("doesNotMatch - successful match"),
			suffix:    doesNotMatchSuffix,
			paramText: "line 3\n",
			expRval:   0,
		},
		{
			ID:        testhelper.MkID("doesNotMatch - bad match"),
			suffix:    doesNotMatchSuffix,
			paramText: "(?m)^line 2$",
			expRval:   1,
			expOutput: `"` + testPath + `" has unexpected content` +
				"\n" +
				"\tmatches:\n" +
				"(?m)^line 2$" +
				"\n",
		},
	}

	for _, tc := range testCases {
		var (
			ccFuncMaker checkContentFuncMaker
			ccFunc      checkContentFunc
			ok          bool
		)

		panicked, panicVal := testhelper.PanicSafe(func() {
			ccFuncMaker, ok = checkTypeMap[tc.suffix]
			if !ok {
				t.Log(tc.IDStr())
				t.Errorf("\t: bad suffix: %q\n", tc.suffix)

				return
			}

			ccFunc = ccFuncMaker(tc.paramText)
		})
		testhelper.CheckExpPanicError(t, panicked, panicVal, tc)

		if !panicked {
			fio, err := testhelper.NewStdioFromString("")
			if err != nil {
				t.Log(tc.IDStr())
				t.Errorf("unexpected error while redirecting IO: %s\n", err)

				continue
			}

			ccFuncRval := ccFunc(testPath, testText)
			testhelper.DiffInt(t,
				tc.IDStr(), "check func return value",
				ccFuncRval, tc.expRval)

			stdout, stderr, err := fio.Done()
			if err != nil {
				t.Log(tc.IDStr())
				t.Errorf("\tunexpected error restoring IO: %s\n", err)

				continue
			}

			if len(stderr) != 0 {
				t.Log(tc.IDStr())
				t.Errorf("\tunexpected error output: %s\n", string(stderr))
			}

			testhelper.DiffString(t,
				tc.IDStr(), "std output",
				string(stdout), tc.expOutput)
		}
	}
}
