package presentation

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"rogue-game/src/datalayer"
	"rogue-game/src/domain/entities"
	"rogue-game/src/domain/gameplay"
)

// ConsoleApp управляет консольной игрой.
type ConsoleApp struct {
	Game    *gameplay.Game
	Storage *datalayer.Storage
	Reader  *bufio.Reader
}

func NewConsoleApp(game *gameplay.Game, st *datalayer.Storage) *ConsoleApp {
	return &ConsoleApp{Game: game, Storage: st, Reader: bufio.NewReader(os.Stdin)}
}

func (a *ConsoleApp) Run() error {
	for {
		a.render()
		if a.Game.IsGameOver {
			a.Game.Stats.Treasures = a.Game.Player.Backpack.TotalTreasure()
			a.Game.Stats.Won = false
			_ = a.Storage.SaveAttempt(a.Game.Stats)
			fmt.Println("Game Over. Нажмите q для выхода.")
			line, _ := a.Reader.ReadString('\n')
			if strings.TrimSpace(line) == "q" {
				return nil
			}
			continue
		}

		fmt.Print("Команда (w/a/s/d, h/j/k/e, l leaderboard, q quit): ")
		line, err := a.Reader.ReadString('\n')
		if err != nil {
			return err
		}
		cmd := strings.TrimSpace(line)
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
		if err := a.Storage.SaveGame(a.Game); err != nil {
			fmt.Println("Ошибка сохранения:", err)
		}
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
	fmt.Println("Нажмите Enter...")
	_, _ = a.Reader.ReadString('\n')
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
	line, _ := a.Reader.ReadString('\n')
	line = strings.TrimSpace(line)
	if line == "" {
		return
	}
	if kind == "h" && line == "0" {
		cc := gameplay.NewCharacterController(a.Game.Player, a.Game)
		if !cc.UnequipWeapon() {
			fmt.Println("Недостаточно места в рюкзаке")
		}
		return
	}
	choice := int(line[0] - '1')
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
	fmt.Println("Нажмите Enter...")
	_, _ = a.Reader.ReadString('\n')
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
