package gameplay

import (
	"math/rand"
	"rogue-game/src/domain/entities"
)

// ItemController управляет предметами на карте и их взаимодействием.
type ItemController struct {
	Game *Game
}

// NewItemController создаёт контроллер предметов.
func NewItemController(g *Game) *ItemController {
	return &ItemController{
		Game: g,
	}
}

// UseItem использует предмет из рюкзака персонажа.
func (ic *ItemController) UseItem(item *entities.Item, character *entities.Character) bool {
	switch item.Type {
	case entities.ItemTypeFood:
		return ic.useFood(item, character)
	case entities.ItemTypePotion:
		return ic.usePotion(item, character)
	case entities.ItemTypeScroll:
		return ic.useScroll(item, character)
	case entities.ItemTypeWeapon:
		return ic.useWeapon(item, character)
	default:
		return false
	}
}

// useFood восстанавливает здоровье персонажа.
func (ic *ItemController) useFood(item *entities.Item, character *entities.Character) bool {
	character.Heal(item.HealthBoost)
	return true
}

// usePotion временно увеличивает характеристику персонажа.
func (ic *ItemController) usePotion(item *entities.Item, character *entities.Character) bool {
	if item.StrengthBoost > 0 {
		character.Strength += item.StrengthBoost
		// TODO: добавить временный эффект с последующим снятием
	}
	if item.DexterityBoost > 0 {
		character.Dexterity += item.DexterityBoost
	}
	if item.MaxHealthBoost > 0 {
		character.MaxHealth += item.MaxHealthBoost
		character.Health += item.MaxHealthBoost
	}
	return true
}

// useScroll постоянно увеличивает характеристику персонажа.
func (ic *ItemController) useScroll(item *entities.Item, character *entities.Character) bool {
	if item.StrengthBoost > 0 {
		character.Strength += item.StrengthBoost
	}
	if item.DexterityBoost > 0 {
		character.Dexterity += item.DexterityBoost
	}
	if item.MaxHealthBoost > 0 {
		character.MaxHealth += item.MaxHealthBoost
		character.Health += item.MaxHealthBoost
	}
	return true
}

// useWeapon экипирует оружие.
func (ic *ItemController) useWeapon(item *entities.Item, character *entities.Character) bool {
	// Если уже есть оружие, кладём его на пол
	if character.CurrentWeapon != nil {
		ic.dropWeapon(character.CurrentWeapon, character.Position)
	}
	character.CurrentWeapon = item
	return true
}

// dropWeapon бросает оружие на соседнюю клетку.
func (ic *ItemController) dropWeapon(weapon *entities.Item, pos entities.Position) {
	// Ищем свободную соседнюю клетку
	neighbors := [][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}}
	for _, n := range neighbors {
		x, y := pos.X+n[0], pos.Y+n[1]
		if ic.isTileFree(x, y) {
			// Помещаем оружие на карту
			// weaponCopy := *weapon
			// TODO: добавить предмет в список предметов уровня
			return
		}
	}
	// Если нет свободной клетки, предмет уничтожается
}

// isTileFree проверяет, свободна ли клетка для размещения предмета.
func (ic *ItemController) isTileFree(x, y int) bool {
	// Проверка на стены и другие препятствия
	// TODO: использовать карту уровня
	// Проверка на наличие других предметов
	for _ = range ic.Game.Items {
		// TODO: предметы должны иметь позицию
	}
	// Проверка на наличие врагов
	for _, enemy := range ic.Game.Enemies {
		if enemy.Position.X == x && enemy.Position.Y == y {
			return false
		}
	}
	// Проверка на игрока
	if ic.Game.Player.Position.X == x && ic.Game.Player.Position.Y == y {
		return false
	}
	return true
}

// PickUpItem подбирает предмет с земли.
func (ic *ItemController) PickUpItem(item *entities.Item, character *entities.Character) bool {
	if character.Backpack.AddItem(item) {
		ic.removeItemFromMap(item)
		return true
	}
	return false
}

// removeItemFromMap удаляет предмет с карты.
func (ic *ItemController) removeItemFromMap(item *entities.Item) {
	// TODO: реализовать удаление из списка предметов уровня
}

// GenerateLoot генерирует сокровища после победы над врагом.
func (ic *ItemController) GenerateLoot(enemy *entities.Enemy) *entities.Item {
	base := enemy.MaxHealth/2 + enemy.Strength + enemy.Dexterity + int(enemy.Hostility)*10
	if enemy.Health > 0 {
		base += enemy.Health / 2
	}
	variance := max(1, base/4)
	minValue := max(1, base-variance)
	maxValue := base + variance
	value := minValue
	if maxValue > minValue {
		value += rand.Intn(maxValue - minValue + 1)
	}
	return &entities.Item{
		Type:  entities.ItemTypeTreasure,
		Value: value,
	}
}
