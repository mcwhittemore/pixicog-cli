package main

import (
	"bytes"
	"crypto/rand"
	"crypto/sha1"
	"errors"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
)

type localError struct {
	msg   string
	cause string
}

func (l *localError) Error() string {
	return fmt.Sprintf("Message: %s\nCause: %s", l.msg, l.cause)
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("You must provide an file for pixicog to run.")
		os.Exit(1)
	}

	args := os.Args[2:]

	msg, err := run(os.Args[1], args)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	fmt.Printf("%s", msg)
}

func run(fn string, args []string) ([]byte, error) {
	mainFile, err := buildMainFile(fn)
	if err != nil {
		return nil, &localError{"Building Application", err.Error()}
	}

	tmpName, err := saveTemp([]byte(mainFile), fn)
	if err != nil {
		os.Remove(tmpName)
		return nil, &localError{"Saving Temp", err.Error()}
	}
	log.Println(tmpName)

	msg, err := gorun(tmpName, args)
	os.Remove(tmpName)
	if err != nil {
		return nil, err
	}

	return msg, nil
}

func buildMainFile(fn string) (string, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, fn, nil, 0)
	if err != nil {
		return "", err
	}

	funcs := getFuncs(f, fset)

	err = checkImports(f, fset, runnerImports)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	format.Node(&buf, fset, f)

	mainStr := buildMainFunc(funcs)
	if err != nil {
		return "", err
	}

	ctx := fmt.Sprintf("%s\n%s\n%s", buf.String(), mainStr, runnerBody)

	return ctx, nil
}

func buildMainFunc(funcs [][]string) string {
	body := buildMainBody(funcs)
	return fmt.Sprintf(`func main() {%s
}`, body)
}

func buildMainBody(funcs [][]string) string {
	checkedFuncs := ""

	for i := 0; i < len(funcs); i++ {
		checkedFuncs += fmt.Sprintf(`
		toRun, err = cog.HasPreviousRun("%s")
		if err != nil {
			panic(err)
		}
		if toRun == false {
		  cog.%s()
			cog.SaveState()
		}
    `, funcs[i][1], funcs[i][0])
	}

	return fmt.Sprintf(`
    var toRun bool
		var err error
		cog := NewRunner(os.Args[1:])
    %s`, checkedFuncs)
}

func checkImports(file *ast.File, fset *token.FileSet, imports []string) error {

	data := make(map[string]struct{})

	holder := struct{}{}

	for _, is := range file.Imports {
		data[is.Path.Value] = holder
	}

	missing := make([]string, 0)
	for _, v := range imports {
		_, ok := data[fmt.Sprintf("\"%s\"", v)]
		if ok == false {
			missing = append(missing, v)
		}
	}

	if len(missing) > 0 {
		return errors.New(fmt.Sprintf("Missing required imports: %s", strings.Join(missing, ", ")))
	}

	return nil
}

func getFuncs(file *ast.File, fset *token.FileSet) [][]string {
	var funcs [][]string

	for _, d := range file.Decls {
		if f, ok := d.(*ast.FuncDecl); ok {
			var oneFunc = make([]string, 2)
			oneFunc[0] = f.Name.Name
			oneFunc[1] = getFuncHash(f, fset)
			if strings.Index(oneFunc[0], "Process") == 0 {
				funcs = append(funcs, oneFunc)
			}

		}
	}

	return funcs
}

func getFuncHash(fun *ast.FuncDecl, fset *token.FileSet) string {
	var buf bytes.Buffer
	format.Node(&buf, fset, fun)

	h := sha1.New()
	h.Write([]byte(buf.String()))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func randomID() (string, error) {
	b := make([]byte, 10)
	_, err := rand.Read(b)
	if err != nil {
		return "", nil
	}

	h := sha1.New()
	h.Write(b)
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func saveTemp(content []byte, fn string) (string, error) {
	id, err := randomID()
	if err != nil {
		return "", err
	}

	fnLen := len(fn)
	prefix := fn[:fnLen-3]

	tmpFileName := fmt.Sprintf("%s.%s.go", prefix, id)

	err = ioutil.WriteFile(tmpFileName, content, 0644)
	if err != nil {
		return "", err
	}

	return tmpFileName, nil
}

func gorun(fn string, args []string) ([]byte, error) {
	ps := string(os.PathSeparator)
	parts := strings.Split(fn, ps)
	pn := len(parts)
	dir := strings.Join(parts[:pn-1], ps)
	args = append([]string{"run", parts[pn-1]}, args...)

	cmd := exec.Command("go", args...)
	cwd, err := os.Getwd()
	if err != nil {
		return nil, &localError{"Could not get CWD", err.Error()}
	}

	cmd.Dir = fmt.Sprintf("%s%s%s", cwd, ps, dir)
	fmt.Println(args, cmd.Dir)
	stderr, err := cmd.StderrPipe()
	stdout, err := cmd.StdoutPipe()

	if err := cmd.Start(); err != nil {
		return nil, &localError{"Failed to start application", err.Error()}
	}

	errStr, err := ioutil.ReadAll(stderr)
	if err != nil {
		return nil, &localError{"Failed to read stderr from application", err.Error()}
	}

	outStr, err := ioutil.ReadAll(stdout)
	if err != nil {
		return nil, &localError{"Failed to read stdout application", err.Error()}
	}

	if err := cmd.Wait(); err != nil {
		cause := fmt.Sprintf("Application exited: %d", cmd.ProcessState.ExitCode())
		return nil, &localError{cause, string(errStr)}
	}
	return outStr, nil
}
