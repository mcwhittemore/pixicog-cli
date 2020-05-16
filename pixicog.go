package main

import (
  "encoding/base64"
	"crypto/rand"
  "strings"
  "log"
  "bytes"
  "go/ast"
	"go/parser"
	"go/token"
  "go/format"
  "fmt"
  "os"
  "os/exec"
  "crypto/sha1"
  "io/ioutil"
)

type localError struct {
  msg string
  cause string
}

func (l *localError) Error() string {
  return fmt.Sprintf("Message: %s\nCause: %s", l.msg, l.cause)
}

func main() {
  if len(os.Args) < 2 {
    fmt.Println("You must provide an file for pixicog to run.");
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

  funcs := getFuncs(f, fset);

  var buf bytes.Buffer
  format.Node(&buf, fset, f)

  mainStr := buildMainFunc(funcs)
  if err != nil {
    return "", err
  }

  helpers := `
  func checkProgress(hash string, src, state pixicog.ImageList) (bool, pixicog.ImageList, pixicog.ImageList) {
    return true, src, state
  }

  func saveProgress(hash, funcHash string, src, state pixicog.ImageList) string {
    return hash + funcHash
  }
  `

  ctx := fmt.Sprintf("%s\n%s\n%s", buf.String(), mainStr, helpers)

  return ctx, nil
}

func buildMainFunc(funcs [][]string) (string) {
  body := buildMainBody(funcs)
  return fmt.Sprintf(`func main() {%s
}`, body)
}

func buildMainBody(funcs [][]string) string {
  checkedFuncs := ""

  for i := 1; i< len(funcs); i++ {
    checkedFuncs += fmt.Sprintf(`
    shouldRun, src, state = checkProgress(hash, src, state)
    if shouldRun == true {
      src, state = %s(src, state)
    }
    hash = saveProgress(hash, "%s", src, state)
    `, funcs[i][0], funcs[i][1])
  }

  return fmt.Sprintf(`
    var hash string
    var shouldRun bool
    src, state := %s(nil, nil)
    %s`, funcs[0][0], checkedFuncs);
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

func randomId() (string, error) {
  b := make([]byte, 10)
	_, err := rand.Read(b)
	if err != nil {
		return "", nil
	}
  return base64.StdEncoding.WithPadding(-1).EncodeToString(b), nil
}

func saveTemp(content []byte, fn string) (string, error) {
  id, err := randomId()
  if err != nil {
    return "", err
  }

  fnLen := len(fn)
  prefix := fn[:fnLen-3]

  tmpFileName := fmt.Sprintf("%s.%s.go", prefix, id);

  err = ioutil.WriteFile(tmpFileName, content, 0644)
  if err != nil {
    return "", err
  }

  return tmpFileName, nil
}


func gorun(fn string, args []string) ([]byte, error) {
  ps := string(os.PathSeparator);
  parts := strings.Split(fn, ps)
  pn := len(parts)
  dir := strings.Join(parts[:pn-1], ps)
  args = append([]string{"run", parts[pn-1]}, args...);

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
