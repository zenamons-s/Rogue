// Пакет generation содержит алгоритмы генерации коридоров и минимального остовного дерева (MST).
package generation

import (
	"sort"
	"rogue-game/src/domain/entities"
)

// edge представляет ребро между двумя комнатами с весом.
type edge struct {
	from, to int
	weight   int
}

// generateEdges создаёт все возможные рёбра между комнатами.
func (lg *LevelGenerator) generateEdges(rooms []*entities.Room) []edge {
	var edges []edge
	for i := 0; i < len(rooms); i++ {
		for j := i + 1; j < len(rooms); j++ {
			weight := lg.distanceBetweenRooms(rooms[i], rooms[j])
			edges = append(edges, edge{from: i, to: j, weight: weight})
		}
	}
	return edges
}

// distanceBetweenRooms вычисляет манхэттенское расстояние между центрами комнат.
func (lg *LevelGenerator) distanceBetweenRooms(a, b *entities.Room) int {
	centerAX := a.X + a.Width/2
	centerAY := a.Y + a.Height/2
	centerBX := b.X + b.Width/2
	centerBY := b.Y + b.Height/2
	return abs(centerAX-centerBX) + abs(centerAY-centerBY)
}

// find возвращает корень множества в DSU.
func find(parent []int, x int) int {
	if parent[x] != x {
		parent[x] = find(parent, parent[x])
	}
	return parent[x]
}

// union объединяет два множества в DSU.
func union(parent, rank []int, x, y int) {
	rootX := find(parent, x)
	rootY := find(parent, y)
	if rootX != rootY {
		if rank[rootX] < rank[rootY] {
			parent[rootX] = rootY
		} else if rank[rootX] > rank[rootY] {
			parent[rootY] = rootX
		} else {
			parent[rootY] = rootX
			rank[rootX]++
		}
	}
}

// generateMST возвращает рёбра минимального остовного дерева (алгоритм Крускала).
func (lg *LevelGenerator) generateMST(edges []edge, numRooms int) []edge {
	// Сортируем рёбра по весу
	sort.Slice(edges, func(i, j int) bool {
		return edges[i].weight < edges[j].weight
	})

	parent := make([]int, numRooms)
	rank := make([]int, numRooms)
	for i := 0; i < numRooms; i++ {
		parent[i] = i
	}

	var mst []edge
	for _, e := range edges {
		if find(parent, e.from) != find(parent, e.to) {
			union(parent, rank, e.from, e.to)
			mst = append(mst, e)
			if len(mst) == numRooms-1 {
				break
			}
		}
	}
	return mst
}

// generateCorridors создаёт коридоры между комнатами на основе MST.
func (lg *LevelGenerator) generateCorridors(rooms []*entities.Room) []*entities.Corridor {
	edges := lg.generateEdges(rooms)
	mst := lg.generateMST(edges, len(rooms))

	var corridors []*entities.Corridor
	for _, e := range mst {
		corridor := lg.createCorridorBetweenRooms(rooms[e.from], rooms[e.to])
		corridor.ID = len(corridors)
		corridors = append(corridors, corridor)
	}
	return corridors
}

// abs возвращает абсолютное значение.
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}