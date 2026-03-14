// Package iolint provides analyzers for detecting io.Reader/io.Writer usage issues.
// It identifies cases where concrete types (like *os.File) are used when
// interface types (io.Reader, io.Writer) would provide better composability.
package iolint

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"

	"github.com/curtbushko/go-ai-lint/internal/domain"
	"github.com/curtbushko/go-ai-lint/internal/ports"
)

// Diagnostics contains the diagnostic templates for iolint.
var Diagnostics = map[string]domain.DiagnosticTemplate{
	"AIL140": {
		ID:       "AIL140",
		Name:     "concrete-file-param",
		Severity: domain.SeverityMedium,
		Category: domain.CategoryIO,
		Message:  "parameter uses concrete *os.File when io.Reader/io.Writer would suffice",
		Why: `Using *os.File as a parameter type limits the function to only work with
file system files. By accepting io.Reader or io.Writer instead, the function
becomes composable with buffers, network connections, compression streams,
HTTP bodies, and any other io implementation.

The io.Reader and io.Writer interfaces are the most important interfaces in Go.
Their power comes from their simplicity: just one method each (Read and Write),
which makes them endlessly composable.`,
		Fix: `Change the parameter type from *os.File to io.Reader (for read operations)
or io.Writer (for write operations). If both are needed, use io.ReadWriter.

If the function uses File-specific methods like Stat(), Seek(), Name(), Fd(),
Sync(), Truncate(), or Chmod(), then *os.File is justified.`,
		Example: domain.FixExample{
			Bad: `func ProcessData(f *os.File) ([]byte, error) {
    return io.ReadAll(f)
}`,
			Good: `func ProcessData(r io.Reader) ([]byte, error) {
    return io.ReadAll(r)
}

// Now works with files, buffers, HTTP bodies, etc.
f, _ := os.Open("data.txt")
ProcessData(f)

var buf bytes.Buffer
ProcessData(&buf)

ProcessData(resp.Body)`,
			Explanation: "By accepting io.Reader, the function works with any data source, not just files.",
		},
		CommonMistakes: []string{
			"WRONG: Keeping *os.File when only Read/Write/Close are used",
			"WRONG: Using *bytes.Buffer parameter - accept io.Reader/io.Writer instead",
			"WRONG: Changing to io.Reader when File-specific methods like Stat() are needed",
		},
	},
	"AIL141": {
		ID:       "AIL141",
		Name:     "concrete-buffer-param",
		Severity: domain.SeverityMedium,
		Category: domain.CategoryIO,
		Message:  "parameter uses concrete *bytes.Buffer when io.Reader/io.Writer would suffice",
		Why: `Using *bytes.Buffer as a parameter type limits the function to only work
with in-memory buffers. By accepting io.Reader or io.Writer instead, the
function becomes composable with files, network connections, and any other
io implementation.`,
		Fix: `Change the parameter type from *bytes.Buffer to io.Reader (for read operations)
or io.Writer (for write operations).`,
		Example: domain.FixExample{
			Bad: `func WriteData(buf *bytes.Buffer, data []byte) error {
    _, err := buf.Write(data)
    return err
}`,
			Good: `func WriteData(w io.Writer, data []byte) error {
    _, err := w.Write(data)
    return err
}`,
			Explanation: "By accepting io.Writer, the function works with any destination, not just buffers.",
		},
		CommonMistakes: []string{
			"WRONG: Keeping *bytes.Buffer when only Read/Write are used",
			"WRONG: Using *bytes.Buffer for the return type - return io.Reader if callers only read",
		},
	},
	"AIL142": {
		ID:       "AIL142",
		Name:     "concrete-response-param",
		Severity: domain.SeverityMedium,
		Category: domain.CategoryIO,
		Message:  "parameter uses concrete *http.Response when only Body is accessed; consider using io.ReadCloser",
		Why: `Using *http.Response as a parameter type when only the Body field is accessed
limits testability and composability. The function can accept io.ReadCloser
instead, making it easier to test with mock data and compose with other
io implementations.

The resp.Body field is an io.ReadCloser. If that's all you need, accept
io.ReadCloser directly. This makes the function's dependencies explicit
and allows callers to pass any io.ReadCloser, not just HTTP response bodies.`,
		Fix: `Change the parameter type from *http.Response to io.ReadCloser.
If you need access to Response-specific fields like StatusCode, Header,
Status, ContentLength, etc., then *http.Response is justified.`,
		Example: domain.FixExample{
			Bad: `func ProcessResponse(resp *http.Response) ([]byte, error) {
    defer resp.Body.Close()
    return io.ReadAll(resp.Body)
}`,
			Good: `func ProcessBody(body io.ReadCloser) ([]byte, error) {
    defer body.Close()
    return io.ReadAll(body)
}

// Called as: ProcessBody(resp.Body)`,
			Explanation: "By accepting io.ReadCloser, the function is easier to test and more composable.",
		},
		CommonMistakes: []string{
			"WRONG: Keeping *http.Response when only Body is used",
			"WRONG: Changing to io.ReadCloser when StatusCode or Header are needed",
		},
	},
}

// analyzer implements the iolint analyzer.
type analyzer struct {
	analysis *analysis.Analyzer
}

// New creates a new iolint analyzer.
func New() ports.Analyzer {
	ioAnalyzer := &analyzer{}
	ioAnalyzer.analysis = &analysis.Analyzer{
		Name:     "iolint",
		Doc:      "detects io.Reader/io.Writer usage issues in AI-generated Go code",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      ioAnalyzer.run,
	}
	return ioAnalyzer
}

// Name returns the analyzer name.
func (a *analyzer) Name() string {
	return "iolint"
}

// Analyzer returns the go/analysis.Analyzer.
func (a *analyzer) Analyzer() *analysis.Analyzer {
	return a.analysis
}

// fileSpecificMethods is the set of methods that are specific to *os.File
// and justify using the concrete type rather than io.Reader/io.Writer.
// Note: While Seek is part of io.Seeker, using it typically indicates
// the need for file-specific behavior beyond simple streaming.
var fileSpecificMethods = map[string]bool{
	"Stat":             true,
	"Fd":               true,
	"Sync":             true,
	"Name":             true,
	"Chdir":            true,
	"Chown":            true,
	"Chmod":            true,
	"Truncate":         true,
	"Readdir":          true,
	"Readdirnames":     true,
	"ReadDir":          true,
	"SetDeadline":      true,
	"SetReadDeadline":  true,
	"SetWriteDeadline": true,
	"SyscallConn":      true,
	"Seek":             true,
}

// bufferSpecificMethods is the set of methods that are specific to *bytes.Buffer
// and justify using the concrete type rather than io.Reader/io.Writer.
// Methods like Read, Write, ReadFrom, WriteTo are standard io interfaces.
var bufferSpecificMethods = map[string]bool{
	"Bytes":           true,
	"String":          true,
	"Len":             true,
	"Cap":             true,
	"Reset":           true,
	"Grow":            true,
	"Next":            true,
	"Truncate":        true,
	"UnreadByte":      true,
	"UnreadRune":      true,
	"ReadBytes":       true,
	"ReadString":      true,
	"WriteByte":       true,
	"WriteRune":       true,
	"ReadByte":        true,
	"ReadRune":        true,
	"Available":       true,
	"AvailableBuffer": true,
}

// responseSpecificFields is the set of fields that are specific to *http.Response
// and justify using the concrete type rather than io.ReadCloser.
// If only Body is accessed, io.ReadCloser should be used instead.
var responseSpecificFields = map[string]bool{
	"StatusCode":       true,
	"Status":           true,
	"Header":           true,
	"ContentLength":    true,
	"TransferEncoding": true,
	"Close":            true,
	"Uncompressed":     true,
	"Trailer":          true,
	"Request":          true,
	"TLS":              true,
	"Proto":            true,
	"ProtoMajor":       true,
	"ProtoMinor":       true,
}

// concreteTypeCheck defines parameters for checking concrete type usage.
type concreteTypeCheck struct {
	diagID          string
	typeChecker     func(*analysis.Pass, ast.Expr) bool
	specificMethods map[string]bool
}

// run performs the analysis.
func (a *analyzer) run(pass *analysis.Pass) (any, error) {
	insp, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !ok {
		return nil, nil
	}

	nodeFilter := []ast.Node{
		(*ast.FuncDecl)(nil),
	}

	// Define checks for concrete types
	checks := []concreteTypeCheck{
		{
			diagID:          "AIL140",
			typeChecker:     isOsFilePointer,
			specificMethods: fileSpecificMethods,
		},
		{
			diagID:          "AIL141",
			typeChecker:     isBytesBufferPointer,
			specificMethods: bufferSpecificMethods,
		},
	}

	insp.Preorder(nodeFilter, func(node ast.Node) {
		funcDecl, ok := node.(*ast.FuncDecl)
		if !ok {
			return
		}
		for _, check := range checks {
			a.checkConcreteTypeParams(pass, funcDecl, check)
		}
		// Check for *http.Response parameters
		a.checkConcreteResponseParams(pass, funcDecl)
	})

	return nil, nil
}

// concreteParam represents a parameter with a concrete type.
type concreteParam struct {
	ident *ast.Ident
	field *ast.Field
}

// checkConcreteTypeParams checks for concrete type parameters that could use io interfaces.
func (a *analyzer) checkConcreteTypeParams(
	pass *analysis.Pass,
	funcDecl *ast.FuncDecl,
	check concreteTypeCheck,
) {
	if funcDecl.Type.Params == nil {
		return
	}

	// Find all matching concrete type parameters
	var params []concreteParam

	for _, field := range funcDecl.Type.Params.List {
		if !check.typeChecker(pass, field.Type) {
			continue
		}
		for _, name := range field.Names {
			params = append(params, concreteParam{ident: name, field: field})
		}
	}

	if len(params) == 0 {
		return
	}

	// If no function body, we can't analyze method usage
	if funcDecl.Body == nil {
		return
	}

	// Track which parameters use type-specific methods
	paramsUsingSpecificMethods := make(map[string]bool)

	ast.Inspect(funcDecl.Body, func(node ast.Node) bool {
		callExpr, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}

		selExpr, ok := callExpr.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}

		// Check if this is a method call on one of our parameters
		ident, ok := selExpr.X.(*ast.Ident)
		if !ok {
			return true
		}

		// Check if this identifier is one of our parameters
		for _, p := range params {
			if ident.Name == p.ident.Name {
				// Check if the method is type-specific
				if check.specificMethods[selExpr.Sel.Name] {
					paramsUsingSpecificMethods[ident.Name] = true
				}
			}
		}

		return true
	})

	// Report diagnostic for parameters that don't use type-specific methods
	diag := Diagnostics[check.diagID]
	for _, p := range params {
		if !paramsUsingSpecificMethods[p.ident.Name] {
			domain.Report(pass, analysis.Diagnostic{
				Pos:      p.field.Pos(),
				Category: string(diag.Category),
				Message:  check.diagID + ": " + diag.Message,
			})
		}
	}
}

// isOsFilePointer checks if a type expression is *os.File.
func isOsFilePointer(pass *analysis.Pass, expr ast.Expr) bool {
	return isPointerToType(pass, expr, "os", "File")
}

// isBytesBufferPointer checks if a type expression is *bytes.Buffer.
func isBytesBufferPointer(pass *analysis.Pass, expr ast.Expr) bool {
	return isPointerToType(pass, expr, "bytes", "Buffer")
}

// isPointerToType checks if a type expression is a pointer to a specific type.
func isPointerToType(pass *analysis.Pass, expr ast.Expr, pkgPath, typeName string) bool {
	starExpr, ok := expr.(*ast.StarExpr)
	if !ok {
		return false
	}

	selExpr, ok := starExpr.X.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	_, ok = selExpr.X.(*ast.Ident)
	if !ok {
		return false
	}

	// Check if selector matches expected type name
	if selExpr.Sel.Name != typeName {
		return false
	}

	// Check using type information
	obj := pass.TypesInfo.ObjectOf(selExpr.Sel)
	if obj == nil {
		return false
	}

	named, ok := obj.Type().(*types.Named)
	if !ok {
		return false
	}

	// Check package path and type name
	if named.Obj().Pkg() == nil {
		return false
	}

	return named.Obj().Pkg().Path() == pkgPath
}

// isHTTPResponsePointer checks if a type expression is *http.Response.
func isHTTPResponsePointer(pass *analysis.Pass, expr ast.Expr) bool {
	return isPointerToType(pass, expr, "net/http", "Response")
}

// checkConcreteResponseParams checks for *http.Response parameters that could use io.ReadCloser.
func (a *analyzer) checkConcreteResponseParams(pass *analysis.Pass, funcDecl *ast.FuncDecl) {
	if funcDecl.Type.Params == nil {
		return
	}

	// Find all *http.Response parameters
	var params []concreteParam

	for _, field := range funcDecl.Type.Params.List {
		if !isHTTPResponsePointer(pass, field.Type) {
			continue
		}
		for _, name := range field.Names {
			params = append(params, concreteParam{ident: name, field: field})
		}
	}

	if len(params) == 0 {
		return
	}

	// If no function body, we can't analyze field usage
	if funcDecl.Body == nil {
		return
	}

	// Track which parameters use response-specific fields (other than Body)
	paramsUsingSpecificFields := make(map[string]bool)

	ast.Inspect(funcDecl.Body, func(node ast.Node) bool {
		selExpr, ok := node.(*ast.SelectorExpr)
		if !ok {
			return true
		}

		// Check if this is a direct field access on one of our parameters (resp.Field)
		// NOT chained accesses like resp.Body.Close
		ident, ok := selExpr.X.(*ast.Ident)
		if !ok {
			return true
		}

		// Check if this identifier is one of our parameters
		for _, p := range params {
			if ident.Name == p.ident.Name {
				// Check if the field is response-specific (not Body)
				fieldName := selExpr.Sel.Name
				if responseSpecificFields[fieldName] {
					paramsUsingSpecificFields[ident.Name] = true
				}
			}
		}

		return true
	})

	// Report diagnostic for parameters that don't use response-specific fields
	diag := Diagnostics["AIL142"]
	for _, p := range params {
		if !paramsUsingSpecificFields[p.ident.Name] {
			domain.Report(pass, analysis.Diagnostic{
				Pos:      p.field.Pos(),
				Category: string(diag.Category),
				Message:  "AIL142: " + diag.Message,
			})
		}
	}
}
