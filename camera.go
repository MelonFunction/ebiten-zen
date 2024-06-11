// Package zen is the root for all ebiten-zen files
package zen

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

// Camera can look at positions, zoom and rotate.
type Camera struct {
	ScreenRotation float64
	Scale          float64
	Position       *Vector
	Width, Height  int
	Surface        *ebiten.Image

	WorldRotation float64 // used by renderisometric to rotate sprites around a point
}

// NewCamera returns a new Camera
func NewCamera(width, height int, x, y, rotation, zoom float64) *Camera {
	return &Camera{
		Position:       NewVector(x, y),
		Width:          width,
		Height:         height,
		ScreenRotation: rotation,
		Scale:          zoom,
		Surface:        ebiten.NewImage(width, height),
		WorldRotation:  0,
	}
}

// Deallocate deallocates the Surface
func (c *Camera) Deallocate() *Camera {
	c.Surface.Deallocate()
	return c
}

// SetPosition looks at a position
func (c *Camera) SetPosition(x, y float64) *Camera {
	c.Position.X = x
	c.Position.Y = y
	return c
}

// MovePosition moves the Camera by x and y.
// Use SetPosition if you want to set the position
func (c *Camera) MovePosition(x, y float64) *Camera {
	c.Position.X += x
	c.Position.Y += y
	return c
}

// GetPosition returns the Camera's Position Vector
func (c *Camera) GetPosition() *Vector {
	return c.Position
}

// RotateScreen rotates by phi
func (c *Camera) RotateScreen(phi float64) *Camera {
	c.ScreenRotation += phi
	return c
}

// SetScreenRotation sets the rotation to rot
func (c *Camera) SetScreenRotation(rot float64) *Camera {
	c.ScreenRotation = rot
	return c
}

// RotateWorld rotates by phi
func (c *Camera) RotateWorld(phi float64) *Camera {
	c.WorldRotation += phi
	return c
}

// SetWorldRotation sets the rotation to rot
func (c *Camera) SetWorldRotation(rot float64) *Camera {
	c.WorldRotation = rot
	return c
}

// Zoom *= the current zoom
func (c *Camera) Zoom(mul float64) *Camera {
	c.Scale *= mul
	if c.Scale <= 0.01 {
		c.Scale = 0.01
	}
	// TODO don't resize canvas
	c.Resize(c.Width, c.Height)
	return c
}

// SetZoom sets the zoom
func (c *Camera) SetZoom(zoom float64) *Camera {
	c.Scale = zoom
	if c.Scale <= 0.01 {
		c.Scale = 0.01
	}
	c.Resize(c.Width, c.Height)
	return c
}

// Resize resizes the camera Surface
func (c *Camera) Resize(w, h int) *Camera {
	c.Width = w
	c.Height = h
	newW := int(float64(w) * 1.0 / c.Scale)
	newH := int(float64(h) * 1.0 / c.Scale)
	if newW <= 16384 && newH <= 16384 {
		c.Surface.Deallocate()
		c.Surface = ebiten.NewImage(newW, newH)
	}
	return c
}

// GetTranslation alters the provided *ebiten.DrawImageOptions' translation based on the given x,y offset and the
// camera's position
func (c *Camera) GetTranslation(ops *ebiten.DrawImageOptions, x, y float64) *ebiten.DrawImageOptions {
	surfaceSize := c.Surface.Bounds().Size()
	ops.GeoM.Translate(float64(surfaceSize.X)/2, float64(surfaceSize.Y)/2)
	ops.GeoM.Translate(-c.Position.X+x, -c.Position.Y+y)
	return ops
}

// GetRotation alters the provided *ebiten.DrawImageOptions' rotation using the provided originX and originY args
func (c *Camera) GetRotation(ops *ebiten.DrawImageOptions, rot, originX, originY float64) *ebiten.DrawImageOptions {
	ops.GeoM.Translate(originX, originY)
	ops.GeoM.Rotate(rot)
	ops.GeoM.Translate(-originX, -originY)
	return ops
}

// GetScale alters the provided *ebiten.DrawImageOptions' scale
func (c *Camera) GetScale(ops *ebiten.DrawImageOptions, scaleX, scaleY float64) *ebiten.DrawImageOptions {
	ops.GeoM.Scale(scaleX, scaleY)
	return ops
}

// GetSkew alters the provided *ebiten.DrawImageOptions' skew
func (c *Camera) GetSkew(ops *ebiten.DrawImageOptions, skewX, skewY float64) *ebiten.DrawImageOptions {
	ops.GeoM.Skew(skewX, skewY)
	return ops
}

// Blit draws the camera's surface to the passed *ebiten.Image and applies zoom
func (c *Camera) Blit(surface *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	surfaceSize := c.Surface.Bounds().Size()
	cx := float64(surfaceSize.X) / 2.0
	cy := float64(surfaceSize.Y) / 2.0

	op.GeoM.Translate(-cx, -cy)
	op.GeoM.Scale(c.Scale, c.Scale)
	op.GeoM.Rotate(c.ScreenRotation)
	op.GeoM.Translate(cx*c.Scale, cy*c.Scale)

	surface.DrawImage(c.Surface, op)
}

// GetScreenCoords converts world coords into screen coords
func (c *Camera) GetScreenCoords(x, y float64) (float64, float64) {
	w, h := c.Width, c.Height
	co := math.Cos(c.ScreenRotation)
	si := math.Sin(c.ScreenRotation)

	x, y = x-c.Position.X, y-c.Position.Y
	x, y = co*x-si*y, si*x+co*y

	return x*c.Scale + float64(w)/2, y*c.Scale + float64(h)/2
}

// GetWorldCoords converts screen coords into world coords
func (c *Camera) GetWorldCoords(x, y float64) (float64, float64) {
	w, h := c.Width, c.Height
	co := math.Cos(-c.ScreenRotation)
	si := math.Sin(-c.ScreenRotation)

	x, y = (x-float64(w)/2)/c.Scale, (y-float64(h)/2)/c.Scale
	x, y = co*x-si*y, si*x+co*y

	return x + c.Position.X, y + c.Position.Y
}

// GetCursorCoords converts cursor/screen coords into world coords
func (c *Camera) GetCursorCoords() (float64, float64) {
	cx, cy := ebiten.CursorPosition()
	return c.GetWorldCoords(float64(cx), float64(cy))
}
