// Пакет presentation содержит консольный интерфейс игры (raw‑mode и line‑mode).
package presentation

import (
	"fmt"
	"strings"

	"rogue-game/src/domain/entities"
	"rogue-game/src/domain/gameplay"
	"rogue-game/src/presentation/i18n"
)

// openQuickInventory открывает быстрый инвентарь для предметов определённого типа (raw‑mode).
func (a *ConsoleApp) openQuickInventory(kind string) {
	indices, title, emptyMessage := a.itemsByKind(kind)
	if len(indices) == 0 {
		a.renderMessageScreen(emptyMessage)
		return
	}
	a.renderInventorySelection(title, indices)
}

// openQuickInventoryLineMode открывает быстрый инвентарь в line‑mode.
func (a *ConsoleApp) openQuickInventoryLineMode(kind string) {
	indices, title, emptyMessage := a.itemsByKind(kind)
	if len(indices) == 0 {
		fmt.Println(emptyMessage)
		return
	}
	a.renderInventorySelectionLineMode(title, indices)
}

// itemsByKind возвращает индексы предметов заданного типа, заголовок и сообщение об отсутствии.
func (a *ConsoleApp) itemsByKind(kind string) ([]int, string, string) {
	filtered := make([]int, 0)
	title := i18n.InventoryTitle
	emptyMessage := ""
	for i, it := range a.Game.Player.Backpack.Slots {
		switch kind {
		case "h":
			title = i18n.QuickWeaponTitle
			emptyMessage = i18n.NoWeaponsInBackpack
			if it.Type == entities.ItemTypeWeapon {
				filtered = append(filtered, i)
			}
		case "j":
			title = i18n.QuickFoodTitle
			emptyMessage = i18n.NoFoodInBackpack
			if it.Type == entities.ItemTypeFood {
				filtered = append(filtered, i)
			}
		case "k":
			title = i18n.QuickPotionTitle
			emptyMessage = i18n.NoPotionsInBackpack
			if it.Type == entities.ItemTypePotion {
				filtered = append(filtered, i)
			}
		case "e":
			title = i18n.QuickScrollTitle
			emptyMessage = i18n.NoScrollsInBackpack
			if it.Type == entities.ItemTypeScroll {
				filtered = append(filtered, i)
			}
		}
	}
	return filtered, title, emptyMessage
}

// renderInventorySelection отображает экран выбора предмета из списка (raw‑mode).
func (a *ConsoleApp) renderInventorySelection(title string, indices []int) {
	for {
		fmt.Print("\033[H\033[2J")
		fmt.Println(title)
		fmt.Println()
		for idx, realIdx := range indices {
			fmt.Printf("%d. %s\n", idx+1, a.inventoryLineByIndex(realIdx))
		}
		if a.hasWeaponInHandsOrList(indices) {
			fmt.Println(i18n.UnequipWeapon)
		}
		fmt.Println()
		fmt.Println(i18n.ChooseItemNumber)
		fmt.Println(i18n.InventoryScreenHint)

		key, err := a.readControlKey()
		if err != nil {
			return
		}
		if key == 'q' || key == 0x1b || key == '\n' {
			return
		}
		if key == '0' && a.hasWeaponInHandsOrList(indices) {
			cc := gameplay.NewCharacterController(a.Game.Player, a.Game)
			if !cc.UnequipWeapon() {
				a.renderMessageScreen(i18n.BackpackFull)
			}
			return
		}
		if key < '1' || key > '9' {
			continue
		}
		choice := int(key - '1')
		if choice < 0 || choice >= len(indices) {
			continue
		}
		cc := gameplay.NewCharacterController(a.Game.Player, a.Game)
		_ = cc.UseItem(indices[choice])
		return
	}
}

// renderInventorySelectionLineMode отображает экран выбора предмета в line‑mode.
func (a *ConsoleApp) renderInventorySelectionLineMode(title string, indices []int) {
	fmt.Println(title)
	for idx, realIdx := range indices {
		fmt.Printf("%d. %s\n", idx+1, a.inventoryLineByIndex(realIdx))
	}
	if a.hasWeaponInHandsOrList(indices) {
		fmt.Println(i18n.UnequipWeapon)
	}
	fmt.Println(i18n.ChooseItemNumber + " (1..9, q — выход)")
	line, _ := a.reader.ReadString('\n')
	line = strings.TrimSpace(strings.ToLower(line))
	if line == "" || line == "q" {
		return
	}
	if line == "0" && a.hasWeaponInHandsOrList(indices) {
		cc := gameplay.NewCharacterController(a.Game.Player, a.Game)
		if !cc.UnequipWeapon() {
			fmt.Println(i18n.BackpackFull)
		}
		return
	}
	if len(line) != 1 || line[0] < '1' || line[0] > '9' {
		return
	}
	choice := int(line[0] - '1')
	if choice < 0 || choice >= len(indices) {
		return
	}
	cc := gameplay.NewCharacterController(a.Game.Player, a.Game)
	_ = cc.UseItem(indices[choice])
}

// renderBackpackScreen отображает полный экран рюкзака с разбивкой по категориям (raw‑mode).
func (a *ConsoleApp) renderBackpackScreen() {
	for {
		fmt.Print("\033[H\033[2J")
		fmt.Println(i18n.InventoryTitle)
		fmt.Println()

		entries := a.printBackpackSections()
		fmt.Println()
		fmt.Println(i18n.InventoryControl1)
		fmt.Println(i18n.InventoryControl2)
		fmt.Println(i18n.InventoryControl3)
		fmt.Println(i18n.InventoryControl4)
		if a.Game.Player.CurrentWeapon != nil {
			fmt.Printf("\nОружие в руках: %s\n", formatItemRu(a.Game.Player.CurrentWeapon))
		}

		key, err := a.readControlKey()
		if err != nil {
			return
		}
		if key == 'q' || key == 0x1b || key == '\n' {
			return
		}
		if key == '0' {
			cc := gameplay.NewCharacterController(a.Game.Player, a.Game)
			if !cc.UnequipWeapon() {
				a.renderMessageScreen(i18n.BackpackFull)
			}
			continue
		}
		if key < '1' || key > '9' {
			continue
		}
		choice := int(key - '1')
		if choice < 0 || choice >= len(entries) {
			continue
		}
		cc := gameplay.NewCharacterController(a.Game.Player, a.Game)
		_ = cc.UseItem(entries[choice])
	}
}

// renderBackpackScreenLineMode отображает экран рюкзака в line‑mode.
func (a *ConsoleApp) renderBackpackScreenLineMode() {
	fmt.Println(i18n.InventoryTitle)
	entries := a.printBackpackSections()
	fmt.Println(i18n.InventoryControl2)
	fmt.Println(i18n.InventoryControl3)
	fmt.Println("- q — закрыть рюкзак")
	line, _ := a.reader.ReadString('\n')
	line = strings.TrimSpace(strings.ToLower(line))
	if line == "" || line == "q" {
		return
	}
	if line == "0" {
		cc := gameplay.NewCharacterController(a.Game.Player, a.Game)
		if !cc.UnequipWeapon() {
			fmt.Println(i18n.BackpackFull)
		}
		return
	}
	if len(line) != 1 || line[0] < '1' || line[0] > '9' {
		return
	}
	choice := int(line[0] - '1')
	if choice < 0 || choice >= len(entries) {
		return
	}
	cc := gameplay.NewCharacterController(a.Game.Player, a.Game)
	_ = cc.UseItem(entries[choice])
}

// printBackpackSections выводит содержимое рюкзака, сгруппированное по типам предметов.
// Возвращает срез индексов предметов в порядке их отображения.
func (a *ConsoleApp) printBackpackSections() []int {
	entries := make([]int, 0, len(a.Game.Player.Backpack.Slots))
	number := 1
	sections := []struct {
		title string
		typee entities.ItemType
	}{
		{"Оружие:", entities.ItemTypeWeapon},
		{"Еда:", entities.ItemTypeFood},
		{"Эликсиры:", entities.ItemTypePotion},
		{"Свитки:", entities.ItemTypeScroll},
	}
	for _, sec := range sections {
		fmt.Println(sec.title)
		has := false
		for i, it := range a.Game.Player.Backpack.Slots {
			if it.Type != sec.typee {
				continue
			}
			has = true
			fmt.Printf("%d. %s\n", number, a.inventoryLineByIndex(i))
			entries = append(entries, i)
			number++
		}
		if !has {
			fmt.Println(i18n.InventoryEmptyLine)
		}
		fmt.Println()
	}
	if a.Game.Player.CurrentWeapon != nil {
		fmt.Println(i18n.UnequipWeapon)
	}
	return entries
}

// hasWeaponInHandsOrList проверяет, есть ли у игрока экипированное оружие или оружие в указанном списке.
func (a *ConsoleApp) hasWeaponInHandsOrList(indices []int) bool {
	if a.Game.Player.CurrentWeapon != nil {
		return true
	}
	for _, idx := range indices {
		if a.Game.Player.Backpack.Slots[idx].Type == entities.ItemTypeWeapon {
			return true
		}
	}
	return false
}

// inventoryLineByIndex возвращает отформатированную строку для предмета по его индексу в рюкзаке.
func (a *ConsoleApp) inventoryLineByIndex(index int) string {
	item := a.Game.Player.Backpack.Slots[index]
	line := formatItemRu(item)
	if item.Type == entities.ItemTypeWeapon && a.Game.Player.CurrentWeapon == item {
		line += " " + i18n.EquippedMarker
	}
	return line
}