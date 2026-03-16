package main

import (
	// Стандартные пакеты для ввода-вывода и работы со строками
	"bufio"
	"fmt"
	"os"
	"strings"

	// Внутренние пакеты проекта
	"rogue-game/src/datalayer"      // Работа с сохранениями и статистикой
	"rogue-game/src/domain/gameplay" // Логика игры: генерация, персонажи, бой
	"rogue-game/src/presentation"    // Консольный интерфейс
	"rogue-game/src/presentation/i18n" // Локализация (русские тексты)
)

// main - точка входа в приложение Rogue-like игры.
// Управляет инициализацией хранилища, загрузкой сохранения, созданием игры и запуском консольного интерфейса.
func main() {
	// Инициализация хранилища для работы с сохранениями и статистикой
	storage := datalayer.NewStorage("data/save.json", "data/stats.json")

	// Подготовка чтения пользовательского ввода
	reader := bufio.NewReader(os.Stdin)

	// Запрос у пользователя: загрузить сохранение или начать новую игру
	fmt.Print(i18n.LoadSavePrompt)
	line, _ := reader.ReadString('\n')
	line = strings.TrimSpace(strings.ToLower(line))

	// Выход, если пользователь ввёл 'q'
	if line == "q" {
		return
	}

	var game *gameplay.Game

	// Если пользователь выбрал загрузку (1 или y), пытаемся загрузить сохранение
	if line == "1" || line == "y" {
		loaded, err := storage.LoadGame()
		if err == nil {
			game = loaded
			fmt.Println(i18n.SaveLoaded)
		} else {
			fmt.Println(i18n.SaveNotFoundNewGame)
		}
	}

	// Если игра не была загружена (пользователь отказался или сохранение не найдено),
	// создаём новую сгенерированную игру
	if game == nil {
		// NewGeneratedGame создаёт игровой мир с заданными размерами и сидом:
		// ширина = 60, высота = 25, seed = 0 (случайная генерация)
		game = gameplay.NewGeneratedGame(60, 25, 0)
	}

	// Создаём консольное приложение и запускаем основной цикл
	app := presentation.NewConsoleApp(game, storage)
	if err := app.Run(); err != nil {
		fmt.Println(i18n.AppErrorPrefix+":", err)
		os.Exit(1)
	}
}
