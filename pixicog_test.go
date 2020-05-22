package main

import (
	"os"
	"testing"
)

func TestMain(t *testing.T) {
	msg, err := run("./test-fixtures/job.go", []string{"arg1", "arg2"})
	if err != nil {
		t.Fatalf("Unexpected error %v", err)
	}
	exp := "Hello\nSaving state: cache/15f978c84d9d12bc033c6e8e9602404b9f0ac559.pgs\nWorld\nSaving state: cache/ba1553157af7908efcd4de57065af5a50d645335.pgs\n"

	os.Remove("./test-fixtures/cache/15f978c84d9d12bc033c6e8e9602404b9f0ac559.pgs")
	os.Remove("./test-fixtures/cache/ba1553157af7908efcd4de57065af5a50d645335.pgs")

	if string(msg) != exp {
		t.Fatalf("Wrong message. Expected [%s]. Got [%s]", exp, msg)
	}

}
