// Package zen is the root for all ebiten-zen files
package zen

import "math"

// TODO
// - vector_test.go
// - {name}InPlace() to prevent creating a new Vector2 every time a function is called

// Vector2 represents a point in space
type Vector2 struct {
	X, Y float64
}

// NewVector2 returns a new *Vector2
func NewVector2(x, y float64) *Vector2 {
	return &Vector2{x, y}
}

// Clone returns a copy of the vector
func (v *Vector2) Clone() *Vector2 {
	return &Vector2{
		v.X,
		v.Y,
	}
}

// Mult multiplies v by scalar and returns a new Vector2 for chaining
func (v *Vector2) Mult(scalar float64) *Vector2 {
	return &Vector2{
		v.X * scalar,
		v.Y * scalar,
	}
}

// Add adds the values of o and v together and returns a new Vector2
func (v *Vector2) Add(o *Vector2) *Vector2 {
	return &Vector2{
		v.X + o.X,
		v.Y + o.Y,
	}
}

// Sub subtracts o from v and returns a new Vector2
func (v *Vector2) Sub(o *Vector2) *Vector2 {
	return &Vector2{
		v.X - o.X,
		v.Y - o.Y,
	}
}

// Length returns the length of the Vector2
func (v *Vector2) Length() float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y)
}

// Normalize returns a new Vector2 representing the normal of v
func (v *Vector2) Normalize() *Vector2 {
	l := v.Length()
	if l > 0 {
		return &Vector2{
			v.X / l,
			v.Y / l,
		}
	}
	return NewVector2(v.X, v.Y)
}

// Rotate rotates a point about 0,0 and returns v for chaining
func (v *Vector2) Rotate(phi float64) *Vector2 {
	c, s := math.Cos(phi), math.Sin(phi)
	return &Vector2{
		X: c*v.X - s*v.Y,
		Y: s*v.X + c*v.Y,
	}
}

// RotateAround rotates a Vector2 about another Vector2 and returns v for chaining
func (v *Vector2) RotateAround(phi float64, o *Vector2) *Vector2 {
	c, s := math.Cos(phi), math.Sin(phi)
	n := NewVector2(v.X, v.Y).Sub(o)
	return NewVector2(
		c*n.X-s*n.Y,
		s*n.X+c*n.Y,
	).Add(o)
}

// AngleTo returns the angle between v to other
func (v *Vector2) AngleTo(other *Vector2) float64 {
	return math.Atan2(v.Y-other.Y, v.X-other.X)
}

// Unpack returns the Vector2's components
func (v *Vector2) Unpack() (float64, float64) {
	return v.X, v.Y
}
