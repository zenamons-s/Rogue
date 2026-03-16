// Пакет presentation содержит консольный интерфейс игры (raw‑mode и line‑mode).
package presentation

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"rogue-game/src/datalayer"
	"rogue-game/src/domain/gameplay"
	"rogue-game/src/presentation/i18n"
)

// ConsoleApp управляет консольной игрой.
type ConsoleApp struct {
	Game         *gameplay.Game
	Storage      *datalayer.Storage
	stdin        *os.File
	rawState     *syscall.Termios
	restoreOnce  sync.Once
	reader       *bufio.Reader
	attemptSaved bool
}

// NewConsoleApp создаёт новый экземпляр консольного приложения.
func NewConsoleApp(game *gameplay.Game, st *datalayer.Storage) *ConsoleApp {
	return &ConsoleApp{Game: game, Storage: st, stdin: os.Stdin, reader: bufio.NewReader(os.Stdin)}
}

func (a *ConsoleApp) Run() error {
	if err := a.enterRawMode(); err != nil {
		return a.runLineMode()
	}
	defer a.restoreTerminal()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	stopSignals := make(chan struct{})
	defer signal.Stop(sigCh)
	defer close(stopSignals)

	go func() {
		select {
		case sig := <-sigCh:
			fmt.Println()
			a.restoreTerminal()
			if sig == syscall.SIGINT {
				os.Exit(130)
			}
			os.Exit(143)
		case <-stopSignals:
			return
		}
	}()

	for {
		a.render()
		if a.Game.IsGameOver {
			done, err := a.handleGameOverRaw()
			if err != nil {
				return err
			}
			if done {
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
			done, err := a.handleGameOverLineMode()
			if err != nil {
				return err
			}
			if done {
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













