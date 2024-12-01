package main

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"

	"github.com/bwmarrin/discordgo"
)

func (c *ServerCloner) copyGuildSettings(source, target *discordgo.Guild) error {
	// Скачиваем иконку сервера
	var iconData []byte
	if source.Icon != "" {
		iconURL := source.IconURL("512") // Получаем URL иконки в высоком качестве
		resp, err := http.Get(iconURL)
		if err != nil {
			fmt.Printf("[!] Ошибка при скачивании иконки: %v\n", err)
		} else {
			defer resp.Body.Close()
			iconData, err = ioutil.ReadAll(resp.Body)
			if err != nil {
				fmt.Printf("[!] Ошибка при чтении иконки: %v\n", err)
			}
		}
	}

	// Создаем base64 строку из иконки
	var iconB64 string
	if len(iconData) > 0 {
		iconB64 = "data:image/png;base64," + base64.StdEncoding.EncodeToString(iconData)
	}

	guildParams := &discordgo.GuildParams{
		Name: source.Name,
		Icon: iconB64,
	}

	_, err := c.session.GuildEdit(target.ID, guildParams)
	if err != nil {
		return fmt.Errorf("ошибка обновления настроек сервера: %w", err)
	}

	fmt.Println("[+] Настройки сервера скопированы")
	if iconB64 != "" {
		fmt.Println("[+] Аватар сервера скопирован")
	}
	return nil
}

func (c *ServerCloner) deleteAllRoles(guild *discordgo.Guild) error {
	for _, role := range guild.Roles {
		if role.Name != "@everyone" {
			if err := c.session.GuildRoleDelete(guild.ID, role.ID); err != nil {
				fmt.Printf("[!] Ошибка при удалении роли %s: %v\n", role.Name, err)
				continue
			}
			fmt.Printf("[-] Удалена роль: %s\n", role.Name)
		}
	}
	return nil
}

func (c *ServerCloner) copyRoles(source, target *discordgo.Guild) error {
	// Получаем все роли и сортируем их по позиции (сверху вниз)
	roles := make([]*discordgo.Role, len(source.Roles))
	copy(roles, source.Roles)
	sort.Slice(roles, func(i, j int) bool {
		return roles[i].Position > roles[j].Position
	})

	// Создаем мапу для хранения соответствия старых и новых ID ролей
	roleMap := make(map[string]string)

	// Копируем роли
	for _, role := range roles {
		if role.Name != "@everyone" {
			perms := role.Permissions
			roleParams := &discordgo.RoleParams{
				Name:        role.Name,
				Color:       &role.Color,
				Hoist:       &role.Hoist,
				Permissions: &perms,
				Mentionable: &role.Mentionable,
			}

			newRole, err := c.session.GuildRoleCreate(target.ID, roleParams)
			if err != nil {
				fmt.Printf("[!] Ошибка при создании роли %s: %v\n", role.Name, err)
				continue
			}

			// Сохраняем соответствие ID ролей
			roleMap[role.ID] = newRole.ID

			// Обновляем роль с теми же настройками
			_, err = c.session.GuildRoleEdit(target.ID, newRole.ID, roleParams)
			if err != nil {
				fmt.Printf("[!] Ошибка при обновлении роли %s: %v\n", role.Name, err)
			}

			c.stats.Roles++
			fmt.Printf("[+] Создана роль: %s (ID: %s) с правами: %d\n",
				newRole.Name, newRole.ID, perms)
		}
	}

	return nil
}

func (c *ServerCloner) copyChannels(source, target *discordgo.Guild) error {
	// Получаем все каналы сервера
	channels, err := c.session.GuildChannels(source.ID)
	if err != nil {
		return fmt.Errorf("ошибка получения каналов: %w", err)
	}

	// Сначала создаем категории
	categoryMap := make(map[string]string) // старый ID -> новый ID

	// Сортируем каналы, чтобы категории создавались первыми
	for _, channel := range channels {
		if channel.Type == discordgo.ChannelTypeGuildCategory {
			// Копируем разрешения канала
			permissionOverwrites := make([]*discordgo.PermissionOverwrite, len(channel.PermissionOverwrites))
			for i, perm := range channel.PermissionOverwrites {
				permissionOverwrites[i] = &discordgo.PermissionOverwrite{
					ID:    perm.ID,
					Type:  perm.Type,
					Allow: perm.Allow,
					Deny:  perm.Deny,
				}
			}

			channelCreate := &discordgo.GuildChannelCreateData{
				Name:                 channel.Name,
				Type:                 channel.Type,
				PermissionOverwrites: permissionOverwrites,
				Position:             channel.Position,
			}

			created, err := c.session.GuildChannelCreateComplex(target.ID, *channelCreate)
			if err != nil {
				fmt.Printf("[!] Ошибка при создании категории %s: %v\n", channel.Name, err)
				continue
			}
			categoryMap[channel.ID] = created.ID
			c.stats.Categories++
			fmt.Printf("[+] Создана категория: %s (с %d разрешениями)\n",
				channel.Name, len(permissionOverwrites))
		}
	}

	// Затем создаем остальные каналы
	for _, channel := range channels {
		if channel.Type != discordgo.ChannelTypeGuildCategory {
			// Копируем разрешения канала
			permissionOverwrites := make([]*discordgo.PermissionOverwrite, len(channel.PermissionOverwrites))
			for i, perm := range channel.PermissionOverwrites {
				permissionOverwrites[i] = &discordgo.PermissionOverwrite{
					ID:    perm.ID,
					Type:  perm.Type,
					Allow: perm.Allow,
					Deny:  perm.Deny,
				}
			}

			channelCreate := &discordgo.GuildChannelCreateData{
				Name:                 channel.Name,
				Type:                 channel.Type,
				Topic:                channel.Topic,
				Bitrate:              channel.Bitrate,
				UserLimit:            channel.UserLimit,
				RateLimitPerUser:     channel.RateLimitPerUser,
				Position:             channel.Position,
				PermissionOverwrites: permissionOverwrites,
				ParentID:             "",
				NSFW:                 channel.NSFW,
			}

			// Если у канала есть родительская категория
			if channel.ParentID != "" {
				if newParentID, ok := categoryMap[channel.ParentID]; ok {
					channelCreate.ParentID = newParentID
				}
			}

			newChannel, err := c.session.GuildChannelCreateComplex(target.ID, *channelCreate)
			if err != nil {
				fmt.Printf("[!] Ошибка при создании канала %s: %v\n", channel.Name, err)
				continue
			}

			switch channel.Type {
			case discordgo.ChannelTypeGuildText:
				c.stats.TextChannels++
				fmt.Printf("[+] Создан текстовый канал: %s (ID: %s) с %d разрешениями\n",
					newChannel.Name, newChannel.ID, len(permissionOverwrites))
			case discordgo.ChannelTypeGuildVoice:
				c.stats.VoiceChannels++
				fmt.Printf("[+] Создан голосовой канал: %s (ID: %s) с %d разрешениями\n",
					newChannel.Name, newChannel.ID, len(permissionOverwrites))
			}
		}
	}

	return nil
}

func (c *ServerCloner) deleteAllChannels(guild *discordgo.Guild) error {
	// Получаем актуальный список каналов
	channels, err := c.session.GuildChannels(guild.ID)
	if err != nil {
		return fmt.Errorf("ошибка получения каналов: %w", err)
	}

	// Сначала удаляем все каналы, кроме категорий
	for _, channel := range channels {
		if channel.Type != discordgo.ChannelTypeGuildCategory {
			_, err := c.session.ChannelDelete(channel.ID)
			if err != nil {
				fmt.Printf("[!] Ошибка при удалении канала %s: %v\n", channel.Name, err)
				continue
			}
			fmt.Printf("[-] Удален канал: %s\n", channel.Name)
		}
	}

	// Затем удаляем категории
	for _, channel := range channels {
		if channel.Type == discordgo.ChannelTypeGuildCategory {
			_, err := c.session.ChannelDelete(channel.ID)
			if err != nil {
				fmt.Printf("[!] Ошибка при удалении категории %s: %v\n", channel.Name, err)
				continue
			}
			fmt.Printf("[-] Удалена категория: %s\n", channel.Name)
		}
	}

	return nil
}
