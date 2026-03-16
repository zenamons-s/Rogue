// Пакет entities определяет основные сущности игрового мира:
// персонажи, враги, предметы, уровни, клетки и т.д.
package entities

// GameSession представляет игровую сессию.
type GameSession struct {
	ID           string
	Player       *Character
	Level        *Level
	CurrentFloor int
	Score        int
	IsActive     bool
}

// Level представляет игровой уровень (этаж).
type Level struct {
	ID        int
	Rooms     []*Room
	Corridors []*Corridor
	Width     int
	Height    int
}

// Room представляет комнату на уровне.
type Room struct {
	ID      int
	X, Y    int // верхний левый угол
	Width   int
	Height  int
	Enemies []*Enemy
	Items   []*Item
	IsStart bool
	IsExit  bool
}

// Corridor представляет коридор, соединяющий две комнаты.
type Corridor struct {
	ID       int
	FromRoom int
	ToRoom   int
	Tiles    []Tile // последовательность клеток коридора
}

// Tile представляет клетку на карте.
type Tile struct {
	X, Y int
	Type TileType
}

// TileType определяет тип клетки.
type TileType int

const (
	TileWall TileType = iota      // Стена — непроходимое препятствие
	TileFloor                     // Пол — проходимая клетка внутри комнаты
	TileDoor                      // Дверь — проходимая клетка, соединяющая комнаты
	TileCorridor                  // Коридор — проходимая клетка коридора
)

// Character представляет игрового персонажа.
type Character struct {
	MaxHealth     int
	Health        int
	Dexterity     int
	Strength      int
	CurrentWeapon *Item
	Backpack      *Backpack
	Position      Position
}

// Position представляет координаты на карте.
type Position struct {
	X, Y int
}

// Backpack представляет рюкзак персонажа.
type Backpack struct {
	Slots    []*Item
	Capacity int
}

// Enemy представляет противника.
type Enemy struct {
	Type      EnemyType
	MaxHealth int
	Health    int
	Dexterity int
	Strength  int
	Hostility HostilityLevel
	Position  Position
}

// EnemyType определяет тип врага.
type EnemyType int

const (
	EnemyZombie EnemyType = iota   // Зомби — медленный, наносит средний урон
	EnemyVampire                   // Вампир — восстанавливает здоровье при атаке
	EnemyGhost                     // Призрак — может проходить сквозь стены
	EnemyOgre                      // Огр — мощный, но медленный
	EnemySnakeMage                 // Змеиный маг — атакует на расстоянии
)

// HostilityLevel определяет уровень враждебности врага.
type HostilityLevel int

const (
	HostilityPassive HostilityLevel = iota   // Пассивный — не атакует первым
	HostilityNeutral                         // Нейтральный — атакует при приближении
	HostilityAggressive                      // Агрессивный — преследует игрока
)

// Item представляет предмет.
type Item struct {
	Type           ItemType
	Subtype        ItemSubtype
	HealthBoost    int // для еды
	MaxHealthBoost int // для свитков и эликсиров
	DexterityBoost int
	StrengthBoost  int
	Value          int // стоимость для сокровищ
}

// ItemType определяет тип предмета.
type ItemType int

const (
	ItemTypeWeapon ItemType = iota   // Оружие — увеличивает силу атаки
	ItemTypeFood                     // Еда — восстанавливает здоровье
	ItemTypePotion                   // Зелье — даёт временный эффект
	ItemTypeScroll                   // Свиток — постоянное улучшение характеристики
	ItemTypeTreasure                 // Сокровище — увеличивает счёт
)

// ItemSubtype определяет подтип предмета.
type ItemSubtype int

const (
	SubtypeSword ItemSubtype = iota          // Меч — оружие ближнего боя
	SubtypeBow                               // Лук — оружие дальнего боя
	SubtypeBread                             // Хлеб — еда, восстанавливает 5 HP
	SubtypeApple                             // Яблоко — еда, восстанавливает 3 HP
	SubtypeHealthPotion                      // Зелье здоровья — восстанавливает 15 HP
	SubtypeStrengthPotion                    // Зелье силы — временно увеличивает силу
	SubtypeScrollOfStrength                  // Свиток силы — постоянно увеличивает силу
	SubtypeScrollOfDexterity                 // Свиток ловкости — постоянно увеличивает ловкость
	SubtypeGold                              // Золото — сокровище, стоимость 10
	SubtypeGem                               // Драгоценный камень — сокровище, стоимость 50
)

// Heal восстанавливает здоровье персонажа на указанное количество, но не выше MaxHealth.
func (c *Character) Heal(amount int) {
	c.Health += amount
	if c.Health > c.MaxHealth {
		c.Health = c.MaxHealth
	}
}

// TakeDamage наносит урон персонажу.
func (c *Character) TakeDamage(amount int) {
	c.Health -= amount
	if c.Health < 0 {
		c.Health = 0
	}
}

// AddItem добавляет предмет в рюкзак, если есть свободное место.
func (b *Backpack) AddItem(item *Item) bool {
	if len(b.Slots) >= b.Capacity {
		return false
	}
	if item == nil {
		return false
	}
	if item.Type == ItemTypeTreasure {
		for _, it := range b.Slots {
			if it.IsTreasure() {
				it.Value += item.Value
				return true
			}
		}
	}
	if item.Type != ItemTypeTreasure {
		countByType := 0
		for _, it := range b.Slots {
			if it.Type == item.Type {
				countByType++
			}
		}
		if countByType >= 9 {
			return false
		}
	}
	b.Slots = append(b.Slots, item)
	return true
}

// RemoveItem удаляет предмет по индексу и возвращает его.
func (b *Backpack) RemoveItem(index int) *Item {
	if index < 0 || index >= len(b.Slots) {
		return nil
	}
	item := b.Slots[index]
	b.Slots = append(b.Slots[:index], b.Slots[index+1:]...)
	return item
}

// IsAlive возвращает true, если здоровье врага больше нуля.
func (e *Enemy) IsAlive() bool {
	return e.Health > 0
}

// IsTreasure возвращает true, если предмет является сокровищем.
func (i *Item) IsTreasure() bool {
	return i.Type == ItemTypeTreasure
}

// IsWalkable возвращает true, если по клетке можно ходить.
func (t *Tile) IsWalkable() bool {
	return t.Type == TileFloor || t.Type == TileDoor || t.Type == TileCorridor
}

// IsAlive возвращает true, если здоровье персонажа больше нуля.
func (c *Character) IsAlive() bool {
	return c.Health > 0
}

// AddTreasure добавляет сокровище в рюкзак (увеличивает общую сумму).
func (b *Backpack) AddTreasure(value int) {
	// Ищем слот с сокровищами
	for _, item := range b.Slots {
		if item.IsTreasure() {
			item.Value += value
			return
		}
	}
	// Если нет слота с сокровищами, создаём новый
	if len(b.Slots) < b.Capacity {
		b.Slots = append(b.Slots, &Item{
			Type:  ItemTypeTreasure,
			Value: value,
		})
	}
}

// TotalTreasure возвращает общую стоимость сокровищ в рюкзаке.
func (b *Backpack) TotalTreasure() int {
	total := 0
	for _, item := range b.Slots {
		if item.IsTreasure() {
			total += item.Value
		}
	}
	return total
}
