package main

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"github.com/mcwhittemore/pixicog"
	"image"
	"image/color"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

func (r *Runner) ProcessLoadData() {
	printer("Hello")
}

func (r *Runner) ProcessMergeData() {
	printer("World")
}

func printer(str string) {
	fmt.Println(str)
}
