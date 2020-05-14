package main

import (
  "testing"
  "strings"
)

func TestMain(t *testing.T) {
  msg, err := run("./test-fixtures/job.go", "arg1", "arg2")
  if err != nil {
    t.Fatalf("Unexpected error %v", err)
  }
  exp := "Hello\nWorld\n"
  if string(msg) != exp {
    t.Fatalf("Wrong message. Expected [%s]. Got [%s]", exp, msg)
  }
}

func TestBuildMainFunc(t *testing.T) {
  funcs := [][]string{{"one", "onehash"}, {"two", "twohash"}, {"three", "threehash"} }

  expected := strings.ReplaceAll(`func main() {
    var hash string
    src, state := one(nil, nil)

    shouldRun, src, state := checkProgress(hash, src, state)
    if shouldRun == true {
      src, state = two(src, state)
    }
    hash = saveProgress(hash, "twohash", src, state)

    shouldRun, src, state := checkProgress(hash, src, state)
    if shouldRun == true {
      src, state = three(src, state)
    }
    hash = saveProgress(hash, "threehash", src, state)

}`, " ", "")

  mainFunc := strings.ReplaceAll(buildMainFunc(funcs), " ", "")

  if mainFunc != expected {
    t.Fatalf("Wrong message. Expected [%s]. Got [%s]", expected, mainFunc)
  }
}

