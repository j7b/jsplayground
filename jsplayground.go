// +build js

package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/scanner"
	"go/token"
	"go/types"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/gopherjs/gopherjs/compiler"
	"github.com/gopherjs/gopherjs/js"
	"github.com/j7b/jsplayground/important"
)

var errAgain = fmt.Errorf("try again")

type formatter struct {
	code    []byte
	imports bool
}

func (f *formatter) format(resolve, reject func(interface{})) {
	var out []byte
	var err error
	switch f.imports {
	case true:
		out, err = important.Process(f.code)
	case false:
		out, err = format.Source(f.code)
	}
	if err == nil {
		resolve(string(out))
		return
	}
	reject(err.Error())
}

func promise(f func(resolve, reject func(interface{}))) *js.Object {
	return js.Global.Get("Promise").New(f)
}

type Go struct {
	packages      map[string]*compiler.Archive
	packagerr     map[string]error
	importContext *compiler.ImportContext
	code          []byte
	packageuri    string
}

func (g *Go) loadpkg(path string) {
	if g.packagerr == nil {
		g.packagerr = make(map[string]error)
	}
	if _, ok := g.packages[path]; ok {
		return
	}
	if _, ok := g.packagerr[path]; ok {
		return
	}
	res, err := http.Get(g.packageuri + "pkg/" + path + ".a")
	if err != nil {
		g.packagerr[path] = err
		return
	}
	if res.StatusCode != http.StatusOK {
		g.packagerr[path] = fmt.Errorf(res.Status)
		return
	}
	defer res.Body.Close()
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		g.packagerr[path] = err
		return
	}
	p, err := compiler.ReadArchive(path+".a", path, bytes.NewReader(b), g.importContext.Packages)
	if err != nil {
		g.packagerr[path] = err
		return
	}
	g.packages[path] = p
}

func (g *Go) PackageURI(uri string) {
	g.packageuri = uri
}

func (g *Go) RedirectConsole(f func(string)) {
	js.Global.Set("goPrintToConsole", js.InternalObject(func(b []byte) {
		f(string(b))
	}))
}

func (g *Go) Compile(src string) *js.Object {
	g.code = []byte(src)
	return promise(g.compile)
}

func (g *Go) compile(resolve, reject func(interface{})) {
	go func() {
		fileSet := token.NewFileSet()
		defer func() {
			if r := recover(); r != nil {
				reject(fmt.Sprintf("PANIC: %#v", r))
			}
		}()
		file, err := parser.ParseFile(fileSet, "prog.go", g.code, parser.ParseComments)
		if err != nil {
			if list, ok := err.(scanner.ErrorList); ok {
				errors := make([]string, len(list))
				for i, entry := range list {
					errors[i] = entry.Error()
				}
				reject(strings.Join(errors, "\n"))
				return
			}
			reject(err.Error())
			return
		}
		mainPkg, err := compiler.Compile("main", []*ast.File{file}, fileSet, g.importContext, false)
		g.packages["main"] = mainPkg
		if err != nil {
			if list, ok := err.(compiler.ErrorList); ok {
				errors := make([]string, len(list))
				for i, entry := range list {
					errors[i] = entry.Error()
				}
				reject(strings.Join(errors, "\n"))
				return
			}
			reject(err.Error())
			return
		}
		var allPkgs []*compiler.Archive
		allPkgs, err = compiler.ImportDependencies(mainPkg, g.importContext.Import)
		if err != nil {
			reject(err.Error())
			return
		}
		jsCode := new(bytes.Buffer)
		compiler.WriteProgramCode(allPkgs, &compiler.SourceMapFilter{Writer: jsCode})
		resolve(jsCode.String())
	}()
}

func (g *Go) Format(src string, imports bool) *js.Object {
	code := []byte(src)
	f := &formatter{code: code, imports: imports}
	return promise(f.format)
}

func imports() {
	if err := important.Imports(); err != nil {
		js.Global.Get("console").Call("warn", "additional imports: "+err.Error())
	}
}

func main() {
	go imports()
	g := new(Go)
	g.packages = make(map[string]*compiler.Archive)
	g.importContext = &compiler.ImportContext{
		Packages: make(map[string]*types.Package),
		Import: func(path string) (a *compiler.Archive, err error) {
			if pkg, found := g.packages[path]; found {
				return pkg, nil
			}
			if err, found := g.packagerr[path]; found {
				return nil, err
			}
			res, err := http.Get("pkg/" + path + ".a")
			if err != nil {
				g.packagerr[path] = err
				return nil, err
			}
			defer res.Body.Close()
			if res.StatusCode != http.StatusOK {
				g.packagerr[path] = fmt.Errorf(res.Status)
				return nil, err
			}
			b, err := ioutil.ReadAll(res.Body)
			if err != nil {
				g.packagerr[path] = err
				return nil, err
			}
			p, err := compiler.ReadArchive(path+".a", path, bytes.NewReader(b), g.importContext.Packages)
			if err != nil {
				g.packagerr[path] = err
				return nil, err
			}
			g.packages[path] = p
			return p, nil
		},
	}
	js.Global.Set("Go", js.MakeWrapper(g))
}
