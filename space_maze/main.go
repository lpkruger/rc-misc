package main

import "fmt"
import "image/color"
import "image/jpeg"
import "image/png"
import "image"
import "sort"
import "os"
import "os/exec"

type XY struct {
	x int
	y int
}

type Path struct {
	xy XY
	next *Path
}

func bounds(xx int, yy int) bool {
	return (xx >= 135 && xx <= 1145 && yy > 0)
}

// 150 is an acceptable threshold for distinguishing white from black
// in this image.
func isit(img *image.RGBA, xx int, yy int) uint8 {
	g, _, _, _ := img.At(xx, yy).RGBA()
	g >>= 8
	if g > 150 {
		return 0xff
	}
	return 0
}

func drawSquare(img *image.RGBA, xx int, yy int, c color.Color) {
	d := 2
	for y := yy - d; y <= yy+d; y++ {
		for x := xx - d; x <= xx+d; x++ {
			if isit(img, x, y) != 0 {
				img.Set(x, y, c)
			}
		}
	}
}

func solve(img *image.RGBA) {
	red := color.RGBA{0xff, 0x10, 0x10, 0xff}

	source_xy := XY{135, 145}
	target_xy := XY{1145, 1285}

	queue := make([]Path, 0)
	queue = append(queue, Path{source_xy, nil})
	visited := make(map[XY]bool)
	solution := queue[0]
	for {
		hd := queue[0]
		//fmt.Println(hd)
		queue = queue[1:]
		xy := hd.xy

		if xy.x == target_xy.x && xy.y == target_xy.y {
			solution = hd
			break
		}

		//fmt.Println(bounds(xy.x, xy.y), isit(img, xy.x, xy.y))

		if !bounds(xy.x, xy.y) || isit(img, xy.x, xy.y) == 0 || visited[xy] {
			continue
		}
		visited[xy] = true

		enq := func(q []Path, x int, y int) []Path {
			if (!bounds(x, y) || isit(img, x, y) == 0 || visited[XY{x, y}]) {
				return q
			}
			newpath := Path{XY{x, y}, &hd}
			return append(q, newpath)
		}

		directions := [...](func(xy XY) XY){
			func(xy XY) XY { return XY{xy.x - 1, xy.y} },
			func(xy XY) XY { return XY{xy.x + 1, xy.y} },
			func(xy XY) XY { return XY{xy.x, xy.y - 1} },
			func(xy XY) XY { return XY{xy.x, xy.y + 1} },
			/*
			func(xy XY) XY { return XY{xy.x + 1, xy.y + 1} },
			func(xy XY) XY { return XY{xy.x + 1, xy.y - 1} },
			func(xy XY) XY { return XY{xy.x - 1, xy.y + 1} },
			func(xy XY) XY { return XY{xy.x - 1, xy.y - 1} },
			*/

		}

		stride := 2

		next_points := make([]XY, 0)
		for _, dir := range directions {
			ok := true
			newxy := xy
			for i := 0; i < stride; i++ {
				newxy = dir(newxy)
				if isit(img, newxy.x, newxy.y) == 0 {
					ok = false
					break
				}
			}

			if ok {
				next_points = append(next_points, newxy)
			}
		}

		pretty_score := func(xy XY) int {
			min := 100000
			for _, dir := range directions {
				count := 0
				newxy := xy
				for {
					if count > min {
						break
					}
					newxy = dir(newxy)
					if isit(img, newxy.x, newxy.y) != 0 && bounds(newxy.x, newxy.y) {
						count++
						continue
					}
					if count < min {
						min = count
					}
					break
				}
			}
			return min
		}

		sort.Slice(next_points, func(i, j int) bool {
			return pretty_score(next_points[i]) > pretty_score(next_points[j])
		})

		for _, newxy := range next_points {
			queue = enq(queue, newxy.x, newxy.y)
		}

		fmt.Println(len(queue))
	}

	for next := &solution; next != nil; next = next.next {
		drawSquare(img, next.xy.x, next.xy.y, red)
	}

	drawSquare(img, 100, 470, red)
	drawSquare(img, target_xy.x, target_xy.y, red)
}

func convert(img image.Image) *image.RGBA {
	b := img.Bounds()
	fmt.Println(b)
	imgSet := image.NewRGBA(b)

	for y := 0; y < b.Max.Y; y++ {
		for x := 0; x < b.Max.X; x++ {
			oldPixel := img.At(x, y)
			r, g, b, _ := oldPixel.RGBA()
			gv := uint8(((r + g + b) / 3) >> 8)

			/*
				if (gv > 150) {
					gv = 0xff;
				} else {
					gv = 0;
				}
			*/

			//pixel := color.RGBA{uint8(r>>8), uint8(g>>8), uint8(b>>8), uint8(a>>8)}
			pixel := color.RGBA{gv, gv, gv, 0xff}
			imgSet.Set(x, y, pixel)
		}
	}

	//fmt.Println(img.At(1500,2000))
	return imgSet
}

func main() {
	file, _ := os.Open("RC-Maze.png")
	defer file.Close()
	img, _ := png.Decode(file)

	img2 := convert(img)
	solve(img2)

	output, _ := os.Create("./solution.jpg")
	defer output.Close()
	jpeg.Encode(output, img2, nil)
	output.Close()

	// on MacOS, this displays the image, remove or change if it causes problems
	exec.Command("open", "./solution.jpg").Run()
}
