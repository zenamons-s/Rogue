// Пакет gameplay содержит контроллеры поведения врагов.
package gameplay

import (
	"math/rand"
	"rogue-game/src/domain/entities"
)

// EnemyController управляет поведением врагов.
type EnemyController struct {
	Enemy *entities.Enemy
	Game  *Game
}

// NewEnemyController создаёт контроллер для врага.
func NewEnemyController(e *entities.Enemy, g *Game) *EnemyController {
	return &EnemyController{
		Enemy: e,
		Game:  g,
	}
}

// TakeTurn выполняет ход врага в зависимости от его типа.
func (ec *EnemyController) TakeTurn() {
	if !ec.Enemy.IsAlive() {
		return
	}

	// Проверяем, находится ли игрок в радиусе враждебности
	if ec.isPlayerInHostilityRange() {
		ec.chasePlayer()
	} else {
		ec.patrol()
	}
}

// isPlayerInHostilityRange определяет, находится ли игрок в радиусе враждебности.
func (ec *EnemyController) isPlayerInHostilityRange() bool {
	px, py := ec.Game.Player.Position.X, ec.Game.Player.Position.Y
	ex, ey := ec.Enemy.Position.X, ec.Enemy.Position.Y
	dist := abs(px-ex) + abs(py-ey) // манхэттенское расстояние
	return dist <= int(ec.Enemy.Hostility)
}

// chasePlayer двигает врага по направлению к игроку.
func (ec *EnemyController) chasePlayer() {
	nx, ny, ok := ec.nextStepToPlayer()
	if !ok {
		ec.patrol()
		return
	}
	ec.Enemy.Position.X = nx
	ec.Enemy.Position.Y = ny
}

// patrol выполняет патрулирование (случайное движение) в соответствии с типом врага.
func (ec *EnemyController) patrol() {
	switch ec.Enemy.Type {
	case entities.EnemyZombie:
		ec.randomMove()
	case entities.EnemyVampire:
		ec.randomMove()
	case entities.EnemyGhost:
		ec.teleport()
	case entities.EnemyOgre:
		ec.doubleMove()
	case entities.EnemySnakeMage:
		ec.diagonalMove()
	default:
		ec.randomMove()
	}
}

// randomMove двигает врага в случайном направлении.
func (ec *EnemyController) randomMove() {
	dirs := [][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}}
	rand.Shuffle(len(dirs), func(i, j int) { dirs[i], dirs[j] = dirs[j], dirs[i] })
	for _, d := range dirs {
		if ec.canMoveTo(ec.Enemy.Position.X+d[0], ec.Enemy.Position.Y+d[1]) {
			ec.move(d[0], d[1])
			return
		}
	}
}

// teleport телепортирует привидение в случайную позицию внутри комнаты.
func (ec *EnemyController) teleport() {
	room := ec.findRoomByPosition(ec.Enemy.Position)
	if room == nil {
		ec.randomMove()
		return
	}
	newX := room.X + 1 + rand.Intn(max(1, room.Width-2))
	newY := room.Y + 1 + rand.Intn(max(1, room.Height-2))
	if ec.canMoveTo(newX, newY) {
		ec.Enemy.Position.X = newX
		ec.Enemy.Position.Y = newY
	}
}

// doubleMove двигает огра на две клетки.
func (ec *EnemyController) doubleMove() {
	// Первый шаг
	ec.randomMove()
	// Второй шаг, если возможно
	ec.randomMove()
	idx := ec.Game.enemyIndex(ec.Enemy)
	ec.Game.OgreRestTurns[idx] = 1
}

// diagonalMove двигает змея-мага по диагонали.
func (ec *EnemyController) diagonalMove() {
	idx := ec.Game.enemyIndex(ec.Enemy)
	dirs := [][2]int{{1, 1}, {1, -1}, {-1, 1}, {-1, -1}}
	if ec.Game.SnakeSideLeft[idx] {
		dirs = [][2]int{{-1, 1}, {-1, -1}, {1, 1}, {1, -1}}
	}
	ec.Game.SnakeSideLeft[idx] = !ec.Game.SnakeSideLeft[idx]
	for _, d := range dirs {
		if ec.canMoveTo(ec.Enemy.Position.X+d[0], ec.Enemy.Position.Y+d[1]) {
			ec.move(d[0], d[1])
			return
		}
	}
	// Если диагональное движение невозможно, двигаемся случайно
	ec.randomMove()
}

// canMoveTo проверяет, может ли враг переместиться в указанную клетку.
func (ec *EnemyController) canMoveTo(x, y int) bool {
	// Проверка выхода за границы уровня
	if x < 0 || y < 0 || x >= ec.Game.CurrentLevel.Width || y >= ec.Game.CurrentLevel.Height {
		return false
	}
	// Проверка на проходимость клетки (стена и т.д.)
	if !ec.Game.isTileWalkable(x, y) {
		return false
	}
	for _, e := range ec.Game.Enemies {
		if e != ec.Enemy && e.Position.X == x && e.Position.Y == y {
			return false
		}
	}
	// Нельзя вставать на клетку с игроком (это будет атака)
	if ec.Game.Player.Position.X == x && ec.Game.Player.Position.Y == y {
		return false
	}
	return true
}

// move перемещает врага на указанное смещение.
func (ec *EnemyController) move(dx, dy int) {
	ec.Enemy.Position.X += dx
	ec.Enemy.Position.Y += dy
}

// Attack наносит удар игроку.
func (ec *EnemyController) Attack(player *entities.Character) bool {
	cs := NewCombatSystem(ec.Game)
	return cs.Attack(ec.Enemy, player)
}

// calculateDamage вычисляет урон, наносимый врагом.
func (ec *EnemyController) calculateDamage() int {
	baseDamage := ec.Enemy.Strength
	// Модификаторы в зависимости от типа
	switch ec.Enemy.Type {
	case entities.EnemyOgre:
		baseDamage *= 2
	}
	return baseDamage
}

func (ec *EnemyController) findRoomByPosition(p entities.Position) *entities.Room {
	for _, room := range ec.Game.CurrentLevel.Rooms {
		if p.X >= room.X && p.X < room.X+room.Width && p.Y >= room.Y && p.Y < room.Y+room.Height {
			return room
		}
	}
	return nil
}

func (ec *EnemyController) nextStepToPlayer() (int, int, bool) {
	start := ec.Enemy.Position
	goal := ec.Game.Player.Position
	type node struct{ x, y int }
	q := []node{{start.X, start.Y}}
	prev := map[node]node{}
	seen := map[node]bool{{start.X, start.Y}: true}
	dirs := [][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}}
	for len(q) > 0 {
		cur := q[0]
		q = q[1:]
		if cur.x == goal.X && cur.y == goal.Y {
			break
		}
		for _, d := range dirs {
			n := node{cur.x + d[0], cur.y + d[1]}
			if seen[n] {
				continue
			}
			if !(n.x == goal.X && n.y == goal.Y) && !ec.canMoveTo(n.x, n.y) {
				continue
			}
			seen[n] = true
			prev[n] = cur
			q = append(q, n)
		}
	}
	t := node{goal.X, goal.Y}
	if !seen[t] {
		return 0, 0, false
	}
	for {
		p, ok := prev[t]
		if !ok {
			return t.x, t.y, true
		}
		if p.x == start.X && p.y == start.Y {
			return t.x, t.y, true
		}
		t = p
	}
}

// abs возвращает абсолютное значение целого числа.
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
