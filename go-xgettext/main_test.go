// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2015-2016 Canonical Ltd
 *
 * Permission is hereby granted, free of charge, to any person
 * obtaining a copy of this software and associated documentation
 * files (the "Software"), to deal in the Software without
 * restriction, including without limitation the rights to use, copy,
 * modify, merge, publish, distribute, sublicense, and/or sell copies
 * of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be
 * included in all copies or substantial portions of the Software.

 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
 * EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
 * MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
 * NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS
 * BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN
 * ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
 * CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */

package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	. "gopkg.in/check.v1"
)

// Hook up check.v1 into the "go test" runner
func Test(t *testing.T) { TestingT(t) }

type xgettextTestSuite struct {
}

var _ = Suite(&xgettextTestSuite{})

// test helper
func makeGoSourceFile(c *C, content []byte) string {
	fname := filepath.Join(c.MkDir(), "foo.go")
	err := os.WriteFile(fname, []byte(content), 0644)
	c.Assert(err, IsNil)

	return fname
}

func (s *xgettextTestSuite) SetUpTest(c *C) {
	// our test defaults
	*noLocation = false
	*addCommentsTag = "TRANSLATORS:"
	*keyword = "i18n.G"
	*keywordPlural = "i18n.NG"
	*keywordContextual = "i18n.CG"
	*sortOutput = true
	*packageName = "snappy"
	*msgIDBugsAddress = "snappy-devel@lists.ubuntu.com"

	// mock time
	formatTime = func() string {
		return "2015-06-30 14:48+0200"
	}
}

func (s *xgettextTestSuite) TestFormatComment(c *C) {
	var tests = []struct {
		in  string
		out string
	}{
		{in: "// foo ", out: "#. foo\n"},
		{in: "/* foo */", out: "#. foo\n"},
		{in: "/* foo\n */", out: "#. foo\n"},
		{in: "/* foo\nbar   */", out: "#. foo\n#. bar\n"},
	}

	for _, test := range tests {
		c.Assert(formatComment(test.in), Equals, test.out)
	}
}

func (s *xgettextTestSuite) TestProcessFilesSimple(c *C) {
	fname := makeGoSourceFile(c, []byte(`package main

func main() {
    // TRANSLATORS: foo comment
    i18n.G("foo")
}
`))
	err := processFiles([]string{fname})
	c.Assert(err, IsNil)

	c.Assert(msgIDs, DeepEquals, map[msgKey][]msgData{
		{"", "foo"}: {
			{
				comment: "#. TRANSLATORS: foo comment\n",
				fname:   fname,
				line:    5,
			},
		},
	})
}

func (s *xgettextTestSuite) TestProcessFilesMultiple(c *C) {
	fname := makeGoSourceFile(c, []byte(`package main

func main() {
    // TRANSLATORS: foo comment
    i18n.G("foo")

    // TRANSLATORS: bar comment
    i18n.G("foo")
}
`))
	err := processFiles([]string{fname})
	c.Assert(err, IsNil)

	c.Assert(msgIDs, DeepEquals, map[msgKey][]msgData{
		{"", "foo"}: {
			{
				comment: "#. TRANSLATORS: foo comment\n",
				fname:   fname,
				line:    5,
			},
			{
				comment: "#. TRANSLATORS: bar comment\n",
				fname:   fname,
				line:    8,
			},
		},
	})
}

const header = `# SOME DESCRIPTIVE TITLE.
# Copyright (C) YEAR THE PACKAGE'S COPYRIGHT HOLDER
# This file is distributed under the same license as the PACKAGE package.
# FIRST AUTHOR <EMAIL@ADDRESS>, YEAR.
#
#, fuzzy
msgid   ""
msgstr  "Project-Id-Version: snappy\n"
        "Report-Msgid-Bugs-To: snappy-devel@lists.ubuntu.com\n"
        "POT-Creation-Date: 2015-06-30 14:48+0200\n"
        "MIME-Version: 1.0\n"
        "Content-Type: text/plain; charset=utf-8\n"
        "Content-Transfer-Encoding: 8bit\n"
`

func (s *xgettextTestSuite) TestWriteOutputSimple(c *C) {
	msgIDs = map[msgKey][]msgData{
		{"", "foo"}: {
			{
				fname:   "fname",
				line:    2,
				comment: "#. foo\n",
			},
		},
	}
	out := bytes.NewBuffer([]byte(""))
	writePotFile(out)

	expected := fmt.Sprintf(`%s
#. foo
#: fname:2
msgid   "foo"
msgstr  ""

`, header)
	c.Assert(out.String(), Equals, expected)
}

func (s *xgettextTestSuite) TestWriteOutputMultiple(c *C) {
	msgIDs = map[msgKey][]msgData{
		{"", "foo"}: {
			{
				fname:   "fname",
				line:    2,
				comment: "#. comment1\n",
			},
			{
				fname:   "fname",
				line:    4,
				comment: "#. comment2\n",
			},
		},
	}
	out := bytes.NewBuffer([]byte(""))
	writePotFile(out)

	expected := fmt.Sprintf(`%s
#. comment1
#. comment2
#: fname:2 fname:4
msgid   "foo"
msgstr  ""

`, header)
	c.Assert(out.String(), Equals, expected)
}

func (s *xgettextTestSuite) TestWriteOutputNoComment(c *C) {
	msgIDs = map[msgKey][]msgData{
		{"", "foo"}: {
			{
				fname: "fname",
				line:  2,
			},
		},
	}
	out := bytes.NewBuffer([]byte(""))
	writePotFile(out)

	expected := fmt.Sprintf(`%s
#: fname:2
msgid   "foo"
msgstr  ""

`, header)
	c.Assert(out.String(), Equals, expected)
}

func (s *xgettextTestSuite) TestWriteOutputNoLocation(c *C) {
	msgIDs = map[msgKey][]msgData{
		{"", "foo"}: {
			{
				fname: "fname",
				line:  2,
			},
		},
	}

	*noLocation = true
	out := bytes.NewBuffer([]byte(""))
	writePotFile(out)

	expected := fmt.Sprintf(`%s
msgid   "foo"
msgstr  ""

`, header)
	c.Assert(out.String(), Equals, expected)
}

func (s *xgettextTestSuite) TestWriteOutputFormatHint(c *C) {
	msgIDs = map[msgKey][]msgData{
		{"", "foo"}: {
			{
				fname:      "fname",
				line:       2,
				formatHint: "c-format",
			},
		},
	}

	out := bytes.NewBuffer([]byte(""))
	writePotFile(out)

	expected := fmt.Sprintf(`%s
#: fname:2
#, c-format
msgid   "foo"
msgstr  ""

`, header)
	c.Assert(out.String(), Equals, expected)
}

func (s *xgettextTestSuite) TestWriteOutputPlural(c *C) {
	msgIDs = map[msgKey][]msgData{
		{"", "foo"}: {
			{
				msgidPlural: "plural",
				fname:       "fname",
				line:        2,
			},
		},
	}

	out := bytes.NewBuffer([]byte(""))
	writePotFile(out)

	expected := fmt.Sprintf(`%s
#: fname:2
msgid   "foo"
msgid_plural   "plural"
msgstr[0]  ""
msgstr[1]  ""

`, header)
	c.Assert(out.String(), Equals, expected)
}

func (s *xgettextTestSuite) TestWriteOutputSorted(c *C) {
	msgIDs = map[msgKey][]msgData{
		{"", "aaa"}: {
			{
				fname: "fname",
				line:  2,
			},
		},
		{"", "zzz"}: {
			{
				fname: "fname",
				line:  2,
			},
		},
	}

	*sortOutput = true
	// we need to run this a bunch of times as the ordering might
	// be right by pure chance
	for i := 0; i < 10; i++ {
		out := bytes.NewBuffer([]byte(""))
		writePotFile(out)

		expected := fmt.Sprintf(`%s
#: fname:2
msgid   "aaa"
msgstr  ""

#: fname:2
msgid   "zzz"
msgstr  ""

`, header)
		c.Assert(out.String(), Equals, expected)
	}
}

func (s *xgettextTestSuite) TestIntegration(c *C) {
	fname := makeGoSourceFile(c, []byte(`package main

func main() {
    // TRANSLATORS: foo comment
    //              with multiple lines
    i18n.G("foo")

    // this comment has no translators tag
    i18n.G("abc")

    // TRANSLATORS: plural
    i18n.NG("singular", "plural", 99)

    i18n.G("zz %s")
}
`))

	// a real integration test :)
	outName := filepath.Join(c.MkDir(), "snappy.pot")
	os.Args = []string{"test-binary",
		"--output", outName,
		"--keyword", "i18n.G",
		"--keyword-plural", "i18n.NG",
		"--msgid-bugs-address", "snappy-devel@lists.ubuntu.com",
		"--package-name", "snappy",
		fname,
	}
	main()

	// verify its what we expect
	got, err := os.ReadFile(outName)
	c.Assert(err, IsNil)
	expected := fmt.Sprintf(`%s
#: %[2]s:9
msgid   "abc"
msgstr  ""

#. TRANSLATORS: foo comment
#. with multiple lines
#: %[2]s:6
msgid   "foo"
msgstr  ""

#. TRANSLATORS: plural
#: %[2]s:12
msgid   "singular"
msgid_plural   "plural"
msgstr[0]  ""
msgstr[1]  ""

#: %[2]s:14
msgid   "zz %%s"
msgstr  ""

`, header, fname)
	c.Assert(string(got), Equals, expected)
}

func (s *xgettextTestSuite) TestProcessFilesConcat(c *C) {
	fname := makeGoSourceFile(c, []byte(`package main

func main() {
    // TRANSLATORS: foo comment
    i18n.G("foo\n" + "bar\n" + "baz")
}
`))
	err := processFiles([]string{fname})
	c.Assert(err, IsNil)

	c.Assert(msgIDs, DeepEquals, map[msgKey][]msgData{
		{"", "foo\\nbar\\nbaz"}: {
			{
				comment: "#. TRANSLATORS: foo comment\n",
				fname:   fname,
				line:    5,
			},
		},
	})
}

func (s *xgettextTestSuite) TestProcessFilesWithQuote(c *C) {
	fname := makeGoSourceFile(c, []byte(fmt.Sprintf(`package main

func main() {
    i18n.G(%[1]s foo "bar"%[1]s)
}
`, "`")))
	err := processFiles([]string{fname})
	c.Assert(err, IsNil)

	out := bytes.NewBuffer([]byte(""))
	writePotFile(out)

	expected := fmt.Sprintf(`%s
#: %[2]s:4
msgid   " foo \"bar\""
msgstr  ""

`, header, fname)
	c.Check(out.String(), Equals, expected)

}

func (s *xgettextTestSuite) TestWriteOutputMultilines(c *C) {
	msgIDs = map[msgKey][]msgData{
		{"", "foo\\nbar\\nbaz"}: {
			{
				fname:   "fname",
				line:    2,
				comment: "#. foo\n",
			},
		},
	}
	out := bytes.NewBuffer([]byte(""))
	writePotFile(out)
	expected := fmt.Sprintf(`%s
#. foo
#: fname:2
msgid   "foo\n"
        "bar\n"
        "baz"
msgstr  ""

`, header)
	c.Assert(out.String(), Equals, expected)
}

func (s *xgettextTestSuite) TestWriteOutputTidy(c *C) {
	msgIDs = map[msgKey][]msgData{
		{"", "foo\\nbar\\nbaz"}: {
			{
				fname: "fname",
				line:  2,
			},
		},
		{"", "zzz\\n"}: {
			{
				fname: "fname",
				line:  4,
			},
		},
	}
	out := bytes.NewBuffer([]byte(""))
	writePotFile(out)
	expected := fmt.Sprintf(`%s
#: fname:2
msgid   "foo\n"
        "bar\n"
        "baz"
msgstr  ""

#: fname:4
msgid   "zzz\n"
msgstr  ""

`, header)
	c.Assert(out.String(), Equals, expected)
}

func (s *xgettextTestSuite) TestProcessFilesWithDoubleQuote(c *C) {
	fname := makeGoSourceFile(c, []byte(`package main

func main() {
    i18n.G("foo \"bar\"")
}
`))
	err := processFiles([]string{fname})
	c.Assert(err, IsNil)

	out := bytes.NewBuffer([]byte(""))
	writePotFile(out)

	expected := fmt.Sprintf(`%s
#: %[2]s:4
msgid   "foo \"bar\""
msgstr  ""

`, header, fname)
	c.Check(out.String(), Equals, expected)

}

func (s *xgettextTestSuite) TestSkipArgs(c *C) {
	fname := makeGoSourceFile(c, []byte(`package main

func main() {
    i18n.G("arg-to-skip", "foo")
}
`))
	*skipArgs = 1
	err := processFiles([]string{fname})
	c.Assert(err, IsNil)

	out := bytes.NewBuffer([]byte(""))
	writePotFile(out)

	expected := fmt.Sprintf(`%s
#: %[2]s:4
msgid   "foo"
msgstr  ""

`, header, fname)
	c.Check(out.String(), Equals, expected)

}

func (s *xgettextTestSuite) TestMsgCtxt(c *C) {
	fname := makeGoSourceFile(c, []byte(`package main

func main() {
    i18n.CG("ctx1", "foo")
}
`))
	err := processFiles([]string{fname})
	c.Assert(err, IsNil)

	out := bytes.NewBuffer([]byte(""))
	writePotFile(out)

	expected := fmt.Sprintf(`%s
#: %[2]s:4
msgctxt "ctx1"
msgid   "foo"
msgstr  ""

`, header, fname)
	c.Check(out.String(), Equals, expected)

}
