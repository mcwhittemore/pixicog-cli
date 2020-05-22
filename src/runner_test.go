package src

import (
	"fmt"
	"github.com/mcwhittemore/pixicog"
	"image"
	"image/color"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestCache(t *testing.T) {
	r1 := NewRunner([]string{"hi"})

	cog := pixicog.ImageList{}
	white := color.RGBA{255, 255, 255, 255}
	black := color.RGBA{0, 0, 0, 255}
	gray := color.RGBA{128, 128, 128, 255}

	cog = append(cog, FlatImage(1, 1, white)) // white
	cog = append(cog, FlatImage(1, 1, black)) // black
	cog = append(cog, FlatImage(1, 1, gray))  // gray

	r1.state["one"] = cog

	cf := r1.CacheFile(createHash(r1.hash, "f1"))

	// incase it wasn't removed last time for some reason
	os.Remove(cf)

	did, err := r1.HasPreviousRun("f1")
	if err != nil {
		t.Fatalf("Unexpected Error: %v", err)
	}
	if did {
		t.Fatalf("Expected true but got false")
	}

	err = r1.SaveState()
	if err != nil {
		t.Fatalf("Unexpected Error: %v", err)
	}

	r2 := NewRunner([]string{"hi"})

	did, err = r2.HasPreviousRun("f1")
	if err != nil {
		t.Fatalf("Unexpected Error: %v", err)
	}
	if did == false {
		t.Fatalf("Expected false but got true")
	}

	did, err = r2.HasPreviousRun("f2")
	if err != nil {
		t.Fatalf("Unexpected Error: %v", err)
	}
	if did == true {
		t.Fatalf("Expected true but got false")
	}

	if reflect.DeepEqual(r1.state, r2.state) == false {
		t.Fatalf("Expected: %v, Received: %v\n", r1.state, r2.state)
	}

	os.Remove(cf)
}

func TestCopy(t *testing.T) {
	data, err := ioutil.ReadFile("runner.go")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	file := string(data)

	importS := strings.Index(file, "(")
	importE := strings.Index(file, ")")
	imports := strings.Split(strings.TrimSpace(file[importS+1:importE]), "\n")

	for i, s := range imports {
		imports[i] = strings.TrimSpace(s)
	}

	importStr := fmt.Sprintf("var runnerImports = []string{%s}", strings.Join(imports, ","))

	bodyStr := fmt.Sprintf("var runnerBody = `%s`", strings.TrimSpace(file[importE+1:]))

	payload := fmt.Sprintf("package main\n%s\n%s", importStr, bodyStr)

	ioutil.WriteFile("../runner_built.go", []byte(payload), 0664)
}

func FlatImage(width, height int, c color.RGBA) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			img.Set(x, y, c)
		}
	}

	return img
}
