# Архитектурная диаграмма для Rogue-like на Go

```mermaid
graph TB
    subgraph "Presentation Layer"
        UI[UI Module]
        Renderer[Renderer]
        Controller[Controller]
        Views[Views]
        Input[Input Handler]
    end

    subgraph "Domain Layer"
        Entities[Entities]
        Services[Services]
        RepoInterfaces[Repository Interfaces]
        GameLogic[Game Logic]
    end

    subgraph "Data Layer"
        RepoImpl[Repository Implementations]
        JSONStorage[JSON Storage]
        FileOps[File Operations]
    end

    subgraph "Infrastructure"
        Terminal[Terminal Library tcell]
        Time[Time Utilities]
        Random[Random Generator]
    end

    subgraph "Application"
        UseCases[Use Cases]
        GameLoop[Game Loop]
        Main[Main Entry Point]
    end

    Main --> GameLoop
    GameLoop --> Controller
    Controller --> Input
    Controller --> Renderer
    Renderer --> Terminal
    Controller --> UseCases
    UseCases --> Services
    UseCases --> RepoInterfaces
    Services --> Entities
    Services --> GameLogic
    RepoInterfaces --> RepoImpl
    RepoImpl --> JSONStorage
    JSONStorage --> FileOps
    GameLogic --> Random
    GameLogic --> Time
    Views --> Renderer
```

## Описание компонентов

### Presentation Layer
- **UI Module**: координация UI компонентов.
- **Renderer**: абстракция для отрисовки игрового состояния в терминале.
- **Controller**: обработка пользовательского ввода и управление потоком игры.
- **Views**: различные экраны (игра, меню, статистика).
- **Input Handler**: чтение и интерпретация нажатий клавиш.

### Domain Layer
- **Entities**: основные структуры данных (Player, Monster, Item, Room, Level, etc.).
- **Services**: сервисы, реализующие игровую логику (генерация, бой, движение).
- **Repository Interfaces**: интерфейсы для доступа к данным (например, GameRepository).
- **Game Logic**: чистые функции игровой механики.

### Data Layer
- **Repository Implementations**: реализации репозиториев для хранения в JSON.
- **JSON Storage**: сериализация/десериализация данных.
- **File Operations**: чтение/запись файлов.

### Infrastructure
- **Terminal Library**: внешняя библиотека для работы с терминалом (tcell).
- **Time Utilities**: утилиты для работы со временем.
- **Random Generator**: генерация случайных чисел.

### Application
- **Use Cases**: основные сценарии приложения (начать новую игру, загрузить игру, сделать ход).
- **Game Loop**: основной цикл игры, связывающий все компоненты.
- **Main Entry Point**: точка входа приложения.

## Поток данных
1. Пользовательский ввод через **Input Handler** передаётся в **Controller**.
2. **Controller** вызывает соответствующий **Use Case**.
3. **Use Case** использует **Services** для изменения состояния **Entities**.
4. Изменённое состояние сохраняется через **Repository Interfaces** в **Data Layer**.
5. **Controller** обновляет **Views** и передаёт данные в **Renderer**.
6. **Renderer** использует **Terminal Library** для отрисовки на экране.