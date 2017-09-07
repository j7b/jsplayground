package important

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"path"
	"strings"

	"golang.org/x/tools/go/ast/astutil"
)

// importPathToNameBasic assumes the package name is the base of import path.
func importPathToName(importPath string) (packageName string) {
	return path.Base(importPath)
}

func Process(code []byte) ([]byte, error) {
	fset := new(token.FileSet)
	f, err := parser.ParseFile(fset, "prog.go", code, parser.ParseComments|parser.AllErrors)
	if err != nil {
		return nil, err
	}
	if _, err = FixImports(fset, f); err != nil {
		return nil, err
	}
	buf := new(bytes.Buffer)
	if err = format.Node(buf, fset, f); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func FixImports(fset *token.FileSet, f *ast.File) (added []string, err error) {
	v := new(Visitor)
	ast.Walk(v, f)
	unusedImport := map[string]bool{}
	for pkg, is := range v.decls {
		if v.refs[pkg] == nil && pkg != "_" && pkg != "." {
			unusedImport[strings.Trim(is.Path.Value, `"`)] = true
		}
	}
	for ipath := range unusedImport {
		if ipath == "C" {
			// Don't remove cgo stuff.
			continue
		}
		astutil.DeleteImport(fset, f, ipath)
	}
	for pkgName, symbols := range v.refs {
		if len(symbols) == 0 {
			continue // skip over packages already imported
		}
		ipath, rename, err := findImport(pkgName, symbols)
		if err != nil {
			return nil, err
		}
		switch rename {
		case true:
			astutil.AddNamedImport(fset, f, pkgName, ipath)
		default:
			astutil.AddImport(fset, f, ipath)
		}
		added = append(added, ipath)
	}
	return
}

type Visitor struct {
	refs  map[string]map[string]bool
	decls map[string]*ast.ImportSpec
}

func (v *Visitor) Visit(node ast.Node) (w ast.Visitor) {
	if v.refs == nil {
		v.refs = make(map[string]map[string]bool)
	}
	if v.decls == nil {
		v.decls = make(map[string]*ast.ImportSpec)
	}
	if node == nil {
		return v
	}
	switch t := node.(type) {
	case *ast.ImportSpec:
		if t.Name != nil {
			v.decls[t.Name.Name] = t
		} else {
			local := importPathToName(strings.Trim(t.Path.Value, `\"`))
			v.decls[local] = t
		}
	case *ast.SelectorExpr:
		xident, ok := t.X.(*ast.Ident)
		if !ok {
			break
		}
		if xident.Obj != nil {
			// if the parser can resolve it, it's not a package ref
			break
		}
		pkgName := xident.Name
		if v.refs[pkgName] == nil {
			v.refs[pkgName] = make(map[string]bool)
		}
		if v.decls[pkgName] == nil {
			v.refs[pkgName][t.Sel.Name] = true
		}
	}
	return v
}

func findImport(shortPkg string, symbols map[string]bool) (importPath string, rename bool, err error) {
	for symbol := range symbols {
		path := stdlib[shortPkg+"."+symbol]
		if path == "" {
			return "", false, nil
		}
		if importPath != "" && importPath != path {
			// Ambiguous. Symbols pointed to different things.
			return "", false, nil
		}
		importPath = path
	}
	return importPath, false, nil
}
