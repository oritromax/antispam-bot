package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/icza/gox/stringsx"
)

var (
	Token          string
	GuildID        = flag.String("guild", "", "Test guild ID. If not passed - bot registers commands globally")
	RemoveCommands = flag.Bool("rmcmd", true, "Remove all commands after shutdowning or not")
	spamWords      []string
)

var (
	command = []*discordgo.ApplicationCommand{
		{
			Name:        "stats",
			Type:        discordgo.ChatApplicationCommand,
			Description: "Current Bot Status",
		},
	}

	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"stats": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Online Since: , Total Filter Size: " + strconv.Itoa(len(spamWords)) + " Words",
				},
			})
		},
	}
)

func init() {

	flag.StringVar(&Token, "t", "OTkyMTk0NTEzMTg5NzQ0Nzgx.GRHYva.iDwOPIsoYJvGZaEBGCpKiIR8Sx3aYhCV52n43o", "Bot Token")
	flag.Parse()

}

func main() {

	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	jsonFile, err := ioutil.ReadFile("words.json")

	log.Println("Loading JSON Word List")

	err = json.Unmarshal([]byte(jsonFile), &spamWords)

	if err != nil {
		fmt.Println(err)
	}

	log.Println("JSON file list Loaded")

	dg.AddHandler(messageCreate)

	dg.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})

	dg.Identify.Intents = discordgo.IntentsGuildMessages

	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}
	// Wait here until CTRL-C or other term signal is received.
	log.Println("Bot is now running.  Press CTRL-C to exit.")
	log.Println("Command Adding")
	registeredCommands := make([]*discordgo.ApplicationCommand, len(command))

	for i, v := range command {
		cmd, err := dg.ApplicationCommandCreate(dg.State.User.ID, *GuildID, v)
		if err != nil {
			log.Panicf("Cannot create '%v' command: %v", v.Name, err)
		}
		registeredCommands[i] = cmd
	}
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	if *RemoveCommands {
		log.Println("Removing commands...")

		for _, v := range registeredCommands {
			err := dg.ApplicationCommandDelete(dg.State.User.ID, *GuildID, v.ID)
			if err != nil {
				log.Panicf("Cannot delete '%v' command: %v", v.Name, err)
			}
		}
	}
	// Cleanly close down the Discord session.
	dg.Close()
}

func check(userMsg string) bool {
	for _, v := range spamWords {
		if strings.Contains(userMsg, v) {
			return true
		}
	}

	return false
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	if m.Author.ID == s.State.User.ID {
		return
	}

	if check(stringsx.Clean(m.Content)) {
		s.ChannelMessageDelete(m.ChannelID, m.ID)
		s.ChannelMessageSend(m.ChannelID, "Message by <@!"+m.Author.ID+"> was deleted because of SPAM Protection")
		log.Println("Sensored Word: " + m.Content)
	}
	// if m.Content == "ping" {

	// }
	if m.Content == "!spambot" {
		s.ChannelMessageSend(m.ChannelID, "Spam Protection Bot by MCM Studio")
	}
}
