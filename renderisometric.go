// Package zen is the root for all ebiten-zen files
package zen

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

// Tile is an interface that Wall and Floor must satisfy
type Tile interface {
	Draw(camera *Camera)
}

// Floor represents a floor tile in the world
type Floor struct {
	Position *Vector
	Size     float64

	Rotation      float64 // individual rotation
	RotationPoint *Vector
	RotatedPos    *Vector

	Sprite *ebiten.Image
}

// Wall represents a wall tile in the world
type Wall struct {
	Position *Vector
	Size     float64
	Height   float64 // height of the top face of the tile

	Rotation      float64 // individual rotation
	RotationPoint *Vector
	RotatedPos    *Vector

	TopSprite   *ebiten.Image
	WallSprites []*ebiten.Image // if len() == 1, same will be used for all walls
}

// Draw draws a rotated texture
func (f *Floor) Draw(camera *Camera) {
	rotation := camera.WorldRotation + f.Rotation
	rotation = math.Atan2(math.Sin(rotation), math.Cos(rotation))
	worldRotationPoint := camera.Position
	f.RotatedPos = f.Position.RotateAround(camera.WorldRotation, worldRotationPoint)
	w, d := float64(f.Sprite.Bounds().Dx()), float64(f.Sprite.Bounds().Dy())
	op := &ebiten.DrawImageOptions{}
	op.ColorScale.Scale(1, 1, 1, 1)
	op = camera.GetRotation(op, rotation, -w/2, -d/2)
	op = camera.GetTranslation(op, f.RotatedPos.X-w/2, f.RotatedPos.Y-d/2)
	camera.Surface.DrawImage(f.Sprite, op)
}

// Draw draws a textured cube
func (t *Wall) Draw(camera *Camera) {
	rotation := camera.WorldRotation + t.Rotation
	rotation = math.Atan2(math.Sin(rotation), math.Cos(rotation)) // clip rotation value
	worldRotationPoint := camera.Position
	t.RotatedPos = t.Position.RotateAround(camera.WorldRotation, worldRotationPoint)
	op := &ebiten.DrawImageOptions{}
	w, d, h := float64(t.TopSprite.Bounds().Dx()), float64(t.TopSprite.Bounds().Dy()), t.Height

	tr := NewVector(t.RotatedPos.X+w/2, t.RotatedPos.Y-d/2).RotateAround(t.Rotation, t.RotationPoint).RotateAround(camera.WorldRotation, t.RotatedPos)
	bl := NewVector(t.RotatedPos.X-w/2, t.RotatedPos.Y+d/2).RotateAround(t.Rotation, t.RotationPoint).RotateAround(camera.WorldRotation, t.RotatedPos)
	br := NewVector(t.RotatedPos.X+w/2, t.RotatedPos.Y+d/2).RotateAround(t.Rotation, t.RotationPoint).RotateAround(camera.WorldRotation, t.RotatedPos)
	tl := NewVector(t.RotatedPos.X-w/2, t.RotatedPos.Y-d/2).RotateAround(t.Rotation, t.RotationPoint).RotateAround(camera.WorldRotation, t.RotatedPos)

	// draw faces clockwise to prevent image flipping
	drawFace := func(p1, p2 *Vector, img *ebiten.Image) {
		op = &ebiten.DrawImageOptions{}
		op.ColorScale.Scale(1, 1, 1, 1)
		op = camera.GetScale(op, (p2.X-p1.X)/(float64(t.TopSprite.Bounds().Dx())), 1)
		op = camera.GetSkew(op, 0, p1.AngleTo(p2))
		op = camera.GetTranslation(op, p1.X, p1.Y-d)
		camera.Surface.DrawImage(img, op)
	}

	if math.Abs(rotation) <= math.Pi/2 {
		// front
		drawFace(bl, br, t.WallSprites[0])
	}
	if rotation > 0 && rotation <= math.Pi {
		// right
		drawFace(br, tr, t.WallSprites[3])
	}
	if math.Abs(rotation) >= math.Pi/2 {
		// back
		drawFace(tr, tl, t.WallSprites[2])
	}
	if rotation < 0 {
		// left
		drawFace(tl, bl, t.WallSprites[1])
	}

	w, d = float64(t.TopSprite.Bounds().Dx()), float64(t.TopSprite.Bounds().Dy())
	op = &ebiten.DrawImageOptions{}
	op.ColorScale.Scale(1, 1, 1, 1)
	op = camera.GetRotation(op, rotation, -w/2, -d/2)
	op = camera.GetTranslation(op, t.RotatedPos.X-w/2, t.RotatedPos.Y-d/2-h)
	camera.Surface.DrawImage(t.TopSprite, op)
}
