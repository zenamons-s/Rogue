package datalayer

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"

	"rogue-game/src/domain/entities"
	"rogue-game/src/domain/gameplay"
)

// SaveData хранит восстановимое состояние игры.
type SaveData struct {
	Session     *entities.GameSession  `json:"session"`
	GroundItems []*gameplay.GroundItem `json:"ground_items"`
	Turn        int                    `json:"turn"`
	Stats       gameplay.AttemptStats  `json:"stats"`
	Seed        int64                  `json:"seed"`
}

// StatsFile хранит статистику всех попыток.
type StatsFile struct {
	Attempts []gameplay.AttemptStats `json:"attempts"`
}

// Storage файловый репозиторий save/load + статистика.
type Storage struct {
	SavePath  string
	StatsPath string
}

func NewStorage(savePath, statsPath string) *Storage {
	return &Storage{SavePath: savePath, StatsPath: statsPath}
}

func (s *Storage) SaveGame(g *gameplay.Game) error {
	payload := SaveData{
		Session:     g.Session,
		GroundItems: g.GroundItems,
		Turn:        g.Turn,
		Stats:       g.Stats,
		Seed:        g.Seed,
	}
	return writeJSON(s.SavePath, payload)
}

func (s *Storage) LoadGame() (*gameplay.Game, error) {
	var payload SaveData
	if err := readJSON(s.SavePath, &payload); err != nil {
		return nil, err
	}
	g := gameplay.NewGame(payload.Session)
	g.GroundItems = payload.GroundItems
	g.Turn = payload.Turn
	g.Stats = payload.Stats
	g.Seed = payload.Seed
	return g, nil
}

func (s *Storage) SaveAttempt(st gameplay.AttemptStats) error {
	stats, err := s.LoadStats()
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		stats = StatsFile{Attempts: []gameplay.AttemptStats{}}
	}
	stats.Attempts = append(stats.Attempts, st)
	return writeJSON(s.StatsPath, stats)
}

func (s *Storage) LoadStats() (StatsFile, error) {
	var sf StatsFile
	err := readJSON(s.StatsPath, &sf)
	return sf, err
}

func (s *Storage) Leaderboard(limit int) ([]gameplay.AttemptStats, error) {
	sf, err := s.LoadStats()
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	sort.Slice(sf.Attempts, func(i, j int) bool {
		if sf.Attempts[i].Treasures == sf.Attempts[j].Treasures {
			return sf.Attempts[i].ReachedLevel > sf.Attempts[j].ReachedLevel
		}
		return sf.Attempts[i].Treasures > sf.Attempts[j].Treasures
	})
	if limit > 0 && len(sf.Attempts) > limit {
		return sf.Attempts[:limit], nil
	}
	return sf.Attempts, nil
}

func writeJSON(path string, v interface{}) error {
	dir := filepath.Dir(path)
	if dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}

	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}

func readJSON(path string, v interface{}) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, v)
}
