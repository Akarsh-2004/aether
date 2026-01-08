package engine

// AABB represents an Axis-Aligned Bounding Box.
type AABB struct {
	CenterX float64
	CenterY float64
	HalfDim float64 // Half the dimension (width/height) of the box
}

// NewAABB creates a new AABB.
func NewAABB(centerX, centerY, halfDim float64) *AABB {
	return &AABB{
		CenterX: centerX,
		CenterY: centerY,
		HalfDim: halfDim,
	}
}

// ContainsPoint checks if a point is within the bounding box.
func (aabb *AABB) ContainsPoint(point *Vec2) bool {
	return (point.X >= aabb.CenterX-aabb.HalfDim &&
		point.X <= aabb.CenterX+aabb.HalfDim &&
		point.Y >= aabb.CenterY-aabb.HalfDim &&
		point.Y <= aabb.CenterY+aabb.HalfDim)
}

// IntersectsAABB checks if two bounding boxes intersect.
func (aabb *AABB) IntersectsAABB(other *AABB) bool {
	// Check if the boxes are separate. If they are, they don't intersect.
	if aabb.CenterX+aabb.HalfDim < other.CenterX-other.HalfDim ||
		aabb.CenterX-aabb.HalfDim > other.CenterX+other.HalfDim {
		return false
	}
	if aabb.CenterY+aabb.HalfDim < other.CenterY-other.HalfDim ||
		aabb.CenterY-aabb.HalfDim > other.CenterY+other.HalfDim {
		return false
	}
	return true
}

// Quadtree represents a node in the quadtree.
type Quadtree struct {
	boundary *AABB
	capacity int // Max number of entities before subdividing
	entities []*Entity

	// Children
	northWest *Quadtree
	northEast *Quadtree
	southWest *Quadtree
	southEast *Quadtree

	divided bool
}

// NewQuadtree creates a new Quadtree node.
func NewQuadtree(boundary *AABB, capacity int) *Quadtree {
	return &Quadtree{
		boundary: boundary,
		capacity: capacity,
		entities: make([]*Entity, 0, capacity),
	}
}

func (qt *Quadtree) subdivide() {
	centerX := qt.boundary.CenterX
	centerY := qt.boundary.CenterY
	halfDim := qt.boundary.HalfDim / 2

	nwAABB := NewAABB(centerX-halfDim, centerY+halfDim, halfDim)
	qt.northWest = NewQuadtree(nwAABB, qt.capacity)

	neAABB := NewAABB(centerX+halfDim, centerY+halfDim, halfDim)
	qt.northEast = NewQuadtree(neAABB, qt.capacity)

	swAABB := NewAABB(centerX-halfDim, centerY-halfDim, halfDim)
	qt.southWest = NewQuadtree(swAABB, qt.capacity)

	seAABB := NewAABB(centerX+halfDim, centerY-halfDim, halfDim)
	qt.southEast = NewQuadtree(seAABB, qt.capacity)

	qt.divided = true
}

// Insert adds an entity to the quadtree.
func (qt *Quadtree) Insert(entity *Entity) bool {
	// If the entity is not in this quadtree's boundary, do nothing.
	if !qt.boundary.ContainsPoint(&entity.Position) {
		return false
	}

	// If there is space in this quadtree, add the entity here.
	if len(qt.entities) < qt.capacity {
		qt.entities = append(qt.entities, entity)
		return true
	}

	// If the quadtree is full, subdivide it.
	if !qt.divided {
		qt.subdivide()
	}

	// Try to insert the entity into the appropriate child quadtree.
	if qt.northWest.Insert(entity) {
		return true
	}
	if qt.northEast.Insert(entity) {
		return true
	}
	if qt.southWest.Insert(entity) {
		return true
	}
	if qt.southEast.Insert(entity) {
		return true
	}

	// If it cannot fit in any child, keep it in the parent.
	// This can happen if the entity is on the border between quadrants.
	qt.entities = append(qt.entities, entity)
	return true
}

// QueryAABB returns all entities within a given AABB.
func (qt *Quadtree) QueryAABB(queryRange *AABB, found []*Entity) []*Entity {
	// If the query range does not intersect this quadtree's boundary, do nothing.
	if !qt.boundary.IntersectsAABB(queryRange) {
		return found
	}

	// Check objects at this level.
	for _, entity := range qt.entities {
		if queryRange.ContainsPoint(&entity.Position) {
			found = append(found, entity)
		}
	}

	// If the quadtree is subdivided, check its children.
	if qt.divided {
		found = qt.northWest.QueryAABB(queryRange, found)
		found = qt.northEast.QueryAABB(queryRange, found)
		found = qt.southWest.QueryAABB(queryRange, found)
		found = qt.southEast.QueryAABB(queryRange, found)
	}

	return found
}

// Circle represents a circular area for queries.
type Circle struct {
	CenterX float64
	CenterY float64
	Radius  float64
}

// ContainsPoint checks if a point is within the circle.
func (c *Circle) ContainsPoint(point *Vec2) bool {
	dx := point.X - c.CenterX
	dy := point.Y - c.CenterY
	return dx*dx+dy*dy <= c.Radius*c.Radius
}

// IntersectsAABB checks if the circle intersects with an AABB.
func (c *Circle) IntersectsAABB(aabb *AABB) bool {
	// Find the closest point on the AABB to the center of the circle
	closestX := clamp(c.CenterX, aabb.CenterX-aabb.HalfDim, aabb.CenterX+aabb.HalfDim)
	closestY := clamp(c.CenterY, aabb.CenterY-aabb.HalfDim, aabb.CenterY+aabb.HalfDim)

	// Calculate the distance between the circle's center and this closest point
	distanceX := c.CenterX - closestX
	distanceY := c.CenterY - closestY

	// If the distance is less than the circle's radius, there is an intersection
	return (distanceX*distanceX + distanceY*distanceY) < (c.Radius * c.Radius)
}

func clamp(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// QueryCircle returns all entities within a given circle.
func (qt *Quadtree) QueryCircle(queryRange *Circle, found []*Entity) []*Entity {
	// If the query range does not intersect this quadtree's boundary, do nothing.
	if !queryRange.IntersectsAABB(qt.boundary) {
		return found
	}

	// Check objects at this level.
	for _, entity := range qt.entities {
		if queryRange.ContainsPoint(&entity.Position) {
			found = append(found, entity)
		}
	}

	// If the quadtree is subdivided, check its children.
	if qt.divided {
		found = qt.northWest.QueryCircle(queryRange, found)
		found = qt.northEast.QueryCircle(queryRange, found)
		found = qt.southWest.QueryCircle(queryRange, found)
		found = qt.southEast.QueryCircle(queryRange, found)
	}

	return found
}

// Remove removes an entity from the quadtree.
func (qt *Quadtree) Remove(entityID string) bool {
	// Check objects at this level.
	for i, entity := range qt.entities {
		if entity.ID == entityID {
			// Remove the entity from the slice.
			qt.entities = append(qt.entities[:i], qt.entities[i+1:]...)
			return true
		}
	}

	// If the quadtree is subdivided, check its children.
	if qt.divided {
		if qt.northWest.Remove(entityID) {
			return true
		}
		if qt.northEast.Remove(entityID) {
			return true
		}
		if qt.southWest.Remove(entityID) {
			return true
		}
		if qt.southEast.Remove(entityID) {
			return true
		}
	}

	return false
}

// Update updates an entity's position in the quadtree.
// For now, it simply removes and re-inserts it if it has moved outside its current quadrant.
// A more optimized version would only do this if it leaves the current boundary.
func (qt *Quadtree) Update(entity *Entity) bool {
	if qt.Remove(entity.ID) {
		return qt.Insert(entity)
	}
	return false
}
