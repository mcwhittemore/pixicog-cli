package src

import (
  "github.com/mcwhittemore/pixicog"
  "image/color"
  "image"
  "testing"
  "os"
  "reflect"
)

func TestCache(t *testing.T) {
  r1 := NewRunner([]string{"hi"})

  cog := pixicog.ImageList{}
  white := color.RGBA{255, 255, 255, 255};
  black := color.RGBA{0, 0, 0, 255};
  gray := color.RGBA{128, 128, 128, 255};

  cog = append(cog, FlatImage(1, 1, white)) // white
  cog = append(cog, FlatImage(1, 1, black)) // black
  cog = append(cog, FlatImage(1, 1, gray)) // gray

  r1.state["one"] = cog

  cf := r1.CacheFile("f1")
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

func FlatImage(width, height int, c color.RGBA) *image.RGBA {
  img := image.NewRGBA(image.Rect(0,0,width,height))

  for x := 0; x < width; x++ {
    for y := 0; y < height; y++ {
      img.Set(x, y, c)
    }
  }

  return img
}
