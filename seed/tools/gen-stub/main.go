package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"io"
	"io/fs"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime/debug"
	"sort"
	"strconv"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/spf13/pflag"
	"golang.org/x/tools/go/ast/astutil"
)

type Options struct {
	SrcDir     string
	BasePkg    string
	DstDirList []string
	DocName    string
	Prefix     string

	Debug bool
}

func main() {
	log.SetFlags(0)
	log.SetPrefix("** ")

	var options Options
	pflag.StringVar(&options.BasePkg, "base", "", "base package")
	pflag.StringVar(&options.SrcDir, "src", "", "source directory")
	pflag.StringSliceVar(&options.DstDirList, "dst", nil, "destination directory")
	pflag.StringVar(&options.DocName, "doc", "", "target openapi doc (OAS3.0)")
	pflag.StringVar(&options.Prefix, "prefix", "", "prefix")
	pflag.BoolVar(&options.Debug, "debug", false, "debug")
	pflag.Parse()

	if options.BasePkg == "" {
		if info, ok := debug.ReadBuildInfo(); ok {
			options.BasePkg = info.Main.Path
		}
		if options.Debug {
			log.Printf("[DEBUG] base package is %q (from buildinfo)", options.BasePkg)
		}
	}

	if options.SrcDir == "" || options.DstDirList == nil || options.DocName == "" {
		flag.Usage()
		os.Exit(1)
	}
	if err := run(options); err != nil {
		log.Fatalf("!! %+v", err)
	}
}

func run(options Options) error {
	basePkg := options.BasePkg
	srcdir := options.SrcDir
	dstdirList := options.DstDirList
	docname := options.DocName
	prefix := options.Prefix

	fset := token.NewFileSet()

	// scan openapi doc
	doc, err := openapi3.NewLoader().LoadFromFile(docname)
	if err != nil {
		return fmt.Errorf("load openapi doc: %w", err)
	}

	state := newState(doc)

	ignoreTestFile := func(info fs.FileInfo) bool {
		return !strings.HasSuffix(info.Name(), "_test.go")
	}

	// scan dst dir
	dstFiles := map[string]*file{} // normalized tag -> file
	{
		for _, dstdir := range dstdirList {
			if err := os.MkdirAll(dstdir, 0744); err != nil {
				return fmt.Errorf("get or create dst dir: %w", err)
			}
			pkgs, err := parser.ParseDir(fset, dstdir, ignoreTestFile, parser.AllErrors|parser.ParseComments|parser.SkipObjectResolution)
			if err != nil {
				return fmt.Errorf("parse impl packages: %w", err)
			}
			for _, pkg := range pkgs {
				for filename, f := range pkg.Files {
					tag := TagFromFileName(filename)

					methods := map[string]*ast.FuncDecl{}
					for _, decl := range f.Decls {
						decl, ok := decl.(*ast.FuncDecl)
						if !ok {
							continue
						}
						if decl.Recv == nil || len(decl.Recv.List) == 0 {
							continue
						}

						typeName := ""
						switch typ := decl.Recv.List[0].Type.(type) {
						case *ast.StarExpr:
							typeName = typ.X.(*ast.Ident).Name
						case *ast.Ident:
							typeName = typ.Name
						default:
							fmt.Fprintln(os.Stderr, "----------------------------------------") // nolint
							printer.Fprint(os.Stderr, fset, decl)                               // nolint
							fmt.Fprintln(os.Stderr, "----------------------------------------") // nolint
							panic(fmt.Sprintf("unexpected type: %T", typ))
						}
						if !strings.HasSuffix(typeName, "Controller") {
							continue

						}
						if !ast.IsExported(decl.Name.Name) {
							continue
						}
						methods[decl.Name.Name] = decl
					}
					dstFiles[tag] = &file{path: filename, syntax: f, methods: methods}
				}
			}
		}
	}

	// scan src dir
	pkgs, err := parser.ParseDir(fset, srcdir, ignoreTestFile, parser.AllErrors)
	if err != nil {
		return fmt.Errorf("parse oapigen packages: %w", err)
	}

	// https://github.com/josharian/impl like generate skeleton
	for _, pkg := range pkgs {
		for _, f := range pkg.Files {
			ob := f.Scope.Lookup("StrictServerInterface")
			if ob == nil {
				continue
			}
			decl, ok := ob.Decl.(*ast.TypeSpec) // nolint
			if !ok {
				continue
			}
			typ, ok := decl.Type.(*ast.InterfaceType)
			if !ok {
				continue
			}
			for _, p := range typ.Methods.List {
				state.Add(p)
			}
		}
	}

	// emit

	var w io.Writer = os.Stderr
	printer := &printer.Config{Tabwidth: 8}
	buf := new(bytes.Buffer)

	type typeSet struct {
		xGoPackage string
		typeNames  []string
	}
	dstdirToTypeSetMap := map[string]*typeSet{}

	if err := state.EachByTag(func(tag string, defs []*def) error {
		xGoPackage := defs[0].xGoPackage
		// log.Printf("\t🔢 x-go-package:%q\ttag:%q", xGoPackage, tag)
		dstdir := SelectDirByXGoPackage(xGoPackage, dstdirList)
		pkgname := filepath.Base(dstdir)

		normalziedTag := NormalizeTag(tag)
		f, updated := dstFiles[normalziedTag]
		if !updated { // ファイルが存在しない場合
			filename := filepath.Join(dstdir, FileNameFromTag(normalziedTag))
			source := fmt.Sprintf("package %s", pkgname)
			syntax, err := parser.ParseFile(fset, tag+".go", source, parser.AllErrors)
			if err != nil {
				return fmt.Errorf("create file, tag=%s", tag)
			}
			f = &file{path: filename, syntax: syntax, methods: map[string]*ast.FuncDecl{}}
		}

		{
			// TODO: 特定の位置以外から実行したときに壊れる
			found := false
			for _, im := range f.syntax.Imports {
				if im.Name != nil && im.Name.Name == "oapigen" {
					found = true
					break
				}
			}
			if !found {
				astutil.AddImport(fset, f.syntax, "context")
				astutil.AddNamedImport(fset, f.syntax, "oapigen", path.Join(basePkg, srcdir))
			}
		}

		parts := strings.Split(tag, "-")
		camelized := make([]string, len(parts))
		for i, x := range parts {
			camelized[i] = ToTitle(x)
		}
		typeName := fmt.Sprintf("%sController", strings.Join(camelized, "")) // tag:foo-bar -> FooBarController

		typeset, ok := dstdirToTypeSetMap[dstdir]
		if !ok {
			typeset = &typeSet{
				xGoPackage: xGoPackage,
			}
			dstdirToTypeSetMap[dstdir] = typeset
		}
		typeset.typeNames = append(typeset.typeNames, typeName)

		fileBuf := new(bytes.Buffer)
		tmpBuf := new(bytes.Buffer)
		w = fileBuf

		// type XXXController struct {}
		// func NewXXXController() *XXXController { ... }
		if len(f.methods) == 0 {
			fmt.Fprintln(w, "")                                                                       // nolint
			fmt.Fprintf(w, "type %s struct { }\n\n", typeName)                                        // nolint
			fmt.Fprintf(w, "func New%s() *%s {\n\treturn &%s{}\n}\n\n", typeName, typeName, typeName) // nolint
		}

		modified := false
		for _, def := range defs {
			buf.Reset()

			// extract method parameters signature
			fn := def.field.Type.(*ast.FuncType)
			if len(fn.Params.List[0].Names) > 0 {
				for _, p := range fn.Params.List[1:] {
					if len(p.Names) > 0 {
						name := p.Names[0].Name
						if strings.HasSuffix(name, "Id") {
							p.Names[0].Name = name[:len(name)-2] + "ID"
						} else if strings.HasSuffix(name, "IdParams") {
							p.Names[0].Name = name[:len(name)-8] + "IDParams"
						} else if strings.HasSuffix(name, "IdParam") {
							p.Names[0].Name = name[:len(name)-7] + "ID"
						}
						if typ, ok := p.Type.(*ast.Ident); ok && ast.IsExported(typ.Name) {
							p.Type = &ast.SelectorExpr{X: &ast.Ident{NamePos: p.Type.Pos(), Name: "oapigen"}, Sel: typ}
						}
					}
				}
			}
			// transform to use named-return
			if results := fn.Results; results != nil && results.List != nil {
				if len(results.List) == 2 {
					r0 := results.List[0]
					if len(r0.Names) == 0 {
						r0.Names = []*ast.Ident{{Name: "output", NamePos: results.Pos() + 1}}
					}
					if typename, ok := r0.Type.(*ast.Ident); ok {
						r0.Type = &ast.SelectorExpr{X: &ast.Ident{Name: "oapigen", NamePos: typename.Pos()}, Sel: typename}
					}
					if len(results.List[1].Names) == 0 {
						results.List[1].Names = []*ast.Ident{{Name: "err", NamePos: results.Pos() + 2}}
					}
				}
			}

			name := def.field.Names[0].Name
			if err := printer.Fprint(buf, fset, fn); err != nil {
				return fmt.Errorf("extract signature, tag=%s, method=%s", tag, name)
			}

			// doc string
			var comments []string
			{
				comments = append(comments, fmt.Sprintf("%s is endpoint of %s %s", name, def.method, def.path))
				doc := def.op.Summary
				if doc == "" {
					doc = def.op.Description
				}
				for _, line := range strings.Split(doc, "\n") {
					if line != "" {
						comments = append(comments, line)
					}
				}
				// parametersの情報を追加
				// e.g.) * body :requestBody -- "need: var body oapigen.CreateKnowledgeJSONBody; gctx.ShouldBindJSON(&body); "
				if len(def.op.Parameters) > 0 || def.op.RequestBody != nil {
					comments = append(comments, "")
					for _, p := range def.op.Parameters {
						pname := p.Value.Name
						if p.Value != nil && p.Value.Schema != nil && p.Value.Schema.Value != nil && p.Value.Schema.Value.Default != nil {
							defaultValue := p.Value.Schema.Value.Default
							if v, ok := defaultValue.(string); ok {
								pname = fmt.Sprintf("%s default=%q", pname, v)
							} else {
								pname = fmt.Sprintf("%s default=%v", pname, defaultValue)
							}
						} else if !p.Value.Required {
							pname = fmt.Sprintf("%s default=nil", pname)
						}
						comments = append(comments, fmt.Sprintf("* %-6s:%-35s -- %q", p.Value.In, pname, p.Value.Description))
					}
					if def.op.RequestBody != nil {
						comments = append(comments, fmt.Sprintf("* %-6s:%-35s -- %q", "body", "requestBody", fmt.Sprintf("need: var body oapigen.%sJSONBody; gctx.ShouldBindJSON(&body); ", name)))
					}
				}
			}

			// already defined
			if decl, ok := f.methods[name]; ok {
				// modify comments, if need
				if strings.Join(comments, "\n") != strings.TrimSpace(decl.Doc.Text()) {
					modified = true

					prevComment := decl.Doc
					clines := make([]*ast.Comment, len(comments))
					pos := decl.Pos() - 1 // hack
					for i, line := range comments {
						clines[i] = &ast.Comment{Slash: pos, Text: "// " + line}
					}
					decl.Doc = &ast.CommentGroup{List: clines}
					for i, cg := range f.syntax.Comments {
						if cg == prevComment {
							f.syntax.Comments[i] = decl.Doc
							break
						}
					}
				}

				// update method signature, if need
				{

					// 名前を既存のコードに合わせる
					defParams := def.field.Type.(*ast.FuncType).Params.List // interfaceのparameters (oapigen)
					implParams := decl.Type.Params.List                     // 実装コードのparameters
					for i, p := range implParams {
						if len(defParams) == i {
							break
						}
						if len(p.Names) > 0 && len(defParams[i].Names) > 0 {
							defParams[i].Names[0].Name = p.Names[0].Name
						}
					}

					tmpBuf.Reset()
					if err := printer.Fprint(tmpBuf, fset, def.field.Type); err != nil {
						return err
					}
					defSig := tmpBuf.String()
					tmpBuf.Reset()
					if err := printer.Fprint(tmpBuf, fset, decl.Type); err != nil {
						return err
					}
					implSig := tmpBuf.String()
					isSignatureChanged := defSig != implSig
					if isSignatureChanged {
						modified = true

						// hack: interfaceのmethodのposを調整する (正しく調整しないとコメントの位置がズレる)
						pos := decl.Type.Params.Pos() + 1
						end := decl.Type.Params.End() - 1
						ast.Inspect(def.field.Type, func(node ast.Node) bool {
							switch n := node.(type) {
							case *ast.Comment:
								n.Slash = pos
							case *ast.FieldList:
								n.Opening = pos
								n.Closing = end
							case *ast.Ident:
								n.NamePos = pos
							case *ast.Ellipsis:
								n.Ellipsis = pos
							case *ast.BasicLit:
								n.ValuePos = pos
							case *ast.ParenExpr:
								n.Lparen = pos
								n.Rparen = end
							case *ast.StarExpr:
								n.Star = pos
							case *ast.UnaryExpr:
								n.OpPos = pos
							case *ast.ArrayType:
								n.Lbrack = pos
							case *ast.InterfaceType:
								n.Interface = pos
							case *ast.MapType:
								n.Map = pos
							case *ast.ChanType:
								n.Begin = pos
							}
							return true
						})
						decl.Type = def.field.Type.(*ast.FuncType)
					}
				}
				continue
			}

			// not defined yet

			fmt.Fprintln(w, "\n// "+strings.TrimSpace(strings.Join(comments, "\n// "))) // nolint
			fmt.Fprintf(w, "func (c *%s) %s", typeName, name)                           // nolint
			fmt.Fprint(w, buf.String()[4:])                                             // nolint
			fmt.Fprintln(w, " {")                                                       // nolint
			fmt.Fprintln(w, "\treturn")                                                 // nolint
			fmt.Fprintln(w, "}")                                                        // nolint
		}

		if fileBuf.Len() == 0 && !modified {
			if options.Debug {
				log.Printf("emit skip :: create=%5v tag=%-25s filename=%s", !updated, tag, f.path)
			}
			return nil
		}

		if created := !updated; options.Debug || created {
			log.Printf("emit file :: create=%5v tag=%-25s filename=%s", created, tag, f.path)
		}
		wf, err := os.Create(f.path + ".mv")
		if err != nil {
			return fmt.Errorf("create %s: %w", f.path, err)
		}
		if err := printer.Fprint(wf, fset, f.syntax); err != nil {
			return fmt.Errorf("write %s: %w", f.path, err)
		}
		if _, err := io.Copy(wf, fileBuf); err != nil {
			return fmt.Errorf("write %s: %w", f.path, err)
		}
		if err := os.Rename(wf.Name(), f.path); err != nil {
			return fmt.Errorf("rename %s: %w", f.path, err)
		}
		return nil
	}); err != nil {
		return err
	}

	// controllerをmountするfile
	for dstdir, typeset := range dstdirToTypeSetMap {
		pkgname := filepath.Base(dstdir)
		controllerTypeNames := typeset.typeNames
		if len(controllerTypeNames) == 0 {
			continue
		}

		filename := filepath.Join(dstdir, "controller.go")
		log.Printf("emit file :: filename=%s", filename)

		wf, err := os.Create(filename)
		if err != nil {
			return fmt.Errorf("create %s", filename)
		}
		defer wf.Close() // nolint

		w = wf
		if typeset.xGoPackage != "" || len(dstdirList) == 1 { // --dstの先頭のディレクトリは全てのcontrollerを埋め込んだものとして扱う
			sort.Strings(controllerTypeNames)

			prefix := ToTitle(prefix)
			if typeset.xGoPackage != "" {
				prefix = ToTitle(typeset.xGoPackage)
			}
			pkgname := filepath.Base(dstdir)
			prefix = prefix + ToTitle(pkgname)
			typeName := fmt.Sprintf("%sController", prefix)
			fmt.Fprintf(w, "// Generated by swagger/tools/gen-stub %s\n", strings.Join(os.Args[1:], " ")) // nolint
			fmt.Fprintf(w, "package %s", pkgname)                                                         // nolint
			fmt.Fprintln(w, "")                                                                           // nolint
			fmt.Fprintln(w, "")                                                                           // nolint
			// CODE -- define struct:
			fmt.Fprintf(w, "// %s :\n", typeName)          // nolint
			fmt.Fprintf(w, "type %s struct {\n", typeName) // nolint
			for _, name := range controllerTypeNames {
				fmt.Fprintf(w, "\t*%s\n", name) // nolint
			}
			fmt.Fprintf(w, "}\n") // nolint
			fmt.Fprintln(w, "")   // nolint
			// CODE --  define factory function:
			fmt.Fprintf(w, "// New%s :\n", typeName)                  // nolint
			fmt.Fprintf(w, "func New%s() *%s{\n", typeName, typeName) // nolint
			fmt.Fprintf(w, "\treturn &%s{\n", typeName)               // nolint
			for _, name := range controllerTypeNames {
				fmt.Fprintf(w, "\t\t%s: New%s(),\n", name, name) // nolint
			}
			fmt.Fprintf(w, "\t}\n") // nolint
			fmt.Fprintf(w, "}\n")   // nolint
		} else { // root dstのpackageの場合には諸々をimportしてembeddedする (importが必要になる)
			sort.Strings(controllerTypeNames)
			typeName := fmt.Sprintf("%s%sController", ToTitle(prefix), ToTitle(pkgname))
			fmt.Fprintf(w, "// Generated by seed/tools/gen-stub %s\n", strings.Join(os.Args[1:], " ")) // nolint
			fmt.Fprintf(w, "package %s", pkgname)                                                      // nolint
			fmt.Fprintln(w, "")                                                                        // nolint
			// CODE --  import:
			fmt.Fprintln(w, "import (") // nolint
			for _, dstdir := range dstdirList {
				typeset, ok := dstdirToTypeSetMap[dstdir]
				if !ok || typeset.xGoPackage == "" || len(typeset.typeNames) == 0 {
					continue
				}
				fmt.Fprintf(w, "\t%s %q\n", typeset.xGoPackage, path.Join(basePkg, dstdir)) // nolint
			}
			fmt.Fprintln(w, ")") // nolint
			fmt.Fprintln(w, "")  // nolint
			// CODE --  define struct:
			fmt.Fprintf(w, "// %s :\n", typeName)          // nolint
			fmt.Fprintf(w, "type %s struct {\n", typeName) // nolint
			for _, name := range controllerTypeNames {
				fmt.Fprintf(w, "\t*%s\n", name) // nolint
			}
			fmt.Fprintln(w, "") // nolint
			for _, dstdir := range dstdirList {
				typeset, ok := dstdirToTypeSetMap[dstdir]
				if !ok || typeset.xGoPackage == "" || len(typeset.typeNames) == 0 {
					continue
				}
				prefix := ToTitle(prefix)
				if typeset.xGoPackage != "" {
					prefix = ToTitle(typeset.xGoPackage)
				}
				pkgname := filepath.Base(dstdir)
				prefix = prefix + ToTitle(pkgname)
				name := fmt.Sprintf("%sController", prefix)
				fmt.Fprintf(w, "\t*%s.%s\n", typeset.xGoPackage, name) // nolint
			}
			fmt.Fprintf(w, "}\n") // nolint
			fmt.Fprintln(w, "")   // nolint
			// CODE --  define factory function:
			fmt.Fprintf(w, "// New%s :\n", typeName)                   // nolint
			fmt.Fprintf(w, "func New%s() *%s {\n", typeName, typeName) // nolint
			fmt.Fprintf(w, "\treturn &%s{\n", typeName)                // nolint
			for _, name := range controllerTypeNames {
				fmt.Fprintf(w, "\t\t%s: New%s(),\n", name, name) // nolint
			}
			for _, dstdir := range dstdirList {
				typeset, ok := dstdirToTypeSetMap[dstdir]
				if !ok || typeset.xGoPackage == "" || len(typeset.typeNames) == 0 {
					continue
				}
				prefix := ToTitle(prefix)
				if typeset.xGoPackage != "" {
					prefix = ToTitle(typeset.xGoPackage)
				}
				pkgname := filepath.Base(dstdir)
				prefix = prefix + ToTitle(pkgname)
				name := fmt.Sprintf("%sController", prefix)
				fmt.Fprintf(w, "\t\t%s: %s.New%s(),\n", name, typeset.xGoPackage, name) // nolint
			}
			fmt.Fprintf(w, "\t}\n") // nolint
			fmt.Fprintf(w, "}\n")   // nolint
		}
	}
	return nil
}

type file struct {
	path    string
	syntax  *ast.File
	methods map[string]*ast.FuncDecl
}

type def struct {
	field *ast.Field

	method     string
	path       string
	xGoPackage string // --dstで渡す先頭のディレクトリ以外の場所に出力したい場合にここに "" 以外の値が入る (e.g. "seoboard")
	op         *openapi3.Operation
}

type endpoint struct {
	method     string
	path       string
	xGoPackage string // --dstで渡す先頭のディレクトリ以外の場所に出力したい場合にここに "" 以外の値が入る (e.g. "seoboard")
	*openapi3.Operation
}

type state struct {
	endpoints map[string]*endpoint

	tags []string
	defs map[string][]*def
}

func newState(doc *openapi3.T) *state {
	endpoints := map[string]*endpoint{}
	for path, pathItem := range doc.Paths {
		for method, op := range pathItem.Operations() {
			id := ToTitle(op.OperationID) // foo -> Foo
			xGoPackage := ""
			if len(op.Extensions) > 0 {
				if v, ok := op.Extensions["x-go-package"]; ok {
					xGoPackage = v.(string)
					if v, err := strconv.Unquote(xGoPackage); err == nil {
						xGoPackage = v
					}
				}
			}
			// log.Printf("ℹ️ x-go-package:%q\ttag:%v\toperationid:%q", xGoPackage, op.Tags, op.OperationID)
			endpoints[id] = &endpoint{Operation: op, method: method, path: path, xGoPackage: xGoPackage}
		}
	}
	return &state{defs: map[string][]*def{}, endpoints: endpoints}
}

func (s *state) Add(method *ast.Field) {
	op, ok := s.endpoints[method.Names[0].Name]
	if !ok {
		panic(fmt.Sprintf("%q is not found, maybe --doc option is invalid?", method.Names[0].Name))
	}
	tag := "notags" // tags[0]が存在しない場合
	if len(op.Tags) > 0 {
		tag = strings.TrimSpace(op.Tags[0])
	}
	ops, found := s.defs[tag]
	if !found {
		s.tags = append(s.tags, tag)
	}
	s.defs[tag] = append(ops, &def{op: op.Operation, field: method, method: op.method, path: op.path, xGoPackage: op.xGoPackage})
}

func (s *state) EachByTag(fn func(string, []*def) error) error {
	sort.Strings(s.tags)
	for _, tag := range s.tags {
		if err := fn(tag, s.defs[tag]); err != nil {
			return err
		}
	}
	return nil
}

// SelectDirByXGoPackage
func SelectDirByXGoPackage(xGoPackage string, dirs []string) string {
	// x-go-packageの指定がない場合には`--dst`で渡された先頭の値を利用する
	if xGoPackage == "" {
		return dirs[0]
	}

	// x-go-packageの指定がある場合には、`x-go-package`の値が含んだディレクトリを利用する
	sep := string(filepath.Separator) // "/"
	for _, dir := range dirs {
		for _, x := range strings.Split(dir, sep) {
			if x == "" {
				continue
			}
			if x == xGoPackage {
				return dir
			}
		}
	}
	log.Printf("dstdir for x-go-package=%q is not found (in %v)", xGoPackage, dirs)
	return dirs[0]
}

func TagFromFileName(filename string) string {
	base := strings.TrimSuffix(filepath.Base(filename), ".go")

	// 特別扱い (TODO: 後で治したい)
	switch base {
	case "growth_topic":
		return "topic"
	}

	tag := strings.ReplaceAll(base, "_", "-") // foo_bar -> foo-bar
	return tag
}
func FileNameFromTag(normalizedTag string) string {
	return strings.ReplaceAll(normalizedTag, "-", "_") + ".go"
}
func NormalizeTag(tag string) string {
	return strings.ReplaceAll(ToSnakeCase(tag), "_", "-") // fooBar -> foo-bar, foo_bar -> foo-bar
}

var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

func ToSnakeCase(str string) string { // fooBar -> foo_bar
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}

func ToTitle(str string) string {
	if str == "" {
		return str
	}
	return strings.ToUpper(str[0:1]) + str[1:] // foo -> Foo
}
