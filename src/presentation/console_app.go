package presentation

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"
	"unsafe"

	"rogue-game/src/datalayer"
	"rogue-game/src/domain/entities"
	"rogue-game/src/domain/gameplay"
	"rogue-game/src/presentation/i18n"
)

// ConsoleApp управляет консольной игрой.
type ConsoleApp struct {
	Game         *gameplay.Game
	Storage      *datalayer.Storage
	stdin        *os.File
	rawState     *syscall.Termios
	reader       *bufio.Reader
	attemptSaved bool
}

func NewConsoleApp(game *gameplay.Game, st *datalayer.Storage) *ConsoleApp {
	return &ConsoleApp{Game: game, Storage: st, stdin: os.Stdin, reader: bufio.NewReader(os.Stdin)}
}

func (a *ConsoleApp) Run() error {
	if err := a.enableRawInput(); err != nil {
		return a.runLineMode()
	}
	defer a.disableRawInput()

	for {
		a.render()
		if a.Game.IsGameOver {
			a.Game.Stats.Treasures = a.Game.Player.Backpack.TotalTreasure()
			if !a.Game.Stats.Won {
				a.Game.Stats.Won = false
			}
			if !a.attemptSaved {
				_ = a.Storage.SaveAttempt(a.Game.Stats)
				a.attemptSaved = true
			}
			if a.Game.Stats.Won {
				fmt.Println(i18n.MsgVictoryExit)
			} else {
				fmt.Println(i18n.MsgGameOverExit)
			}
			key, err := a.readKey()
			if err != nil {
				return err
			}
			if key == 'q' {
				return nil
			}
			continue
		}

		fmt.Print(i18n.PromptCommandRaw)
		prevFloor := a.Game.Session.CurrentFloor
		cmd, err := a.readKey()
		if err != nil {
			return err
		}
		if cmd == 0 {
			continue
		}
		switch cmd {
		case 'w':
			a.Game.MovePlayer(0, -1)
		case 'a':
			a.Game.MovePlayer(-1, 0)
		case 's':
			a.Game.MovePlayer(0, 1)
		case 'd':
			a.Game.MovePlayer(1, 0)
		case 'h', 'j', 'k', 'e':
			a.openQuickInventory(string(cmd))
		case 'b':
			a.renderBackpackScreen()
		case 'l':
			a.renderLeaderboard()
		case 't':
			a.renderCurrentStats()
		case '?', 'i':
			a.renderHelp()
		case 'q':
			a.Game.Stats.Treasures = a.Game.Player.Backpack.TotalTreasure()
			_ = a.Storage.SaveAttempt(a.Game.Stats)
			return nil
		}
		if a.Game.Session.CurrentFloor > prevFloor {
			_ = a.Storage.SaveAttempt(a.Game.Stats)
		}
		if err := a.Storage.SaveGame(a.Game); err != nil {
			fmt.Println(i18n.MsgSaveFailed+":", err)
		}
	}
}

func (a *ConsoleApp) runLineMode() error {
	for {
		a.render()
		if a.Game.IsGameOver {
			a.Game.Stats.Treasures = a.Game.Player.Backpack.TotalTreasure()
			if !a.Game.Stats.Won {
				a.Game.Stats.Won = false
			}
			if !a.attemptSaved {
				_ = a.Storage.SaveAttempt(a.Game.Stats)
				a.attemptSaved = true
			}
			if a.Game.Stats.Won {
				fmt.Println(i18n.MsgVictoryExit)
			} else {
				fmt.Println(i18n.MsgGameOverExit)
			}
			line, _ := a.reader.ReadString('\n')
			if strings.TrimSpace(strings.ToLower(line)) == "q" {
				return nil
			}
			continue
		}

		fmt.Print(i18n.PromptCommandLine)
		prevFloor := a.Game.Session.CurrentFloor
		line, err := a.reader.ReadString('\n')
		if err != nil {
			return err
		}
		cmd := strings.TrimSpace(strings.ToLower(line))
		if cmd == "" {
			continue
		}
		switch cmd {
		case "w":
			a.Game.MovePlayer(0, -1)
		case "a":
			a.Game.MovePlayer(-1, 0)
		case "s":
			a.Game.MovePlayer(0, 1)
		case "d":
			a.Game.MovePlayer(1, 0)
		case "h", "j", "k", "e":
			a.openQuickInventoryLineMode(cmd)
		case "b":
			a.renderBackpackScreenLineMode()
		case "l":
			a.renderLeaderboard()
		case "t":
			a.renderCurrentStats()
		case "?", "i":
			a.renderHelpLineMode()
		case "q":
			a.Game.Stats.Treasures = a.Game.Player.Backpack.TotalTreasure()
			_ = a.Storage.SaveAttempt(a.Game.Stats)
			return nil
		}
		if a.Game.Session.CurrentFloor > prevFloor {
			_ = a.Storage.SaveAttempt(a.Game.Stats)
		}
		if err := a.Storage.SaveGame(a.Game); err != nil {
			fmt.Println(i18n.MsgSaveFailed+":", err)
		}
	}
}

func (a *ConsoleApp) enableRawInput() error {
	fd := int(a.stdin.Fd())
	oldState, err := getTermios(fd)
	if err != nil {
		return err
	}
	raw := *oldState
	raw.Lflag &^= syscall.ICANON | syscall.ECHO
	raw.Iflag &^= syscall.ICRNL | syscall.INLCR
	raw.Cc[syscall.VMIN] = 1
	raw.Cc[syscall.VTIME] = 0
	if err := setTermios(fd, &raw); err != nil {
		return err
	}
	a.rawState = oldState
	return nil
}

func (a *ConsoleApp) disableRawInput() {
	if a.rawState != nil {
		_ = setTermios(int(a.stdin.Fd()), a.rawState)
		a.rawState = nil
	}
}

func (a *ConsoleApp) readKey() (rune, error) {
	var b [1]byte
	for {
		_, err := a.stdin.Read(b[:])
		if err != nil {
			return 0, err
		}
		ch := b[0]
		if ch == 0x1b || ch == '\r' || ch == '\n' {
			return 0, nil
		}
		if ch >= 'A' && ch <= 'Z' {
			ch += 'a' - 'A'
		}
		return rune(ch), nil
	}
}

func (a *ConsoleApp) renderCurrentStats() {
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
	)
	fmt.Println(i18n.PressAnyKey)
	_, _ = a.readKey()
}

func (a *ConsoleApp) openQuickInventory(kind string) {
	indices, title, emptyMessage := a.itemsByKind(kind)
	if len(indices) == 0 {
		a.renderMessageScreen(emptyMessage)
		return
	}
	a.renderInventorySelection(title, indices)
}

func (a *ConsoleApp) openQuickInventoryLineMode(kind string) {
	indices, title, emptyMessage := a.itemsByKind(kind)
	if len(indices) == 0 {
		fmt.Println(emptyMessage)
		return
	}
	a.renderInventorySelectionLineMode(title, indices)
}

func (a *ConsoleApp) itemsByKind(kind string) ([]int, string, string) {
	filtered := make([]int, 0)
	title := i18n.InventoryTitle
	emptyMessage := ""
	for i, it := range a.Game.Player.Backpack.Slots {
		switch kind {
		case "h":
			title = i18n.QuickWeaponTitle
			emptyMessage = i18n.NoWeaponsInBackpack
			if it.Type == entities.ItemTypeWeapon {
				filtered = append(filtered, i)
			}
		case "j":
			title = i18n.QuickFoodTitle
			emptyMessage = i18n.NoFoodInBackpack
			if it.Type == entities.ItemTypeFood {
				filtered = append(filtered, i)
			}
		case "k":
			title = i18n.QuickPotionTitle
			emptyMessage = i18n.NoPotionsInBackpack
			if it.Type == entities.ItemTypePotion {
				filtered = append(filtered, i)
			}
		case "e":
			title = i18n.QuickScrollTitle
			emptyMessage = i18n.NoScrollsInBackpack
			if it.Type == entities.ItemTypeScroll {
				filtered = append(filtered, i)
			}
		}
	}
	return filtered, title, emptyMessage
}

func (a *ConsoleApp) renderInventorySelection(title string, indices []int) {
	for {
		fmt.Print("\033[H\033[2J")
		fmt.Println(title)
		fmt.Println()
		for idx, realIdx := range indices {
			fmt.Printf("%d. %s\n", idx+1, a.inventoryLineByIndex(realIdx))
		}
		if a.hasWeaponInHandsOrList(indices) {
			fmt.Println(i18n.UnequipWeapon)
		}
		fmt.Println()
		fmt.Println(i18n.ChooseItemNumber)
		fmt.Println(i18n.InventoryScreenHint)

		key, err := a.readControlKey()
		if err != nil {
			return
		}
		if key == 'q' || key == 0x1b || key == '\n' {
			return
		}
		if key == '0' && a.hasWeaponInHandsOrList(indices) {
			cc := gameplay.NewCharacterController(a.Game.Player, a.Game)
			if !cc.UnequipWeapon() {
				a.renderMessageScreen(i18n.BackpackFull)
			}
			return
		}
		if key < '1' || key > '9' {
			continue
		}
		choice := int(key - '1')
		if choice < 0 || choice >= len(indices) {
			continue
		}
		cc := gameplay.NewCharacterController(a.Game.Player, a.Game)
		_ = cc.UseItem(indices[choice])
		return
	}
}

func (a *ConsoleApp) renderInventorySelectionLineMode(title string, indices []int) {
	fmt.Println(title)
	for idx, realIdx := range indices {
		fmt.Printf("%d. %s\n", idx+1, a.inventoryLineByIndex(realIdx))
	}
	if a.hasWeaponInHandsOrList(indices) {
		fmt.Println(i18n.UnequipWeapon)
	}
	fmt.Println(i18n.ChooseItemNumber + " (1..9, q — выход)")
	line, _ := a.reader.ReadString('\n')
	line = strings.TrimSpace(strings.ToLower(line))
	if line == "" || line == "q" {
		return
	}
	if line == "0" && a.hasWeaponInHandsOrList(indices) {
		cc := gameplay.NewCharacterController(a.Game.Player, a.Game)
		if !cc.UnequipWeapon() {
			fmt.Println(i18n.BackpackFull)
		}
		return
	}
	if len(line) != 1 || line[0] < '1' || line[0] > '9' {
		return
	}
	choice := int(line[0] - '1')
	if choice < 0 || choice >= len(indices) {
		return
	}
	cc := gameplay.NewCharacterController(a.Game.Player, a.Game)
	_ = cc.UseItem(indices[choice])
}

func (a *ConsoleApp) renderBackpackScreen() {
	for {
		fmt.Print("\033[H\033[2J")
		fmt.Println(i18n.InventoryTitle)
		fmt.Println()

		entries := a.printBackpackSections()
		fmt.Println()
		fmt.Println(i18n.InventoryControl1)
		fmt.Println(i18n.InventoryControl2)
		fmt.Println(i18n.InventoryControl3)
		fmt.Println(i18n.InventoryControl4)
		if a.Game.Player.CurrentWeapon != nil {
			fmt.Printf("\nОружие в руках: %s\n", formatItemRu(a.Game.Player.CurrentWeapon))
		}

		key, err := a.readControlKey()
		if err != nil {
			return
		}
		if key == 'q' || key == 0x1b || key == '\n' {
			return
		}
		if key == '0' {
			cc := gameplay.NewCharacterController(a.Game.Player, a.Game)
			if !cc.UnequipWeapon() {
				a.renderMessageScreen(i18n.BackpackFull)
			}
			continue
		}
		if key < '1' || key > '9' {
			continue
		}
		choice := int(key - '1')
		if choice < 0 || choice >= len(entries) {
			continue
		}
		cc := gameplay.NewCharacterController(a.Game.Player, a.Game)
		_ = cc.UseItem(entries[choice])
	}
}

func (a *ConsoleApp) renderBackpackScreenLineMode() {
	fmt.Println(i18n.InventoryTitle)
	entries := a.printBackpackSections()
	fmt.Println(i18n.InventoryControl2)
	fmt.Println(i18n.InventoryControl3)
	fmt.Println("- q — закрыть рюкзак")
	line, _ := a.reader.ReadString('\n')
	line = strings.TrimSpace(strings.ToLower(line))
	if line == "" || line == "q" {
		return
	}
	if line == "0" {
		cc := gameplay.NewCharacterController(a.Game.Player, a.Game)
		if !cc.UnequipWeapon() {
			fmt.Println(i18n.BackpackFull)
		}
		return
	}
	if len(line) != 1 || line[0] < '1' || line[0] > '9' {
		return
	}
	choice := int(line[0] - '1')
	if choice < 0 || choice >= len(entries) {
		return
	}
	cc := gameplay.NewCharacterController(a.Game.Player, a.Game)
	_ = cc.UseItem(entries[choice])
}

func (a *ConsoleApp) printBackpackSections() []int {
	entries := make([]int, 0, len(a.Game.Player.Backpack.Slots))
	number := 1
	sections := []struct {
		title string
		typee entities.ItemType
	}{
		{"Оружие:", entities.ItemTypeWeapon},
		{"Еда:", entities.ItemTypeFood},
		{"Эликсиры:", entities.ItemTypePotion},
		{"Свитки:", entities.ItemTypeScroll},
	}
	for _, sec := range sections {
		fmt.Println(sec.title)
		has := false
		for i, it := range a.Game.Player.Backpack.Slots {
			if it.Type != sec.typee {
				continue
			}
			has = true
			fmt.Printf("%d. %s\n", number, a.inventoryLineByIndex(i))
			entries = append(entries, i)
			number++
		}
		if !has {
			fmt.Println(i18n.InventoryEmptyLine)
		}
		fmt.Println()
	}
	if a.Game.Player.CurrentWeapon != nil {
		fmt.Println(i18n.UnequipWeapon)
	}
	return entries
}

func (a *ConsoleApp) hasWeaponInHandsOrList(indices []int) bool {
	if a.Game.Player.CurrentWeapon != nil {
		return true
	}
	for _, idx := range indices {
		if a.Game.Player.Backpack.Slots[idx].Type == entities.ItemTypeWeapon {
			return true
		}
	}
	return false
}

func (a *ConsoleApp) inventoryLineByIndex(index int) string {
	item := a.Game.Player.Backpack.Slots[index]
	line := formatItemRu(item)
	if item.Type == entities.ItemTypeWeapon && a.Game.Player.CurrentWeapon == item {
		line += " " + i18n.EquippedMarker
	}
	return line
}

func (a *ConsoleApp) renderMessageScreen(msg string) {
	fmt.Print("\033[H\033[2J")
	fmt.Println(msg)
	fmt.Println(i18n.PressAnyKey)
	_, _ = a.readControlKey()
}

func (a *ConsoleApp) render() {
	fmt.Print("\033[H\033[2J")
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

func (a *ConsoleApp) renderLeaderboard() {
	fmt.Println(i18n.LeaderboardTitle)
	rows, err := a.Storage.Leaderboard(10)
	if err != nil {
		fmt.Println(i18n.MsgReadStatsFail+":", err)
		return
	}
	for i, r := range rows {
		fmt.Printf(i18n.LeaderboardLine, i+1, r.Treasures, r.ReachedLevel, r.DefeatedEnemies, r.TilesWalked)
	}
	if len(rows) == 0 {
		fmt.Println(i18n.LeaderboardEmpty)
	}
	fmt.Println(i18n.PressAnyKey)
	_, _ = a.readKey()
}

func (a *ConsoleApp) renderHelp() {
	fmt.Print("\033[H\033[2J")
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
	fmt.Print("\033[H\033[2J")
	for _, line := range i18n.HelpLines {
		fmt.Println(line)
	}
	line, _ := a.reader.ReadString('\n')
	line = strings.TrimSpace(strings.ToLower(line))
	if line == "q" || line == "" {
		return
	}
}

func (a *ConsoleApp) readControlKey() (rune, error) {
	var b [1]byte
	_, err := a.stdin.Read(b[:])
	if err != nil {
		return 0, err
	}
	ch := b[0]
	if ch >= 'A' && ch <= 'Z' {
		ch += 'a' - 'A'
	}
	if ch == 0x1b {
		return 0x1b, nil
	}
	if ch == '\r' || ch == '\n' {
		return '\n', nil
	}
	return rune(ch), nil
}

func getTermios(fd int) (*syscall.Termios, error) {
	state := &syscall.Termios{}
	_, _, errno := syscall.Syscall6(syscall.SYS_IOCTL, uintptr(fd), uintptr(syscall.TCGETS), uintptr(unsafe.Pointer(state)), 0, 0, 0)
	if errno != 0 {
		return nil, errno
	}
	return state, nil
}

func setTermios(fd int, state *syscall.Termios) error {
	_, _, errno := syscall.Syscall6(syscall.SYS_IOCTL, uintptr(fd), uintptr(syscall.TCSETS), uintptr(unsafe.Pointer(state)), 0, 0, 0)
	if errno != 0 {
		return errno
	}
	return nil
}

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

func formatItemRu(i *entities.Item) string {
	if i == nil {
		return "нет"
	}
	switch i.Type {
	case entities.ItemTypeWeapon:
		sub := "оружие"
		if i.Subtype == entities.SubtypeSword {
			sub = "меч"
		} else if i.Subtype == entities.SubtypeBow {
			sub = "лук"
		}
		return fmt.Sprintf("%s (сила +%d)", sub, i.StrengthBoost)
	case entities.ItemTypeFood:
		sub := "еда"
		if i.Subtype == entities.SubtypeBread {
			sub = "хлеб"
		} else if i.Subtype == entities.SubtypeApple {
			sub = "яблоко"
		}
		return fmt.Sprintf("%s (лечение +%d)", sub, i.HealthBoost)
	case entities.ItemTypePotion:
		sub := "эликсир"
		if i.Subtype == entities.SubtypeHealthPotion {
			sub = "эликсир здоровья"
		} else if i.Subtype == entities.SubtypeStrengthPotion {
			sub = "эликсир силы"
		}
		return fmt.Sprintf("%s (сила +%d, ловкость +%d, макс.здоровье +%d)", sub, i.StrengthBoost, i.DexterityBoost, i.MaxHealthBoost)
	case entities.ItemTypeScroll:
		sub := "свиток"
		if i.Subtype == entities.SubtypeScrollOfStrength {
			sub = "свиток силы"
		} else if i.Subtype == entities.SubtypeScrollOfDexterity {
			sub = "свиток ловкости"
		}
		return fmt.Sprintf("%s (сила +%d, ловкость +%d, макс.здоровье +%d)", sub, i.StrengthBoost, i.DexterityBoost, i.MaxHealthBoost)
	case entities.ItemTypeTreasure:
		return fmt.Sprintf("сокровище (ценность %d)", i.Value)
	default:
		return "неизвестный предмет"
	}
}
