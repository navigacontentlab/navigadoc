package navigadoc_test

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Infomaker/etree"
	"github.com/navigacontentlab/navigadoc/doc"
)

type TestData struct {
	xml          string
	json         string
	root         string
	customConfig string
	typeConfig   string
	expectError  bool
}

func safeClose(t *testing.T, name string, c io.Closer) {
	t.Helper()

	if err := c.Close(); err != nil {
		t.Errorf("failed to close %s: %v", name, err)
	}
}

func stripchars(str, chr string) string {
	return strings.Map(func(r rune) rune {
		if !strings.ContainsRune(chr, r) {
			return r
		}
		return -1
	}, str)
}

func ensureDebugDir(t *testing.T) string {
	t.Helper()

	outDir := filepath.Join("testdebug", t.Name())
	must(t, os.MkdirAll(outDir, 0770), "failed to create test debug output directory")

	return outDir
}

func dumpStructComparison(t *testing.T, diff string) {
	t.Helper()

	outDir := ensureDebugDir(t)
	diffPath := filepath.Join(outDir, "struct-diff.txt")

	err := ioutil.WriteFile(diffPath, []byte(diff), 0600)
	must(t, err, "failed to create struct diff output file")

	t.Logf("struct diff: %s", diffPath)
}

func getCDATAString(e *etree.Element) string {
	if len(e.Child) == 0 {
		return ""
	}

	buf := strings.Builder{}
	for i := range e.Child {
		switch c := e.Child[i].(type) {
		case *etree.CharData:
			if c.IsWhitespace() {
				continue
			}
			buf.WriteString(c.Data)
		case *etree.Element:
			return buf.String()
		}
	}

	return buf.String()
}

func dumpNavigaDoc(t *testing.T, document *doc.Document, docPath string) {
	docFile, err := os.Create(docPath)
	must(t, err, "failed to create navigadoc output file")

	defer func(t *testing.T, c io.Closer) {
		err := c.Close()
		if err != nil {
			t.Logf("failed to close navigadoc.json: %s", err.Error())
		}
	}(t, docFile)

	enc := json.NewEncoder(docFile)

	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")

	must(t, enc.Encode(document), "failed to write navigadoc output to file")
}
