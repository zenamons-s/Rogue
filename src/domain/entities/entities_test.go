package entities

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCharacter_Heal(t *testing.T) {
	chr := &Character{
		MaxHealth: 100,
		Health:    50,
	}
	chr.Heal(30)
	assert.Equal(t, 80, chr.Health)
	chr.Heal(100)
	assert.Equal(t, 100, chr.Health) // не превышает максимум
}

func TestCharacter_TakeDamage(t *testing.T) {
	chr := &Character{
		MaxHealth: 100,
		Health:    80,
	}
	chr.TakeDamage(30)
	assert.Equal(t, 50, chr.Health)
	chr.TakeDamage(100)
	assert.Equal(t, 0, chr.Health) // здоровье не может быть отрицательным
}

func TestBackpack_AddItem(t *testing.T) {
	bp := &Backpack{
		Capacity: 5,
		Slots:    make([]*Item, 0),
	}
	item := &Item{Type: ItemTypeWeapon}
	ok := bp.AddItem(item)
	assert.True(t, ok)
	assert.Equal(t, 1, len(bp.Slots))
	// Добавляем ещё 5 предметов (всего 6) при capacity=5
	for i := 0; i < 5; i++ {
		bp.AddItem(&Item{Type: ItemTypeFood})
	}
	assert.Equal(t, 5, len(bp.Slots)) // не больше capacity
}

func TestBackpack_RemoveItem(t *testing.T) {
	bp := &Backpack{
		Capacity: 5,
		Slots:    []*Item{{Type: ItemTypeWeapon}, {Type: ItemTypeFood}},
	}
	removed := bp.RemoveItem(0)
	assert.NotNil(t, removed)
	assert.Equal(t, ItemTypeWeapon, removed.Type)
	assert.Equal(t, 1, len(bp.Slots))
	// Удаление несуществующего индекса
	removed = bp.RemoveItem(10)
	assert.Nil(t, removed)
}

func TestEnemy_IsAlive(t *testing.T) {
	enemy := &Enemy{Health: 10}
	assert.True(t, enemy.IsAlive())
	enemy.Health = 0
	assert.False(t, enemy.IsAlive())
	enemy.Health = -5
	assert.False(t, enemy.IsAlive())
}

func TestItem_IsTreasure(t *testing.T) {
	item := &Item{Type: ItemTypeTreasure}
	assert.True(t, item.IsTreasure())
	item.Type = ItemTypeWeapon
	assert.False(t, item.IsTreasure())
}

func TestTile_IsWalkable(t *testing.T) {
	tile := Tile{Type: TileFloor}
	assert.True(t, tile.IsWalkable())
	tile.Type = TileWall
	assert.False(t, tile.IsWalkable())
	tile.Type = TileDoor
	assert.True(t, tile.IsWalkable())
	tile.Type = TileCorridor
	assert.True(t, tile.IsWalkable())
}
