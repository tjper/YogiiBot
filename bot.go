package main

import (
	"bufio"
	"fmt"
	"github.com/penutty/YogiiBot/commands"
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
	lastmsg    int64
	maxMsgTime int64
}

func NewBot() *Bot {
	return &Bot{
		server:  "irc.twitch.tv",
		port:    "6667",
		nick:    "YogiiBot", //Change to your Twitch username
		channel: "penutty",  //Change to your channel
		conn:    nil,        //Don't change this
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
	channel := "penutty"
	nick := "yogiibot"

	ircbot := NewBot()
	go ircbot.ConsoleInput()
	ircbot.Connect()

	pass1, err := ioutil.ReadFile("twitch_pass.txt")
	pass := strings.Replace(string(pass1), "\n", "", 0)
	if err != nil {
		fmt.Println("Error reading from twitch_pass.txt.  Maybe it isn't created?")
		os.Exit(1)
	}

	if !ircbot.readSettingsDB(channel) {
		ircbot.nick = nick
		ircbot.channel = "#" + channel
		ircbot.writeSettingsDB()
	}

	fmt.Fprintf(ircbot.conn, "CAP REQ :twitch.tv/commands\n")
	fmt.Fprintf(ircbot.conn, "CAP REQ :twitch.tv/tags\n")
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
		if strings.Contains(line, "PING") {
			pongdata := strings.Split(line, "PING ")
			fmt.Fprintf(ircbot.conn, "PONG %s\r\n", pongdata[1])
		} else if strings.Contains(line, ".tmi.twitch.tv PRIVMSG "+ircbot.channel) {

			command, err := commands.NewCommand(ircbot, line)
			if err != nil {
				continue
			}
			go command.Exec()
		}
	}
}

func (bot *Bot) readSettingsDB(channel string) bool {
	settings, err := ioutil.ReadFile("settings#" + channel + ".ini")
	bot.channel = "#" + channel
	if err != nil {
		fmt.Println("Unable to read SettingsDB from " + channel)
		return false
	}
	split1 := strings.Split(string(settings), "\n")
	for _, splitted1 := range split1 {
		split2 := strings.Split(splitted1, "|")
		if split2[0] == "nickname" {
			bot.nick = split2[1]
		}

	}
	return true
}

func (bot *Bot) writeSettingsDB() {
	dst, err := os.Create("settings" + bot.channel + ".ini")
	defer dst.Close()
	if err != nil {
		fmt.Println("Can't write to SettingsDB from " + bot.channel)
		return
	}
	fmt.Fprintf(dst, "nickname|"+bot.nick+"\n")
}
