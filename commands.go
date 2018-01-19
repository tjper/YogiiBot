// commands
package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

func (bot *Bot) CmdInterpreter(username string, usermessage string) {
	message := strings.ToLower(usermessage)
	yogiibot_error := fmt.Sprintf("@%s - YogiiBot is being a bad dog. Let @penutty_ know and he'll spend some more time training him.", username)

	var m string
	ok, err := bot.UserExists(username)
	if err != nil && err != sql.ErrNoRows {
		bot.Message(yogiibot_error)
		return
	}
	if ok {
		switch {
		case strings.HasPrefix(message, "!nuts"):

			nuts, err := bot.SelectNuts(username)
			if err != nil {
				bot.Message(yogiibot_error)
				return
			}
			m = fmt.Sprintf("@%s - %s", username, nuts)

		case strings.HasPrefix(message, "!thanks"):
			referencedByUserName := strings.Split(message, "!thanks ")
			ok, err := bot.UserExists(referencedByUserName[1])
			if err != nil && err != sql.ErrNoRows {
				bot.Message(yogiibot_error)
				return
			}

			if ok {
				if err = bot.CreateReference(username, referencedByUserName[1]); err != nil {
					bot.Message(yogiibot_error)
					break
				}
				if err = bot.AddNuts(referencedByUserName[1], 5); err != nil {
					bot.Message(yogiibot_error)
					break
				}

				m = fmt.Sprintf("Thanks %s, for recommending penutty_ to %s!", referencedByUserName[1], username)
			} else {
				m = fmt.Sprintf("Sorry %s, I've never smelled %s before!", username, referencedByUserName[1])
			}
		}

	} else {
		switch {
		case strings.HasPrefix(message, "!getnutty"):
			if err = bot.CreateUser(username); err != nil {
				bot.Message(yogiibot_error)
				break
			}
			m = fmt.Sprintf("Rufffff! Welcome to penutty_'s channel @%s!", username)
		default:
			m = fmt.Sprintf("@%s - type !getnutty to start giving commands to Yogiibot! He's a good dog.", username)
		}
	}

	bot.Message(m)

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
