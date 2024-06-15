// Package zen is the root for all ebiten-zen files
package zen

import (
	"errors"
	"math"
)

var ErrShapeNotFound = errors.New("Couldn't remove shape from SpatialHash; not found")

// Shape interface describes an object that can be added to the hash
type Shape interface {
	GetPosition() *Vector2 // get the position
	GetBounds() (float64, float64, float64, float64)
	MovePosition(x, y float64) // move by amount
	SetPosition(x, y float64)  // move to position
	SetHash(s *SpatialHash)    // sets ref to hash
	GetHash() *SpatialHash     // gets	 ref to hash

	SetParent(i interface{})
	GetParent() interface{}
}

// CircleShape shape
type CircleShape struct {
	// Center point
	Pos         *Vector2
	Radius      float64
	SpatialHash *SpatialHash
	Parent      interface{}
}

// RectangleShape shape
type RectangleShape struct {
	// Center point
	Pos           *Vector2
	Width, Height float64
	SpatialHash   *SpatialHash
	Parent        interface{}
}

// PointShape is a RectangleShape but with 0 width and height
type PointShape struct{ *RectangleShape }

// Cell contains shapes
type Cell struct {
	Shapes map[Shape]Shape
}

// CellCoord is a coordinate used to index the hash
type CellCoord struct {
	X, Y int
}

// SpatialHash contains cells
type SpatialHash struct {
	// Size of the grid/cell/partition
	CellSize int
	// Store shapes in a cell depending on their bounds
	Hash map[CellCoord]*Cell
	// Backref for shapes to find its containing cells
	Backref map[Shape][]*Cell
}

// NewSpatialHash returns a new *SpatialHash
func NewSpatialHash(cellSize int) *SpatialHash {
	return &SpatialHash{
		CellSize: cellSize,
		Hash:     make(map[CellCoord]*Cell),
		Backref:  make(map[Shape][]*Cell),
	}
}

// GetXYWH converts bounds coords into X,Y,W,H
// Also returns float32s for use with ebiten's vector.DrawFilledRect since that's
// the most likely use for this function.
// You can also just type cast the interface to RectangleShape (if it is one) to
// get access to Pos, Width and Height
func (s *SpatialHash) GetXYWH(shape Shape) (float32, float32, float32, float32) {
	x1, y1, x2, y2 := shape.GetBounds()
	w, h := x2-x1, y2-y1
	return float32(x1), float32(y1), float32(w), float32(h)
}

// Add adds a shape to the spatial hash
func (s *SpatialHash) Add(shape Shape) {
	x1, y1, x2, y2 := shape.GetBounds()

	// make sure big shapes are constrained properly
	xStep := x2 - x1
	if xStep > float64(s.CellSize) {
		xStep = float64(s.CellSize)
	}
	yStep := y2 - y1
	if yStep > float64(s.CellSize) {
		yStep = float64(s.CellSize)
	}
	for x := x1; x <= x2; x += xStep {
		for y := y1; y <= y2; y += yStep {
			hashPos := CellCoord{
				int(math.Floor(x / float64(s.CellSize))),
				int(math.Floor(y / float64(s.CellSize))),
			}
			if _, ok := s.Hash[hashPos]; !ok {
				s.Hash[hashPos] = &Cell{Shapes: make(map[Shape]Shape)}
			}
			s.Hash[hashPos].Shapes[shape] = shape                        // add shape to cell
			s.Backref[shape] = append(s.Backref[shape], s.Hash[hashPos]) // add cell to backref

			if xStep == 0 || yStep == 0 {
				goto done
			}
		}
	}
done:
	shape.SetHash(s)
}

// Remove removes a shape from the spatial hash
func (s *SpatialHash) Remove(shape Shape) error {
	if cells, ok := s.Backref[shape]; ok {
		for _, cell := range cells {
			delete(cell.Shapes, shape)
		}
		delete(s.Backref, shape)
	}

	return ErrShapeNotFound
}

// GetCollisionCandidates returns a list of all shapes in the same cells as shape
func (s *SpatialHash) GetCollisionCandidates(shape Shape) []Shape {
	shapesMap := make(map[Shape]struct{})
	if cells, ok := s.Backref[shape]; ok {
		for _, cell := range cells {
			for _, sh := range cell.Shapes {
				shapesMap[sh] = struct{}{}
			}
		}
	}
	delete(shapesMap, shape)
	shapes := make([]Shape, 0)
	for k := range shapesMap {
		shapes = append(shapes, k)
	}
	return shapes
}

// CollisionData contains information about the collision
type CollisionData struct {
	Target           Shape
	Other            Shape
	SeparatingVector *Vector2
}

func collisionRectRect(r1, r2 *RectangleShape) *Vector2 {
	r1Left, r1Up, r1Right, r1Down := r1.GetBounds()
	r2Left, r2Up, r2Right, r2Down := r2.GetBounds()

	// TODO is this ok?
	if !(((r1Right >= r2Left && r1Right <= r2Right) || (r1Left >= r2Left && r1Left <= r2Right) || (r1Left >= r2Left && r1Right <= r2Right) || (r2Left >= r1Left && r2Right <= r1Right)) &&
		((r1Up <= r2Down && r1Up >= r2Up) || (r1Down <= r2Down && r1Down >= r2Up) || (r1Up >= r2Up && r1Down <= r2Down) || (r2Up >= r1Up && r2Down <= r1Down))) {

		return &Vector2{0, 0}
	}

	var dx, dy float64
	if r1.Pos.X < r2.Pos.X {
		dx = r2.Pos.X - r2.Width/2 - r1.Pos.X - r1.Width/2
	} else {
		dx = r2.Pos.X + r2.Width/2 - r1.Pos.X + r1.Width/2
	}
	if r1.Pos.Y < r2.Pos.Y {
		dy = r2.Pos.Y - r2.Height/2 - r1.Pos.Y - r1.Height/2
	} else {
		dy = r2.Pos.Y + r2.Height/2 - r1.Pos.Y + r1.Height/2
	}

	if math.Abs(dx) < math.Abs(dy) {
		dy = 0
	} else {
		dx = 0
	}
	return &Vector2{dx, dy}
}

func collisionRectCirc(r1 *RectangleShape, c1 *CircleShape) *Vector2 {
	// Check bbox of circle
	rr := collisionRectRect(
		r1,
		&RectangleShape{
			Pos:    c1.Pos,
			Width:  c1.Radius * 2,
			Height: c1.Radius * 2,
		})
	if rr.Length() == 0 {
		return rr
	}

	// Get nearest corner, return if midpoint of c1 is inside rect
	left, up, right, down := r1.GetBounds()
	var co *Vector2
	if r1.Pos.X > c1.Pos.X { // left
		if r1.Pos.Y > c1.Pos.Y { // top
			if c1.Pos.X > left || c1.Pos.Y > up {
				return rr
			}
			co = NewVector2(left, up)
		} else { // bottom
			if c1.Pos.X > left || c1.Pos.Y < down {
				return rr
			}
			co = NewVector2(left, down)
		}
	} else { // right
		if r1.Pos.Y > c1.Pos.Y { // top
			if c1.Pos.X < right || c1.Pos.Y > up {
				return rr
			}
			co = NewVector2(right, up)
		} else { // bottom
			if c1.Pos.X < right || c1.Pos.Y < down {
				return rr
			}
			co = NewVector2(right, down)
		}
	}

	// Resolve circle/point collision
	cc := collisionCircCirc(
		&CircleShape{
			Pos:    co,
			Radius: 0,
		},
		c1)
	return cc

}

func collisionCircCirc(c1, c2 *CircleShape) *Vector2 {
	dist := c1.Pos.Sub(c2.Pos)
	depth := c1.Radius + c2.Radius - dist.Length()
	if depth < 0 {
		return &Vector2{0, 0}
	}

	return dist.Normalize().Mult(depth)
}

// ResolveCollisions returns the correct separating Vector2 using []CollisionData as input.
// You can get the collisions by calling CheckCollisions, and then pass the output into
// this function (you may also want to filter what collisions go through to this function,
// as you may have some shapes that don't affect how the shape being passed into CheckCollisions
// moves, for example, a shape which represents an enemy).
// This function is a little broken, and it works best if you use a CircleShape as the target
// with the collisions being RectangleShapes.
func (s *SpatialHash) ResolveCollisions(target Shape, collisions []CollisionData) {
	sep := NewVector2(0, 0)
	switch target.(type) {
	case *CircleShape:
		for i, collision := range collisions {
			// log.Println(i, collision.SeparatingVector)
			if i == 0 {
				sep = collision.SeparatingVector
			} else {
				switch collision.Other.(type) {
				case *CircleShape:
					sep = sep.Add(collision.SeparatingVector)
				case *RectangleShape:
					if collision.SeparatingVector.X == 0 || collision.SeparatingVector.Y == 0 {
						if sep.X == 0 || sep.Y == 0 {
							sep = sep.Add(collision.SeparatingVector)
						} else {
							sep = collision.SeparatingVector
						}
					}
				}
			}
		}
		// if len(collisions) > 0 {
		// 	log.Println("sep", sep)
		// 	log.Println()
		// }
	case *RectangleShape:
		for i, collision := range collisions {
			// log.Println(i, collision.SeparatingVector)
			if i == 0 {
				sep = collision.SeparatingVector
			} else {
				switch collision.Other.(type) {
				case *CircleShape:
					sep = sep.Add(collision.SeparatingVector)
				case *RectangleShape:
					// TODO fix shape getting snagged on two rects next to each other
					if collision.SeparatingVector.X == 0 || collision.SeparatingVector.Y == 0 {
						if math.Abs(collision.SeparatingVector.X) > math.Abs(sep.X) {
							sep.X = collision.SeparatingVector.X
						}
						if math.Abs(collision.SeparatingVector.Y) > math.Abs(sep.Y) {
							sep.Y = collision.SeparatingVector.Y
						}
					}
					// if collision.SeparatingVector.Length() > sep.Length() {
					// 	sep = collision.SeparatingVector
					// }
				}
			}
		}
		// if len(collisions) > 0 {
		// 	log.Println("sep", sep)
		// 	log.Println()
		// }
	}
	target.MovePosition(sep.Unpack())
}

// CheckCollisions returns a list of all shapes and their separating vector
func (s *SpatialHash) CheckCollisions(shape Shape) []CollisionData {
	collisions := make([]CollisionData, 0)
	candidates := s.GetCollisionCandidates(shape)

	switch typed := shape.(type) {
	case *RectangleShape:
		for _, candidate := range candidates {
			var col *Vector2
			switch other := candidate.(type) {
			case *RectangleShape:
				col = collisionRectRect(typed, other)
			case *CircleShape:
				col = collisionRectCirc(typed, other)
			default:
				// TODO error
			}
			if col != nil && col.Length() > 0 {
				collisions = append(collisions, CollisionData{Other: candidate, SeparatingVector: col})
			}
		}
	case *CircleShape:
		for _, candidate := range candidates {
			var col *Vector2
			switch other := candidate.(type) {
			case *RectangleShape:
				col = collisionRectCirc(other, typed).Mult(-1)
			case *CircleShape:
				col = collisionCircCirc(typed, other)
			default:
				// TODO error
			}
			if col != nil && col.Length() > 0 {
				collisions = append(collisions, CollisionData{Other: candidate, SeparatingVector: col})
			}
		}
	default:
		// TODO error
	}

	return collisions
}

// NewCircleShape creates, then adds a new CircleShape to the hash before returning it
func (s *SpatialHash) NewCircleShape(x, y, r float64) *CircleShape {
	ci := &CircleShape{
		Pos:    &Vector2{x, y},
		Radius: r,
	}
	s.Add(ci)
	return ci
}

// GetPosition returns the Point of the CircleShape
func (ci *CircleShape) GetPosition() *Vector2 {
	return ci.Pos
}

// GetBounds returns the Bounds of the CircleShape
func (ci *CircleShape) GetBounds() (float64, float64, float64, float64) {
	return ci.Pos.X - ci.Radius,
		ci.Pos.Y - ci.Radius,
		ci.Pos.X + ci.Radius,
		ci.Pos.Y + ci.Radius
}

// MovePosition moves the CircleShape by x and y
func (ci *CircleShape) MovePosition(x, y float64) {
	ci.Pos.X += x
	ci.Pos.Y += y
	hash := ci.GetHash()
	hash.Remove(ci)
	hash.Add(ci)
}

// SetPosition moves the CircleShape to x and y
func (ci *CircleShape) SetPosition(x, y float64) {
	ci.Pos.X = x
	ci.Pos.Y = y
	hash := ci.GetHash()
	hash.Remove(ci)
	hash.Add(ci)
}

// SetHash sets the hash
func (ci *CircleShape) SetHash(s *SpatialHash) {
	ci.SpatialHash = s
}

// GetHash gets the hash
func (ci *CircleShape) GetHash() *SpatialHash {
	return ci.SpatialHash
}

// SetParent sets the parent
func (ci *CircleShape) SetParent(i interface{}) {
	ci.Parent = i
}

// GetParent gets the parent
func (ci *CircleShape) GetParent() interface{} {
	return ci.Parent
}

// NewRectangleShape creates, then adds a new RectangleShape to the hash before returning it
func (s *SpatialHash) NewRectangleShape(x, y, w, h float64) *RectangleShape {
	re := &RectangleShape{
		Pos:    &Vector2{x, y},
		Width:  w,
		Height: h,
	}
	s.Add(re)
	return re
}

// GetPosition returns the Point of the RectangleShape
func (re *RectangleShape) GetPosition() *Vector2 {
	return re.Pos
}

// GetBounds returns the Bounds of the RectangleShape
func (re *RectangleShape) GetBounds() (float64, float64, float64, float64) {
	return re.Pos.X - re.Width/2,
		re.Pos.Y - re.Height/2,
		re.Pos.X + re.Width/2,
		re.Pos.Y + re.Height/2
}

// MovePosition moves the RectangleShape by x and y
func (re *RectangleShape) MovePosition(x, y float64) {
	re.Pos.X += x
	re.Pos.Y += y
	hash := re.GetHash()
	hash.Remove(re)
	hash.Add(re)
}

// SetPosition moves the RectangleShape to x and y
func (re *RectangleShape) SetPosition(x, y float64) {
	re.Pos.X = x
	re.Pos.Y = y
	hash := re.GetHash()
	hash.Remove(re)
	hash.Add(re)
}

// SetHash sets the hash
func (re *RectangleShape) SetHash(s *SpatialHash) {
	re.SpatialHash = s
}

// GetHash gets the hash
func (re *RectangleShape) GetHash() *SpatialHash {
	return re.SpatialHash
}

// SetParent sets the parent
func (re *RectangleShape) SetParent(i interface{}) {
	re.Parent = i
}

// GetParent gets the parent
func (re *RectangleShape) GetParent() interface{} {
	return re.Parent
}

// NewPointShape creates, then adds a new RectangleShape to the hash before returning it
func (s *SpatialHash) NewPointShape(x, y float64) *PointShape {
	return &PointShape{s.NewRectangleShape(x, y, 0, 0)}
}
