// Package main 👍
package main

import (
	"bytes"
	"embed"
	"errors"
	"image/color"
	"image/png"
	"log"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"

	zen "github.com/melonfunction/ebiten-zen"
)

//go:embed tiles.png
var embedded embed.FS

// vars
var (
	WindowWidth  = 640 * 2
	WindowHeight = 480 * 2

	SpriteSheet *zen.SpriteSheet

	alreadyDrew  bool
	surface      *ebiten.Image
	drawDebug    = true
	debugKeyDown = false

	ErrNormalExit = errors.New("Normal exit")
)

// Game implements ebiten.Game interface.
type Game struct{}

// Update proceeds the game state.
// Update is called every tick (1/60 [s] by default).
func (g *Game) Update() error {
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		return ErrNormalExit
	}
	if ebiten.IsKeyPressed(ebiten.KeyD) {
		if !debugKeyDown {
			drawDebug = !drawDebug
			debugKeyDown = true
		}
	} else {
		debugKeyDown = false
	}

	return nil
}

// Draw draws the game screen.
// Draw is called every frame (typically 1/60[s] for 60Hz display).
func (g *Game) Draw(screen *ebiten.Image) {
	w, h := ebiten.WindowSize()

	// draw some tiles to the surface
	if alreadyDrew {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(4, 4)
		op.GeoM.Translate(-float64(w), -float64(h))
		op.GeoM.Rotate(math.Pi / 4)
		screen.DrawImage(surface, op)

		if drawDebug {
			fill := ebiten.NewImageFromImage(SpriteSheet.PaddedImage)
			fill.Fill(color.RGBA{128, 0, 128, 128})
			op = &ebiten.DrawImageOptions{}
			op.GeoM.Scale(8, 8)
			screen.DrawImage(fill, op)

			op = &ebiten.DrawImageOptions{}
			op.GeoM.Scale(8, 8)
			screen.DrawImage(SpriteSheet.PaddedImage, op)
		}

		return
	}
	alreadyDrew = true

	for x := 0; x < w/int(float64(SpriteSheet.SpriteWidth)); x++ {
		for y := 0; y < h/int(float64(SpriteSheet.SpriteHeight)); y++ {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(
				float64(SpriteSheet.SpriteWidth)*float64(x)*2, // add a lil space between tiles
				float64(SpriteSheet.SpriteHeight)*float64(y)*2)
			// surface.DrawImage(SpriteSheet.GetSprite(rand.Int()%2, rand.Int()%2), op)
			surface.DrawImage(SpriteSheet.GetSprite(1, 1+rand.Int()%3), op)
		}
	}
}

// Layout sets window size
func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	WindowWidth = outsideWidth
	WindowHeight = outsideHeight
	return outsideWidth, outsideHeight
}

func main() {
	game := &Game{}
	ebiten.SetWindowSize(WindowWidth, WindowHeight)
	ebiten.SetWindowTitle("Scale and outline example - Press D to toggle debug overlay")
	ebiten.SetWindowResizable(true)

	surface = ebiten.NewImage(ebiten.WindowSize())

	if b, err := embedded.ReadFile("tiles.png"); err == nil {
		if s, err := png.Decode(bytes.NewReader(b)); err == nil {
			sprites := ebiten.NewImageFromImage(s)
			SpriteSheet = zen.NewSpriteSheet(sprites, 8, 8, zen.SpriteSheetOptions{
				Scale:            2,
				OutlineThickness: 1,
				OutlineColor:     color.RGBA{128, 128, 128, 128},
			})
		}
	} else {
		log.Fatal(err)
	}

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
