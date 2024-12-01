package main

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

type ServerCloner struct {
	session      *discordgo.Session
	stats        CloneStats
	lastSourceID string
	lastTargetID string
}

func NewServerCloner() *ServerCloner {
	return &ServerCloner{
		stats: CloneStats{},
	}
}

func (c *ServerCloner) Login(token string) error {
	session, err := discordgo.New(token)
	if err != nil {
		return fmt.Errorf("ошибка создания сессии: %w", err)
	}

	c.session = session
	if err := c.session.Open(); err != nil {
		return fmt.Errorf("ошибка подключения к Discord: %w", err)
	}

	fmt.Println("[+] Успешный вход в аккаунт")
	return nil
}

func (c *ServerCloner) CloneServer(sourceID, targetID string) error {
	sourceGuild, err := c.session.Guild(sourceID)
	if err != nil {
		return fmt.Errorf("ошибка получения исходного сервера: %w", err)
	}

	targetGuild, err := c.session.Guild(targetID)
	if err != nil {
		return fmt.Errorf("ошибка получения целевого сервера: %w", err)
	}

	fmt.Println("\n=== Начало клонирования ===")
	fmt.Printf("[+] Исходный сервер: %s\n", sourceGuild.Name)
	fmt.Printf("[+] Целевой сервер: %s\n\n", targetGuild.Name)

	if err := c.copyGuildSettings(sourceGuild, targetGuild); err != nil {
		return err
	}

	fmt.Println("\n[+] Очистка целевого сервера...")
	if err := c.deleteAllRoles(targetGuild); err != nil {
		return err
	}
	if err := c.deleteAllChannels(targetGuild); err != nil {
		return err
	}

	fmt.Println("\n[+] Копирование ролей...")
	if err := c.copyRoles(sourceGuild, targetGuild); err != nil {
		return err
	}

	fmt.Println("\n[+] Копирование каналов...")
	if err := c.copyChannels(sourceGuild, targetGuild); err != nil {
		return err
	}

	fmt.Println("\n[+] Копирование эмодзи и стикеров...")
	if err := c.copyEmojisAndStickers(sourceGuild, targetGuild); err != nil {
		return err
	}

	c.printStats()
	return nil
}

func (c *ServerCloner) printStats() {
	fmt.Println("\n=== Статистика клонирования ===")
	fmt.Printf("[+] Скопировано ролей: %d\n", c.stats.Roles)
	fmt.Printf("[+] Создано категорий: %d\n", c.stats.Categories)
	fmt.Printf("[+] Создано текстовых каналов: %d\n", c.stats.TextChannels)
	fmt.Printf("[+] Создано голосовых каналов: %d\n", c.stats.VoiceChannels)
	fmt.Printf("[+] Скопировано эмодзи: %d\n", c.stats.Emojis)
	fmt.Println("==============================")
}
