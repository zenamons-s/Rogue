// Пакет presentation содержит консольный интерфейс игры (raw‑mode и line‑mode).
package presentation

import (
	"fmt"
	"os"
	"strings"
	"syscall"
	"unsafe"

	"rogue-game/src/domain/entities"
	"rogue-game/src/presentation/i18n"
)

func clearScreen() {
	fmt.Print("\033[H\033[2J")
}

type winsize struct {
	Row    uint16
	Col    uint16
	Xpixel uint16
	Ypixel uint16
}

func terminalWidth() int {
	ws := &winsize{}
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, os.Stdout.Fd(), uintptr(syscall.TIOCGWINSZ), uintptr(unsafe.Pointer(ws)))
	if errno != 0 || ws.Col == 0 {
		return 100
	}
	return int(ws.Col)
}

func centerLine(line string, termWidth int) string {
	if line == "" || termWidth <= len(line) {
		return line
	}
	padding := (termWidth - len(line)) / 2
	return strings.Repeat(" ", padding) + line
}

func printCentered(lines []string, termWidth int) {
	for _, line := range lines {
		fmt.Println(centerLine(strings.TrimRight(line, "\n"), termWidth))
	}
}

// render отрисовывает игровое поле, HUD и подсказки.
func (a *ConsoleApp) render() {
	clearScreen()
	termWidth := terminalWidth()
	printCentered([]string{"=== ИГРА ===", ""}, termWidth)
	printCentered(a.renderMapBlock(), termWidth)
	fmt.Println()
	printCentered(a.renderHUDBlock(), termWidth)
}

func (a *ConsoleApp) renderMapBlock() []string {
	lines := make([]string, 0, a.Game.CurrentLevel.Height)
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
		lines = append(lines, row.String())
	}
	return lines
}

func (a *ConsoleApp) renderHUDBlock() []string {
	lines := []string{"=== СТАТУС ==="}
	lines = append(lines, fmt.Sprintf(i18n.HUDLineFormat,
		a.Game.Player.Health,
		a.Game.Player.MaxHealth,
		a.Game.Player.Strength,
		a.Game.Player.Dexterity,
		a.Game.Session.CurrentFloor,
		a.Game.Session.Score,
		a.Game.Player.Backpack.TotalTreasure(),
		a.Game.Turn,
	))
	lines = append(lines, fmt.Sprintf(i18n.InventoryItems, len(a.Game.Player.Backpack.Slots)))
	if a.Game.Player.CurrentWeapon != nil {
		lines = append(lines, fmt.Sprintf(i18n.WeaponEquipped, formatItemRu(a.Game.Player.CurrentWeapon)))
	} else {
		lines = append(lines, i18n.WeaponEquippedNone)
	}
	lines = append(lines, i18n.HUDHintLine1, i18n.HUDHintLine2)
	return lines
}

// renderCurrentStats выводит на экран статистику текущей игровой сессии.
func (a *ConsoleApp) renderCurrentStats() {
	clearScreen()
	termWidth := terminalWidth()
	result := "поражение"
	if a.Game.Stats.Won {
		result = "победа"
	}
	lines := []string{
		"=== СТАТИСТИКА ===",
		"",
		fmt.Sprintf("Сокровища: %d", a.Game.Player.Backpack.TotalTreasure()),
		fmt.Sprintf("Достигнутый уровень: %d", a.Game.Stats.ReachedLevel),
		fmt.Sprintf("Побеждённые враги: %d", a.Game.Stats.DefeatedEnemies),
		fmt.Sprintf("Использовано еды/эликсиров/свитков: %d/%d/%d", a.Game.Stats.UsedFood, a.Game.Stats.UsedPotions, a.Game.Stats.UsedScrolls),
		fmt.Sprintf("Ударов нанесено/урона получено: %d/%d", a.Game.Stats.HitsDealt, a.Game.Stats.HitsTaken),
		fmt.Sprintf("Клеток пройдено: %d", a.Game.Stats.TilesWalked),
		fmt.Sprintf("Итог попытки: %s", result),
		"",
		i18n.PressAnyKey,
	}
	printCentered(lines, termWidth)
	_, _ = a.readKey()
}

// renderLeaderboard выводит таблицу лидеров (топ‑10 попыток).
func (a *ConsoleApp) renderLeaderboard() {
	clearScreen()
	termWidth := terminalWidth()
	lines := []string{"=== ТАБЛИЦА ЛУЧШИХ ===", ""}
	rows, err := a.Storage.Leaderboard(10)
	if err != nil {
		printCentered([]string{i18n.MsgReadStatsFail + ": " + err.Error()}, termWidth)
		return
	}
	lines = append(lines, " #  Сокр  Ур  Враги  Еда  Элк  Свт  Уд  Проп  Ходы  Итог")
	for i, r := range rows {
		result := "поражение"
		if r.Won {
			result = "победа"
		}
		lines = append(lines, fmt.Sprintf("%2d) %4d %3d %6d %4d %4d %4d %3d %5d %5d  %s",
			i+1, r.Treasures, r.ReachedLevel, r.DefeatedEnemies, r.UsedFood, r.UsedPotions, r.UsedScrolls, r.HitsDealt, r.HitsTaken, r.TilesWalked, result))
	}
	if len(rows) == 0 {
		lines = append(lines, i18n.LeaderboardEmpty)
	}
	lines = append(lines, "", i18n.PressAnyKey)
	printCentered(lines, termWidth)
	_, _ = a.readKey()
}

// renderHelp отображает экран с подсказками по управлению (raw‑mode).
func (a *ConsoleApp) renderHelp() {
	clearScreen()
	termWidth := terminalWidth()
	lines := append([]string{"=== СПРАВКА ===", ""}, i18n.HelpLines...)
	printCentered(lines, termWidth)
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
	termWidth := terminalWidth()
	lines := append([]string{"=== СПРАВКА ===", ""}, i18n.HelpLines...)
	printCentered(lines, termWidth)
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
