// Package main
package main

import (
	"errors"
	"fmt"
	"image/color"
	"log"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	zen "github.com/melonfunction/ebiten-zen"
)

// vars
var (
	camera   *zen.Camera
	collider *zen.SpatialHash
	player   *zen.CircleShape

	playerDirection float64

	WindowWidth  = 640 * 2
	WindowHeight = 480 * 2

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

	if ebiten.IsKeyPressed(ebiten.KeyR) {
		playerDirection += math.Pi / 100
	}
	if ebiten.IsKeyPressed(ebiten.KeyG) {
		playerDirection -= math.Pi / 100
	}

	dir := zen.NewVector2(0, 0)
	speed := 5.0
	if ebiten.IsKeyPressed(ebiten.KeyH) {
		dir.X--
	}
	if ebiten.IsKeyPressed(ebiten.KeyN) {
		dir.X++
	}
	if ebiten.IsKeyPressed(ebiten.KeyC) {
		dir.Y--
	}
	if ebiten.IsKeyPressed(ebiten.KeyT) {
		dir.Y++
	}
	dir = dir.Normalize().Mult(speed).Rotate(playerDirection)
	player.MovePosition(dir.X, dir.Y)

	collisions := collider.CheckCollisions(player)
	for _, collision := range collisions {
		sep := collision.SeparatingVector
		player.MovePosition(sep.X, sep.Y)
	}

	camera.SetPosition(player.Pos.X, player.Pos.Y)

	return nil
}

// Draw draws the game screen.
// Draw is called every frame (typically 1/60[s] for 60Hz display).
func (g *Game) Draw(screen *ebiten.Image) {
	camera.Surface.Clear()

	for k := range collider.Backref {
		switch s := k.(type) {
		case *zen.RectangleShape:
			w, h := s.Width, s.Height
			x, y := camera.GetScreenCoords(s.Pos.X-w/2, s.Pos.Y-h/2)
			vector.DrawFilledRect(camera.Surface, float32(x), float32(y), float32(w), float32(h), color.RGBA{64, 0, 0, 32}, true)
			vector.StrokeRect(camera.Surface, float32(x), float32(y), float32(w), float32(h), 2, color.RGBA{128, 0, 0, 64}, true)
		case *zen.CircleShape:
			x, y := camera.GetScreenCoords(s.Pos.X, s.Pos.Y)
			r := float32(s.Radius)
			vector.DrawFilledCircle(camera.Surface, float32(x), float32(y), r, color.RGBA{64, 0, 0, 32}, true)
			vector.StrokeCircle(camera.Surface, float32(x), float32(y), r, 2, color.RGBA{128, 0, 0, 64}, true)
		}
	}

	// draw the player look direction too
	x1, y1 := camera.GetScreenCoords(player.Pos.X, player.Pos.Y)
	r := float32(player.Radius) * 2
	x2, y2 := zen.NewVector2(0, -1).Rotate(playerDirection).Mult(float64(r)).Add(zen.NewVector2(x1, y1)).Unpack()
	vector.StrokeLine(camera.Surface, float32(x1), float32(y1), float32(x2), float32(y2), 2, color.RGBA{128, 0, 0, 64}, true)

	mx, my := ebiten.CursorPosition()
	wx, wy := camera.GetWorldCoords(float64(mx), float64(my))
	ebitenutil.DebugPrintAt(camera.Surface, fmt.Sprintf("%d, %d", int(wx), int(wy)), mx, my-16)
	ebitenutil.DebugPrintAt(camera.Surface, fmt.Sprintf("%f, %f", player.Pos.X, player.Pos.Y), 0, 0)
	camera.Blit(screen)
}

// Layout sets window size
func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	WindowWidth = outsideWidth
	WindowHeight = outsideHeight

	// update camera bounds in the collider

	return outsideWidth, outsideHeight
}

func main() {
	log.SetFlags(log.Lshortfile)

	game := &Game{}
	ebiten.SetWindowSize(WindowWidth, WindowHeight)
	ebiten.SetWindowTitle("Collisions")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	camera = zen.NewCamera(WindowWidth, WindowHeight, 0, 0, 0, 1)

	collider = zen.NewSpatialHash(100)
	collider.NewRectangleShape(100, 100, 100, 100)
	collider.NewRectangleShape(200, 100, 100, 100)
	player = collider.NewCircleShape(100, 250, 16)

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
