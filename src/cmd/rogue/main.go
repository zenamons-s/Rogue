package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"rogue-game/src/datalayer"
	"rogue-game/src/domain/gameplay"
	"rogue-game/src/presentation"
	"rogue-game/src/presentation/i18n"
)

func main() {
	storage := datalayer.NewStorage("data/save.json", "data/stats.json")
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(i18n.LoadSavePrompt)
	line, _ := reader.ReadString('\n')
	line = strings.TrimSpace(strings.ToLower(line))
	if line == "q" {
		return
	}

	var game *gameplay.Game
	if line == "1" || line == "y" {
		loaded, err := storage.LoadGame()
		if err == nil {
			game = loaded
			fmt.Println(i18n.SaveLoaded)
		} else {
			fmt.Println(i18n.SaveNotFoundNewGame)
		}
	}
	if game == nil {
		game = gameplay.NewGeneratedGame(60, 25, 0)
	}
	app := presentation.NewConsoleApp(game, storage)
	if err := app.Run(); err != nil {
		fmt.Println(i18n.AppErrorPrefix+":", err)
		os.Exit(1)
	}
}
