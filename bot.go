package main

import (
	"bufio"
	"database/sql"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"net/textproto"
	"os"
	"strings"
	"time"
)

type Bot struct {
	server     string
	port       string
	nick       string
	channel    string
	conn       net.Conn
	mods       map[string]bool
	lastmsg    int64
	maxMsgTime int64

	dbconn *sql.DB
}

func NewBot() *Bot {
	return &Bot{
		server:  "irc.twitch.tv",
		port:    "6667",
		nick:    "YogiiBot", //Change to your Twitch username
		channel: "penutty_", //Change to your channel
		conn:    nil,        //Don't change this
		mods:    make(map[string]bool),
	}
}

func (bot *Bot) Connect() {
	var err error
	fmt.Printf("Attempting to connect to server...\n")
	bot.conn, err = net.Dial("tcp", bot.server+":"+bot.port)
	if err != nil {
		fmt.Printf("Unable to connect to Twitch IRC server! Reconnecting in 10 seconds...\n")
		time.Sleep(10 * time.Second)
		bot.Connect()
	}
	fmt.Printf("Connected to IRC server %s\n", bot.server)
}

func (bot *Bot) Message(message string) {
	if message == "" {
		return
	}
	if bot.lastmsg+bot.maxMsgTime <= time.Now().Unix() {
		fmt.Printf("Bot: " + message + "\n")
		fmt.Fprintf(bot.conn, "PRIVMSG "+bot.channel+" :"+message+"\r\n")
		bot.lastmsg = time.Now().Unix()
	} else {
		fmt.Println("Attempted to spam message")
	}
}

func (bot *Bot) ConsoleInput() {
	reader := bufio.NewReader(os.Stdin)
	for {
		text, _ := reader.ReadString('\n')
		if text == "/quit" {
			bot.conn.Close()
			os.Exit(0)
		}
		if text != "" {
			bot.Message(text)
		}
	}
}

func main() {
	channel := flag.String("channel", "penutty_", "Sets the channel for the bot to go into.")
	nick := flag.String("nickname", "yogiibot", "The username of the bot.")
	flag.Parse()
	fmt.Printf("Twitch IRC Bot made in Go! https://github.com/Vaultpls/Twitch-IRC-Bot\n")

	ircbot := NewBot()
	go ircbot.ConsoleInput()
	ircbot.Connect()
	messagesCount := 0

	pass1, err := ioutil.ReadFile("twitch_pass.txt")
	pass := strings.Replace(string(pass1), "\n", "", 0)
	if err != nil {
		fmt.Println("Error reading from twitch_pass.txt.  Maybe it isn't created?")
		os.Exit(1)
	}

	//Prep everything
	if !ircbot.readSettingsDB(*channel) {
		ircbot.nick = *nick
		ircbot.channel = "#" + *channel
		ircbot.writeSettingsDB()
	}
	ircbot.OpenNuttyDB()
	defer ircbot.CloseNuttyDB()
	//

	fmt.Fprintf(ircbot.conn, "CAP REQ :twitch.tv/membership")
	fmt.Fprintf(ircbot.conn, "USER %s 8 * :%s\r\n", ircbot.nick, ircbot.nick)
	fmt.Fprintf(ircbot.conn, "PASS %s\r\n", pass)
	fmt.Fprintf(ircbot.conn, "NICK %s\r\n", ircbot.nick)
	fmt.Fprintf(ircbot.conn, "JOIN %s\r\n", ircbot.channel)
	fmt.Printf("Inserted information to server...\n")
	fmt.Printf("If you don't see the stream chat it probably means the Twitch oAuth password is wrong\n")
	fmt.Printf("Channel: " + ircbot.channel + "\n")
	defer ircbot.conn.Close()
	reader := bufio.NewReader(ircbot.conn)
	tp := textproto.NewReader(reader)
	go ircbot.ConsoleInput()
	for {
		line, err := tp.ReadLine()
		if err != nil {
			break
		}
		fmt.Println(line)
		if strings.Contains(line, ".tmi.twitch.tv PRIVMSG "+ircbot.channel) {
			messagesCount++
			userdata := strings.Split(line, ".tmi.twitch.tv PRIVMSG "+ircbot.channel)
			username := strings.Split(userdata[0], "@")
			usermessage := strings.Replace(userdata[1], " :", "", 1)
			fmt.Printf(username[1] + ": " + usermessage + "\n")
			go ircbot.CmdInterpreter(username[1], usermessage)

		} else if strings.Contains(line, ".tmi.twitch.tv JOIN "+ircbot.channel) {
			userjoindata := strings.Split(line, ".tmi.twitch.tv JOIN "+ircbot.channel)
			userjoined := strings.Split(userjoindata[0], "@")
			if !ircbot.UserExists(userjoined[1]) {
				ircbot.CreateUser(userjoined[1])
			}
		}
	}

}
