package presentation

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"rogue-game/src/datalayer"
	"rogue-game/src/domain/gameplay"
)

func TestNewConsoleApp_MarksFinishedGameAsSaved(t *testing.T) {
	game := gameplay.NewGeneratedGame(60, 25, 1)
	game.AttemptSaved = true

	app := NewConsoleApp(game, datalayer.NewStorage("save.json", "stats.json"))

	if !app.Game.AttemptSaved {
		t.Fatalf("expected attemptSaved=true for loaded game")
	}
}

func TestLoadSavedGame_SetsAttemptSavedByGameState(t *testing.T) {
	tmp := t.TempDir()
	storage := datalayer.NewStorage(filepath.Join(tmp, "save.json"), filepath.Join(tmp, "stats.json"))

	saved := gameplay.NewGeneratedGame(60, 25, 1)
	saved.IsGameOver = true
	saved.AttemptSaved = true
	if err := storage.SaveGame(saved); err != nil {
		t.Fatalf("save game: %v", err)
	}

	app := NewConsoleApp(gameplay.NewGeneratedGame(60, 25, 2), storage)
	app.loadSavedGame()
	if !app.Game.AttemptSaved {
		t.Fatalf("expected attemptSaved=true after loading finished save")
	}

	saved.IsGameOver = false
	saved.AttemptSaved = false
	if err := storage.SaveGame(saved); err != nil {
		t.Fatalf("save game: %v", err)
	}
	app.loadSavedGame()
	if app.Game.AttemptSaved {
		t.Fatalf("expected attemptSaved=false after loading active save")
	}
}

func TestPersistAttemptIfNeeded_WritesFlagToSave(t *testing.T) {
	tmp := t.TempDir()
	storage := datalayer.NewStorage(filepath.Join(tmp, "save.json"), filepath.Join(tmp, "stats.json"))
	app := NewConsoleApp(gameplay.NewGeneratedGame(60, 25, 3), storage)
	app.Game.IsGameOver = true
	app.Game.AttemptSaved = false

	app.persistAttemptIfNeeded()

	loaded, err := storage.LoadGame()
	if err != nil {
		t.Fatalf("load game: %v", err)
	}
	if !loaded.AttemptSaved {
		t.Fatalf("expected attempt_saved=true in persisted save data")
	}
}

func TestRenderLeaderboard_ClearsScreenBeforeRender(t *testing.T) {
	tmp := t.TempDir()
	storage := datalayer.NewStorage(filepath.Join(tmp, "save.json"), filepath.Join(tmp, "stats.json"))
	app := NewConsoleApp(gameplay.NewGeneratedGame(60, 25, 1), storage)
	app.reader = bufio.NewReader(strings.NewReader("x"))

	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = w

	app.renderLeaderboard()

	_ = w.Close()
	os.Stdout = oldStdout
	out, _ := io.ReadAll(r)
	_ = r.Close()

	if !bytes.HasPrefix(out, []byte("\033[H\033[2J")) {
		t.Fatalf("expected clear-screen sequence at output start, got %q", string(out))
	}
}
