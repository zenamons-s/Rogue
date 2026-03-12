package gameplay

import (
	"rogue-game/src/domain/entities"
)

// CharacterController управляет действиями персонажа.
type CharacterController struct {
	Character *entities.Character
	Game      *Game
}

// NewCharacterController создаёт контроллер для персонажа.
func NewCharacterController(ch *entities.Character, g *Game) *CharacterController {
	return &CharacterController{
		Character: ch,
		Game:      g,
	}
}

// UseItem использует предмет из рюкзака по индексу.
func (cc *CharacterController) UseItem(index int) bool {
	if index < 0 || index >= len(cc.Character.Backpack.Slots) {
		return false
	}
	item := cc.Character.Backpack.Slots[index]
	switch item.Type {
	case entities.ItemTypeFood:
		cc.Character.Heal(item.HealthBoost)
		cc.Game.Stats.UsedFood++
		cc.removeItemFromBackpack(index)
		return true
	case entities.ItemTypePotion:
		cc.applyPotion(item)
		cc.Game.Stats.UsedPotions++
		cc.removeItemFromBackpack(index)
		return true
	case entities.ItemTypeScroll:
		cc.applyScroll(item)
		cc.Game.Stats.UsedScrolls++
		cc.removeItemFromBackpack(index)
		return true
	case entities.ItemTypeWeapon:
		cc.equipWeapon(item)
		return true
	default:
		// Сокровища нельзя использовать
		return false
	}
}

// applyPotion применяет эффект зелья (временное усиление).
func (cc *CharacterController) applyPotion(item *entities.Item) {
	// TODO: реализовать временные модификаторы
	// Пока просто увеличиваем характеристики
	if item.StrengthBoost > 0 {
		cc.Character.Strength += item.StrengthBoost
	}
	if item.DexterityBoost > 0 {
		cc.Character.Dexterity += item.DexterityBoost
	}
	if item.MaxHealthBoost > 0 {
		cc.Character.MaxHealth += item.MaxHealthBoost
		cc.Character.Health += item.MaxHealthBoost
	}
}

// applyScroll применяет эффект свитка (постоянное усиление).
func (cc *CharacterController) applyScroll(item *entities.Item) {
	if item.StrengthBoost > 0 {
		cc.Character.Strength += item.StrengthBoost
	}
	if item.DexterityBoost > 0 {
		cc.Character.Dexterity += item.DexterityBoost
	}
	if item.MaxHealthBoost > 0 {
		cc.Character.MaxHealth += item.MaxHealthBoost
		cc.Character.Health += item.MaxHealthBoost
	}
}

// equipWeapon экипирует оружие.
func (cc *CharacterController) equipWeapon(item *entities.Item) {
	// Если уже есть оружие, кладём его на пол
	if cc.Character.CurrentWeapon != nil {
		cc.dropWeapon()
	}
	cc.Character.CurrentWeapon = item
	// Удаляем из рюкзака
	cc.removeItemFromBackpackByItem(item)
}

// dropWeapon бросает текущее оружие на соседнюю клетку.
func (cc *CharacterController) dropWeapon() {
	if cc.Character.CurrentWeapon == nil {
		return
	}
	// TODO: определить свободную соседнюю клетку и поместить туда оружие
	// Пока просто удаляем из инвентаря (для упрощения)
	cc.Character.CurrentWeapon = nil
}

// removeItemFromBackpack удаляет предмет из рюкзака по индексу.
func (cc *CharacterController) removeItemFromBackpack(index int) {
	cc.Character.Backpack.RemoveItem(index)
}

// removeItemFromBackpackByItem удаляет конкретный предмет из рюкзака.
func (cc *CharacterController) removeItemFromBackpackByItem(item *entities.Item) {
	for i, it := range cc.Character.Backpack.Slots {
		if it == item {
			cc.Character.Backpack.Slots = append(cc.Character.Backpack.Slots[:i], cc.Character.Backpack.Slots[i+1:]...)
			break
		}
	}
}

// PickUpItem подбирает предмет с земли и добавляет в рюкзак.
func (cc *CharacterController) PickUpItem(item *entities.Item) bool {
	if cc.Character.Backpack.AddItem(item) {
		// Удаляем предмет с карты (TODO)
		// cc.Game.removeItemFromMap(item)
		return true
	}
	return false
}

// Attack наносит удар врагу.
func (cc *CharacterController) Attack(enemy *entities.Enemy) {
	// TODO: реализовать логику боя (проверка попадания, расчёт урона)
	damage := cc.calculateDamage()
	enemy.Health -= damage
	if enemy.Health < 0 {
		enemy.Health = 0
	}
	// Если враг убит, добавляем сокровища
	if !enemy.IsAlive() {
		treasure := cc.calculateTreasure(enemy)
		cc.Character.Backpack.AddTreasure(treasure)
	}
}

// calculateDamage вычисляет урон, наносимый персонажем.
func (cc *CharacterController) calculateDamage() int {
	baseDamage := cc.Character.Strength
	if cc.Character.CurrentWeapon != nil {
		baseDamage += cc.Character.CurrentWeapon.StrengthBoost
	}
	return baseDamage
}

// calculateTreasure вычисляет количество сокровищ, выпадающих с врага.
func (cc *CharacterController) calculateTreasure(enemy *entities.Enemy) int {
	// Формула: зависит от характеристик врага
	return enemy.Health + enemy.Strength + enemy.Dexterity + int(enemy.Hostility)*10
}
