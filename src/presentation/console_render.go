// Пакет presentation содержит консольный интерфейс игры (raw‑mode и line‑mode).
package presentation

import (
	"fmt"
	"strings"

	"rogue-game/src/domain/entities"
	"rogue-game/src/presentation/i18n"
)

func clearScreen() {
	fmt.Print("\033[H\033[2J")
}

// render отрисовывает игровое поле, HUD и подсказки.
func (a *ConsoleApp) render() {
	clearScreen()
	for y := 0; y < a.Game.CurrentLevel.Height; y++ {
		row := strings.Builder{}
		for x := 0; x < a.Game.CurrentLevel.Width; x++ {
			if !a.Game.Explored[y][x] {
				row.WriteRune(' ')
				continue
			}
			if !a.Game.Visible[y][x] {
				if a.Game.TileMap[y][x] == entities.TileWall {
					row.WriteRune('#')
				} else {
					row.WriteRune(' ')
				}
				continue
			}
			ch := tileRune(a.Game.TileMap[y][x])
			if a.Game.ExitPos.X == x && a.Game.ExitPos.Y == y {
				ch = '>'
			}
			for _, gi := range a.Game.GroundItems {
				if !gi.Collected && gi.Position.X == x && gi.Position.Y == y {
					ch = itemRune(gi.Item)
				}
			}
			for _, e := range a.Game.Enemies {
				if e.IsAlive() && e.Position.X == x && e.Position.Y == y {
					ch = enemyRune(e)
				}
			}
			if a.Game.Player.Position.X == x && a.Game.Player.Position.Y == y {
				ch = '@'
			}
			row.WriteRune(ch)
		}
		fmt.Println(row.String())
	}
	fmt.Printf(i18n.HUDLineFormat,
		a.Game.Player.Health,
		a.Game.Player.MaxHealth,
		a.Game.Player.Strength,
		a.Game.Player.Dexterity,
		a.Game.Session.CurrentFloor,
		a.Game.Session.Score,
		a.Game.Player.Backpack.TotalTreasure(),
		a.Game.Turn,
	)
	fmt.Printf(i18n.InventoryItems, len(a.Game.Player.Backpack.Slots))
	if a.Game.Player.CurrentWeapon != nil {
		fmt.Printf(i18n.WeaponEquipped, formatItemRu(a.Game.Player.CurrentWeapon))
	} else {
		fmt.Println(i18n.WeaponEquippedNone)
	}
	fmt.Print(i18n.HUDHintLine1)
	fmt.Print(i18n.HUDHintLine2)
}

// renderCurrentStats выводит на экран статистику текущей игровой сессии.
func (a *ConsoleApp) renderCurrentStats() {
	result := "поражение"
	if a.Game.Stats.Won {
		result = "победа"
	}
	fmt.Println(i18n.StatsTitle)
	fmt.Printf(i18n.StatsLineFormat,
		a.Game.Player.Backpack.TotalTreasure(),
		a.Game.Stats.ReachedLevel,
		a.Game.Stats.DefeatedEnemies,
		a.Game.Stats.UsedFood,
		a.Game.Stats.UsedPotions,
		a.Game.Stats.UsedScrolls,
		a.Game.Stats.HitsDealt,
		a.Game.Stats.HitsTaken,
		a.Game.Stats.TilesWalked,
		result,
	)
	fmt.Println(i18n.PressAnyKey)
	_, _ = a.readKey()
}

// renderLeaderboard выводит таблицу лидеров (топ‑10 попыток).
func (a *ConsoleApp) renderLeaderboard() {
	clearScreen()
	fmt.Println(i18n.LeaderboardTitle)
	rows, err := a.Storage.Leaderboard(10)
	if err != nil {
		fmt.Println(i18n.MsgReadStatsFail+":", err)
		return
	}
	for i, r := range rows {
		result := "поражение"
		if r.Won {
			result = "победа"
		}
		fmt.Printf(i18n.LeaderboardLine, i+1, r.Treasures, r.ReachedLevel, r.DefeatedEnemies, r.UsedFood, r.UsedPotions, r.UsedScrolls, r.HitsDealt, r.HitsTaken, r.TilesWalked, result)
	}
	if len(rows) == 0 {
		fmt.Println(i18n.LeaderboardEmpty)
	}
	fmt.Println(i18n.PressAnyKey)
	_, _ = a.readKey()
}

// renderHelp отображает экран с подсказками по управлению (raw‑mode).
func (a *ConsoleApp) renderHelp() {
	clearScreen()
	for _, line := range i18n.HelpLines {
		fmt.Println(line)
	}
	for {
		key, err := a.readControlKey()
		if err != nil {
			return
		}
		if key == 'q' || key == '\n' || key == '\r' || key == 0x1b {
			return
		}
	}
}

func (a *ConsoleApp) renderHelpLineMode() {
	clearScreen()
	for _, line := range i18n.HelpLines {
		fmt.Println(line)
	}
	line, _ := a.reader.ReadString('\n')
	line = strings.TrimSpace(strings.ToLower(line))
	if line == "q" || line == "" {
		return
	}
}

// renderMessageScreen выводит сообщение на весь экран и ждёт нажатия любой клавиши.
func (a *ConsoleApp) renderMessageScreen(msg string) {
	clearScreen()
	fmt.Println(msg)
	fmt.Println(i18n.PressAnyKey)
	_, _ = a.readControlKey()
}

// tileRune возвращает символ для отображения типа тайла.
func tileRune(t entities.TileType) rune {
	switch t {
	case entities.TileWall:
		return '#'
	case entities.TileFloor:
		return '.'
	case entities.TileCorridor:
		return '+'
	case entities.TileDoor:
		return 'D'
	default:
		return ' '
	}
}

// enemyRune возвращает символ для отображения типа врага.
func enemyRune(e *entities.Enemy) rune {
	switch e.Type {
	case entities.EnemyZombie:
		return 'z'
	case entities.EnemyVampire:
		return 'v'
	case entities.EnemyGhost:
		return 'g'
	case entities.EnemyOgre:
		return 'O'
	case entities.EnemySnakeMage:
		return 's'
	default:
		return 'e'
	}
}

// itemRune возвращает символ для отображения типа предмета.
func itemRune(i *entities.Item) rune {
	switch i.Type {
	case entities.ItemTypeFood:
		return 'f'
	case entities.ItemTypePotion:
		return 'p'
	case entities.ItemTypeScroll:
		return 'r'
	case entities.ItemTypeWeapon:
		return 'w'
	case entities.ItemTypeTreasure:
		return '$'
	default:
		return '?'
	}
}

// formatItemRu возвращает читаемое описание предмета на русском языке.
func formatItemRu(i *entities.Item) string {
	if i == nil {
		return "нет"
	}
	switch i.Type {
	case entities.ItemTypeWeapon:
		sub := "оружие"
		switch i.Subtype {
		case entities.SubtypeSword:
			sub = "меч"
		case entities.SubtypeBow:
			sub = "лук"
		}
		return fmt.Sprintf("%s (сила +%d)", sub, i.StrengthBoost)
	case entities.ItemTypeFood:
		sub := "еда"
		switch i.Subtype {
		case entities.SubtypeBread:
			sub = "хлеб"
		case entities.SubtypeApple:
			sub = "яблоко"
		}
		return fmt.Sprintf("%s (лечение +%d)", sub, i.HealthBoost)
	case entities.ItemTypePotion:
		sub := "эликсир"
		switch i.Subtype {
		case entities.SubtypeHealthPotion:
			sub = "эликсир здоровья"
		case entities.SubtypeStrengthPotion:
			sub = "эликсир силы"
		}
		return fmt.Sprintf("%s (сила +%d, ловкость +%d, макс.здоровье +%d)", sub, i.StrengthBoost, i.DexterityBoost, i.MaxHealthBoost)
	case entities.ItemTypeScroll:
		sub := "свиток"
		switch i.Subtype {
		case entities.SubtypeScrollOfStrength:
			sub = "свиток силы"
		case entities.SubtypeScrollOfDexterity:
			sub = "свиток ловкости"
		}
		return fmt.Sprintf("%s (сила +%d, ловкость +%d, макс.здоровье +%d)", sub, i.StrengthBoost, i.DexterityBoost, i.MaxHealthBoost)
	case entities.ItemTypeTreasure:
		return fmt.Sprintf("сокровище (ценность %d)", i.Value)
	default:
		return "неизвестный предмет"
	}
}
