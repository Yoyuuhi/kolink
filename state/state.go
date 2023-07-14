package state

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"yoyuuhi/kolink/request"

	"github.com/juju/errors"
)

type FuncState struct {
	File     string
	Function string
}

func GenerateStateMap(repositoryName string, requestDef request.RequestDef) (map[string][]FuncState, error) {
	// collect type and func/method in callee files
	callee := requestDef.Callee
	calleeDir := callee.Dir
	calleeFileFuncMap := map[string]map[string]bool{}
	calleeFuncFileMap := map[string]string{}
	calleeTypeFileMap := map[string]string{}
	if err := filepath.Walk(calleeDir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		fileName := info.Name()
		if !regexp.MustCompile(`.+(\.go)`).Match([]byte(fileName)) {
			return nil
		}
		if callee.IgnoreTest && regexp.MustCompile(`.+(test)`).Match([]byte(fileName)) {
			return nil
		}
		if len(callee.FocusFileMap) > 0 && !callee.FocusFileMap[fileName] {
			return nil
		}

		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			return errors.Trace(err)
		}

		funcMap := map[string]bool{}
		for _, decl := range file.Decls {
			funcDecl, ok := decl.(*ast.FuncDecl)
			if !ok {
				continue
			}

			if callee.IgnoreNewFunc && regexp.MustCompile(`^(New).+`).Match([]byte(funcDecl.Name.Name)) {
				continue
			}
			if len(callee.FocusFuncMap) > 0 && !callee.FocusFuncMap[funcDecl.Name.Name] {
				continue
			}

			if regexp.MustCompile(`^([A-Z]).+`).Match([]byte(funcDecl.Name.Name)) {
				funcMap[funcDecl.Name.Name] = true
				calleeFuncFileMap[funcDecl.Name.Name] = fileName
			}
		}
		if len(funcMap) > 0 {
			calleeFileFuncMap[fileName] = funcMap
		}

		ast.Inspect(file, func(node ast.Node) bool {
			switch node.(type) {
			case *ast.TypeSpec:
				typeSpec := node.(*ast.TypeSpec)
				_, isFunc := typeSpec.Type.(*ast.FuncType)
				if isFunc {
					return true
				}
				if !regexp.MustCompile(`^([A-Z]).+`).Match([]byte(typeSpec.Name.Name)) {
					return true
				}
				calleeTypeFileMap[typeSpec.Name.Name] = fileName
			}
			return true
		})

		return nil
	}); err != nil {
		fmt.Println("Failed to walk directory:", err)
		return nil, errors.Trace(err)
	}

	// analyze caller
	caller := requestDef.Caller
	callerFileStateMap := map[string][]FuncState{}
	callerDir := caller.Dir
	if err := filepath.Walk(callerDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return errors.Trace(err)
		}
		if info.IsDir() {
			return nil
		}
		fileName := info.Name()
		if !regexp.MustCompile(`.+(\.go)`).Match([]byte(fileName)) {
			return nil
		}
		if caller.IgnoreTest && regexp.MustCompile(`.+(test)`).Match([]byte(fileName)) {
			return nil
		}
		if len(caller.FocusFileMap) > 0 && !caller.FocusFileMap[fileName] {
			return nil
		}

		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, path, nil, parser.AllErrors)
		if err != nil {
			return errors.Trace(err)
		}

		variableTypeMap := map[string]string{}
		usedCalleeFiles := map[string]bool{}
		ast.Inspect(file, func(node ast.Node) bool {
			if typeSpec, ok := node.(*ast.TypeSpec); ok {
				if structType, ok := typeSpec.Type.(*ast.StructType); ok {
					for _, field := range structType.Fields.List {
						switch field.Type.(type) {
						case *ast.SelectorExpr:
							selectorExpr := field.Type.(*ast.SelectorExpr)
							if _, e := calleeTypeFileMap[selectorExpr.Sel.Name]; e {
								if len(field.Names) > 0 {
									variableTypeMap[field.Names[0].Name] = selectorExpr.Sel.Name
								}
								usedCalleeFiles[calleeTypeFileMap[selectorExpr.Sel.Name]] = true

								continue
							}
						}
					}
				}
			}

			return true
		})

		packageNameMap := map[string]string{}
		for _, importSpec := range file.Imports {
			importPath, err := strconv.Unquote(importSpec.Path.Value)
			if err != nil {
				return errors.Trace(err)
			}

			splitPath := strings.Split(importPath, "/")
			pkgName := splitPath[len(splitPath)-1]
			if importSpec.Name != nil {
				pkgName = importSpec.Name.Name
			}
			packageNameMap[pkgName] = importPath
		}

		funcStates := []FuncState{}
		ast.Inspect(file, func(node ast.Node) bool {
			switch node.(type) {
			case *ast.SelectorExpr:
				selectorExpr := node.(*ast.SelectorExpr)
				x := selectorExpr.X
				ident := selectorExpr.Sel

				if xSelectorExpr, ok := x.(*ast.SelectorExpr); ok {
					if helperType, e := variableTypeMap[xSelectorExpr.Sel.Name]; e {
						if len(callee.FocusFuncMap) > 0 && !callee.FocusFuncMap[ident.Name] {
							return true
						}
						funcStates = append(funcStates, FuncState{
							File:     calleeTypeFileMap[helperType],
							Function: ident.Name,
						})
					}
				}
			case *ast.CallExpr:
				callExpr := node.(*ast.CallExpr)
				fun := callExpr.Fun

				if selectorExpr, ok := fun.(*ast.SelectorExpr); ok {
					if packageName, ok := selectorExpr.X.(*ast.Ident); ok {
						if _, e := packageNameMap[packageName.Name]; e {
							funcName := selectorExpr.Sel.Name
							if calleeFile, e := calleeFuncFileMap[funcName]; e {
								if len(callee.FocusFuncMap) > 0 && !callee.FocusFuncMap[funcName] {
									return true
								}
								funcStates = append(funcStates, FuncState{
									File:     calleeFile,
									Function: funcName,
								})
							}
						}

						for calleeFile := range usedCalleeFiles {
							for funcName := range calleeFileFuncMap[calleeFile] {
								if funcName == selectorExpr.Sel.Name {
									if calleeFile, e := calleeFuncFileMap[funcName]; e {
										if len(callee.FocusFuncMap) > 0 && !callee.FocusFuncMap[funcName] {
											return true
										}
										funcStates = append(funcStates, FuncState{
											File:     calleeFile,
											Function: funcName,
										})
									}
								}
							}
						}
					}
				}
			}

			return true
		})
		callerFileStateMap[fileName] = funcStates

		return nil
	}); err != nil {
		fmt.Println("Failed to walk directory:", err)
		return nil, errors.Trace(err)
	}
	return callerFileStateMap, nil
}
