package main

import (
  "fmt"
  "github.com/mcwhittemore/pixicog-go"
)

func ProcessLoadData(src, state pixicog.ImageList) (pixicog.ImageList, pixicog.ImageList) {
  fmt.Println("Hello")
  return src, state
}

func ProcessMergeData(src, state pixicog.ImageList) (pixicog.ImageList, pixicog.ImageList) {
  fmt.Println("World")
  return src, state
}

