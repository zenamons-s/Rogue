package datalayer

import (
	"path/filepath"
	"testing"

	"rogue-game/src/domain/gameplay"
)

func TestSaveAttempt_PreservesWonFlag(t *testing.T) {
	tmp := t.TempDir()
	storage := NewStorage(filepath.Join(tmp, "save.json"), filepath.Join(tmp, "stats.json"))

	if err := storage.SaveAttempt(gameplay.AttemptStats{
		Treasures:    123,
		ReachedLevel: 21,
		Won:          true,
	}); err != nil {
		t.Fatalf("save attempt: %v", err)
	}

	rows, err := storage.Leaderboard(10)
	if err != nil {
		t.Fatalf("leaderboard: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected one row, got %d", len(rows))
	}
	if !rows[0].Won {
		t.Fatalf("expected won=true after save/load roundtrip")
	}
}
