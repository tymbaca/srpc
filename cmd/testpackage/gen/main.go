package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
)

func main() {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "./cmd/testpackage/service.go", nil, parser.SkipObjectResolution)
	assertNil(err)

	ast.Inspect(f, func(none ast.Node) bool {
		switch n := none.(type) {
		case *ast.TypeSpec:
			switch iface := n.Type.(type) {
			case *ast.InterfaceType:
				fmt.Printf("n.Name: %v\n", n.Name)
				inspectInterface(iface)
			}
		}
		return true
	})
}

func inspectInterface(iface *ast.InterfaceType) {
	fmt.Printf("expr: %#v\n", iface)
	for _, method := range iface.Methods.List {
		fmt.Printf("method.Names: %v\n", method.Names)
		fmt.Printf("method.Type: %#v\n", method.Type)
		methodFunc := method.Type.(*ast.FuncType)

		for _, in := range methodFunc.Params.List {
			fmt.Printf("in.Names: %v\n", in.Names)
			fmt.Printf("in.Type: %#v\n", in.Type)
		}
	}
}

func assertNil(v any) {
	if v != nil {
		log.Panicf("%s", v)
	}
}

func assert(cond bool) {
	if !cond {
		panic("assertion failure")
	}
}
