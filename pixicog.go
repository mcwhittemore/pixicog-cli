package main

import (
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

func main() {
  fmt.Println(os.Args[1])
  msg, err := run(os.Args[1], os.Args[2], os.Args[3])
  fmt.Printf("%s", msg)
  if err != nil {
    fmt.Println("err", err)
    panic(1)
  }
}

func run(fn, input, output string) ([]byte, error) {
  mainFile, err := buildMainFile(fn)
  if err != nil {
    return nil, err
  }

  tmpName, err := saveTemp([]byte(mainFile))
  if err != nil {
    os.Remove(tmpName)
    return nil, err
  }
  log.Println(tmpName)

  msg, err := gorun(tmpName, input, output)
  //os.Remove(tmpName)
  if err != nil {
    return nil, err
  }

  return msg, err
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

func saveTemp(content []byte) (string, error) {
  tmpFile, err := ioutil.TempFile("", "pixicog-script*.go")
  if err != nil {
    return "", err
  }

  if _, err := tmpFile.Write(content); err != nil {
    return "", err
  }

  if err := tmpFile.Close(); err != nil {
    return "", err
  }

  return tmpFile.Name(), nil
}


func gorun(fn, input, output string) ([]byte, error) {
  cmd := exec.Command("go", "run", fn, input, output)
  msg, err := cmd.CombinedOutput()
  return msg, err
}
