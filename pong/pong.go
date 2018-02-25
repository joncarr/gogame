package main

//TODO: (jon)
// Refactor / optimize
// AI is invincible, currently unnbeatable
// Handle window resizing
// use bitmaps for graphics?
// Add ability for 2 player?

import (
	"math"
	"time"

	"github.com/joncarr/gogame/noise"
	"github.com/veandco/go-sdl2/sdl"
)

const (
	winWidth  = 800
	winHeight = 600
)

type GameState int

const (
	Start GameState = iota
	Play
)

var state = Start

var nums = [][]byte{
	{
		1, 1, 1,
		1, 0, 1,
		1, 0, 1,
		1, 0, 1,
		1, 1, 1,
	},
	{
		1, 1, 0,
		0, 1, 0,
		0, 1, 0,
		0, 1, 0,
		1, 1, 1,
	},
	{
		1, 1, 1,
		0, 0, 1,
		1, 1, 1,
		1, 0, 0,
		1, 1, 1,
	},
	{
		1, 1, 1,
		0, 0, 1,
		1, 1, 1,
		0, 0, 1,
		1, 1, 1,
	},
}

// Color represented as RGB value
type Color struct {
	r, g, b byte
}

// Pos represents 2 dimensional location (x, y)
type Pos struct {
	x, y float32
}

// Ball is the pong ball.
// pos - is the ball's position
// radius - ball's radius
// xv - is the ball's x axis velocity
// yv - is the ball's y axis velocity
// color - ball color (RGB)
type Ball struct {
	Pos
	radius float32
	xv     float32
	yv     float32
	color  Color
}

func drawNumber(p Pos, c Color, size int, num int, pixels []byte) {
	startX := int(p.x) - (size*3)/2
	startY := int(p.y) - (size*5)/2

	for i, v := range nums[num] {
		if v == 1 {
			for y := startY; y < startY+size; y++ {
				for x := startX; x < startX+size; x++ {
					setPixel(x, y, c, pixels)
				}
			}
		}
		startX += size
		if (i+1)%3 == 0 {
			startY += size
			startX -= size * 3
		}
	}
}

func (b *Ball) draw(pixels []byte) {
	for y := -b.radius; y < b.radius; y++ {
		for x := -b.radius; x < b.radius; x++ {
			if x*x+y*y < b.radius*b.radius {
				setPixel(int(b.x+x), int(b.y+y), b.color, pixels)
			}
		}
	}
}

func getCenter() Pos {
	return Pos{
		float32(winWidth / 2),
		float32(winHeight / 2),
	}
}

func (b *Ball) update(pLeft, pRight *Paddle, elapsedTime float32) {
	b.x += b.xv * elapsedTime
	b.y += b.yv * elapsedTime

	if b.y-b.radius < 0 || b.y+b.radius > winHeight {
		b.yv = -b.yv
	}

	if b.x < 0 {
		pRight.score++
		b.Pos = getCenter()
		state = Start
	} else if b.x > winWidth {
		pLeft.score++
		b.Pos = getCenter()
		state = Start
	}

	if b.x-b.radius < pLeft.x+pLeft.w/2 {
		if b.y > pLeft.y-pLeft.h/2 && b.y < pLeft.y+pLeft.h/2 {
			b.xv = -b.xv
			b.x = pLeft.x + pLeft.w/2.0 + b.radius
		}
	}

	if b.x+b.radius > pRight.x+pRight.w/2 {
		if b.y > pRight.y-pRight.h/2 && b.y < pRight.y+pRight.h/2 {
			b.xv = -b.xv
			b.x = pRight.x - pRight.w/2.0 - b.radius

		}
	}

}

// Paddle is a game paddle
// Pos - the paddles position
// w - widtrh of the paddle
// h - height of the paddle
// color - the paddles' color
type Paddle struct {
	Pos
	w     float32
	h     float32
	speed float32
	score int
	color Color
}

// Lerp is a linear interpolation helper, it
// returns value of start point 'a' plus the provided
// percentage 'pct' times the difference of start point
// 'a' and end point 'b'
func flerp(a, b, pct float32) float32 {
	return a + pct*(b-a)
}

func (p *Paddle) draw(pixels []byte) {
	startX := int(p.x - p.w/2)
	startY := int(p.y - p.h/2)

	for y := 0; y < int(p.h); y++ {
		for x := 0; x < int(p.w); x++ {
			setPixel(startX+x, startY+y, p.color, pixels)
		}
	}

	numX := flerp(p.x, getCenter().x, 0.5)
	drawNumber(Pos{numX, 35}, p.color, 10, p.score, pixels)

}

func (p *Paddle) update(keyState []uint8, controllerAxis int16, elapsedTime float32) {
	if keyState[sdl.SCANCODE_UP] != 0 {
		p.y -= p.speed * elapsedTime
	}
	if keyState[sdl.SCANCODE_DOWN] != 0 {
		p.y += p.speed * elapsedTime
	}

	// Joystick calculation
	if math.Abs(float64(controllerAxis)) > 1500 {
		pct := float32(controllerAxis) / 32767.0
		p.y += p.speed * pct * elapsedTime
	}
}

func (p *Paddle) aiUpdate(b *Ball, elapsedTime float32) {
	p.y = b.y
}

func clear(pixels []byte) {
	for i := range pixels {
		pixels[i] = 0
	}
}

func lerp(b1, b2 byte, pct float32) byte {
	return byte(float32(b1) + pct*(float32(b2)-float32(b1)))
}

func colorLerp(c1, c2 Color, pct float32) Color {
	return Color{
		lerp(c1.r, c2.r, pct),
		lerp(c1.g, c2.g, pct),
		lerp(c1.b, c2.b, pct),
	}
}

func getGradient(c1, c2 Color) []Color {
	r := make([]Color, 256)
	for i := range r {
		pct := float32(i) / float32(255)
		r[i] = colorLerp(c1, c2, pct)
	}
	return r
}

func getDualGradient(c1, c2, c3, c4 Color) []Color {
	r := make([]Color, 256)
	for i := range r {
		pct := float32(i) / float32(255)
		if pct < 0.5 {
			r[i] = colorLerp(c1, c2, pct*float32(2))
		} else {
			r[i] = colorLerp(c3, c4, pct*float32(1.5)-float32(0.5))
		}

	}
	return r
}

func clamp(min, max, v int) int {
	if v < min {
		v = min
	} else if v > max {
		v = max
	}
	return v
}

func rescaleAndDraw(noise []float32, min, max float32, gradient []Color, width, height int) []byte {
	result := make([]byte, width*height*4)
	scale := 255.0 / (max - min)
	offset := min * scale

	for i := range noise {
		noise[i] = noise[i]*scale - offset
		c := gradient[clamp(0, 255, int(noise[i]))]
		p := i * 4
		result[p] = c.r
		result[p+1] = c.g
		result[p+2] = c.b
	}
	return result
}

func setPixel(x, y int, c Color, pixels []byte) {

	index := (y*winWidth + x) * 4

	if index < len(pixels)-4 && index >= 0 {
		pixels[index] = c.r
		pixels[index+1] = c.g
		pixels[index+2] = c.b
	}
}

func main() {

	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		panic(err)
	}
	defer sdl.Quit()

	window, err := sdl.CreateWindow("PONKEY PONG", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, winWidth, winHeight, sdl.WINDOW_SHOWN)
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

	var controllerHandlers []*sdl.GameController
	for i := 0; i < sdl.NumJoysticks(); i++ {
		controllerHandlers = append(controllerHandlers, sdl.GameControllerOpen(i))
		defer controllerHandlers[i].Close()
	}

	pixels := make([]byte, winWidth*winHeight*4)

	p1 := Paddle{Pos{50, 100}, 20, 100, 400, 0, Color{225, 225, 225}}
	p2 := Paddle{Pos{float32(winWidth) - 50, 5}, 20, 100, 400, 0, Color{225, 225, 225}}
	ball := Ball{Pos{300, 300}, 15, 400, 400, Color{225, 225, 225}}

	keyState := sdl.GetKeyboardState()

	noise, min, max := noise.MakeNoise(noise.FBM, .01, 0.2, 2, 3, winWidth, winHeight)
	gradient := getGradient(Color{255, 0, 0}, Color{0, 0, 0})
	noisePixels := rescaleAndDraw(noise, min, max, gradient, winWidth, winHeight)

	var frameStart time.Time
	var elapsedTime float32
	var controllerAxis int16

	// Infinite loop
	for {
		frameStart = time.Now()

		// Event polling loop
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch event.(type) {
			case *sdl.QuitEvent:
				return
			}
		}

		// Check joysticks
		for _, controller := range controllerHandlers {
			if controller != nil {
				controllerAxis = controller.GetAxis(sdl.CONTROLLER_AXIS_LEFTY)
			}
		}

		if state == Play {
			p1.update(keyState, controllerAxis, elapsedTime)
			p2.aiUpdate(&ball, elapsedTime)
			ball.update(&p1, &p2, elapsedTime)
		} else if state == Start {
			if keyState[sdl.SCANCODE_SPACE] != 0 {
				if p1.score == 3 || p2.score == 3 {
					p1.score = 0
					p2.score = 0
				}
				state = Play
			}
		}

		for i := range noisePixels {
			pixels[i] = noisePixels[i]
		}
		p1.draw(pixels)
		p2.draw(pixels)
		ball.draw(pixels)

		texture.Update(nil, pixels, winWidth*4)
		renderer.Copy(texture, nil, nil)
		renderer.Present()

		elapsedTime = float32(time.Since(frameStart).Seconds())

		if elapsedTime < .005 {
			sdl.Delay(5 - uint32(elapsedTime*1000.0))
			elapsedTime = float32(time.Since(frameStart).Seconds())
		}

	}

}
