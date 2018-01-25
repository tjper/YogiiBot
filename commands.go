// commands
package main

import (
	"database/sql"
	"errors"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	nuts     = regexp.MustCompile(`^(!nuts)$`)
	thanks   = regexp.MustCompile(`^(\!thanks)(\s){1}([a-zA-Z0-9_]){4,25}$`)
	getnutty = regexp.MustCompile(`^(\!getnutty)$`)

	duel       = regexp.MustCompile(`^(\!duel)$`)
	duelwinner = regexp.MustCompile(`^(\!duelwinner)(\s){1}(penutty|opponent){1}$`)
	duelcancel = regexp.MustCompile(`^(\!duelcancel)$`)
	penutty    = regexp.MustCompile(`^(\!penutty)$`)
	opponent   = regexp.MustCompile(`^(\!opponent)$`)

	findYogi = regexp.MustCompile(`^(\!findyogi)(\s){1}([a-zA-z]){5}$`)
)

func yogiibot_error(username string) string {
	return fmt.Sprintf("@%s - YogiiBot is being a bad dog. Let @penutty_ know and he'll spend some more time training him.", username)
}

func (bot *Bot) CmdInterpreter(m map[string]string, usermessage string) {
	message := strings.ToLower(usermessage)

	switch {
	case duel.MatchString(message):
		bot.Duel(m)
	case duelwinner.MatchString(message):
		bot.DuelWinner(m, message)
	case duelcancel.MatchString(message):
		bot.DuelCancel(m)
	case nuts.MatchString(message):
		bot.Nuts(m)
	case penutty.MatchString(message):
		bot.Penutty(m)
	case opponent.MatchString(message):
		bot.Opponent(m)
	case thanks.MatchString(message):
		bot.Thanks(m, message)
	case getnutty.MatchString(message):
		bot.GetNutty(m)
	case findYogi.MatchString(message):
		bot.FindYogi(m, message)
	default:
		bot.Default(m)
	}
}

func (bot *Bot) Duel(m map[string]string) {
	if !bot.isMod(m) && !bot.isBroadcaster(m) {
		return
	}
	if bot.duelOpen {
		return
	}

	bot.duel = make(map[string][]Vote)
	bot.duelOpen = true
	bot.duelStart = time.Now()
	bot.Message("DUEL START")
	bot.Message("!penutty - bet on penutty")
	bot.Message("!opponent - bet on the opponent")
}

func (bot *Bot) DuelWinner(m map[string]string, message string) {
	if !bot.duelOpen {
		return
	}
	if !bot.isMod(m) && !bot.isBroadcaster(m) {
		return
	}

	a := strings.Split(message, "!duelwinner ")
	winner := a[1]
	bot.duelOpen = false
	duelLength := time.Since(bot.duelStart).Seconds()

	var nuts, winners float64
	for _, a := range bot.duel {
		nuts += float64(len(a))
	}
	winners = float64(len(bot.duel[winner]))
	minPoints := 1.00
	maxPoints := float64(math.Ceil(nuts / winners))

	for _, v := range bot.duel[winner] {
		points := (v.dt.Sub(bot.duelStart).Seconds() / duelLength) * maxPoints
		if points < minPoints {
			points = minPoints
		}
		if err := bot.AddNuts(v.userID, points, points-1); err != nil {
			fmt.Printf("Error: %s", err)
			return
		}
	}

	bot.duel = make(map[string][]Vote)
	bot.votees = make(map[int]bool)
	bot.Message(fmt.Sprintf("%s has won!", winner))
}

func (bot *Bot) DuelCancel(m map[string]string) {
	if !bot.duelOpen {
		return
	}
	if !bot.isMod(m) && !bot.isBroadcaster(m) {
		return
	}
	for _, v := range bot.duel["penutty"] {
		if err := bot.AddNuts(v.userID, 1.00, 0.00); err != nil {
			fmt.Printf("Error: %s", err)
			return
		}
	}
	for _, v := range bot.duel["opponent"] {
		if err := bot.AddNuts(v.userID, 1.00, 0.00); err != nil {
			fmt.Printf("Error: %s", err)
			return
		}

	}
	bot.duel = make(map[string][]Vote)
	bot.votees = make(map[int]bool)
	bot.duelOpen = false
	bot.Message("DUEL CANCELLED")
}

func (bot *Bot) Nuts(m map[string]string) {
	userID, userName, err := getIdentifiers(m)
	if err != nil {
		fmt.Printf("Error: %s", err)
		return
	}

	if !bot.isNutty(userID, userName) {
		return
	}

	nuts, err := bot.SelectNuts(userID)
	if err != nil {
		bot.Message(yogiibot_error(userName))
		return
	}
	bot.Message(fmt.Sprintf("@%s - %s", userName, nuts))
}

func (bot *Bot) Penutty(m map[string]string) {
	userID, userName, err := getIdentifiers(m)
	if err != nil {
		fmt.Printf("Error: %s", err)
		return
	}

	if !bot.duelOpen {
		return
	}
	if !bot.isNutty(userID, userName) {
		return
	}
	_, ok := bot.votees[userID]
	if ok {
		return
	}

	vote := Vote{userID, time.Now()}
	bot.duel["penutty"] = append(bot.duel["penutty"], vote)
	if err := bot.RemoveNuts(userID, 1.00); err != nil {
		fmt.Printf("Error: %s", err)
	}
	bot.votees[userID] = true
}

func (bot *Bot) Opponent(m map[string]string) {
	userID, userName, err := getIdentifiers(m)
	if err != nil {
		fmt.Printf("Error: %s", err)
		return
	}

	if !bot.duelOpen {
		return
	}
	if !bot.isNutty(userID, userName) {
		return
	}
	_, ok := bot.votees[userID]
	if ok {
		return
	}

	vote := Vote{userID, time.Now()}
	bot.duel["opponent"] = append(bot.duel["opponent"], vote)
	if err := bot.RemoveNuts(userID, 1.00); err != nil {
		fmt.Printf("Error: %s", err)
	}
	bot.votees[userID] = true
}

func (bot *Bot) Thanks(m map[string]string, message string) {
	userID, userName, err := getIdentifiers(m)
	if err != nil {
		fmt.Printf("Error: %s", err)
		return
	}

	reward := 5.00
	a := strings.Split(message, "!thanks ")
	referencedByUserName := a[1]

	referencedByUserID, err := bot.SelectUserID(referencedByUserName)
	if err != nil && err != sql.ErrNoRows {
		bot.Message(yogiibot_error(userName))
		return
	}

	switch {
	case !bot.isNutty(userID, userName):
		return
	case userName == referencedByUserName:
		bot.Message(fmt.Sprintf("Rufffff! You can't thank yourself @%s!", userName))
		return
	case !bot.isNutty(referencedByUserName, userName):
		bot.Message(fmt.Sprintf("Sorry @%s, I don't know @%s.", userName, referencedByUserName))
		return
	}

	ok, err := bot.ReferenceExists(userID, referencedByUserID)
	switch {
	case err != nil && err != sql.ErrNoRows:
		bot.Message(yogiibot_error(userName))
		return
	case ok:
		bot.Message(fmt.Sprintf("@%s you already thanked @%s!", userName, referencedByUserName))
		return
	}

	if err := bot.CreateReference(userID, referencedByUserID); err != nil && err != sql.ErrNoRows {
		bot.Message(yogiibot_error(userName))
		return
	}
	if err := bot.AddNuts(referencedByUserID, reward, reward); err != nil {
		bot.Message(yogiibot_error(userName))
		return
	}

	bot.Message(fmt.Sprintf("Thanks @%s, for recommending penutty_ to @%s! You've earned %v nuts!", referencedByUserName, userName, reward))

}

func (bot *Bot) GetNutty(m map[string]string) {
	userID, userName, err := getIdentifiers(m)
	if err != nil {
		fmt.Printf("Error: %s", err)
		return
	}

	if bot.isNutty(userID, userName) {
		bot.Message(fmt.Sprintf("You're already Nuttyyyyy @%s!", userName))
		return
	}
	if err := bot.CreateUser(userName, userID); err != nil {
		bot.Message(yogiibot_error(userName))
		return
	}
	bot.Message(fmt.Sprintf("Rufffff! Welcome @%s! Scroll down to the info section to see what commands you can use!", userName))

}

func (bot *Bot) FindYogi(m map[string]string, message string) {
	userID, userName, err := getIdentifiers(m)
	if err != nil {
		fmt.Printf("Error: %s", err)
		return
	}

	reward := 3.00
	a := strings.Split(message, "!findyogi ")
	hash := a[1]

	fmt.Printf("hash = %s\n", hash)
	found, ok := bot.yogihashs[hash]
	if found || !ok {
		return
	}

	bot.yogihashs[hash] = true
	if err := bot.AddNuts(userID, reward, reward); err != nil {
		fmt.Errorf("Error: %s\n", err)
		return
	}
	bot.Message(fmt.Sprintf("@%s found Yogi!!! You've been rewarded %v nuts.", userName, reward))
}

func (bot *Bot) Default(m map[string]string) {
	userID, userName, err := getIdentifiers(m)
	if err != nil {
		fmt.Printf("Error: %s", err)
		return
	}
	lastMsg, ok := bot.lastMsg[userID]
	if ok && time.Since(lastMsg) <= 10*time.Minute {
		return
	}

	reward := 1.00
	if !bot.isNutty(userID, userName) {
		return
	}
	if err := bot.AddNuts(userID, reward, reward); err != nil {
		bot.Message(yogiibot_error(userName))
		return
	}

	bot.lastMsg[userID] = time.Now()
}

var ErrorInvalidIdentifiers = errors.New("Invalid user identifiers.")

func getIdentifiers(m map[string]string) (userID int, userName string, err error) {
	tuserID, ok := m["user-id"]
	if !ok {
		return 0, "", ErrorInvalidIdentifiers
	}
	userID, err = strconv.Atoi(tuserID)
	if err != nil {
		return 0, "", err
	}

	userName, ok = m["display-name"]
	if !ok {
		return 0, "", ErrorInvalidIdentifiers
	}

	return userID, userName, nil
}

func (bot *Bot) isNutty(i interface{}, username string) bool {
	var ok bool
	var err error

	switch v := i.(type) {
	case int:
		ok, err = bot.UserIDExists(v)
	case string:
		ok, err = bot.UserNameExists(v)
	}

	if err != nil && err != sql.ErrNoRows {
		fmt.Printf("Error: %s", err)
		bot.Message(yogiibot_error(username))
	}
	if !ok {
		bot.Message(fmt.Sprintf("@%s - !getnutty in order to start earning nuts and give the YogiiBot commands!", username))
	}
	return ok
}

func (bot *Bot) isMod(m map[string]string) bool {
	if m["mod"] != "1" {
		return false
	}
	return true
}
func (bot *Bot) isBroadcaster(m map[string]string) bool {
	if m["broadcaster"] != "1" {
		return false
	}
	return true
}
