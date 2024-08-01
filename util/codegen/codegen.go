// Copyright (c) Tailscale Inc & AUTHORS
// SPDX-License-Identifier: BSD-3-Clause

// Package codegen contains shared utilities for generating code.
package codegen

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"io"
	"os"
	"reflect"
	"strings"

	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/imports"
	"tailscale.com/util/mak"
)

var flagCopyright = flag.Bool("copyright", true, "add Tailscale copyright to generated file headers")

// LoadTypes returns all named types in pkgName, keyed by their type name.
func LoadTypes(buildTags string, pkgName string) (*packages.Package, map[string]*types.Named, error) {
	cfg := &packages.Config{
		Mode:  packages.NeedTypes | packages.NeedTypesInfo | packages.NeedSyntax | packages.NeedName,
		Tests: buildTags == "test",
	}
	if buildTags != "" && !cfg.Tests {
		cfg.BuildFlags = []string{"-tags=" + buildTags}
	}

	pkgs, err := packages.Load(cfg, pkgName)
	if err != nil {
		return nil, nil, err
	}
	if cfg.Tests {
		pkgs = testPackages(pkgs)
	}
	if len(pkgs) != 1 {
		return nil, nil, fmt.Errorf("wrong number of packages: %d", len(pkgs))
	}
	pkg := pkgs[0]
	return pkg, namedTypes(pkg), nil
}

func testPackages(pkgs []*packages.Package) []*packages.Package {
	var testPackages []*packages.Package
	for _, pkg := range pkgs {
		testPackageID := fmt.Sprintf("%[1]s [%[1]s.test]", pkg.PkgPath)
		if pkg.ID == testPackageID {
			testPackages = append(testPackages, pkg)
		}
	}
	return testPackages
}

// HasNoClone reports whether the provided tag has `codegen:noclone`.
func HasNoClone(structTag string) bool {
	val := reflect.StructTag(structTag).Get("codegen")
	for _, v := range strings.Split(val, ",") {
		if v == "noclone" {
			return true
		}
	}
	return false
}

const copyrightHeader = `// Copyright (c) Tailscale Inc & AUTHORS
// SPDX-License-Identifier: BSD-3-Clause

`

const genAndPackageHeader = `// Code generated by %v; DO NOT EDIT.

package %s
`

func NewImportTracker(thisPkg *types.Package) *ImportTracker {
	return &ImportTracker{
		thisPkg: thisPkg,
	}
}

// ImportTracker provides a mechanism to track and build import paths.
type ImportTracker struct {
	thisPkg  *types.Package
	packages map[string]bool
}

func (it *ImportTracker) Import(pkg string) {
	if pkg != "" && !it.packages[pkg] {
		mak.Set(&it.packages, pkg, true)
	}
}

func (it *ImportTracker) qualifier(pkg *types.Package) string {
	if it.thisPkg == pkg {
		return ""
	}
	it.Import(pkg.Path())
	// TODO(maisem): handle conflicts?
	return pkg.Name()
}

// QualifiedName returns the string representation of t in the package.
func (it *ImportTracker) QualifiedName(t types.Type) string {
	return types.TypeString(t, it.qualifier)
}

// PackagePrefix returns the prefix to be used when referencing named objects from pkg.
func (it *ImportTracker) PackagePrefix(pkg *types.Package) string {
	if s := it.qualifier(pkg); s != "" {
		return s + "."
	}
	return ""
}

// Write prints all the tracked imports in a single import block to w.
func (it *ImportTracker) Write(w io.Writer) {
	fmt.Fprintf(w, "import (\n")
	for s := range it.packages {
		fmt.Fprintf(w, "\t%q\n", s)
	}
	fmt.Fprintf(w, ")\n\n")
}

func writeHeader(w io.Writer, tool, pkg string) {
	if *flagCopyright {
		fmt.Fprint(w, copyrightHeader)
	}
	fmt.Fprintf(w, genAndPackageHeader, tool, pkg)
}

// WritePackageFile adds a file with the provided imports and contents to package.
// The tool param is used to identify the tool that generated package file.
func WritePackageFile(tool string, pkg *packages.Package, path string, it *ImportTracker, contents *bytes.Buffer) error {
	buf := new(bytes.Buffer)
	writeHeader(buf, tool, pkg.Name)
	it.Write(buf)
	if _, err := buf.Write(contents.Bytes()); err != nil {
		return err
	}
	return writeFormatted(buf.Bytes(), path)
}

// writeFormatted writes code to path.
// It runs gofmt on it before writing;
// if gofmt fails, it writes code unchanged.
// Errors can include I/O errors and gofmt errors.
//
// The advantage of always writing code to path,
// even if gofmt fails, is that it makes debugging easier.
// The code can be long, but you need it in order to debug.
// It is nicer to work with it in a file than a terminal.
// It is also easier to interpret gofmt errors
// with an editor providing file and line numbers.
func writeFormatted(code []byte, path string) error {
	out, fmterr := imports.Process(path, code, &imports.Options{
		Comments:   true,
		TabIndent:  true,
		TabWidth:   8,
		FormatOnly: true, // fancy gofmt only
	})
	if fmterr != nil {
		out = code
	}
	ioerr := os.WriteFile(path, out, 0644)
	// Prefer I/O errors. They're usually easier to fix,
	// and until they're fixed you can't do much else.
	if ioerr != nil {
		return ioerr
	}
	if fmterr != nil {
		return fmt.Errorf("%s:%v", path, fmterr)
	}
	return nil
}

// namedTypes returns all named types in pkg, keyed by their type name.
func namedTypes(pkg *packages.Package) map[string]*types.Named {
	nt := make(map[string]*types.Named)
	for _, file := range pkg.Syntax {
		for _, d := range file.Decls {
			decl, ok := d.(*ast.GenDecl)
			if !ok || decl.Tok != token.TYPE {
				continue
			}
			for _, s := range decl.Specs {
				spec, ok := s.(*ast.TypeSpec)
				if !ok {
					continue
				}
				typeNameObj, ok := pkg.TypesInfo.Defs[spec.Name]
				if !ok {
					continue
				}
				typ, ok := typeNameObj.Type().(*types.Named)
				if !ok {
					continue
				}
				nt[spec.Name.Name] = typ
			}
		}
	}
	return nt
}

// AssertStructUnchanged generates code that asserts at compile time that type t is unchanged.
// thisPkg is the package containing t.
// tname is the named type corresponding to t.
// ctx is a single-word context for this assertion, such as "Clone".
// If non-nil, AssertStructUnchanged will add elements to imports
// for each package path that the caller must import for the returned code to compile.
func AssertStructUnchanged(t *types.Struct, tname string, params *types.TypeParamList, ctx string, it *ImportTracker) []byte {
	buf := new(bytes.Buffer)
	w := func(format string, args ...any) {
		fmt.Fprintf(buf, format+"\n", args...)
	}
	w("// A compilation failure here means this code must be regenerated, with the command at the top of this file.")

	hasTypeParams := params != nil && params.Len() > 0
	if hasTypeParams {
		constraints, identifiers := FormatTypeParams(params, it)
		w("func _%s%sNeedsRegeneration%s (%s%s) {", tname, ctx, constraints, tname, identifiers)
		w("_%s%sNeedsRegeneration(struct {", tname, ctx)
	} else {
		w("var _%s%sNeedsRegeneration = %s(struct {", tname, ctx, tname)
	}

	for i := range t.NumFields() {
		st := t.Field(i)
		fname := st.Name()
		ft := t.Field(i).Type()
		if IsInvalid(ft) {
			continue
		}
		qname := it.QualifiedName(ft)
		var tag string
		if hasTypeParams {
			tag = t.Tag(i)
			if tag != "" {
				tag = "`" + tag + "`"
			}
		}
		if st.Anonymous() {
			w("\t%s %s", fname, tag)
		} else {
			w("\t%s %s %s", fname, qname, tag)
		}
	}

	if hasTypeParams {
		w("}{})\n}")
	} else {
		w("}{})")
	}
	return buf.Bytes()
}

// IsInvalid reports whether the provided type is invalid. It is used to allow
// codegeneration to run even when the target files have build errors or are
// missing views.
func IsInvalid(t types.Type) bool {
	return t.String() == "invalid type"
}

// ContainsPointers reports whether typ contains any pointers,
// either explicitly or implicitly.
// It has special handling for some types that contain pointers
// that we know are free from memory aliasing/mutation concerns.
func ContainsPointers(typ types.Type) bool {
	switch typ.String() {
	case "time.Time":
		// time.Time contains a pointer that does not need copying
		return false
	case "inet.af/netip.Addr", "net/netip.Addr", "net/netip.Prefix", "net/netip.AddrPort":
		return false
	}
	switch ft := typ.Underlying().(type) {
	case *types.Array:
		return ContainsPointers(ft.Elem())
	case *types.Basic:
		if ft.Kind() == types.UnsafePointer {
			return true
		}
	case *types.Chan:
		return true
	case *types.Interface:
		if ft.Empty() || ft.IsMethodSet() {
			return true
		}
		for i := 0; i < ft.NumEmbeddeds(); i++ {
			if ContainsPointers(ft.EmbeddedType(i)) {
				return true
			}
		}
	case *types.Map:
		return true
	case *types.Pointer:
		return true
	case *types.Slice:
		return true
	case *types.Struct:
		for i := range ft.NumFields() {
			if ContainsPointers(ft.Field(i).Type()) {
				return true
			}
		}
	case *types.Union:
		for i := range ft.Len() {
			if ContainsPointers(ft.Term(i).Type()) {
				return true
			}
		}
	}
	return false
}

// IsViewType reports whether the provided typ is a View.
func IsViewType(typ types.Type) bool {
	t, ok := typ.Underlying().(*types.Struct)
	if !ok {
		return false
	}
	if t.NumFields() != 1 {
		return false
	}
	return t.Field(0).Name() == "ж"
}

// FormatTypeParams formats the specified params and returns two strings:
//   - constraints are comma-separated type parameters and their constraints in square brackets (e.g. [T any, V constraints.Integer])
//   - names are comma-separated type parameter names in square brackets (e.g. [T, V])
//
// If params is nil or empty, both return values are empty strings.
func FormatTypeParams(params *types.TypeParamList, it *ImportTracker) (constraints, names string) {
	if params == nil || params.Len() == 0 {
		return "", ""
	}
	var constraintList, nameList []string
	for i := range params.Len() {
		param := params.At(i)
		name := param.Obj().Name()
		constraint := it.QualifiedName(param.Constraint())
		nameList = append(nameList, name)
		constraintList = append(constraintList, name+" "+constraint)
	}
	constraints = "[" + strings.Join(constraintList, ", ") + "]"
	names = "[" + strings.Join(nameList, ", ") + "]"
	return constraints, names
}

// LookupMethod returns the method with the specified name in t, or nil if the method does not exist.
func LookupMethod(t types.Type, name string) *types.Func {
	if t, ok := t.(*types.Named); ok {
		for i := 0; i < t.NumMethods(); i++ {
			if method := t.Method(i); method.Name() == name {
				return method
			}
		}
	}
	if t, ok := t.Underlying().(*types.Interface); ok {
		for i := 0; i < t.NumMethods(); i++ {
			if method := t.Method(i); method.Name() == name {
				return method
			}
		}
	}
	return nil
}
