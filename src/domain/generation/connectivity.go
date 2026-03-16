// Пакет generation содержит алгоритмы проверки и обеспечения связности уровней.
package generation

import (
	"rogue-game/src/domain/entities"
)

// isGraphConnected проверяет, является ли граф комнат связным на основе коридоров.
func (lg *LevelGenerator) isGraphConnected(rooms []*entities.Room, corridors []*entities.Corridor) bool {
	if len(rooms) == 0 {
		return true
	}
	parent := make([]int, len(rooms))
	for i := range parent {
		parent[i] = i
	}

	var find func(int) int
	find = func(x int) int {
		if parent[x] != x {
			parent[x] = find(parent[x])
		}
		return parent[x]
	}
	union := func(x, y int) {
		rootX := find(x)
		rootY := find(y)
		if rootX != rootY {
			parent[rootY] = rootX
		}
	}

	for _, c := range corridors {
		union(c.FromRoom, c.ToRoom)
	}

	root := find(0)
	for i := 1; i < len(rooms); i++ {
		if find(i) != root {
			return false
		}
	}
	return true
}

// ensureConnectivity добавляет дополнительные коридоры, если граф несвязный.
func (lg *LevelGenerator) ensureConnectivity(rooms []*entities.Room, corridors []*entities.Corridor) []*entities.Corridor {
	if lg.isGraphConnected(rooms, corridors) {
		return corridors
	}
	// Если несвязно, добавляем недостающие рёбра между компонентами
	// Простой алгоритм: соединяем первую комнату каждой компоненты с первой комнатой следующей компоненты
	// Но для простоты просто добавим рёбра между всеми комнатами (неэффективно)
	// Вместо этого используем полный граф и строим MST заново.
	// Перегенерируем коридоры с помощью MST (уже сделано в generateCorridors)
	// Поэтому просто возвращаем новые коридоры, сгенерированные MST.
	return lg.generateCorridors(rooms)
}