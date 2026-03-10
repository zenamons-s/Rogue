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
	px, py := ec.Game.Player.Position.X, ec.Game.Player.Position.Y
	ex, ey := ec.Enemy.Position.X, ec.Enemy.Position.Y

	// Определяем направление движения
	dx, dy := 0, 0
	if px > ex {
		dx = 1
	} else if px < ex {
		dx = -1
	}
	if py > ey {
		dy = 1
	} else if py < ey {
		dy = -1
	}

	// Пытаемся двигаться по направлению к игроку
	if dx != 0 && ec.canMoveTo(ex+dx, ey) {
		ec.move(dx, 0)
	} else if dy != 0 && ec.canMoveTo(ex, ey+dy) {
		ec.move(0, dy)
	} else {
		// Если нельзя двигаться прямо, делаем случайный ход
		ec.randomMove()
	}
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
	// TODO: определить границы комнаты
	// Пока просто случайное перемещение в пределах уровня
	newX := rand.Intn(ec.Game.CurrentLevel.Width)
	newY := rand.Intn(ec.Game.CurrentLevel.Height)
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
}

// diagonalMove двигает змея-мага по диагонали.
func (ec *EnemyController) diagonalMove() {
	dirs := [][2]int{{1, 1}, {1, -1}, {-1, 1}, {-1, -1}}
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
	// TODO: использовать карту уровня
	// Пока считаем все клетки проходимыми, кроме занятых другими врагами
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
func (ec *EnemyController) Attack(player *entities.Character) {
	// TODO: реализовать логику атаки с учётом типа врага
	damage := ec.calculateDamage()
	player.TakeDamage(damage)
}

// calculateDamage вычисляет урон, наносимый врагом.
func (ec *EnemyController) calculateDamage() int {
	baseDamage := ec.Enemy.Strength
	// Модификаторы в зависимости от типа
	switch ec.Enemy.Type {
	case entities.EnemyVampire:
		// Вампир отнимает максимальное здоровье
		// TODO: реализовать
	case entities.EnemyOgre:
		baseDamage *= 2
	}
	return baseDamage
}

// abs возвращает абсолютное значение целого числа.
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}