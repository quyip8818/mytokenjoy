//go:build ignore

// uuid-migrate rewrites Go struct fields and function parameters from
// string/int64 to uuid.UUID based on the whitelist defined in
// docs/adr/uuid-v7-migration.md.
//
// Usage:
//
//	go run scripts/uuid-migrate/main.go
//
// It modifies files in-place. Run `goimports -w ./...` afterwards.
package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// ---------------------------------------------------------------------------
// Configuration
// ---------------------------------------------------------------------------

// Struct field whitelist: field name → set of old type strings it can have.
// Only fields matching BOTH name AND old type will be rewritten.
var fieldWhitelist = map[string]map[string]bool{
	"ID":                {"string": true, "int64": true},
	"CompanyID":         {"int64": true},
	"UserID":            {"string": true},
	"MemberID":          {"string": true, "*string": true},
	"DepartmentID":      {"string": true},
	"ProjectID":         {"string": true, "*string": true},
	"PlatformKeyID":     {"string": true},
	"NodeID":            {"string": true},
	"RoleID":            {"string": true},
	"LotID":             {"string": true},
	"RechargeOrderID":   {"string": true},
	"ApplicantID":       {"string": true},
	"OperatorID":        {"string": true},
	"ManagerID":         {"*string": true},
	"ParentID":          {"*string": true},
	"DefaultModelID":    {"*int64": true},
	"FallbackModelID":   {"*int64": true},
	"OwnerDepartmentID": {"string": true},
	"FIFOHeadLotID":     {"*string": true},
	"RootDeptID":        {"*string": true},
	"LastLedgerID":      {"*string": true},
	"AxisID":            {"string": true},
	"OwnerID":           {"string": true},
	"ModelID":           {"int64": true},
	"CreatedBy":         {"string": true},
}

// Slice/map fields: field name → old type expression → new type expression
var collectionWhitelist = map[string][2]string{
	"ModelWhitelist":  {"[]int64", "[]uuid.UUID"},
	"RequestedModels": {"[]int64", "[]uuid.UUID"},
	"AllowedModelIDs": {"[]int64", "[]uuid.UUID"},
	"NotifyRoleIDs":   {"[]string", "[]uuid.UUID"},
	"MemberIDs":       {"[]string", "[]uuid.UUID"},
	"MemberBudgets":   {"map[string]float64", "map[uuid.UUID]float64"},
}

// Function/method parameter whitelist: param name → old type.
var paramWhitelist = map[string]string{
	"companyID":     "int64",
	"memberID":      "string",
	"projectID":     "string",
	"departmentID":  "string",
	"nodeID":        "string",
	"platformKeyID": "string",
	"keyID":         "string",
	"modelID":       "int64",
	"ownerID":       "string",
	"lotID":         "string",
	"userID":        "string",
	"providerKeyID": "string",
	"operatorID":    "string",
	"deptID":        "string",
}

// Struct names whose `ID` field must NOT be converted.
var excludeStructs = map[string]bool{
	"RawConsumeLog": true,
	"RiverJobView":  true,
	"SSEEvent":      true,
}

// Struct-scoped field whitelist: fields that only convert in specific structs.
// Key = field name, Value = set of struct names where conversion is allowed.
// If a field is in this map, it ONLY converts inside those structs.
var structScopedFields = map[string]map[string]bool{
	"OperatorID":    {"OperationLog": true},
	"CreatedBy":     {"RechargeOrder": true},
	"AxisID":        {"ConsumedDelta": true},
	"OwnerID":       {"ModelAllowlistRow": true},
	"ModelID":       {"ModelAllowlistRow": true, "ModelInfo": true},
	"MemberBudgets": {"Project": true},
}

// Directory prefixes to skip entirely.
var skipDirs = []string{
	"internal/integration/",
	"vendor/",
	".git/",
}

// Field names that must NEVER be converted regardless of type.
var fieldBlacklist = map[string]bool{
	"ExternalID":         true,
	"EmployeeID":         true,
	"NewAPIWalletUserID": true,
	"NewAPIKeyID":        true,
	"NewAPIChannelID":    true,
	"PackageID":          true,
	"CallerID":           true,
	"AccessKeyID":        true,
	"AppID":              true,
	"CorpID":             true,
	"AgentID":            true,
	"TokenID":            true,
	"LogID":              true,
	"InviteCode":         true,
	"IdempotencyKey":     true,
	"KeyHash":            true,
	"KeyPrefix":          true,
	"Operator":           true,
}

// ---------------------------------------------------------------------------
// Main
// ---------------------------------------------------------------------------

func main() {
	root := "."
	modified := 0

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			base := info.Name()
			if base == "vendor" || base == ".git" || base == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		if strings.HasSuffix(path, "_test.go") {
			return nil
		}
		for _, skip := range skipDirs {
			if strings.Contains(path, skip) {
				return nil
			}
		}

		if rewriteFile(path) {
			modified++
			fmt.Println("modified:", path)
		}
		return nil
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "walk error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("\n%d files modified. Run: goimports -w ./...\n", modified)
}

// ---------------------------------------------------------------------------
// File processing
// ---------------------------------------------------------------------------

func rewriteFile(path string) bool {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		fmt.Fprintf(os.Stderr, "parse error %s: %v\n", path, err)
		return false
	}

	changed := false

	// Walk top-level declarations to properly track struct names
	for _, decl := range f.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if ok && genDecl.Tok == token.TYPE {
			for _, spec := range genDecl.Specs {
				ts, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}
				structType, ok := ts.Type.(*ast.StructType)
				if !ok {
					continue
				}
				structName := ts.Name.Name
				if structType.Fields != nil {
					for _, field := range structType.Fields.List {
						if rewriteStructField(field, structName) {
							changed = true
						}
					}
				}
			}
		}

		// Handle function declarations (methods and functions)
		funcDecl, ok := decl.(*ast.FuncDecl)
		if ok && funcDecl.Type.Params != nil {
			for _, field := range funcDecl.Type.Params.List {
				if rewriteFuncParam(field) {
					changed = true
				}
			}
		}
	}

	// Also handle interface method signatures inside type declarations
	for _, decl := range f.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}
		for _, spec := range genDecl.Specs {
			ts, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			ifaceType, ok := ts.Type.(*ast.InterfaceType)
			if !ok {
				continue
			}
			for _, method := range ifaceType.Methods.List {
				funcType, ok := method.Type.(*ast.FuncType)
				if !ok {
					continue
				}
				if funcType.Params != nil {
					for _, field := range funcType.Params.List {
						if rewriteFuncParam(field) {
							changed = true
						}
					}
				}
			}
		}
	}

	if !changed {
		return false
	}

	// Add uuid import if not present
	addUUIDImport(f)

	// Write back
	out, err := os.Create(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "create error %s: %v\n", path, err)
		return false
	}
	defer out.Close()

	cfg := &printer.Config{Mode: printer.UseSpaces | printer.TabIndent, Tabwidth: 8}
	if err := cfg.Fprint(out, fset, f); err != nil {
		fmt.Fprintf(os.Stderr, "print error %s: %v\n", path, err)
		return false
	}
	return true
}

// ---------------------------------------------------------------------------
// Struct field rewriting
// ---------------------------------------------------------------------------

func rewriteStructField(field *ast.Field, structName string) bool {
	if len(field.Names) == 0 {
		return false // embedded field
	}
	name := field.Names[0].Name

	// Check blacklist first
	if fieldBlacklist[name] {
		return false
	}

	// Check struct-scoped fields (only allowed in specific structs)
	if allowedStructs, scoped := structScopedFields[name]; scoped {
		if !allowedStructs[structName] {
			return false
		}
	}

	// Check collection whitelist (matches field name + type + struct scope)
	if spec, ok := collectionWhitelist[name]; ok {
		// MemberBudgets is struct-scoped via structScopedFields above,
		// so if we reach here, the struct check already passed.
		oldType := typeString(field.Type)
		if oldType == spec[0] {
			field.Type = parseTypeExpr(spec[1])
			return true
		}
	}

	// Check regular field whitelist
	allowedTypes, ok := fieldWhitelist[name]
	if !ok {
		return false
	}

	// Special case: ID field in excluded structs
	if name == "ID" && excludeStructs[structName] {
		return false
	}

	oldType := typeString(field.Type)
	if !allowedTypes[oldType] {
		return false
	}

	// Determine new type
	if strings.HasPrefix(oldType, "*") {
		field.Type = &ast.StarExpr{X: uuidSelector()}
	} else {
		field.Type = uuidSelector()
	}
	return true
}

// ---------------------------------------------------------------------------
// Function parameter rewriting
// ---------------------------------------------------------------------------

func rewriteFuncParam(field *ast.Field) bool {
	// If multiple params share a type (e.g., `a, b string`), only convert
	// if ALL param names are in the whitelist with the same expected type.
	// Otherwise skip to avoid breaking one param while corrupting another.
	if len(field.Names) == 0 {
		return false
	}

	actualType := typeString(field.Type)
	allMatch := true
	anyMatch := false

	for _, ident := range field.Names {
		expectedOldType, ok := paramWhitelist[ident.Name]
		if !ok || expectedOldType != actualType {
			allMatch = false
		} else {
			anyMatch = true
		}
	}

	// Only convert if all names in this field list match the whitelist
	if !anyMatch || !allMatch {
		return false
	}

	field.Type = uuidSelector()
	return true
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// typeString returns a simple string representation of a type expression.
func typeString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + typeString(t.X)
	case *ast.ArrayType:
		if t.Len == nil {
			return "[]" + typeString(t.Elt)
		}
		return "[...]" + typeString(t.Elt)
	case *ast.MapType:
		return "map[" + typeString(t.Key) + "]" + typeString(t.Value)
	case *ast.SelectorExpr:
		return typeString(t.X) + "." + t.Sel.Name
	default:
		return ""
	}
}

// uuidSelector returns an AST node for `uuid.UUID`.
func uuidSelector() ast.Expr {
	return &ast.SelectorExpr{
		X:   ast.NewIdent("uuid"),
		Sel: ast.NewIdent("UUID"),
	}
}

// parseTypeExpr creates an AST expression from a simple type string.
func parseTypeExpr(s string) ast.Expr {
	switch {
	case s == "[]uuid.UUID":
		return &ast.ArrayType{Elt: uuidSelector()}
	case s == "map[uuid.UUID]float64":
		return &ast.MapType{Key: uuidSelector(), Value: ast.NewIdent("float64")}
	case s == "*uuid.UUID":
		return &ast.StarExpr{X: uuidSelector()}
	case s == "uuid.UUID":
		return uuidSelector()
	default:
		return ast.NewIdent(s)
	}
}

// addUUIDImport ensures `"github.com/google/uuid"` is in the import list.
func addUUIDImport(f *ast.File) {
	const uuidPath = `"github.com/google/uuid"`

	// Check if already imported
	for _, imp := range f.Imports {
		if imp.Path.Value == uuidPath {
			return
		}
	}

	// Add import
	newImport := &ast.ImportSpec{
		Path: &ast.BasicLit{Kind: token.STRING, Value: uuidPath},
	}

	// Find or create import declaration
	for _, decl := range f.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.IMPORT {
			continue
		}
		genDecl.Specs = append(genDecl.Specs, newImport)
		return
	}

	// No import block exists, create one
	importDecl := &ast.GenDecl{
		Tok:   token.IMPORT,
		Specs: []ast.Spec{newImport},
	}
	f.Decls = append([]ast.Decl{importDecl}, f.Decls...)
}
