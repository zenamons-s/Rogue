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
	lineWidth := textWidth(line)
	if line == "" || termWidth <= lineWidth {
		return line
	}
	padding := (termWidth - lineWidth) / 2
	return strings.Repeat(" ", padding) + line
}

func printCentered(lines []string, termWidth int) {
	for _, line := range lines {
		fmt.Println(centerLine(strings.TrimRight(line, "\n"), termWidth))
	}
}

func printLines(lines []string) {
	for _, line := range lines {
		fmt.Println(strings.TrimRight(line, "\n"))
	}
}

func textWidth(s string) int {
	return len([]rune(s))
}

func padRight(s string, width int) string {
	if width <= textWidth(s) {
		return s
	}
	return s + strings.Repeat(" ", width-textWidth(s))
}

func maxLineWidth(lines []string) int {
	maxW := 0
	for _, line := range lines {
		if w := textWidth(line); w > maxW {
			maxW = w
		}
	}
	return maxW
}

func centeredBlock(lines []string, termWidth int) []string {
	maxW := maxLineWidth(lines)
	if maxW == 0 || termWidth <= maxW {
		return lines
	}
	leftPad := strings.Repeat(" ", (termWidth-maxW)/2)
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		out = append(out, leftPad+line)
	}
	return out
}

func makeRule(title string, width int) string {
	if width < 20 {
		width = 20
	}
	rule := strings.Repeat("=", width)
	if title == "" {
		return rule
	}
	label := " " + strings.ToUpper(title) + " "
	if textWidth(label) >= width {
		return label
	}
	start := (width - textWidth(label)) / 2
	runes := []rune(rule)
	for i, ch := range []rune(label) {
		runes[start+i] = ch
	}
	return string(runes)
}

func boxedSection(title string, body []string, width int) []string {
	if width < 24 {
		width = 24
	}
	insideWidth := width - 4
	if insideWidth < 1 {
		insideWidth = 1
	}
	lines := []string{fmt.Sprintf("+%s+", strings.Repeat("-", width-2))}
	titleLine := fmt.Sprintf("| %s |", padRight(strings.ToUpper(title), insideWidth))
	sepLine := fmt.Sprintf("| %s |", strings.Repeat("-", insideWidth))
	lines = append(lines, titleLine, sepLine)
	for _, line := range body {
		lines = append(lines, fmt.Sprintf("| %s |", padRight(line, insideWidth)))
	}
	lines = append(lines, fmt.Sprintf("+%s+", strings.Repeat("-", width-2)))
	return lines
}

// render отрисовывает игровое поле, HUD и подсказки.
func (a *ConsoleApp) render() {
	clearScreen()
	termWidth := terminalWidth()
	mapLines := a.renderMapBlock()
	statusLines := a.renderHUDBlock()
	controlsLines := []string{
		"WASD — ходьба/атака",
		"b — рюкзак",
		"h/j/k/e — быстрый выбор предметов",
		"t — статистика текущей попытки",
		"l — таблица лучших попыток",
		"? или i — помощь",
		"q — выход с сохранением",
	}

	contentWidth := maxLineWidth(mapLines)
	for _, w := range []int{58, maxLineWidth(statusLines) + 4, maxLineWidth(controlsLines) + 4} {
		if w > contentWidth {
			contentWidth = w
		}
	}

	var lines []string
	lines = append(lines, makeRule("ROGUE", contentWidth))
	lines = append(lines, "")
	lines = append(lines, boxedSection("Карта", mapLines, contentWidth)...)
	lines = append(lines, "")
	lines = append(lines, boxedSection("Статус игрока", statusLines, contentWidth)...)
	lines = append(lines, "")
	lines = append(lines, boxedSection("Управление", controlsLines, contentWidth)...)
	lines = append(lines, "")
	lines = append(lines, "Команда: ")

	printLines(centeredBlock(lines, termWidth))
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
	lines := []string{
		fmt.Sprintf("Здоровье   : %d/%d", a.Game.Player.Health, a.Game.Player.MaxHealth),
		fmt.Sprintf("Сила       : %d", a.Game.Player.Strength),
		fmt.Sprintf("Ловкость   : %d", a.Game.Player.Dexterity),
		fmt.Sprintf("Уровень    : %d", a.Game.Session.CurrentFloor),
		fmt.Sprintf("Счёт       : %d", a.Game.Session.Score),
		fmt.Sprintf("Сокровища  : %d", a.Game.Player.Backpack.TotalTreasure()),
		fmt.Sprintf("Ход        : %d", a.Game.Turn),
		fmt.Sprintf("В рюкзаке  : %d", len(a.Game.Player.Backpack.Slots)),
	}
	if a.Game.Player.CurrentWeapon != nil {
		lines = append(lines, fmt.Sprintf("Оружие     : %s", formatItemRu(a.Game.Player.CurrentWeapon)))
	} else {
		lines = append(lines, "Оружие     : нет")
	}
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
	statsBody := []string{
		fmt.Sprintf("Сокровища           : %d", a.Game.Player.Backpack.TotalTreasure()),
		fmt.Sprintf("Достигнутый уровень : %d", a.Game.Stats.ReachedLevel),
		fmt.Sprintf("Побеждённые враги   : %d", a.Game.Stats.DefeatedEnemies),
		fmt.Sprintf("Еда / эликсиры / свитки : %d / %d / %d", a.Game.Stats.UsedFood, a.Game.Stats.UsedPotions, a.Game.Stats.UsedScrolls),
		fmt.Sprintf("Удары / пропущено   : %d / %d", a.Game.Stats.HitsDealt, a.Game.Stats.HitsTaken),
		fmt.Sprintf("Клеток пройдено     : %d", a.Game.Stats.TilesWalked),
		fmt.Sprintf("Итог попытки        : %s", result),
	}
	boxWidth := maxLineWidth(statsBody) + 6
	lines := []string{makeRule("Статистика текущей попытки", boxWidth), ""}
	lines = append(lines, boxedSection("Итоги", statsBody, boxWidth)...)
	lines = append(lines, "", i18n.PressAnyKey)
	printLines(centeredBlock(lines, termWidth))
	_, _ = a.readKey()
}

// renderLeaderboard выводит таблицу лидеров (топ‑10 попыток).
func (a *ConsoleApp) renderLeaderboard() {
	clearScreen()
	termWidth := terminalWidth()
	lines := []string{}
	rows, err := a.Storage.Leaderboard(10)
	if err != nil {
		printCentered([]string{i18n.MsgReadStatsFail + ": " + err.Error()}, termWidth)
		return
	}
	headers := []string{"№", "Сокр", "Ур", "Враги", "Еда", "Элк", "Свит", "Уд", "Проп", "Клетки", "Итог"}
	widths := []int{2, 5, 3, 5, 3, 3, 4, 3, 4, 6, 9}

	separator := func() string {
		parts := make([]string, 0, len(widths))
		for _, w := range widths {
			parts = append(parts, strings.Repeat("-", w+2))
		}
		return "+" + strings.Join(parts, "+") + "+"
	}
	formatRow := func(values []string) string {
		cells := make([]string, 0, len(values))
		for i, v := range values {
			cells = append(cells, " "+padRight(v, widths[i])+" ")
		}
		return "|" + strings.Join(cells, "|") + "|"
	}

	tableLines := []string{separator(), formatRow(headers), separator()}
	for i, r := range rows {
		result := "поражение"
		if r.Won {
			result = "победа"
		}
		tableLines = append(tableLines, formatRow([]string{
			fmt.Sprintf("%d", i+1),
			fmt.Sprintf("%d", r.Treasures),
			fmt.Sprintf("%d", r.ReachedLevel),
			fmt.Sprintf("%d", r.DefeatedEnemies),
			fmt.Sprintf("%d", r.UsedFood),
			fmt.Sprintf("%d", r.UsedPotions),
			fmt.Sprintf("%d", r.UsedScrolls),
			fmt.Sprintf("%d", r.HitsDealt),
			fmt.Sprintf("%d", r.HitsTaken),
			fmt.Sprintf("%d", r.TilesWalked),
			result,
		}))
	}
	if len(rows) == 0 {
		tableLines = append(tableLines, formatRow([]string{"-", "-", "-", "-", "-", "-", "-", "-", "-", "-", i18n.LeaderboardEmpty}))
	}
	tableLines = append(tableLines, separator())

	tableWidth := maxLineWidth(tableLines)
	lines = append(lines, makeRule("Таблица лучших попыток", tableWidth), "")
	lines = append(lines, tableLines...)
	lines = append(lines, "", i18n.PressAnyKey)
	printLines(centeredBlock(lines, termWidth))
	_, _ = a.readKey()
}

// renderHelp отображает экран с подсказками по управлению (raw‑mode).
func (a *ConsoleApp) renderHelp() {
	clearScreen()
	termWidth := terminalWidth()
	sections := []string{
		"УПРАВЛЕНИЕ",
		"  WASD — движение и атака",
		"  b — рюкзак, h/j/k/e — быстрый список предметов",
		"  t — статистика, l — таблица лучших, ?/i — справка",
		"  q — выход из игры",
		"",
		"АТАКА",
		"  Отдельной кнопки удара нет.",
		"  Подойдите к врагу и нажмите WASD в его сторону.",
		"",
		"ПРЕДМЕТЫ",
		"  Еда лечит, эликсиры временно усиливают параметры.",
		"  Свитки дают мгновенные эффекты, оружие повышает урон.",
		"",
		"СИМВОЛЫ КАРТЫ",
		"  @ игрок, > выход, # стена, . пол, + коридор, D дверь",
		"  z/v/g/O/s враги, f еда, p эликсир, r свиток, w оружие, $ сокровище",
		"",
		"СТАТИСТИКА",
		"  На экране статистики видны итоги текущей попытки.",
		"  Esc / Enter / q — закрыть справку.",
	}
	boxWidth := maxLineWidth(sections) + 6
	lines := []string{makeRule("Справка", boxWidth), ""}
	lines = append(lines, boxedSection("Rogue — подсказки", sections, boxWidth)...)
	lines = append(lines, "", i18n.PressAnyKey)
	printLines(centeredBlock(lines, termWidth))
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
	sections := []string{
		"WASD — движение/атака | b — рюкзак | h/j/k/e — быстрый выбор",
		"t — статистика | l — лучшие попытки | ?/i — справка | q — выход",
		"В бою подойдите к врагу и нажмите WASD в его сторону.",
		"Esc / Enter / q — закрыть справку.",
	}
	boxWidth := maxLineWidth(sections) + 6
	lines := []string{makeRule("Справка", boxWidth), ""}
	lines = append(lines, boxedSection("Кратко", sections, boxWidth)...)
	lines = append(lines, "", "Нажмите Enter (или q), чтобы вернуться.")
	printLines(centeredBlock(lines, termWidth))
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
