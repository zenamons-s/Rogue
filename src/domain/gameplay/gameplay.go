package gameplay

import (
	"rogue-game/src/domain/entities"
)

// Game представляет игровую сессию с геймплейными правилами.
type Game struct {
	Session    *entities.GameSession
	CurrentLevel *entities.Level
	Player     *entities.Character
	Enemies    []*entities.Enemy
	Items      []*entities.Item
	Turn       int
	IsGameOver bool
}

// NewGame создаёт новую игровую сессию с начальным уровнем.
func NewGame(session *entities.GameSession) *Game {
	return &Game{
		Session:    session,
		CurrentLevel: session.Level,
		Player:     session.Player,
		Enemies:    collectEnemies(session.Level),
		Items:      collectItems(session.Level),
		Turn:       0,
		IsGameOver: false,
	}
}

// collectEnemies возвращает всех врагов на уровне.
func collectEnemies(level *entities.Level) []*entities.Enemy {
	var enemies []*entities.Enemy
	for _, room := range level.Rooms {
		enemies = append(enemies, room.Enemies...)
	}
	return enemies
}

// collectItems возвращает все предметы на уровне.
func collectItems(level *entities.Level) []*entities.Item {
	var items []*entities.Item
	for _, room := range level.Rooms {
		items = append(items, room.Items...)
	}
	return items
}

// MovePlayer перемещает персонажа в указанную позицию, если это возможно.
func (g *Game) MovePlayer(dx, dy int) bool {
	newX := g.Player.Position.X + dx
	newY := g.Player.Position.Y + dy

	// Проверка, что клетка проходима
	if !g.isTileWalkable(newX, newY) {
		return false
	}

	// Проверка на столкновение с врагом
	if enemy := g.enemyAt(newX, newY); enemy != nil {
		g.initiateCombat(g.Player, enemy)
		return true
	}

	// Проверка на предмет
	if item := g.itemAt(newX, newY); item != nil {
		g.pickUpItem(item)
	}

	g.Player.Position.X = newX
	g.Player.Position.Y = newY
	g.endPlayerTurn()
	return true
}

// isTileWalkable проверяет, можно ли пройти по клетке.
func (g *Game) isTileWalkable(x, y int) bool {
	// TODO: реализовать проверку по карте уровня
	// Пока возвращаем true для всех координат в пределах уровня
	if x < 0 || y < 0 || x >= g.CurrentLevel.Width || y >= g.CurrentLevel.Height {
		return false
	}
	// Предполагаем, что все клетки, кроме стен, проходимы
	// В будущем нужно учитывать TileType
	return true
}

// enemyAt возвращает врага в указанных координатах.
func (g *Game) enemyAt(x, y int) *entities.Enemy {
	for _, e := range g.Enemies {
		if e.Position.X == x && e.Position.Y == y {
			return e
		}
	}
	return nil
}

// itemAt возвращает предмет в указанных координатах.
func (g *Game) itemAt(x, y int) *entities.Item {
	for _, _ = range g.Items {
		// TODO: предметы должны иметь позицию
		// Пока заглушка
	}
	return nil
}

// initiateCombat начинает бой между персонажем и врагом.
func (g *Game) initiateCombat(attacker, defender interface{}) {
	// TODO: реализовать логику боя
}

// pickUpItem подбирает предмет.
func (g *Game) pickUpItem(item *entities.Item) {
	// TODO: добавить предмет в рюкзак
}

// endPlayerTurn завершает ход игрока и активирует ходы врагов.
func (g *Game) endPlayerTurn() {
	g.Turn++
	g.processEnemyTurns()
	g.checkGameOver()
}

// processEnemyTurns обрабатывает ходы всех врагов.
func (g *Game) processEnemyTurns() {
	for _, enemy := range g.Enemies {
		if enemy.IsAlive() {
			g.enemyTurn(enemy)
		}
	}
}

// enemyTurn выполняет ход врага (движение, атака).
func (g *Game) enemyTurn(enemy *entities.Enemy) {
	// TODO: реализовать AI врага
}

// checkGameOver проверяет условия завершения игры.
func (g *Game) checkGameOver() {
	if g.Player.Health <= 0 {
		g.IsGameOver = true
	}
}