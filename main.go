package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
)

func clearScreen() {
	switch runtime.GOOS {
	case "linux", "darwin":
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		cmd.Run()
	case "windows":
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
}

func printHeader() {
	fmt.Println("\n=== Discord Server Cloner ===")
	fmt.Println("Author: Kyurenoxd")
	fmt.Println("GitHub: github.com/Kyurenoxd")
	fmt.Println("Website: kyureno.dev")
	fmt.Println("===========================")
}

func getUserInput(prompt string) string {
	var input string
	fmt.Print(prompt)
	fmt.Scanln(&input)
	return input
}

func getChoice(choices []string) int {
	fmt.Println("\nВыберите действие:")
	for i, choice := range choices {
		fmt.Printf("%d. %s\n", i+1, choice)
	}

	var choice int
	fmt.Print("\nВаш выбор (1-", len(choices), "): ")
	fmt.Scanln(&choice)
	return choice - 1
}

func main() {
	var cloner *ServerCloner

	for {
		clearScreen()
		printHeader()

		if cloner == nil {
			token := getUserInput("Введите полученный Authorization ID: ")
			cloner = NewServerCloner()
			if err := cloner.Login(token); err != nil {
				log.Printf("[!] Ошибка при входе: %v\n", err)
				choices := []string{"Попробовать снова", "Выйти"}
				if choice := getChoice(choices); choice == 1 {
					fmt.Println("\nПрограмма завершена.")
					break
				}
				continue
			}
		}

		var sourceID, targetID string
		if cloner.lastSourceID != "" && cloner.lastTargetID != "" {
			choices := []string{
				fmt.Sprintf("Использовать последние ID (Исходный: %s, Целевой: %s)",
					cloner.lastSourceID, cloner.lastTargetID),
				"Ввести новые ID",
			}
			if choice := getChoice(choices); choice == 0 {
				sourceID = cloner.lastSourceID
				targetID = cloner.lastTargetID
			}
		}

		if sourceID == "" {
			sourceID = getUserInput("Введите ID исходного сервера: ")
		}
		if targetID == "" {
			targetID = getUserInput("Введите ID целевого сервера: ")
		}

		fmt.Println("\nНачинаем процесс клонирования...")

		// Сохраняем ID для следующего использования
		cloner.lastSourceID = sourceID
		cloner.lastTargetID = targetID

		if err := cloner.CloneServer(sourceID, targetID); err != nil {
			log.Printf("[!] Ошибка при клонировании: %v\n", err)
			choices := []string{"Попробовать снова", "Выйти"}
			if choice := getChoice(choices); choice == 1 {
				fmt.Println("\nПрограмма завершена.")
				break
			}
			continue
		}

		fmt.Println("\n=== Клонирование завершено ===")
		choices := []string{"Клонировать другой сервер", "Выйти"}
		if choice := getChoice(choices); choice == 1 {
			fmt.Println("\nСпасибо за использование ServerSync!")
			break
		}
	}
}
