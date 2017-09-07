package travail

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path"
	"strings"
)

func filtertest(fi os.FileInfo) bool {
	if strings.Contains(fi.Name(), "_test") {
		return false
	}
	return true
}

func Map(ipath, dir string) (map[string]string, error) {
	fset := new(token.FileSet)
	pkgs, err := parser.ParseDir(fset, dir, filtertest, parser.AllErrors)
	if err != nil {
		return nil, err
	}
	base := path.Base(dir)
	pkg := pkgs[base]
	if pkg == nil {
		return nil, fmt.Errorf("map: package %s not found", base)
	}
	ast.PackageExports(pkg)
	m := make(map[string]string)
	for _, f := range pkg.Files {
		for _, d := range f.Decls {
			switch t := d.(type) {
			case *ast.GenDecl:
				for _, s := range t.Specs {
					switch t := s.(type) {
					case *ast.TypeSpec:
						// log.Println(t.Name)
						m[fmt.Sprintf("%s.%s", base, t.Name)] = ipath
					case *ast.ValueSpec:
						for _, v := range t.Names {
							//log.Println(v)
							m[fmt.Sprintf("%s.%s", base, v)] = ipath
						}
					}
				}
			case *ast.FuncDecl:
				//log.Println(t.Name)
				m[fmt.Sprintf("%s.%s", base, t.Name)] = ipath
			}
		}
	}
	return m, nil
}

/*
func main() {
	context := build.Default
	_, err := context.Import("github.com/hajimehoshi/ebiten", "/Volumes/Things/Playground/src/github.com/hajimehoshi/ebiten", build.IgnoreVendor)
	if err != nil {
		log.Fatal(err)
	}
	pkgs, err := parser.ParseDir(fset, "/Volumes/Things/Playground/src/github.com/hajimehoshi/ebiten", nil, parser.AllErrors)
	if err != nil {
		log.Fatal(err)
	}
	pkg := pkgs["ebiten"]
	ast.PackageExports(pkg)
	for _, f := range pkg.Files {
		for _, d := range f.Decls {
			switch t := d.(type) {
			case *ast.GenDecl:
				for _, s := range t.Specs {
					switch t := s.(type) {
					case *ast.TypeSpec:
						log.Println(t.Name)
					case *ast.ValueSpec:
						for _, v := range t.Names {
							log.Println(v)
						}
					}
				}
			case *ast.FuncDecl:
				log.Println(t.Name)
			}
		}
	}
}
*/
