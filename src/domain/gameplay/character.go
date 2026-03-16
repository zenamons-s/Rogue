// Пакет gameplay содержит контроллеры и логику взаимодействия персонажа с игровым миром.
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
	const potionDuration = 20
	if item.StrengthBoost > 0 {
		cc.Character.Strength += item.StrengthBoost
		cc.Game.PotionEffects = append(cc.Game.PotionEffects, TimedEffect{Stat: "str", Amount: item.StrengthBoost, TurnsLeft: potionDuration})
	}
	if item.DexterityBoost > 0 {
		cc.Character.Dexterity += item.DexterityBoost
		cc.Game.PotionEffects = append(cc.Game.PotionEffects, TimedEffect{Stat: "dex", Amount: item.DexterityBoost, TurnsLeft: potionDuration})
	}
	if item.MaxHealthBoost > 0 {
		cc.Character.MaxHealth += item.MaxHealthBoost
		cc.Character.Health += item.MaxHealthBoost
		cc.Game.PotionEffects = append(cc.Game.PotionEffects, TimedEffect{Stat: "maxhp", Amount: item.MaxHealthBoost, TurnsLeft: potionDuration})
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

// UnequipWeapon убирает оружие из рук обратно в рюкзак.
func (cc *CharacterController) UnequipWeapon() bool {
	if cc.Character.CurrentWeapon == nil {
		return true
	}
	if !cc.Character.Backpack.AddItem(cc.Character.CurrentWeapon) {
		return false
	}
	cc.Character.CurrentWeapon = nil
	return true
}

// dropWeapon бросает текущее оружие на соседнюю клетку.
func (cc *CharacterController) dropWeapon() {
	if cc.Character.CurrentWeapon == nil {
		return
	}
	neighbors := [][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}}
	for _, n := range neighbors {
		x := cc.Character.Position.X + n[0]
		y := cc.Character.Position.Y + n[1]
		if !cc.Game.isTileWalkable(x, y) || cc.Game.enemyAt(x, y) != nil || cc.Game.groundItemAt(x, y) != nil {
			continue
		}
		cc.Game.GroundItems = append(cc.Game.GroundItems, &GroundItem{Item: cc.Character.CurrentWeapon, Position: entities.Position{X: x, Y: y}})
		break
	}
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
		return true
	}
	return false
}

// Attack наносит удар врагу.
func (cc *CharacterController) Attack(enemy *entities.Enemy) {
	cs := NewCombatSystem(cc.Game)
	cs.Attack(cc.Character, enemy)
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
