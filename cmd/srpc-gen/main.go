package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/format"
	"go/types"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"text/template"

	_ "embed"

	"golang.org/x/tools/go/packages"
)

const version = "v0.0.1"

// ---------------- Data Models ----------------

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

type fileData struct {
	Version string
	PkgName string
	Target  string
	Imports []importMeta
	Methods []methodMeta
}

// ---------------- Templates ----------------

//go:embed srpc.client.go.tmpl
var clientTmpl string

//go:embed srpc.server.go.tmpl
var serverTmpl string

// ---------------- Main ----------------

func main() {
	target := flag.String("target", "", "name of interface to generate for (required)")
	only := flag.String("only", "", "generate only provided part: [client | server] (optional)")
	clientOut := flag.String("client-out", "", "client filename (optional)")
	serverOut := flag.String("server-out", "", "server filename (optional)")
	flag.Parse()

	if *target == "" {
		failf("missing --target")
	}

	outDir := getOutDir()
	pkg := loadPackage(outDir)
	iface := loadTargetInterface(pkg, *target)

	methods, imports := collectMethods(pkg, iface)

	generateFiles(pkg.Name, *target, outDir, methods, imports, *only, *clientOut, *serverOut)
}

// ---------------- Package Loading ----------------

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

// ---------------- Generation ----------------

func generateFiles(pkgName, target, outDir string, methods []methodMeta, imports []importMeta, only string, clientOut, serverOut string) {
	if clientOut == "" {
		clientOut = fmt.Sprintf("srpc.%s.client.go", target)
	}
	if serverOut == "" {
		serverOut = fmt.Sprintf("srpc.%s.server.go", target)
	}
	clientFile := filepath.Join(outDir, clientOut)
	serverFile := filepath.Join(outDir, serverOut)

	if only == "" || only == "client" {
		if tryGenerateFile(clientFile, func() ([]byte, error) {
			return generateClient(pkgName, target, methods, imports)
		}) {
			slog.Info("generated client", "filename", clientFile)
		} else {
			slog.Info("skipping client (already exists)", "filename", clientFile)
		}
	}

	if only == "" || only == "server" {
		if tryGenerateFile(serverFile, func() ([]byte, error) {
			return generateServer(pkgName, target, methods, imports)
		}) {
			slog.Info("generated server", "filename", serverFile)
		} else {
			slog.Info("skipping server (already exists)", "filename", serverFile)
		}
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

func generateClient(pkgName, target string, methods []methodMeta, imports []importMeta) ([]byte, error) {
	data := fileData{
		Version: version,
		PkgName: pkgName,
		Target:  target,
		Imports: imports,
		Methods: methods,
	}
	return renderTemplate(clientTmpl, data)
}

func generateServer(pkgName, target string, methods []methodMeta, imports []importMeta) ([]byte, error) {
	data := fileData{
		Version: version,
		PkgName: pkgName,
		Target:  target,
		Imports: imports,
		Methods: methods,
	}
	return renderTemplate(serverTmpl, data)
}

func renderTemplate(tmplSrc string, data fileData) ([]byte, error) {
	tmpl, err := template.New("gen").Parse(tmplSrc)
	if err != nil {
		return nil, err
	}
	var b bytes.Buffer
	if err := tmpl.Execute(&b, data); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

// ---------------- Reflection ----------------

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

		validateParams(m, sig)
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

func validateParams(m *types.Func, sig *types.Signature) {
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
	addImportIfExternal(sig.Params().At(1).Type(), pkg, imports)
	addImportIfExternal(sig.Results().At(0).Type(), pkg, imports)

	return methodMeta{
		Name:     m.Name(),
		ReqType:  types.TypeString(sig.Params().At(1).Type(), qualifier),
		RespType: types.TypeString(sig.Results().At(0).Type(), qualifier),
	}
}

func addImportIfExternal(t types.Type, pkg *packages.Package, imports map[string]string,
) {
	if named, ok := t.(*types.Named); ok {
		if typePkg := named.Obj().Pkg(); typePkg != nil && typePkg.Path() != pkg.Types.Path() {
			imports[typePkg.Name()] = typePkg.Path()
		}
	}
}

// ---------------- Utils ----------------

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

func failf(formatStr string, args ...any) {
	slog.Error(fmt.Sprintf(formatStr+"\n", args...))
	os.Exit(1)
}
