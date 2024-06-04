// Package main 👍
package main

import (
	"bytes"
	"embed"
	"errors"
	"image/png"
	"log"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"

	zen "github.com/melonfunction/ebiten-zen"
)

//go:embed sprites.png
var embedded embed.FS

// vars
var (
	WindowWidth  = 640 * 2
	WindowHeight = 480 * 2

	SpriteSheet *zen.SpriteSheet
	Animation   *zen.Animation

	spaceReleased = true

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

	if ebiten.IsKeyPressed(ebiten.KeySpace) {
		if spaceReleased {
			log.Println("pausing/playing")
			spaceReleased = false
			if Animation.Paused {
				Animation.Play()
			} else {
				Animation.Pause()
			}
		}
	} else {
		spaceReleased = true
	}

	Animation.Update()

	return nil
}

// Draw draws the game screen.
// Draw is called every frame (typically 1/60[s] for 60Hz display).
func (g *Game) Draw(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}

	// Draw the first frame at the top left corner of the screen
	screen.DrawImage(SpriteSheet.GetSprite(0, 0), op)

	// Draw the animation in the center of the screen
	w, h := SpriteSheet.SpriteWidth, SpriteSheet.SpriteHeight
	op.GeoM.Translate(float64(WindowWidth)/2-float64(w)/2, float64(WindowHeight)/2-float64(h)/2)
	Animation.Draw(screen, op)

	ebitenutil.DebugPrint(screen, "Press space to pause")
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
	ebiten.SetWindowTitle("Animation example")
	ebiten.SetWindowResizable(true)

	if b, err := embedded.ReadFile("sprites.png"); err == nil {
		if s, err := png.Decode(bytes.NewReader(b)); err == nil {
			sprites := ebiten.NewImageFromImage(s)

			SpriteSheet = zen.NewSpriteSheet(sprites, 8, 10, zen.SpriteSheetOptions{
				Scale: 16,
			})

			duration := time.Second / 20 // 20 fps animation
			frames := make([]zen.Frame, 5)
			for x := 0; x <= 4; x++ {
				if x == 4 {
					// Draw the last frame for a bit longer
					frames[x] = zen.NewFrame(SpriteSheet.GetSprite(x, 0), duration*5)
				} else {
					frames[x] = zen.NewFrame(SpriteSheet.GetSprite(x, 0), duration)
				}
			}
			Animation = zen.NewAnimation(frames)
		}
	} else {
		log.Fatal(err)
	}

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
