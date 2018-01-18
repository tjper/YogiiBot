// commands
package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

func (bot *Bot) CmdInterpreter(username string, usermessage string) {
	message := strings.ToLower(usermessage)
	tempstr := strings.Split(message, " ")

	for _, str := range tempstr {
		if strings.HasPrefix(str, "https://") || strings.HasPrefix(str, "http://") {
			go bot.Message("^ " + webTitle(str))
		} else if isWebsite(str) {
			go bot.Message("^ " + webTitle("http://"+str))
		}
	}

	switch {
	case strings.HasPrefix(message, "!help"):
		bot.Message("For help on the bot please go to http://commandanddemand.com/bot.html")
	}

}

//Website stuff
func webTitle(website string) string {
	response, err := http.Get(website)
	if err != nil {
		return "Error reading website"
	} else {
		defer response.Body.Close()
		contents, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return "Error reading website"
		}
		if strings.Contains(string(contents), "<title>") && strings.Contains(string(contents), "</title>") {
			derp := strings.Split(string(contents), "<title>")
			derpz := strings.Split(derp[1], "</title>")
			return derpz[0]
		}
		return "No title"
	}
}

func isWebsite(website string) bool {
	domains := []string{".com", ".net", ".org", ".info", ".fm", ".gg", ".tv"}
	for _, domain := range domains {
		if strings.Contains(website, domain) {
			return true
		}
	}
	return false
}

//End website stuff

//Mod stuff
func (bot *Bot) isMod(username string) bool {
	temp := strings.Replace(bot.channel, "#", "", 1)
	if bot.mods[username] == true || temp == username || username == "vaultpls" {
		return true
	}
	return false
}

func (bot *Bot) timeout(username string, reason string) {
	if bot.isMod(username) {
		return
	}
	fmt.Fprintf(bot.conn, "PRIVMSG "+bot.channel+" :/timeout "+username+"\r\n")
	bot.Message(username + " was timed out(" + reason + ")!")
}

func (bot *Bot) ban(username string, reason string) {
	if bot.isMod(username) {
		return
	}
	fmt.Fprintf(bot.conn, "PRIVMSG "+bot.channel+" :/ban "+username+"\r\n")
	bot.Message(username + " was banned(" + reason + ")!")
}

//End mod stuff
