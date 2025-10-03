package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/format"
	"go/types"
	"os"
	"path"
	"path/filepath"

	"golang.org/x/tools/go/packages"
)

const version = "v0.0.1"

type methodMeta struct {
	Name     string
	ReqType  string
	RespType string
}

type importMeta struct {
	Name string
	Path string
}

func (i importMeta) ImportString() string {
	if _, nameFromPath := path.Split(i.Path); nameFromPath == i.Name || i.Name == "" {
		return fmt.Sprintf("\"%s\"", i.Path)
	}

	return fmt.Sprintf("%s \"%s\"", i.Name, i.Path)
}

func main() {
	target := flag.String("target", "", "name of interface to generate for (required)")
	flag.Parse()

	if *target == "" {
		failf("missing --target")
	}

	outDir := getOutDir()
	pkg := loadPackage(outDir)
	iface := loadTargetInterface(pkg, *target)

	methods, imports := collectMethods(pkg, iface)

	generateFiles(pkg.Name, *target, outDir, methods, imports)
}

func getOutDir() string {
	gofile := os.Getenv("GOFILE")
	if gofile == "" {
		return "."
	}
	if dir := filepath.Dir(gofile); dir != "" {
		return dir
	}
	return "."
}

func loadPackage(outDir string) *packages.Package {
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedImports,
		Dir:  outDir,
	}

	pkgs, err := packages.Load(cfg, ".")
	if err != nil {
		failf("loading package: %v", err)
	}
	if packages.PrintErrors(pkgs) > 0 || len(pkgs) == 0 {
		failf("failed to load package")
	}
	return pkgs[0]
}

func loadTargetInterface(pkg *packages.Package, target string) *types.Interface {
	obj := pkg.Types.Scope().Lookup(target)
	if obj == nil {
		failf("interface %q not found in package %s", target, pkg.Types.Name())
	}

	iface, ok := obj.Type().Underlying().(*types.Interface)
	if !ok {
		failf("%q is not an interface", target)
	}

	if iface.NumEmbeddeds() > 0 {
		failf("interface %q has embedded interfaces; not supported", target)
	}
	return iface
}

func collectMethods(pkg *packages.Package, iface *types.Interface) ([]methodMeta, []importMeta) {
	qualifier := func(other *types.Package) string {
		if other == nil || other.Path() == pkg.Types.Path() {
			return ""
		}
		return other.Name()
	}

	imports := map[string]string{}
	var methods []methodMeta

	for i := 0; i < iface.NumMethods(); i++ {
		m := iface.Method(i)
		sig, ok := m.Type().(*types.Signature)
		if !ok {
			failf("method %s has no signature", m.Name())
		}

		validateParams(m, sig, qualifier, pkg)
		methods = append(methods, buildMethodMeta(m, sig, qualifier, pkg, imports))
	}

	var importMetas []importMeta
	for name, path := range imports {
		importMetas = append(importMetas, importMeta{
			Name: name,
			Path: path,
		})
	}
	return methods, importMetas
}

func validateParams(m *types.Func, sig *types.Signature, qualifier func(*types.Package) string, pkg *packages.Package) {
	params := sig.Params()
	if params.Len() != 2 {
		failf("method %s: expected 2 parameters, got %d", m.Name(), params.Len())
	}

	if params.At(0).Type().String() != "context.Context" {
		failf("method %s: first parameter must be context.Context", m.Name())
	}

	results := sig.Results()
	if results.Len() != 2 {
		failf("method %s: expected 2 results, got %d", m.Name(), results.Len())
	}

	if results.At(1).Type().String() != "error" {
		failf("method %s: second result must be error", m.Name())
	}
}

func buildMethodMeta(m *types.Func, sig *types.Signature,
	qualifier func(*types.Package) string, pkg *packages.Package,
	imports map[string]string,
) methodMeta {
	addImportIfExternal(sig.Params().At(1).Type(), pkg, qualifier, imports)
	addImportIfExternal(sig.Results().At(0).Type(), pkg, qualifier, imports)

	return methodMeta{
		Name:     m.Name(),
		ReqType:  types.TypeString(sig.Params().At(1).Type(), qualifier),
		RespType: types.TypeString(sig.Results().At(0).Type(), qualifier),
	}
}

func addImportIfExternal(t types.Type, pkg *packages.Package,
	qualifier func(*types.Package) string, imports map[string]string,
) {
	if named, ok := t.(*types.Named); ok {
		if typePkg := named.Obj().Pkg(); typePkg != nil && typePkg.Path() != pkg.Types.Path() {
			imports[typePkg.Name()] = typePkg.Path()
		}
	}
}

func generateFiles(pkgName, target, outDir string, methods []methodMeta, imports []importMeta) {
	clientFile := filepath.Join(outDir, "client.srpc.go")
	serverFile := filepath.Join(outDir, "server.srpc.go")

	if tryGenerateFile(clientFile, func() ([]byte, error) {
		return generateClient(pkgName, target, methods, imports)
	}) {
		fmt.Printf("wrote %s\n", clientFile)
	} else {
		fmt.Printf("skipping %s (already exists)\n", clientFile)
	}

	if tryGenerateFile(serverFile, func() ([]byte, error) {
		return generateServer(pkgName, target, methods, imports)
	}) {
		fmt.Printf("wrote %s\n", serverFile)
	} else {
		fmt.Printf("skipping %s (already exists)\n", serverFile)
	}
}

func tryGenerateFile(path string, gen func() ([]byte, error)) bool {
	if fileExists(path) {
		return false
	}
	src, err := gen()
	if err != nil {
		failf("generate: %v", err)
	}
	if err := writeFormattedFile(path, src); err != nil {
		failf("writing file: %v", err)
	}
	return true
}

func generateImports(b *bytes.Buffer, imports ...importMeta) {
	fmt.Fprintf(b, "import (\n")
	for _, i := range imports {
		fmt.Fprintf(b, "%s\n", i.ImportString())
	}
	fmt.Fprintf(b, ")\n\n")
}

func generateClient(pkgName, target string, methods []methodMeta, imports []importMeta) ([]byte, error) {
	var b bytes.Buffer

	fmt.Fprintf(&b, "// Code generated by srpc-gen %s. DO NOT EDIT.\n\n", version)
	fmt.Fprintf(&b, "package %s\n\n", pkgName)

	imports = append(imports, importPath("context"), importPath("github.com/tymbaca/srpc"))
	generateImports(&b, imports...)

	fmt.Fprintf(&b, "func New%sClient(client *srpc.Client) *%sClient {\n\treturn &%sClient{client: client}\n}\n\n",
		target, target, target)

	fmt.Fprintf(&b, "type %sClient struct {\n\tclient *srpc.Client\n}\n\n", target)

	for _, m := range methods {
		fmt.Fprintf(&b, "func (c *%sClient) %s(ctx context.Context, req %s) (resp %s, err error) {\n",
			target, m.Name, m.ReqType, m.RespType)
		fmt.Fprintf(&b, "\terr = c.client.Call(ctx, \"%s.%s\", req, &resp)\n", target, m.Name)
		fmt.Fprintf(&b, "\treturn resp, err\n}\n\n")
	}
	return b.Bytes(), nil
}

func importPath(s string) importMeta {
	return importMeta{Path: s}
}

func generateServer(pkgName, target string, methods []methodMeta, imports []importMeta) ([]byte, error) {
	var b bytes.Buffer

	fmt.Fprintf(&b, "// Code generated by srpc-gen %s. Edit for your needs.\n\n", version)
	fmt.Fprintf(&b, "package %s\n\n", pkgName)

	imports = append(imports, importPath("context"))
	generateImports(&b, imports...)

	fmt.Fprintf(&b, "type %sServer struct {\n\t// TODO: fill\n}\n\n", target)

	for _, m := range methods {
		fmt.Fprintf(&b, "func (s *%sServer) %s(ctx context.Context, req %s) (%s, error) {\n",
			target, m.Name, m.ReqType, m.RespType)
		fmt.Fprintf(&b, "\tpanic(\"not implemented\") // TODO: Implement\n}\n\n")
	}
	return b.Bytes(), nil
}

func writeFormattedFile(path string, src []byte) error {
	fmtSrc, err := format.Source(src)
	if err != nil {
		return fmt.Errorf("format.Source failed: %w\nunformatted source:\n%s", err, string(src))
	}
	return os.WriteFile(path, fmtSrc, 0o644)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func failf(formatStr string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, formatStr+"\n", args...)
	os.Exit(1)
}
