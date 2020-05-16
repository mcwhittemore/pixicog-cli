package main

import (
  "github.com/mcwhittemore/pixicog"
  "fmt"
)

func ProcessLoadData(src, state pixicog.ImageList) (pixicog.ImageList, pixicog.ImageList) {
  printer("Hello")
  return src, state
}

func ProcessMergeData(src, state pixicog.ImageList) (pixicog.ImageList, pixicog.ImageList) {
  printer("World")
  return src, state
}

func printer(str string) {
  fmt.Println(str)
}
