package main
var runnerImports = []string{"crypto/sha1","encoding/base64","fmt","github.com/mcwhittemore/pixicog","image","image/color","io/ioutil","os","strconv","strings"}
var runnerBody = `type Runner struct {
	state       map[string]pixicog.ImageList
	args        []string
	hash        string
	stateLoaded bool
}

func NewRunner(args []string) Runner {
	state := make(map[string]pixicog.ImageList)
	hash := strings.Join(args, "-")
	return Runner{state, args, hash, false}
}

func (r *Runner) CacheFile(hash string) string {
	return fmt.Sprintf("cache/%s.pgs", hash)
}

func (r *Runner) HasPreviousRun(fnHash string) (bool, error) {

	nHash := createHash(r.hash, fnHash)
	cacheFile := r.CacheFile(nHash)

	if r.stateLoaded == true {
		r.hash = nHash
		return false, nil
	}

	hasHash := fileExists(cacheFile)
	if hasHash || r.stateLoaded == true {
		r.hash = nHash
		return true, nil
	}

	err := r.LoadState(r.CacheFile(r.hash))
	if err != nil {
		return true, err
	}
	r.hash = nHash

	return false, nil
}

func (r *Runner) SaveState() error {
	filename := r.CacheFile(r.hash)
	fmt.Printf("Saving state: %s\n", filename)

	var items []string
	for k, v := range r.state {
		ilStr := imageListToString(v)
		items = append(items, fmt.Sprintf("%s,%s", k, ilStr))
	}

	out := strings.Join(items, "\n") + "\n"

	return ioutil.WriteFile(filename, []byte(out), 0644)
}

func imageListToString(il pixicog.ImageList) string {
	var imgStrs []string
	n := len(il)
	for i := 0; i < n; i++ {
		imgStrs = append(imgStrs, imageToStr(il[i]))
	}

	return strings.Join(imgStrs, ",")
}

func imageToStr(img image.Image) string {
	bounds := img.Bounds()

	width := bounds.Max.X - bounds.Min.X
	height := bounds.Max.Y - bounds.Min.Y
	chanels := 4

	header := fmt.Sprintf("%d|%d|%d", width, height, chanels)

	var parts []byte

	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			c := img.At(x, y)
			r, g, b, a := c.RGBA()
			parts = append(parts, uint8(r), uint8(g), uint8(b), uint8(a))
		}
	}

	imgStr := base64.StdEncoding.WithPadding(base64.NoPadding).EncodeToString(parts)

	return fmt.Sprintf("%s|%s", header, imgStr)

}

func (r *Runner) LoadState(filename string) error {
	r.stateLoaded = true
	if fileExists(filename) == false {
		return nil
	}

	fmt.Printf("Loading state from file: %s\n", filename)

	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")

	for _, l := range lines {
		parts := strings.Split(l, ",")
		name := parts[0]
		imgs := parts[1:]

		il, err := loadImageList(imgs)
		if err != nil {
			return err
		}
		r.state[name] = il
	}

	return nil
}

func loadImageList(imgs []string) (pixicog.ImageList, error) {
	il := pixicog.ImageList{}

	for _, info := range imgs {

		parts := strings.Split(info, "|")

		width, err := strconv.Atoi(parts[0])
		if err != nil {
			return nil, err
		}

		height, err := strconv.Atoi(parts[1])
		if err != nil {
			return nil, err
		}

		channels, err := strconv.Atoi(parts[2])
		if err != nil {
			return nil, err
		}

		rect := image.Rect(0, 0, width, height)
		img := image.NewRGBA(rect)

		data, err := base64.StdEncoding.WithPadding(base64.NoPadding).DecodeString(parts[3])
		if err != nil {
			return nil, err
		}

		x := 0
		y := 0
		for c := 0; c < len(data); c += channels {
			r := uint8(data[c])
			g := uint8(data[c+1])
			b := uint8(data[c+2])
			a := uint8(data[c+3])

			clr := color.RGBA{r, g, b, a}

			img.Set(x, y, clr)

			y++

			if y == height {
				y = 0
				x++
			}

		}

		il = append(il, img)
	}

	return il, nil
}

func createHash(h1, h2 string) string {
	h := sha1.New()
	h.Write([]byte(h1))
	h.Write([]byte(h2))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}`