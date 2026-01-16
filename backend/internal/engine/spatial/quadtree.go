package spatial

import (
	"sync"

	"github.com/akarsh-2004/aether/internal/config"
	"github.com/akarsh-2004/aether/internal/engine/entity"
)

type Quadtree struct {
	bounds     Rectangle
	capacity   int
	depth      int
	maxDepth   int
	entities   []*entity.Entity
	children   [4]*Quadtree
	divided    bool
	mu         sync.RWMutex
}

type Rectangle struct {
	X, Y          float64
	Width, Height float64
}

func NewQuadtree(bounds Rectangle, capacity, depth, maxDepth int) *Quadtree {
	return &Quadtree{
		bounds:   bounds,
		capacity: capacity,
		depth:    depth,
		maxDepth: maxDepth,
		entities: make([]*entity.Entity, 0, capacity),
	}
}

func NewQuadtreeFromConfig(cfg config.EngineConfig) *Quadtree {
	bounds := Rectangle{
		X:      cfg.WorldBounds.MinX,
		Y:      cfg.WorldBounds.MinY,
		Width:  cfg.WorldBounds.MaxX - cfg.WorldBounds.MinX,
		Height: cfg.WorldBounds.MaxY - cfg.WorldBounds.MinY,
	}
	return NewQuadtree(bounds, cfg.QuadtreeCapacity, 0, cfg.QuadtreeDepth)
}

func (qt *Quadtree) Insert(ent *entity.Entity) bool {
	if !qt.bounds.Contains(ent.Position) {
		return false
	}

	qt.mu.Lock()
	defer qt.mu.Unlock()

	if len(qt.entities) < qt.capacity || qt.depth >= qt.maxDepth {
		qt.entities = append(qt.entities, ent)
		return true
	}

	if !qt.divided {
		qt.subdivide()
	}

	return qt.children[0].Insert(ent) ||
		qt.children[1].Insert(ent) ||
		qt.children[2].Insert(ent) ||
		qt.children[3].Insert(ent)
}

func (qt *Quadtree) Remove(ent *entity.Entity) bool {
	qt.mu.Lock()
	defer qt.mu.Unlock()

	return qt.removeInternal(ent)
}

func (qt *Quadtree) removeInternal(ent *entity.Entity) bool {
	// Check current level
	for i, e := range qt.entities {
		if e.ID == ent.ID {
			qt.entities = append(qt.entities[:i], qt.entities[i+1:]...)
			return true
		}
	}

	// Check children if divided
	if qt.divided {
		return qt.children[0].removeInternal(ent) ||
			qt.children[1].removeInternal(ent) ||
			qt.children[2].removeInternal(ent) ||
			qt.children[3].removeInternal(ent)
	}

	return false
}

func (qt *Quadtree) Update(ent *entity.Entity, oldPos entity.Vector2) bool {
	// Remove from old position
	qt.Remove(ent)
	
	// Insert at new position
	return qt.Insert(ent)
}

func (qt *Quadtree) QueryRadius(center entity.Vector2, radius float64) []*entity.Entity {
	var results []*entity.Entity
	radiusSq := radius * radius

	qt.mu.RLock()
	defer qt.mu.RUnlock()

	qt.queryRadiusInternal(center, radiusSq, &results)
	return results
}

func (qt *Quadtree) queryRadiusInternal(center entity.Vector2, radiusSq float64, results *[]*entity.Entity) {
	// Check entities at this level
	for _, ent := range qt.entities {
		distSq := ent.Position.Distance(center)
		if distSq <= radiusSq {
			*results = append(*results, ent)
		}
	}

	// Check children if divided
	if qt.divided {
		for _, child := range qt.children {
			if child.bounds.IntersectsCircle(center, radiusSq) {
				child.queryRadiusInternal(center, radiusSq, results)
			}
		}
	}
}

func (qt *Quadtree) QueryBounds(bounds Rectangle) []*entity.Entity {
	var results []*entity.Entity

	qt.mu.RLock()
	defer qt.mu.RUnlock()

	qt.queryBoundsInternal(bounds, &results)
	return results
}

func (qt *Quadtree) queryBoundsInternal(bounds Rectangle, results *[]*entity.Entity) {
	// Check entities at this level
	for _, ent := range qt.entities {
		if bounds.Contains(ent.Position) {
			*results = append(*results, ent)
		}
	}

	// Check children if divided
	if qt.divided {
		for _, child := range qt.children {
			if child.bounds.Intersects(bounds) {
				child.queryBoundsInternal(bounds, results)
			}
		}
	}
}

func (qt *Quadtree) Clear() {
	qt.mu.Lock()
	defer qt.mu.Unlock()

	qt.entities = qt.entities[:0]
	if qt.divided {
		for _, child := range qt.children {
			child.Clear()
		}
		qt.divided = false
		qt.children = [4]*Quadtree{}
	}
}

func (qt *Quadtree) GetStats() QuadtreeStats {
	qt.mu.RLock()
	defer qt.mu.RUnlock()

	stats := QuadtreeStats{
		Depth:    qt.depth,
		Entities: len(qt.entities),
		Divided:  qt.divided,
	}

	if qt.divided {
		for _, child := range qt.children {
			childStats := child.GetStats()
			stats.TotalEntities += childStats.TotalEntities
			stats.ChildCount++
		}
	} else {
		stats.TotalEntities = stats.Entities
	}

	return stats
}

type QuadtreeStats struct {
	Depth         int
	Entities      int
	TotalEntities int
	Divided       bool
	ChildCount    int
}

func (qt *Quadtree) subdivide() {
	x := qt.bounds.X
	y := qt.bounds.Y
	w := qt.bounds.Width / 2
	h := qt.bounds.Height / 2

	qt.children[0] = NewQuadtree(Rectangle{X: x, Y: y, Width: w, Height: h}, qt.capacity, qt.depth+1, qt.maxDepth) // NW
	qt.children[1] = NewQuadtree(Rectangle{X: x + w, Y: y, Width: w, Height: h}, qt.capacity, qt.depth+1, qt.maxDepth) // NE
	qt.children[2] = NewQuadtree(Rectangle{X: x, Y: y + h, Width: w, Height: h}, qt.capacity, qt.depth+1, qt.maxDepth) // SW
	qt.children[3] = NewQuadtree(Rectangle{X: x + w, Y: y + h, Width: w, Height: h}, qt.capacity, qt.depth+1, qt.maxDepth) // SE

	qt.divided = true

	// Re-insert existing entities into children
	for _, ent := range qt.entities {
		qt.children[0].Insert(ent) ||
			qt.children[1].Insert(ent) ||
			qt.children[2].Insert(ent) ||
			qt.children[3].Insert(ent)
	}

	qt.entities = qt.entities[:0] // Clear current level entities
}

func (r Rectangle) Contains(point entity.Vector2) bool {
	return point.X >= r.X &&
		point.X < r.X+r.Width &&
		point.Y >= r.Y &&
		point.Y < r.Y+r.Height
}

func (r Rectangle) Intersects(other Rectangle) bool {
	return !(r.X >= other.X+other.Width ||
		other.X >= r.X+r.Width ||
		r.Y >= other.Y+other.Height ||
		other.Y >= r.Y+r.Height)
}

func (r Rectangle) IntersectsCircle(center entity.Vector2, radiusSq float64) bool {
	// Find the closest point on the rectangle to the circle center
	closestX := max(r.X, min(center.X, r.X+r.Width))
	closestY := max(r.Y, min(center.Y, r.Y+r.Height))

	// Calculate distance from circle center to closest point
	dx := center.X - closestX
	dy := center.Y - closestY

	return (dx*dx + dy*dy) <= radiusSq
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
