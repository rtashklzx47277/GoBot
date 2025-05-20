package tools

import (
	"database/sql"
	"fmt"
	"net"
	"os"
	"regexp"
	"time"

	"github.com/bwmarrin/discordgo"
)

func Regexp(str, substr string) (string, bool) {
	match := regexp.MustCompile(substr).FindAllStringSubmatch(str, 1)

	if len(match) == 0 {
		return "", false
	}

	return match[0][1], true
}

func IsContain(list []string, target string) bool {
	for _, element := range list {
		if element == target {
			return true
		}
	}

	return false
}

func CheckHealth(db *sql.DB) error {
	conn, err := net.DialTimeout("tcp", "8.8.8.8:53", 1*time.Second)
	if err != nil {
		return fmt.Errorf("failed to connect to external network (8.8.8.8:53) after retries: %v", err)
	}
	defer conn.Close()

	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping DB after retries: %v", err)
	}

	return nil
}

func AddCommands(s *discordgo.Session, check bool) error {
	if !check {
		return nil
	}

	appID, guildID := os.Getenv("DISCORD_APP_ID"), os.Getenv("DISCORD_GUILD_ID")

	registeredCommands, err := s.ApplicationCommands(appID, guildID)
	if err != nil {
		return err
	}

	registeredMap := make(map[string]*discordgo.ApplicationCommand)
	for _, command := range registeredCommands {
		registeredMap[command.Name] = command
	}

	for _, command := range commands {
		if existingCmd, exists := registeredMap[command.Name]; exists {
			if existingCmd.Description != command.Description {
				_, err := s.ApplicationCommandEdit(appID, guildID, existingCmd.ID, command)
				if err != nil {
					return err
				}
			}
		} else {
			_, err := s.ApplicationCommandCreate(appID, guildID, command)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
