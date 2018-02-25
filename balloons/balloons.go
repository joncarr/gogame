package main

import (
	"fmt"
	"image/png"
	"os"
	"time"

	"github.com/veandco/go-sdl2/sdl"
)

const (
	winWidth  = 800
	winHeight = 600
	curDir    = "/home/jec/Code/GoCode/src/github.com/joncarr/gogame/balloons"
)

type texture struct {
	pos
	pixels      []byte
	w, h, pitch int
}

type pos struct {
	x, y float32
}

func clear(pixels []byte) {
	for i := range pixels {
		pixels[i] = 0
	}
}

func (t *texture) draw(pixels []byte) {
	for y := 0; y < t.h; y++ {
		for x := 0; x < t.w; x++ {
			screenY := y + int(t.y)
			screenX := x + int(t.x)
			if screenX >= 0 && screenX < winWidth &&
				screenY >= 0 && screenY < winHeight {
				texIndex := y*t.pitch + x*4
				screenIndex := screenY*winWidth*4 + screenX*4

				pixels[screenIndex] = t.pixels[texIndex]
				pixels[screenIndex+1] = t.pixels[texIndex+1]
				pixels[screenIndex+2] = t.pixels[texIndex+2]
				pixels[screenIndex+3] = t.pixels[texIndex+3]
			}
		}
	}
}

func (t *texture) drawAlpha(pixels []byte) {
	for y := 0; y < t.h; y++ {
		for x := 0; x < t.w; x++ {
			screenY := y + int(t.y)
			screenX := x + int(t.x)
			if screenX >= 0 && screenX < winWidth &&
				screenY >= 0 && screenY < winHeight {
				texIndex := y*t.pitch + x*4
				screenIndex := screenY*winWidth*4 + screenX*4

				srcR := int(t.pixels[texIndex])
				srcG := int(t.pixels[texIndex+1])
				srcB := int(t.pixels[texIndex+2])
				srcA := int(t.pixels[texIndex+3])

				dstR := int(pixels[screenIndex])
				dstG := int(pixels[screenIndex+1])
				dstB := int(pixels[screenIndex+2])
				// dstA := int(pixels[texIndex+3])

				rstR := (srcR*255 + dstR*(255-srcA)) / 255
				rstG := (srcG*255 + dstG*(255-srcA)) / 255
				rstB := (srcB*255 + dstB*(255-srcA)) / 255

				pixels[screenIndex] = byte(rstR)
				pixels[screenIndex+1] = byte(rstG)
				pixels[screenIndex+2] = byte(rstB)
			}
		}
	}
}

type rgba struct {
	r, g, b byte
}

func setPixel(x, y int, c rgba, pixels []byte) {

	index := (y*winWidth + x) * 4

	if index < len(pixels)-4 && index >= 0 {
		pixels[index] = c.r
		pixels[index+1] = c.g
		pixels[index+2] = c.b
	}
}

func loadBalloons() []texture {

	balloonStr := []string{
		"balloon_red.png",
		"balloon_blue.png",
		"balloon_green.png",
	}

	balloonTextures := make([]texture, len(balloonStr))

	for i := range balloonStr {
		file, err := os.Open(curDir +
			"/assets/sprites/" + balloonStr[i])
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}
		defer file.Close()

		image, err := png.Decode(file)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}

		w := image.Bounds().Max.X
		h := image.Bounds().Max.Y

		balloonPixels := make([]byte, w*h*4)
		bIndex := 0
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				r, g, b, a := image.At(x, y).RGBA()
				balloonPixels[bIndex] = byte(r / 256)
				bIndex++
				balloonPixels[bIndex] = byte(g / 256)
				bIndex++
				balloonPixels[bIndex] = byte(b / 256)
				bIndex++
				balloonPixels[bIndex] = byte(a / 256)
				bIndex++
			}
		}

		balloonTextures[i] = texture{pos{0, 0}, balloonPixels, w, h, w * 4}
	}
	return balloonTextures
}

func main() {

	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		panic(err)
	}
	defer sdl.Quit()

	window, err := sdl.CreateWindow("BA-1100-NS", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, winWidth, winHeight, sdl.WINDOW_SHOWN)
	if err != nil {
		panic(err)
	}
	defer window.Destroy()

	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		panic(err)
	}
	defer renderer.Destroy()

	texture, err := renderer.CreateTexture(sdl.PIXELFORMAT_ABGR8888, sdl.TEXTUREACCESS_STREAMING, winWidth, winHeight)
	if err != nil {
		panic(err)
	}
	defer texture.Destroy()

	pixels := make([]byte, winWidth*winHeight*4)
	balloons := loadBalloons()
	dir := 1

	for {

		frameStart := time.Now()

		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch event.(type) {
			case *sdl.QuitEvent:
				return
			}
		}
		clear(pixels)
		for _, tex := range balloons {
			tex.drawAlpha(pixels)
		}

		balloons[1].x += float32(1 * dir)
		if balloons[1].x > 400 || balloons[1].x < 0 {
			dir = dir * -1
		}

		texture.Update(nil, pixels, winWidth*4)
		renderer.Copy(texture, nil, nil)
		renderer.Present()

		elapsedTime := float32(time.Since(frameStart).Seconds() * 1000)
		fmt.Println("ms per frame:", elapsedTime)
		if elapsedTime < 5 {
			sdl.Delay(5 - uint32(elapsedTime))
			elapsedTime = float32(time.Since(frameStart).Seconds())
		}
	}

}
