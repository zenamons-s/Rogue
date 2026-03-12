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
				fmt.Println("Победа! Нажмите q для выхода.")
			} else {
				fmt.Println("Game Over. Нажмите q для выхода.")
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

		fmt.Print("Команда (WASD, h/j/k/e, l leaderboard, t stats, q quit): ")
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
			a.useInventory(string(cmd))
		case 'l':
			a.renderLeaderboard()
		case 't':
			a.renderCurrentStats()
		case 'q':
			a.Game.Stats.Treasures = a.Game.Player.Backpack.TotalTreasure()
			_ = a.Storage.SaveAttempt(a.Game.Stats)
			return nil
		}
		if a.Game.Session.CurrentFloor > prevFloor {
			_ = a.Storage.SaveAttempt(a.Game.Stats)
		}
		if err := a.Storage.SaveGame(a.Game); err != nil {
			fmt.Println("Ошибка сохранения:", err)
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
				fmt.Println("Победа! Нажмите q для выхода.")
			} else {
				fmt.Println("Game Over. Нажмите q для выхода.")
			}
			line, _ := a.reader.ReadString('\n')
			if strings.TrimSpace(strings.ToLower(line)) == "q" {
				return nil
			}
			continue
		}

		fmt.Print("Команда (w/a/s/d, h/j/k/e, l leaderboard, t stats, q quit): ")
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
			a.useInventory(cmd)
		case "l":
			a.renderLeaderboard()
		case "t":
			a.renderCurrentStats()
		case "q":
			a.Game.Stats.Treasures = a.Game.Player.Backpack.TotalTreasure()
			_ = a.Storage.SaveAttempt(a.Game.Stats)
			return nil
		}
		if a.Game.Session.CurrentFloor > prevFloor {
			_ = a.Storage.SaveAttempt(a.Game.Stats)
		}
		if err := a.Storage.SaveGame(a.Game); err != nil {
			fmt.Println("Ошибка сохранения:", err)
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
	fmt.Println("=== Текущая статистика ===")
	fmt.Printf("Treasure=%d ReachedLevel=%d Kills=%d Food=%d Potions=%d Scrolls=%d HitsDealt=%d HitsTaken=%d TilesWalked=%d\n",
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
	fmt.Println("Нажмите любую клавишу...")
	_, _ = a.readKey()
}

func (a *ConsoleApp) useInventory(kind string) {
	filtered := make([]int, 0)
	for i, it := range a.Game.Player.Backpack.Slots {
		switch kind {
		case "h":
			if it.Type == entities.ItemTypeWeapon {
				filtered = append(filtered, i)
			}
		case "j":
			if it.Type == entities.ItemTypeFood {
				filtered = append(filtered, i)
			}
		case "k":
			if it.Type == entities.ItemTypePotion {
				filtered = append(filtered, i)
			}
		case "e":
			if it.Type == entities.ItemTypeScroll {
				filtered = append(filtered, i)
			}
		}
	}
	if len(filtered) == 0 {
		fmt.Println("Нет подходящих предметов")
		return
	}
	fmt.Println("Выберите номер:")
	for idx, realIdx := range filtered {
		fmt.Printf("%d) %+v\n", idx+1, *a.Game.Player.Backpack.Slots[realIdx])
	}
	if kind == "h" {
		fmt.Println("0) Убрать оружие из рук")
	}

	key, err := a.readKey()
	if err != nil || key == 0 {
		return
	}
	if kind == "h" && key == '0' {
		cc := gameplay.NewCharacterController(a.Game.Player, a.Game)
		if !cc.UnequipWeapon() {
			fmt.Println("Недостаточно места в рюкзаке")
		}
		return
	}
	if key < '1' || key > '9' {
		return
	}
	choice := int(key - '1')
	if choice < 0 || choice >= len(filtered) {
		return
	}
	cc := gameplay.NewCharacterController(a.Game.Player, a.Game)
	_ = cc.UseItem(filtered[choice])
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
	fmt.Printf("HP:%d/%d STR:%d DEX:%d Floor:%d Score:%d Treasure:%d Turn:%d\n",
		a.Game.Player.Health,
		a.Game.Player.MaxHealth,
		a.Game.Player.Strength,
		a.Game.Player.Dexterity,
		a.Game.Session.CurrentFloor,
		a.Game.Session.Score,
		a.Game.Player.Backpack.TotalTreasure(),
		a.Game.Turn,
	)
	fmt.Printf("Inventory items: %d\n", len(a.Game.Player.Backpack.Slots))
	if a.Game.Player.CurrentWeapon != nil {
		fmt.Printf("Weapon equipped: %+v\n", *a.Game.Player.CurrentWeapon)
	} else {
		fmt.Println("Weapon equipped: none")
	}
}

func (a *ConsoleApp) renderLeaderboard() {
	fmt.Println("=== Leaderboard ===")
	rows, err := a.Storage.Leaderboard(10)
	if err != nil {
		fmt.Println("Ошибка чтения статистики:", err)
		return
	}
	for i, r := range rows {
		fmt.Printf("%d) treasure=%d level=%d kills=%d walked=%d\n", i+1, r.Treasures, r.ReachedLevel, r.DefeatedEnemies, r.TilesWalked)
	}
	if len(rows) == 0 {
		fmt.Println("Пока пусто")
	}
	fmt.Println("Нажмите любую клавишу...")
	_, _ = a.readKey()
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
