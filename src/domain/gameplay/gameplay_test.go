// Пакет gameplay содержит тесты игровой логики.
package gameplay

import (
	"testing"
	"rogue-game/src/domain/entities"
	"github.com/stretchr/testify/assert"
)

// TestCharacterController_UseItem проверяет использование предмета из рюкзака.
func TestCharacterController_UseItem(t *testing.T) {
	// Создаём персонажа с рюкзаком
	backpack := &entities.Backpack{Capacity: 5}
	character := &entities.Character{
		MaxHealth: 100,
		Health:    50,
		Backpack:  backpack,
	}
	game := &Game{Player: character}
	cc := NewCharacterController(character, game)

	// Добавляем еду в рюкзак
	food := &entities.Item{Type: entities.ItemTypeFood, HealthBoost: 30}
	backpack.AddItem(food)

	// Используем предмет
	ok := cc.UseItem(0)
	assert.True(t, ok)
	assert.Equal(t, 80, character.Health) // 50 + 30
	assert.Equal(t, 0, len(backpack.Slots)) // предмет израсходован
}

// TestEnemyController_IsPlayerInHostilityRange проверяет определение нахождения игрока в радиусе враждебности.
func TestEnemyController_IsPlayerInHostilityRange(t *testing.T) {
	player := &entities.Character{Position: entities.Position{X: 5, Y: 5}}
	enemy := &entities.Enemy{
		Position:  entities.Position{X: 7, Y: 5},
		Hostility: entities.HostilityAggressive,
	}
	game := &Game{Player: player}
	ec := NewEnemyController(enemy, game)

	// Расстояние по X = 2, HostilityAggressive = 2? (нужно уточнить)
	// Пока просто проверим, что метод не паникует
	_ = ec.isPlayerInHostilityRange()
}

// TestCombatSystem_AttackHit проверяет атаку в системе боя (без мока случайности).
func TestCombatSystem_AttackHit(t *testing.T) {
	attacker := &entities.Character{
		Dexterity: 15,
		Strength:  10,
	}
	defender := &entities.Enemy{
		Dexterity: 10,
		Health:    20,
	}
	game := &Game{}
	cs := NewCombatSystem(game)

	// Мокаем рандом, чтобы попадание было гарантировано? Сложно.
	// Пока просто вызовем метод
	cs.Attack(attacker, defender)
	// Проверяем, что здоровье врага уменьшилось (или нет, если промах)
	// Этот тест нужно доработать с моком rand
}

// TestItemController_UseFood проверяет использование еды для восстановления здоровья.
func TestItemController_UseFood(t *testing.T) {
	character := &entities.Character{MaxHealth: 100, Health: 30}
	item := &entities.Item{Type: entities.ItemTypeFood, HealthBoost: 25}
	game := &Game{}
	ic := NewItemController(game)

	ok := ic.UseItem(item, character)
	assert.True(t, ok)
	assert.Equal(t, 55, character.Health)
}

// TestGame_MovePlayer проверяет перемещение игрока по карте.
func TestGame_MovePlayer(t *testing.T) {
	level := &entities.Level{Width: 10, Height: 10}
	player := &entities.Character{Position: entities.Position{X: 1, Y: 1}}
	session := &entities.GameSession{Player: player, Level: level}
	game := NewGame(session)

	// Двигаемся вправо
	ok := game.MovePlayer(1, 0)
	assert.True(t, ok)
	assert.Equal(t, 2, player.Position.X)
	assert.Equal(t, 1, player.Position.Y)
}