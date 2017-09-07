package important

import (
	"bytes"
	"go/parser"
	"go/printer"
	"go/token"
	"testing"
)

const testfile1 = `

package main

import (
	"fmt"
)

func main() {
	println("Hello World")
}

`

const testfile2 = `
package main

func main() {
	fmt.Println("Hello World")
}
`

var fset = new(token.FileSet)

func TestRemove(t *testing.T) {
	f, err := parser.ParseFile(fset, "test1.go", testfile1, parser.AllErrors)
	if err != nil {
		t.Fatal(err)
	}
	_, err = FixImports(fset, f)
	if err != nil {
		t.Fatal(err)
	}
	buf := new(bytes.Buffer)
	err = printer.Fprint(buf, fset, f)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(buf.String())
}

func TestAdd(t *testing.T) {
	f, err := parser.ParseFile(fset, "test2.go", testfile2, parser.AllErrors)
	if err != nil {
		t.Fatal(err)
	}
	_, err = FixImports(fset, f)
	if err != nil {
		t.Fatal(err)
	}
	buf := new(bytes.Buffer)
	err = printer.Fprint(buf, fset, f)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(buf.String())
}
