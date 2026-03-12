package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"rogue-game/src/datalayer"
	"rogue-game/src/domain/gameplay"
	"rogue-game/src/presentation"
)

func main() {
	storage := datalayer.NewStorage("data/save.json", "data/stats.json")
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Загрузить последнюю игру? (y/N): ")
	line, _ := reader.ReadString('\n')
	line = strings.TrimSpace(strings.ToLower(line))

	var game *gameplay.Game
	if line == "y" {
		loaded, err := storage.LoadGame()
		if err == nil {
			game = loaded
		} else {
			fmt.Println("Сохранение не найдено, старт новой игры")
		}
	}
	if game == nil {
		game = gameplay.NewGeneratedGame(60, 25, 0)
	}
	app := presentation.NewConsoleApp(game, storage)
	if err := app.Run(); err != nil {
		fmt.Println("Ошибка:", err)
		os.Exit(1)
	}
}
