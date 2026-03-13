package gameplay

import (
	"math/rand"
	"rogue-game/src/domain/entities"
)

// CombatSystem управляет боем между персонажем и врагом.
type CombatSystem struct {
	Game *Game
}

// NewCombatSystem создаёт систему боя.
func NewCombatSystem(g *Game) *CombatSystem {
	return &CombatSystem{
		Game: g,
	}
}

// Attack выполняет атаку атакующего по защищающемуся.
func (cs *CombatSystem) Attack(attacker, defender interface{}) bool {
	var hitChance float64
	var damage int

	switch a := attacker.(type) {
	case *entities.Character:
		switch d := defender.(type) {
		case *entities.Enemy:
			if d.Type == entities.EnemyVampire {
				// Первый удар по вампиру всегда промах
				if cs.isFirstAttackOnVampire(d) {
					return false
				}
			}
			hitChance = cs.calculateHitChance(a.Dexterity, d.Dexterity)
			if !cs.checkHit(hitChance) {
				return false // промах
			}
			damage = cs.calculateDamage(a.Strength, a.CurrentWeapon)
			d.Health -= damage
			if d.Health < 0 {
				d.Health = 0
			}
			// Если враг убит, генерируем лут
			if !d.IsAlive() {
				cs.generateLoot(d)
			}
			return true
		}
	case *entities.Enemy:
		switch d := defender.(type) {
		case *entities.Character:
			hitChance = cs.calculateHitChance(a.Dexterity, d.Dexterity)
			if !cs.checkHit(hitChance) {
				return false
			}
			damage = cs.calculateEnemyDamage(a)
			d.TakeDamage(damage)
			// Дополнительные эффекты
			cs.applySpecialEffects(a, d)
			return true
		}
	}
	return false
}

// calculateHitChance вычисляет вероятность попадания на основе ловкости.
func (cs *CombatSystem) calculateHitChance(attackerDex, defenderDex int) float64 {
	baseChance := 0.8
	dexDiff := float64(attackerDex-defenderDex) * 0.01
	chance := baseChance + dexDiff
	if chance < 0.1 {
		chance = 0.1
	}
	if chance > 0.95 {
		chance = 0.95
	}
	return chance
}

// checkHit определяет, произошло ли попадание, на основе вероятности.
func (cs *CombatSystem) checkHit(hitChance float64) bool {
	return rand.Float64() < hitChance
}

// calculateDamage вычисляет урон, наносимый персонажем.
func (cs *CombatSystem) calculateDamage(strength int, weapon *entities.Item) int {
	damage := strength
	if weapon != nil {
		damage += weapon.StrengthBoost
	}
	// Добавляем небольшой случайный разброс
	variation := rand.Intn(5) - 2 // от -2 до +2
	damage += variation
	if damage < 1 {
		damage = 1
	}
	return damage
}

// calculateEnemyDamage вычисляет урон, наносимый врагом.
func (cs *CombatSystem) calculateEnemyDamage(enemy *entities.Enemy) int {
	damage := enemy.Strength
	switch enemy.Type {
	case entities.EnemyOgre:
		damage *= 2
	}
	variation := rand.Intn(5) - 2
	damage += variation
	if damage < 1 {
		damage = 1
	}
	return damage
}

// applySpecialEffects применяет специальные эффекты атаки врага.
func (cs *CombatSystem) applySpecialEffects(enemy *entities.Enemy, character *entities.Character) {
	switch enemy.Type {
	case entities.EnemyVampire:
		// Снижение максимального здоровья
		character.MaxHealth -= 5
		if character.MaxHealth < 1 {
			character.MaxHealth = 1
		}
		if character.Health > character.MaxHealth {
			character.Health = character.MaxHealth
		}
	case entities.EnemySnakeMage:
		// Вероятность усыпления
		if rand.Float64() < 0.3 {
			cs.Game.PlayerSleepTurns = 1
		}
	}
}

// isFirstAttackOnVampire проверяет, является ли это первой атакой на данного вампира.
func (cs *CombatSystem) isFirstAttackOnVampire(vampire *entities.Enemy) bool {
	idx := cs.Game.enemyIndex(vampire)
	if idx < 0 {
		return false
	}
	if !cs.Game.VampireFirstMiss[idx] {
		cs.Game.VampireFirstMiss[idx] = true
		return true
	}
	return false
}

// generateLoot создаёт сокровища после убийства врага.
func (cs *CombatSystem) generateLoot(enemy *entities.Enemy) {
	itemController := NewItemController(cs.Game)
	loot := itemController.GenerateLoot(enemy)
	if loot == nil {
		return
	}
	cs.Game.GroundItems = append(cs.Game.GroundItems, &GroundItem{Item: loot, Position: enemy.Position})
}
