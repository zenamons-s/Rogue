// Пакет generation отвечает за процедурную генерацию уровней (комнаты, коридоры).
package generation

import (
	"math/rand"
	"rogue-game/src/domain/entities"
)

// LevelGenerator отвечает за генерацию уровня.
type LevelGenerator struct {
	Width  int
	Height int
	Seed   int64
}

// NewLevelGenerator создаёт новый генератор уровней.
func NewLevelGenerator(width, height int, seed int64) *LevelGenerator {
	if seed == 0 {
		seed = rand.Int63()
	}
	rand.Seed(seed)
	return &LevelGenerator{
		Width:  width,
		Height: height,
		Seed:   seed,
	}
}

// Generate создаёт новый уровень с комнатами и коридорами.
func (lg *LevelGenerator) Generate(levelID int) *entities.Level {
	// Генерация 9 комнат
	rooms := lg.generateRooms()

	// Генерация коридоров, обеспечивающих связность
	corridors := lg.generateCorridors(rooms)

	// Выбор стартовой и конечной комнат
	startRoomIndex, exitRoomIndex := lg.selectStartAndExitRooms(rooms)

	rooms[startRoomIndex].IsStart = true
	rooms[exitRoomIndex].IsExit = true

	// Создание уровня
	level := &entities.Level{
		ID:        levelID,
		Rooms:     rooms,
		Corridors: corridors,
		Width:     lg.Width,
		Height:    lg.Height,
	}
	return level
}

// generateRooms создаёт 9 комнат, размещённых в сетке 3x3.
func (lg *LevelGenerator) generateRooms() []*entities.Room {
	rooms := make([]*entities.Room, 9)
	sectionWidth := lg.Width / 3
	sectionHeight := lg.Height / 3

	for i := 0; i < 9; i++ {
		// Секция (row, col)
		row := i / 3
		col := i % 3

		// Случайные размеры комнаты (минимум 3x3, максимум размер секции - 2)
		minRoomWidth := 3
		minRoomHeight := 3
		maxRoomWidth := sectionWidth - 2
		maxRoomHeight := sectionHeight - 2
		if maxRoomWidth < minRoomWidth {
			maxRoomWidth = minRoomWidth
		}
		if maxRoomHeight < minRoomHeight {
			maxRoomHeight = minRoomHeight
		}
		roomWidth := rand.Intn(maxRoomWidth-minRoomWidth+1) + minRoomWidth
		roomHeight := rand.Intn(maxRoomHeight-minRoomHeight+1) + minRoomHeight

		// Случайное положение внутри секции
		maxX := sectionWidth - roomWidth
		maxY := sectionHeight - roomHeight
		if maxX < 0 {
			maxX = 0
		}
		if maxY < 0 {
			maxY = 0
		}
		offsetX := rand.Intn(maxX + 1)
		offsetY := rand.Intn(maxY + 1)

		// Глобальные координаты
		x := col*sectionWidth + offsetX
		y := row*sectionHeight + offsetY

		rooms[i] = &entities.Room{
			ID:      i,
			X:       x,
			Y:       y,
			Width:   roomWidth,
			Height:  roomHeight,
			Enemies: []*entities.Enemy{},
			Items:   []*entities.Item{},
			IsStart: false,
			IsExit:  false,
		}
	}
	return rooms
}

// createCorridorBetweenRooms создаёт коридор между двумя комнатами.
func (lg *LevelGenerator) createCorridorBetweenRooms(roomA, roomB *entities.Room) *entities.Corridor {
	// Находим центры комнат
	centerAX := roomA.X + roomA.Width/2
	centerAY := roomA.Y + roomA.Height/2
	centerBX := roomB.X + roomB.Width/2
	centerBY := roomB.Y + roomB.Height/2

	// Генерируем L-образный коридор: сначала по горизонтали, потом по вертикали
	var tiles []entities.Tile

	// Горизонтальный сегмент
	startX, endX := centerAX, centerBX
	if startX > endX {
		startX, endX = endX, startX
	}
	for x := startX; x <= endX; x++ {
		tiles = append(tiles, entities.Tile{X: x, Y: centerAY, Type: entities.TileCorridor})
	}

	// Вертикальный сегмент
	startY, endY := centerAY, centerBY
	if startY > endY {
		startY, endY = endY, startY
	}
	for y := startY; y <= endY; y++ {
		// Чтобы не дублировать угол, пропускаем точку пересечения
		if y != centerAY {
			tiles = append(tiles, entities.Tile{X: centerBX, Y: y, Type: entities.TileCorridor})
		}
	}

	return &entities.Corridor{
		ID:       len(tiles), // временный ID
		FromRoom: roomA.ID,
		ToRoom:   roomB.ID,
		Tiles:    tiles,
	}
}

// selectStartAndExitRooms выбирает случайные комнаты для старта и выхода (не одну и ту же).
func (lg *LevelGenerator) selectStartAndExitRooms(rooms []*entities.Room) (start, exit int) {
	start = rand.Intn(len(rooms))
	exit = rand.Intn(len(rooms))
	for exit == start {
		exit = rand.Intn(len(rooms))
	}
	return start, exit
}