package datalayer

import (
	// Стандартные пакеты для работы с JSON, файлами и сортировкой
	"encoding/json"
	"os"
	"path/filepath"
	"sort"

	// Внутренние пакеты проекта
	"rogue-game/src/domain/entities" // Сущности игровой сессии
	"rogue-game/src/domain/gameplay" // Логика игры и статистика
)

// SaveData хранит восстановимое состояние игры.
type SaveData struct {
	Session          *entities.GameSession  `json:"session"`
	GroundItems      []*gameplay.GroundItem `json:"ground_items"`
	Turn             int                    `json:"turn"`
	Stats            gameplay.AttemptStats  `json:"stats"`
	Seed             int64                  `json:"seed"`
	Visible          [][]bool               `json:"visible"`
	Explored         [][]bool               `json:"explored"`
	ExitPos          entities.Position      `json:"exit_pos"`
	IsGameOver       bool                   `json:"is_game_over"`
	PlayerSleepTurns int                    `json:"player_sleep_turns"`
	PotionEffects    []gameplay.TimedEffect `json:"potion_effects"`
	VampireFirstMiss map[int]bool           `json:"vampire_first_miss"`
	OgreRestTurns    map[int]int            `json:"ogre_rest_turns"`
	OgreCounterReady map[int]bool           `json:"ogre_counter_ready"`
	SnakeSideLeft    map[int]bool           `json:"snake_side_left"`
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

// NewStorage создаёт новый экземпляр хранилища с указанными путями к файлам сохранения и статистики.
func NewStorage(savePath, statsPath string) *Storage {
	return &Storage{SavePath: savePath, StatsPath: statsPath}
}

// SaveGame сохраняет текущее состояние игры в JSON-файл.
func (s *Storage) SaveGame(g *gameplay.Game) error {
	payload := SaveData{
		Session:          g.Session,
		GroundItems:      g.GroundItems,
		Turn:             g.Turn,
		Stats:            g.Stats,
		Seed:             g.Seed,
		Visible:          g.Visible,
		Explored:         g.Explored,
		ExitPos:          g.ExitPos,
		IsGameOver:       g.IsGameOver,
		PlayerSleepTurns: g.PlayerSleepTurns,
		PotionEffects:    g.PotionEffects,
		VampireFirstMiss: g.VampireFirstMiss,
		OgreRestTurns:    g.OgreRestTurns,
		OgreCounterReady: g.OgreCounterReady,
		SnakeSideLeft:    g.SnakeSideLeft,
	}
	return writeJSON(s.SavePath, payload)
}

// LoadGame загружает состояние игры из JSON-файла и восстанавливает игровой объект.
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
	g.ExitPos = payload.ExitPos
	g.IsGameOver = payload.IsGameOver
	g.PlayerSleepTurns = payload.PlayerSleepTurns
	g.PotionEffects = payload.PotionEffects
	g.VampireFirstMiss = payload.VampireFirstMiss
	g.OgreRestTurns = payload.OgreRestTurns
	g.OgreCounterReady = payload.OgreCounterReady
	g.SnakeSideLeft = payload.SnakeSideLeft
	// Инициализация карт, если они nil (для обратной совместимости)
	if g.VampireFirstMiss == nil {
		g.VampireFirstMiss = map[int]bool{}
	}
	if g.OgreRestTurns == nil {
		g.OgreRestTurns = map[int]int{}
	}
	if g.OgreCounterReady == nil {
		g.OgreCounterReady = map[int]bool{}
	}
	if g.SnakeSideLeft == nil {
		g.SnakeSideLeft = map[int]bool{}
	}
	// Восстановление видимости и исследованных клеток, если размеры совпадают
	if len(payload.Visible) == g.CurrentLevel.Height && len(payload.Explored) == g.CurrentLevel.Height {
		g.Visible = payload.Visible
		g.Explored = payload.Explored
	}
	return g, nil
}

// SaveAttempt добавляет статистику завершённой попытки в файл статистики.
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

// LoadStats загружает весь файл статистики.
func (s *Storage) LoadStats() (StatsFile, error) {
	var sf StatsFile
	err := readJSON(s.StatsPath, &sf)
	return sf, err
}

// Leaderboard возвращает отсортированный по убыванию сокровищ список попыток (лидерборд).
// Параметр limit ограничивает количество возвращаемых записей (0 = без ограничения).
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

// writeJSON записывает произвольную структуру в JSON-файл с отступами.
// Создаёт директории, если они не существуют.
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

// readJSON читает JSON-файл и парсит его в переданную структуру.
func readJSON(path string, v interface{}) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, v)
}
