package navigadoc_test

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"bitbucket.org/infomaker/doc-format/v2/doc"
	"github.com/Infomaker/etree"
)

const configFile = "testdata/custom-config.json"
const configWithMetaFile = "testdata/custom-config-use-meta.json"

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

func checkNamespaces(t *testing.T, testFile string, xmlDocBytes []byte) {
	t.Helper()

	xmlDoc := etree.NewDocument()
	err := xmlDoc.ReadFromBytes(xmlDocBytes)
	must(t, err, "failed etree.ReadFromBytes")

	xmlns := xmlDoc.Root().SelectAttr("xmlns")
	if xmlns == nil {
		t.Errorf(fmt.Sprintf("generated root is missing xmlns: %s", testFile))
	} else if strings.Compare(xmlns.Value, "http://iptc.org/std/nar/2006-10-01/") != 0 {
		t.Errorf(fmt.Sprintf("generated root does not have expected xmlns: %s", testFile))
	}

	allElements := xmlDoc.FindElements(fmt.Sprintf("/%s//", xmlDoc.Root().Tag))
	for _, element := range allElements {
		switch element.Tag {
		case "links":
			fallthrough
		case "metadata":
			parent := element.Parent().Tag
			if parent == "itemMeta" || parent == "contentMeta" {
				if xmlns := element.SelectAttr("xmlns"); xmlns == nil || xmlns.Value == "" {
					t.Errorf("%s missing namespace: %s", element.Tag, element.GetPath())
				}
			}
		case "idf":
			if xmlns := element.SelectAttr("xmlns"); xmlns == nil || xmlns.Value == "" {
				t.Errorf("idf is missing namespace: %s", element.GetPath())
			}
		}
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

func dumpDocs(t *testing.T, document *doc.Document, generatedDoc *doc.Document, originalXML []byte, generatedXML []byte, errors ...error) {
	t.Helper()

	if !t.Failed() && os.Getenv("DUMP_DOCUMENTS") != "true" {
		return
	}

	outDir := ensureDebugDir(t)

	if len(originalXML) > 0 {
		originalPath := filepath.Join(outDir, "original.xml")
		err := ioutil.WriteFile(originalPath, originalXML, 0600)
		must(t, err, "failed to create original XML output file")
		t.Logf("Original XML: %s", originalPath)
	}

	if len(generatedXML) > 0 {
		generatedPath := filepath.Join(outDir, "generated.xml")
		err := ioutil.WriteFile(generatedPath, generatedXML, 0600)
		must(t, err, "failed to create generated XML output file")
		t.Logf("Generated XML: %s", generatedPath)
	}

	if len(errors) > 0 {
		errorsPath := filepath.Join(outDir, "errors.log")

		logFile, err := os.Create(errorsPath)
		must(t, err, "failed to create error log file")
		defer func() {
			must(t, logFile.Close(),
				"failed to close error log file")
		}()

		for _, lErr := range errors {
			if lErr == nil {
				continue
			}

			_, err := fmt.Fprintln(logFile, lErr.Error())
			must(t, err, "failed to write to error log file")
		}
	}

	if document != nil {
		docPath := filepath.Join(outDir, "navigadoc.json")
		dumpNavigaDoc(t, document, docPath)
		t.Logf("NavigaDoc: %s", docPath)
	}

	if generatedDoc != nil {
		docPath := filepath.Join(outDir, "generateddoc.json")
		dumpNavigaDoc(t, generatedDoc, docPath)
		t.Logf("GeneratedDoc: %s", docPath)
	}
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
