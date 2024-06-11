package tools

import (
	"fmt"
	"log"
	"os"
	"runtime/debug"
	"time"

	"github.com/bwmarrin/discordgo"
)

var logChannel = os.Getenv("DISCORD_LOG_CHANNEL_ID")

func DiscordNotify(s *discordgo.Session, class, user string) {
	message := fmt.Sprintf("<t:%d:f> ***%s %s*** notify failed!", int(time.Now().Unix()), user, class)
	s.ChannelMessageSend(logChannel, message)
}

func ErrorRecord(err any) {
	originalFlags := log.Flags()
	log.Printf("%v\n%s", err, debug.Stack())
	log.SetFlags(0)
	log.Println("====================================================================================================")
	log.SetFlags(originalFlags)
}
