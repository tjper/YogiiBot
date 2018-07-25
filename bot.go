package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"html"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"net/textproto"
	"os"
	"strings"
	"sync"
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

	lastMsg   map[int]time.Time
	duel      map[string][]Vote
	duelOpen  bool
	duelStart time.Time
	votees    map[int]bool

	yogihashs map[string]bool

	bet *betRound

	triviaquestion TriviaQuestion

	duoqueue []*User
	duoopen  bool

	dba *DatabaseAccess
}

type Vote struct {
	userID int
	dt     time.Time
}

type TriviaQuestion struct {
	Question         string
	Answer           string
	IncorrectAnswers []string
}

func NewBot() *Bot {
	return &Bot{
		server:    "irc.twitch.tv",
		port:      "6667",
		nick:      "YogiiBot", //Change to your Twitch username
		channel:   "penutty",  //Change to your channel
		conn:      nil,        //Don't change this
		lastMsg:   make(map[int]time.Time),
		duel:      make(map[string][]Vote),
		votees:    make(map[int]bool),
		yogihashs: make(map[string]bool),
		duoqueue:  make([]*User, 0),
		duoopen:   false,
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

func (bot *Bot) TriviaQuestion() {
	var wg sync.WaitGroup
	for {
		wg.Add(1)
		go func() {
			defer wg.Done()
			n := (rand.Int() % 20) + 10
			time.Sleep(time.Duration(n) * time.Minute)
			r, err := http.Get("https://opentdb.com/api.php?amount=1&category=15&type=multiple")
			if err != nil {
				fmt.Printf("TriviaQuestion - Error: %s\n", err)
				return
			}

			type result struct {
				Category          string
				Type              string
				Difficulty        string
				Incorrect_Answers []string
				Question          string
				Correct_Answer    string
			}
			type body struct {
				Response_code int
				Results       []result
			}

			b := new(body)
			if err = json.NewDecoder(r.Body).Decode(b); err != nil {
				fmt.Printf("TriviaQuestion - Error: %s\n", err)
				return
			}
			bot.triviaquestion = TriviaQuestion{
				Question: html.UnescapeString(b.Results[0].Question),
				Answer:   html.UnescapeString(b.Results[0].Correct_Answer),
			}
			for _, a := range b.Results[0].Incorrect_Answers {
				bot.triviaquestion.IncorrectAnswers = append(bot.triviaquestion.IncorrectAnswers, html.UnescapeString(a))
			}
		}()
		wg.Wait()
	}
}

func (bot *Bot) WildYogi() {
	var wg sync.WaitGroup
	for {
		wg.Add(1)
		go func() {
			defer wg.Done()
			n := rand.Int() % 150
			time.Sleep(time.Duration(n) * time.Minute)
			r := RandomString(5)
			bot.yogihashs[r] = false
			bot.Message(fmt.Sprintf("A wild yogi has appeared! Who will catch him first? Type !findyogi %s", r))
		}()
		wg.Wait()
	}
}

const letterBytes = "abcdefghijklmnopqrstuvwxyz"

func RandomString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
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

	//Prep everything
	if !ircbot.readSettingsDB(channel) {
		ircbot.nick = nick
		ircbot.channel = "#" + channel
		ircbot.writeSettingsDB()
	}
	ircbot.dba, err = NewDatabaseAccess()
	if err != nil {
		return
	}

	defer ircbot.dba.CloseNuttyDB()
	go ircbot.WildYogi()
	go ircbot.TriviaQuestion()

	//
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

			command, err := NewCommand(ircbot, line)
			if err != nil {
				continue
			}
			go command.Exec(ircbot)
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
