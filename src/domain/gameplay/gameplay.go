package gameplay

import (
	"math"
	"math/rand"
	"rogue-game/src/domain/entities"
	"rogue-game/src/domain/generation"
)

// AttemptStats хранит статистику одного прохождения.
type AttemptStats struct {
	Treasures       int  `json:"treasures"`
	ReachedLevel    int  `json:"reached_level"`
	DefeatedEnemies int  `json:"defeated_enemies"`
	UsedFood        int  `json:"used_food"`
	UsedPotions     int  `json:"used_potions"`
	UsedScrolls     int  `json:"used_scrolls"`
	HitsDealt       int  `json:"hits_dealt"`
	HitsTaken       int  `json:"hits_taken"`
	TilesWalked     int  `json:"tiles_walked"`
	Won             bool `json:"won"`
}

// GroundItem описывает предмет на карте с позицией.
type GroundItem struct {
	Item      *entities.Item    `json:"item"`
	Position  entities.Position `json:"position"`
	Collected bool              `json:"collected"`
}

// Game представляет игровую сессию с геймплейными правилами.
type Game struct {
	Session      *entities.GameSession
	CurrentLevel *entities.Level
	Player       *entities.Character
	Enemies      []*entities.Enemy
	Items        []*entities.Item
	GroundItems  []*GroundItem
	ItemPos      map[int]entities.Position
	Turn         int
	IsGameOver   bool
	TileMap      [][]entities.TileType
	Visible      [][]bool
	Explored     [][]bool
	ExitPos      entities.Position
	Stats        AttemptStats
	Seed         int64
}

// NewGame создаёт новую игровую сессию с начальным уровнем.
func NewGame(session *entities.GameSession) *Game {
	g := &Game{
		Session:      session,
		CurrentLevel: session.Level,
		Player:       session.Player,
		Enemies:      collectEnemies(session.Level),
		Items:        collectItems(session.Level),
		GroundItems:  []*GroundItem{},
		ItemPos:      map[int]entities.Position{},
		Turn:         0,
		IsGameOver:   false,
		Stats: AttemptStats{
			ReachedLevel: session.CurrentFloor,
		},
	}
	g.rebuildLevelState()
	return g
}

// NewGeneratedGame создаёт новую игру с генерацией уровня и базовым наполнением.
func NewGeneratedGame(width, height int, seed int64) *Game {
	if seed == 0 {
		seed = rand.Int63()
	}
	lg := generation.NewLevelGenerator(width, height, seed)
	level := lg.Generate(1)
	populateLevel(level)
	start := findStartPosition(level)
	session := &entities.GameSession{
		ID:           "session",
		Player:       defaultPlayer(start),
		Level:        level,
		CurrentFloor: 1,
		Score:        0,
		IsActive:     true,
	}
	g := NewGame(session)
	g.Seed = seed
	return g
}

func defaultPlayer(start entities.Position) *entities.Character {
	return &entities.Character{
		MaxHealth: 100,
		Health:    100,
		Dexterity: 12,
		Strength:  10,
		Backpack: &entities.Backpack{
			Slots:    []*entities.Item{},
			Capacity: 30,
		},
		Position: start,
	}
}

func populateLevel(level *entities.Level) {
	for _, room := range level.Rooms {
		if room.IsStart {
			continue
		}
		enemy := &entities.Enemy{
			Type:      entities.EnemyType(rand.Intn(5)),
			Health:    20 + rand.Intn(20),
			Dexterity: 7 + rand.Intn(8),
			Strength:  5 + rand.Intn(8),
			Hostility: entities.HostilityAggressive,
			Position: entities.Position{
				X: room.X + room.Width/2,
				Y: room.Y + room.Height/2,
			},
		}
		room.Enemies = append(room.Enemies, enemy)

		item := &entities.Item{Type: entities.ItemTypeFood, Subtype: entities.SubtypeBread, HealthBoost: 20}
		if rand.Intn(2) == 0 {
			item = &entities.Item{Type: entities.ItemTypePotion, Subtype: entities.SubtypeHealthPotion, MaxHealthBoost: 5}
		}
		room.Items = append(room.Items, item)
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

func (g *Game) rebuildLevelState() {
	g.TileMap = buildTileMap(g.CurrentLevel)
	g.ExitPos = findExitPosition(g.CurrentLevel)
	h := g.CurrentLevel.Height
	w := g.CurrentLevel.Width
	g.Visible = make([][]bool, h)
	g.Explored = make([][]bool, h)
	for y := 0; y < h; y++ {
		g.Visible[y] = make([]bool, w)
		g.Explored[y] = make([]bool, w)
	}
	g.GroundItems = g.GroundItems[:0]
	g.ItemPos = map[int]entities.Position{}
	idx := 0
	for _, room := range g.CurrentLevel.Rooms {
		for i, it := range room.Items {
			pos := entities.Position{X: room.X + 1 + (i % max(1, room.Width-2)), Y: room.Y + 1 + ((i / max(1, room.Width-2)) % max(1, room.Height-2))}
			g.GroundItems = append(g.GroundItems, &GroundItem{Item: it, Position: pos})
			g.ItemPos[idx] = pos
			idx++
		}
	}
	g.updateVisibility(8)
}

func buildTileMap(level *entities.Level) [][]entities.TileType {
	tiles := make([][]entities.TileType, level.Height)
	for y := 0; y < level.Height; y++ {
		tiles[y] = make([]entities.TileType, level.Width)
		for x := 0; x < level.Width; x++ {
			tiles[y][x] = entities.TileWall
		}
	}
	if len(level.Rooms) == 0 {
		for y := 0; y < level.Height; y++ {
			for x := 0; x < level.Width; x++ {
				tiles[y][x] = entities.TileFloor
			}
		}
		return tiles
	}
	for _, room := range level.Rooms {
		for y := room.Y; y < room.Y+room.Height; y++ {
			for x := room.X; x < room.X+room.Width; x++ {
				if y == room.Y || y == room.Y+room.Height-1 || x == room.X || x == room.X+room.Width-1 {
					tiles[y][x] = entities.TileWall
				} else {
					tiles[y][x] = entities.TileFloor
				}
			}
		}
	}
	for _, c := range level.Corridors {
		for _, t := range c.Tiles {
			if t.Y >= 0 && t.Y < level.Height && t.X >= 0 && t.X < level.Width {
				tiles[t.Y][t.X] = entities.TileCorridor
			}
		}
	}
	return tiles
}

func findStartPosition(level *entities.Level) entities.Position {
	for _, room := range level.Rooms {
		if room.IsStart {
			return entities.Position{X: room.X + room.Width/2, Y: room.Y + room.Height/2}
		}
	}
	return entities.Position{X: 1, Y: 1}
}

func findExitPosition(level *entities.Level) entities.Position {
	for _, room := range level.Rooms {
		if room.IsExit {
			return entities.Position{X: room.X + room.Width/2, Y: room.Y + room.Height/2}
		}
	}
	return entities.Position{X: level.Width - 2, Y: level.Height - 2}
}

// MovePlayer перемещает персонажа в указанную позицию, если это возможно.
func (g *Game) MovePlayer(dx, dy int) bool {
	newX := g.Player.Position.X + dx
	newY := g.Player.Position.Y + dy

	if !g.isTileWalkable(newX, newY) {
		return false
	}

	if enemy := g.enemyAt(newX, newY); enemy != nil {
		g.initiateCombat(g.Player, enemy)
		g.endPlayerTurn()
		return true
	}

	if item := g.groundItemAt(newX, newY); item != nil {
		g.pickUpItem(item)
	}

	g.Player.Position.X = newX
	g.Player.Position.Y = newY
	g.Stats.TilesWalked++
	if newX == g.ExitPos.X && newY == g.ExitPos.Y {
		g.advanceLevel()
		return true
	}
	g.endPlayerTurn()
	return true
}

func (g *Game) advanceLevel() {
	g.Session.CurrentFloor++
	g.Stats.ReachedLevel = g.Session.CurrentFloor
	lg := generation.NewLevelGenerator(g.CurrentLevel.Width, g.CurrentLevel.Height, g.Seed+int64(g.Session.CurrentFloor))
	lvl := lg.Generate(g.Session.CurrentFloor)
	populateLevel(lvl)
	g.CurrentLevel = lvl
	g.Session.Level = lvl
	g.Enemies = collectEnemies(lvl)
	g.Items = collectItems(lvl)
	g.Player.Position = findStartPosition(lvl)
	g.rebuildLevelState()
}

// isTileWalkable проверяет, можно ли пройти по клетке.
func (g *Game) isTileWalkable(x, y int) bool {
	if x < 0 || y < 0 || x >= g.CurrentLevel.Width || y >= g.CurrentLevel.Height {
		return false
	}
	t := g.TileMap[y][x]
	return t == entities.TileFloor || t == entities.TileDoor || t == entities.TileCorridor
}

// enemyAt возвращает врага в указанных координатах.
func (g *Game) enemyAt(x, y int) *entities.Enemy {
	for _, e := range g.Enemies {
		if e.IsAlive() && e.Position.X == x && e.Position.Y == y {
			return e
		}
	}
	return nil
}

func (g *Game) groundItemAt(x, y int) *GroundItem {
	for _, gi := range g.GroundItems {
		if !gi.Collected && gi.Position.X == x && gi.Position.Y == y {
			return gi
		}
	}
	return nil
}

// itemAt возвращает предмет в указанных координатах.
func (g *Game) itemAt(x, y int) *entities.Item {
	if gi := g.groundItemAt(x, y); gi != nil {
		return gi.Item
	}
	return nil
}

// initiateCombat начинает бой между персонажем и врагом.
func (g *Game) initiateCombat(attacker, defender interface{}) {
	cs := NewCombatSystem(g)
	hit := cs.Attack(attacker, defender)
	if hit {
		g.Stats.HitsDealt++
	}
	if d, ok := defender.(*entities.Enemy); ok && !d.IsAlive() {
		g.Stats.DefeatedEnemies++
		g.Session.Score += 10
	}
}

// pickUpItem подбирает предмет.
func (g *Game) pickUpItem(item *GroundItem) {
	if g.Player.Backpack.AddItem(item.Item) {
		item.Collected = true
	}
}

// endPlayerTurn завершает ход игрока и активирует ходы врагов.
func (g *Game) endPlayerTurn() {
	g.Turn++
	g.processEnemyTurns()
	g.checkGameOver()
	g.updateVisibility(8)
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
	ec := NewEnemyController(enemy, g)
	px, py := g.Player.Position.X, g.Player.Position.Y
	if math.Abs(float64(px-enemy.Position.X))+math.Abs(float64(py-enemy.Position.Y)) == 1 {
		ec.Attack(g.Player)
		g.Stats.HitsTaken++
		return
	}
	ec.TakeTurn()
}

// checkGameOver проверяет условия завершения игры.
func (g *Game) checkGameOver() {
	if g.Player.Health <= 0 {
		g.IsGameOver = true
	}
}

func (g *Game) updateVisibility(radius int) {
	for y := range g.Visible {
		for x := range g.Visible[y] {
			g.Visible[y][x] = false
		}
	}
	px, py := g.Player.Position.X, g.Player.Position.Y
	for y := py - radius; y <= py+radius; y++ {
		for x := px - radius; x <= px+radius; x++ {
			if x < 0 || y < 0 || x >= g.CurrentLevel.Width || y >= g.CurrentLevel.Height {
				continue
			}
			if lineOfSight(g.TileMap, px, py, x, y) {
				g.Visible[y][x] = true
				g.Explored[y][x] = true
			}
		}
	}
}

func lineOfSight(grid [][]entities.TileType, x0, y0, x1, y1 int) bool {
	dx := abs(x1 - x0)
	dy := -abs(y1 - y0)
	sx := -1
	if x0 < x1 {
		sx = 1
	}
	sy := -1
	if y0 < y1 {
		sy = 1
	}
	err := dx + dy
	x, y := x0, y0
	for {
		if x == x1 && y == y1 {
			return true
		}
		if !(x == x0 && y == y0) && grid[y][x] == entities.TileWall {
			return false
		}
		e2 := 2 * err
		if e2 >= dy {
			err += dy
			x += sx
		}
		if e2 <= dx {
			err += dx
			y += sy
		}
		if y < 0 || y >= len(grid) || x < 0 || x >= len(grid[0]) {
			return false
		}
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
