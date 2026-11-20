// Package geom provides 2D geometry primitives used throughout the optimizer.
package geom

import (
	"fmt"
	"math"
)

// Point is an integer tile coordinate (top-left of a tile).
type Point struct {
	X, Y int
}

func (p Point) Add(o Point) Point { return Point{p.X + o.X, p.Y + o.Y} }
func (p Point) Sub(o Point) Point { return Point{p.X - o.X, p.Y - o.Y} }
func (p Point) Mul(k int) Point   { return Point{p.X * k, p.Y * k} }

func (p Point) Manhattan(o Point) int {
	return abs(p.X-o.X) + abs(p.Y-o.Y)
}

func (p Point) Dist(o Point) float64 {
	return math.Hypot(float64(p.X-o.X), float64(p.Y-o.Y))
}

func (p Point) String() string { return fmt.Sprintf("(%d,%d)", p.X, p.Y) }

// Tile is an alias for Point used for grid cells.
type Tile = Point

// Direction is one of the four cardinal directions.
// Value16 is the Factorio 2.0 blueprint defines.direction (N=0, E=4, S=8, W=12).
type Direction int

const (
	North Direction = iota
	East
	South
	West
)

func (d Direction) Opposite() Direction { return (d + 2) % 4 }
func (d Direction) CW() Direction       { return (d + 1) % 4 }
func (d Direction) CCW() Direction      { return (d + 3) % 4 }
func (d Direction) Axis() int {
	if d == North || d == South {
		return 1
	}
	return 0
}

func (d Direction) Delta() Point {
	switch d {
	case North:
		return Point{0, -1}
	case East:
		return Point{1, 0}
	case South:
		return Point{0, 1}
	case West:
		return Point{-1, 0}
	}
	return Point{}
}

func (d Direction) Value16() int {
	switch d {
	case North:
		return 0
	case East:
		return 4
	case South:
		return 8
	case West:
		return 12
	}
	return 0
}

func DirectionFrom16(v int) Direction {
	switch v {
	case 0, 1, 2, 15:
		return North
	case 3, 4, 5:
		return East
	case 6, 7, 8, 9:
		return South
	case 10, 11, 12, 13:
		return West
	}
	return North
}

func (d Direction) String() string {
	switch d {
	case North:
		return "N"
	case East:
		return "E"
	case South:
		return "S"
	case West:
		return "W"
	}
	return "?"
}

// AllDirections is N,E,S,W.
var AllDirections = []Direction{North, East, South, West}

// Rect is an axis-aligned integer rectangle [Min, Max).
type Rect struct {
	Min, Max Point // Max is exclusive
}

func NewRect(x, y, w, h int) Rect {
	return Rect{Min: Point{x, y}, Max: Point{x + w, y + h}}
}

func (r Rect) Width() int  { return r.Max.X - r.Min.X }
func (r Rect) Height() int { return r.Max.Y - r.Min.Y }
func (r Rect) Area() int   { return r.Width() * r.Height() }

func (r Rect) Contains(p Point) bool {
	return p.X >= r.Min.X && p.X < r.Max.X && p.Y >= r.Min.Y && p.Y < r.Max.Y
}

func (r Rect) Expand(n int) Rect {
	return Rect{
		Min: Point{r.Min.X - n, r.Min.Y - n},
		Max: Point{r.Max.X + n, r.Max.Y + n},
	}
}

func (r Rect) Intersects(o Rect) bool {
	return r.Min.X < o.Max.X && r.Max.X > o.Min.X &&
		r.Min.Y < o.Max.Y && r.Max.Y > o.Min.Y
}

// Centre returns the geometric centre of the rect.
func (r Rect) Centre() (float64, float64) {
	return float64(r.Min.X+r.Max.X) / 2, float64(r.Min.Y+r.Max.Y) / 2
}

// EdgeTiles returns tiles just outside the given side of the rect.
func (r Rect) EdgeTiles(side Direction) []Point {
	var out []Point
	switch side {
	case North:
		for x := r.Min.X; x < r.Max.X; x++ {
			out = append(out, Point{x, r.Min.Y - 1})
		}
	case South:
		for x := r.Min.X; x < r.Max.X; x++ {
			out = append(out, Point{x, r.Max.Y})
		}
	case West:
		for y := r.Min.Y; y < r.Max.Y; y++ {
			out = append(out, Point{r.Min.X - 1, y})
		}
	case East:
		for y := r.Min.Y; y < r.Max.Y; y++ {
			out = append(out, Point{r.Max.X, y})
		}
	}
	return out
}

// Vec2 is a floating-point 2D vector.
type Vec2 struct {
	X, Y float64
}

func (v Vec2) Add(o Vec2) Vec2     { return Vec2{v.X + o.X, v.Y + o.Y} }
func (v Vec2) Sub(o Vec2) Vec2     { return Vec2{v.X - o.X, v.Y - o.Y} }
func (v Vec2) Scale(k float64) Vec2 { return Vec2{v.X * k, v.Y * k} }
func (v Vec2) Length() float64     { return math.Hypot(v.X, v.Y) }
func (v Vec2) ToPoint() Point      { return Point{int(math.Floor(v.X)), int(math.Floor(v.Y))} }

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
