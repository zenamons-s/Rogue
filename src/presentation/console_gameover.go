// Пакет presentation содержит консольный интерфейс игры (raw‑mode и line‑mode).
package presentation

import (
	"fmt"
	"strings"

	"rogue-game/src/presentation/i18n"
)

// handleGameOverRaw обрабатывает экран завершения игры в raw‑mode.
// Возвращает true, если игрок выбрал выход, иначе false (начата новая или загружена игра).
func (a *ConsoleApp) handleGameOverRaw() (bool, error) {
	a.persistAttemptIfNeeded()
	if a.Game.Stats.Won {
		fmt.Println(i18n.MsgVictoryExit)
	} else {
		fmt.Println(i18n.MsgGameOverExit)
	}
	key, err := a.readKey()
	if err != nil {
		return false, err
	}
	switch key {
	case 'q':
		return true, nil
	case 'n':
		a.startNewGame()
	case 'l':
		a.loadSavedGame()
	}
	return false, nil
}

// handleGameOverLineMode обрабатывает экран завершения игры в line‑mode.
func (a *ConsoleApp) handleGameOverLineMode() (bool, error) {
	a.persistAttemptIfNeeded()
	if a.Game.Stats.Won {
		fmt.Println(i18n.MsgVictoryExit)
	} else {
		fmt.Println(i18n.MsgGameOverExit)
	}
	line, err := a.reader.ReadString('\n')
	if err != nil {
		return false, err
	}
	switch strings.TrimSpace(strings.ToLower(line)) {
	case "q":
		return true, nil
	case "n":
		a.startNewGame()
	case "l":
		a.loadSavedGame()
	}
	return false, nil
}

// persistAttemptIfNeeded сохраняет статистику попытки, если она ещё не была сохранена.
func (a *ConsoleApp) persistAttemptIfNeeded() {
	a.Game.Stats.Treasures = a.Game.Player.Backpack.TotalTreasure()
	if !a.Game.AttemptSaved {
		_ = a.Storage.SaveAttempt(a.Game.Stats)
		a.Game.AttemptSaved = true
		_ = a.Storage.SaveGame(a.Game)
	}
}

// startNewGame сбрасывает игровую сессию и начинает новую игру.
func (a *ConsoleApp) startNewGame() {
	a.Game.ResetAsNewSession()
	_ = a.Storage.SaveGame(a.Game)
}

// loadSavedGame загружает сохранённую игру из хранилища.
func (a *ConsoleApp) loadSavedGame() {
	loaded, err := a.Storage.LoadGame()
	if err != nil {
		fmt.Println(i18n.MsgLoadFailed+":", err)
		return
	}
	a.Game = loaded
}
