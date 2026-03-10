package generation

import (
	"testing"
	"rogue-game/src/domain/entities"
	"github.com/stretchr/testify/assert"
)

func TestLevelGenerator_Generate(t *testing.T) {
	lg := NewLevelGenerator(100, 100, 42)
	level := lg.Generate(1)

	assert.Equal(t, 1, level.ID)
	assert.Equal(t, 100, level.Width)
	assert.Equal(t, 100, level.Height)
	assert.Len(t, level.Rooms, 9)
	assert.True(t, len(level.Corridors) >= 8) // MST для 9 комнат даёт 8 коридоров

	// Проверяем, что есть ровно одна стартовая и одна конечная комната, и они разные
	var startRoom, exitRoom *entities.Room
	for _, room := range level.Rooms {
		if room.IsStart {
			startRoom = room
		}
		if room.IsExit {
			exitRoom = room
		}
	}
	assert.NotNil(t, startRoom)
	assert.NotNil(t, exitRoom)
	assert.NotEqual(t, startRoom.ID, exitRoom.ID)
}

func TestLevelGenerator_GenerateRooms(t *testing.T) {
	lg := NewLevelGenerator(90, 90, 123)
	rooms := lg.generateRooms()

	assert.Len(t, rooms, 9)
	for i, room := range rooms {
		// Комната должна находиться в своей секции
		row := i / 3
		col := i % 3
		sectionWidth := 90 / 3
		sectionHeight := 90 / 3
		assert.True(t, room.X >= col*sectionWidth && room.X < (col+1)*sectionWidth)
		assert.True(t, room.Y >= row*sectionHeight && room.Y < (row+1)*sectionHeight)
		assert.True(t, room.Width >= 3)
		assert.True(t, room.Height >= 3)
		// Комната не должна выходить за границы секции
		assert.True(t, room.X+room.Width <= (col+1)*sectionWidth)
		assert.True(t, room.Y+room.Height <= (row+1)*sectionHeight)
	}
}

func TestLevelGenerator_GenerateCorridors(t *testing.T) {
	lg := NewLevelGenerator(100, 100, 456)
	rooms := lg.generateRooms()
	corridors := lg.generateCorridors(rooms)

	// Должно быть ровно 8 коридоров (9 комнат - 1)
	assert.Len(t, corridors, 8)

	// Проверяем, что каждый коридор соединяет две разные комнаты
	for _, corridor := range corridors {
		assert.NotEqual(t, corridor.FromRoom, corridor.ToRoom)
		assert.True(t, corridor.FromRoom >= 0 && corridor.FromRoom < 9)
		assert.True(t, corridor.ToRoom >= 0 && corridor.ToRoom < 9)
		assert.NotEmpty(t, corridor.Tiles)
	}
}

func TestLevelGenerator_IsGraphConnected(t *testing.T) {
	lg := NewLevelGenerator(100, 100, 789)
	rooms := lg.generateRooms()
	corridors := lg.generateCorridors(rooms)

	connected := lg.isGraphConnected(rooms, corridors)
	assert.True(t, connected)

	// Создаём пустой список коридоров — граф должен быть несвязным
	emptyCorridors := []*entities.Corridor{}
	connected = lg.isGraphConnected(rooms, emptyCorridors)
	assert.False(t, connected)
}