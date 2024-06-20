// Package main
package main

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"image/color"
	"image/png"
	"log"
	"math"
	"math/rand/v2"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	zen "github.com/melonfunction/ebiten-zen"
)

//go:embed car.png
//go:embed tiles.png
//go:embed tall_wall.png
var embedded embed.FS

// vars
var (
	camera *zen.Camera

	floor       *zen.Floor
	wall        *zen.Wall
	tallWall    *zen.Wall
	spriteStack *zen.SpriteStack
	billboard   *zen.Billboard
	stress      []zen.IsometricDrawable

	WindowWidth              = 640 * 2
	WindowHeight             = 480 * 2
	SpriteSheetTiles         *zen.SpriteSheet
	SpriteSheetCar           *zen.SpriteSheet
	SpriteSheetTallWallSides *zen.SpriteSheet
	SpriteSheetTallWallTop   *zen.SpriteSheet

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
	if ebiten.IsKeyPressed(ebiten.KeyG) {
		camera.RotateWorld(-0.025)
		// spriteStack.Rotation -= 0.025
	}
	if ebiten.IsKeyPressed(ebiten.KeyR) {
		camera.RotateWorld(0.025)
		// spriteStack.Rotation += 0.025
	}

	// camera movement isn't adjusted by camera rotation here
	if ebiten.IsKeyPressed(ebiten.KeyH) {
		camera.MovePosition(-10, 0)
	}
	if ebiten.IsKeyPressed(ebiten.KeyN) {
		camera.MovePosition(10, 0)
	}
	if ebiten.IsKeyPressed(ebiten.KeyC) {
		camera.MovePosition(0, -10)
	}
	if ebiten.IsKeyPressed(ebiten.KeyT) {
		camera.MovePosition(0, 10)
	}

	return nil
}

// Draw draws the game screen.
// Draw is called every frame (typically 1/60[s] for 60Hz display).
func (g *Game) Draw(screen *ebiten.Image) {
	camera.Surface.Clear()
	floor.Draw(camera)
	wall.Draw(camera)
	tallWall.Draw(camera)
	spriteStack.Draw(camera)
	billboard.Draw(camera)

	for i := 0; i < len(stress); i++ {
		stress[i].Draw(camera)
	}

	ebiten.ActualTPS()
	ebitenutil.DebugPrintAt(camera.Surface, fmt.Sprintf("%f, %f", ebiten.ActualFPS(), ebiten.ActualTPS()), 0, 0)

	mx, my := ebiten.CursorPosition()
	wx, wy := camera.GetWorldCoords(float64(mx), float64(my))
	ebitenutil.DebugPrintAt(camera.Surface, fmt.Sprintf("%d, %d", int(wx), int(wy)), mx, my-16)
	vector.DrawFilledCircle(camera.Surface, float32(camera.Surface.Bounds().Dx())/2, float32(camera.Surface.Bounds().Dy())/2, 4, color.RGBA{255, 255, 255, 255}, false)
	camera.Blit(screen)
}

// Layout sets window size
func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	WindowWidth = outsideWidth
	WindowHeight = outsideHeight
	return outsideWidth, outsideHeight
}

func main() {
	log.SetFlags(log.Lshortfile)

	if b, err := embedded.ReadFile("tiles.png"); err == nil {
		if s, err := png.Decode(bytes.NewReader(b)); err == nil {
			sprites := ebiten.NewImageFromImage(s)
			SpriteSheetTiles = zen.NewSpriteSheet(sprites, 16, 16, zen.SpriteSheetOptions{
				Scale: 2,
			})
		}
	} else {
		log.Fatal(err)
	}

	if b, err := embedded.ReadFile("car.png"); err == nil {
		if s, err := png.Decode(bytes.NewReader(b)); err == nil {
			sprites := ebiten.NewImageFromImage(s)
			SpriteSheetCar = zen.NewSpriteSheet(sprites, 15, 32, zen.SpriteSheetOptions{
				Scale: 2,
			})
		}
	} else {
		log.Fatal(err)
	}

	if b, err := embedded.ReadFile("tall_wall.png"); err == nil {
		if s, err := png.Decode(bytes.NewReader(b)); err == nil {
			sprites := ebiten.NewImageFromImage(s)
			SpriteSheetTallWallSides = zen.NewSpriteSheet(sprites, 16*2, 16*4, zen.SpriteSheetOptions{
				Scale: 2,
			})
			SpriteSheetTallWallTop = zen.NewSpriteSheet(sprites, 16*2, 16*2, zen.SpriteSheetOptions{
				Scale: 2,
			})
		}
	} else {
		log.Fatal(err)
	}

	game := &Game{}
	ebiten.SetWindowSize(WindowWidth, WindowHeight)
	ebiten.SetWindowTitle("Wall and floor rendering example")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	camera = zen.NewCamera(WindowWidth, WindowHeight, 0, 0, 0, 1)

	// Outlines on IsometricDrawables are dynamic!
	// You can't use a spritesheet with an outline as it breaks how the stack is rendered!
	// You would use a spritesheet outline if you're not using IsometricDrawables
	floor = zen.NewFloor(
		SpriteSheetTiles.GetSprite(4, 0),
		0,
		// zen.NewVector2(0, 0),
		zen.NewVector2(-float64(SpriteSheetTiles.SpriteWidth)*4, 0),
		// zen.NewVector2(0, 0),
		zen.NewVector2(float64(SpriteSheetTiles.SpriteWidth)/2, float64(SpriteSheetTiles.SpriteHeight)/2),
	)
	floor.OutlineColor = color.RGBA{0, 255, 0, 255}
	floor.OutlineThickness = 2

	wall = zen.NewWall(
		SpriteSheetTiles.GetSprite(0, 0),
		[]*ebiten.Image{
			SpriteSheetTiles.GetSprite(2, 0),
			SpriteSheetTiles.GetSprite(1, 0),
			SpriteSheetTiles.GetSprite(1, 0),
			SpriteSheetTiles.GetSprite(1, 0),
		},
		float64(SpriteSheetTiles.SpriteHeight),
		0,
		// zen.NewVector2(0, 0),
		// zen.NewVector2(0, 0),
		zen.NewVector2(-float64(SpriteSheetTiles.SpriteWidth)*8, 0),
		zen.NewVector2(float64(SpriteSheetTiles.SpriteWidth)/2, float64(SpriteSheetTiles.SpriteHeight)/2),
	)
	wall.OutlineColor = color.RGBA{255, 0, 0, 255}
	wall.OutlineThickness = 2

	tallWall = zen.NewWall(
		SpriteSheetTallWallTop.GetSprite(0, 2),
		[]*ebiten.Image{
			SpriteSheetTallWallSides.GetSprite(0, 0),
			SpriteSheetTallWallSides.GetSprite(0, 0),
			SpriteSheetTallWallSides.GetSprite(0, 0),
			SpriteSheetTallWallSides.GetSprite(0, 0),
		},
		float64(SpriteSheetTallWallSides.SpriteHeight),
		0,
		zen.NewVector2(0, 0),
		// zen.NewVector2(-float64(SpriteSheetTiles.SpriteWidth)*4, -float64(SpriteSheetTiles.SpriteWidth)*4),
		zen.NewVector2(float64(SpriteSheetTallWallTop.SpriteWidth)/2, float64(SpriteSheetTallWallTop.SpriteHeight)/2),
	)
	tallWall.OutlineColor = color.RGBA{255, 0, 0, 255}
	tallWall.OutlineThickness = 2

	spriteStack = zen.NewSpriteStack(
		SpriteSheetCar,
		0,
		zen.NewVector2(float64(SpriteSheetTiles.SpriteWidth)*5, 0),
		zen.NewVector2(float64(SpriteSheetCar.SpriteWidth)/2, float64(SpriteSheetCar.SpriteHeight)/2),
	)
	spriteStack.OutlineColor = color.RGBA{0, 255, 0, 255}
	spriteStack.OutlineThickness = 2

	billboard = zen.NewBillboard(
		SpriteSheetTiles.GetSprite(5, 0),
		// zen.NewVector2(0, 0),
		zen.NewVector2(float64(SpriteSheetTiles.SpriteWidth)*2, 0),
		// zen.NewVector2(0, 0),
		zen.NewVector2(float64(SpriteSheetTiles.SpriteWidth)/2, float64(SpriteSheetTiles.SpriteHeight)), // bottom of the sprite
	)
	billboard.OutlineThickness = 2
	billboard.OutlineColor = color.RGBA{0, 255, 0, 255}

	stress = make([]zen.IsometricDrawable, 0)
	for i := 0; i < 0; i++ { // TODO memory leak at 2000, but only when drawing outlines
		x, y := float64(rand.IntN(30)-15), float64(rand.IntN(30)-15)

		w := zen.NewWall(
			SpriteSheetTiles.GetSprite(0, 0),
			[]*ebiten.Image{
				SpriteSheetTiles.GetSprite(2, 0),
				SpriteSheetTiles.GetSprite(1, 0),
				SpriteSheetTiles.GetSprite(1, 0),
				SpriteSheetTiles.GetSprite(1, 0),
			},
			float64(SpriteSheetTiles.SpriteWidth),
			math.Pi/4*rand.Float64(),
			zen.NewVector2(float64(SpriteSheetTiles.SpriteWidth)*x, float64(SpriteSheetTiles.SpriteHeight)*y),
			zen.NewVector2(float64(SpriteSheetTiles.SpriteWidth)/2, float64(SpriteSheetTiles.SpriteHeight)/2),
		)
		// w.OutlineColor = color.RGBA{255, 255, 0, 255}
		// w.OutlineThickness = 2
		stress = append(stress, w)
	}

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
