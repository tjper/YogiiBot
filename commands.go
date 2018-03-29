// commands
package main

import (
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	nuts     = regexp.MustCompile(`^(!nuts)$`)
	points = regexp.MustCompile(`^(!points)$`)
	thanks   = regexp.MustCompile(`^(\!thanks)(\s){1}([a-zA-Z0-9_]){4,25}$`)
	getnutty = regexp.MustCompile(`^(\!getnutty)$`)

	findYogi = regexp.MustCompile(`^(\!findyogi)(\s){1}([a-zA-z]){5}$`)

	win                = regexp.MustCompile(`^(\!win)(\s){1}([0-9]){0,3}(\.)?([0-9]){0,2}$`)
	lose               = regexp.MustCompile(`^(\!lose)(\s){1}([0-9]){0,3}(\.)?([0-9]){0,2}$`)
	fortniteBet        = regexp.MustCompile(`^(\!fortnitebet)$`)
	fortniteEndBet	   = regexp.MustCompile(`^(\!fortniteendbet)$`)
	fortniteResolveBet = regexp.MustCompile(`^(\!fortniteresolvebet)(\s){1}(win|lose){1}$`)
	fortniteCancelBet  = regexp.MustCompile(`^(\!fortnitecancelbet)$`)

	trivia = regexp.MustCompile(`^(\!trivia)(\s){1}(.)+$`)

	leaderboard = regexp.MustCompile(`^(\!leaderboard)$`)
	
	redemduo = regexp.MustCompile(`^(\!redem)(\s){1}(duo)$`)
	redemvbucks = regexp.MustCompile(`^(\!redem)(\s){1}(vbucks)$`)
)

func yogiibot_error(username string, err error) {
	fmt.Printf("\nyogiibot had failed with error: %s on user %s.", err, username)
	return
}

func (bot *Bot) CmdInterpreter(m map[string]string, usermessage string) {
	message := strings.ToLower(usermessage)

	switch {
	case win.MatchString(message):
		bot.Win(m, message)
	case lose.MatchString(message):
		bot.Lose(m, message)
	case fortniteBet.MatchString(message):
		bot.FortniteBet(m)
	case fortniteEndBet.MatchString(message):
		bot.FortniteEndBet(m)
	case fortniteCancelBet.MatchString(message):
		bot.FortniteCancelBet(m)
	case fortniteResolveBet.MatchString(message):
		bot.FortniteResolveBet(m, message)
	case points.MatchString(message):
		bot.Points(m)
	case nuts.MatchString(message):
		bot.Nuts(m)
	case thanks.MatchString(message):
		bot.Thanks(m, message)
	case getnutty.MatchString(message):
		bot.GetNutty(m)
	case findYogi.MatchString(message):
		bot.FindYogi(m, message)
//	case trivia.MatchString(message):
//		bot.Trivia(m, message)
	case leaderboard.MatchString(message):
		bot.LeaderBoard(m)
	case redemduo.MatchString(message):
		bot.RedemDuo(m)
	case redemvbucks.MatchString(message):
		bot.RedemVBucks(m)
	default:
		bot.Default(m)
	}
}

func (bot *Bot) UserNotice(m map[string]string) {
	userID, userName, err := getIdentifiers(m)
	if err != nil {
		fmt.Printf("Error: %s", err)
		return
	}
	if !bot.isNutty(userID, userName) {
		bot.GetNutty(m)
	}

	switch {
	case m["msg-id"] == "sub":
		if err := bot.AddNuts(userID, 1.00); err != nil {
			fmt.Printf("Error - UserNotice: %s", err)
			return
		}
	}
}

func (bot *Bot) RedemDuo(m map[string]string) {
	userID, userName, err := getIdentifiers(m)
	if err != nil {
		fmt.Printf("Error: %s", err)
		return
	}
	if !bot.isNutty(userID, userName) {
		bot.GetNutty(m)
		return
	}

	cost := 2.00
	duo := 1
	nuts, err := bot.SelectNuts(userID)
	if err != nil {
		yogiibot_error(userName, err)
		return
	}
	if nuts < cost {
		return
	}
	if err := bot.InsertRedem(userID, duo, cost); err != nil {
		return
	}
	if err := bot.RemoveNuts(userID, cost); err != nil {
		fmt.Printf("RemoveNuts - Error: %s\n", err)
		return
	}
	bot.Message(fmt.Sprintf("@%s has redeemed DUOS!", userName))
}

func (bot *Bot) RedemVBucks(m map[string]string) {
	userID, userName, err := getIdentifiers(m)
	if err != nil {
		fmt.Printf("Error: %s", err)
		return
	}
	if !bot.isNutty(userID, userName) {
		bot.GetNutty(m)
		return
	}
	cost := 8.00
	vbucks := 2
	nuts, err := bot.SelectNuts(userID)
	if err != nil {
		fmt.Printf("SelectNuts - Error: %s\n", err)
		return
	}
	if nuts < cost {
		return	
	}
	if err := bot.InsertRedem(userID, vbucks, cost); err != nil {
		return
	}
	if err := bot.RemoveNuts(userID, cost); err != nil {
		fmt.Printf("RemoveNuts - Error: %s\n", err)
		return
	}
	bot.Message(fmt.Sprintf("@%s has redeemed VBUCKS!", userName))
}

func (bot *Bot) LeaderBoard(m map[string]string) {
	userID, userName, err := getIdentifiers(m)
	if err != nil {
		fmt.Printf("Error: %s", err)
		return
	}

	if !bot.isNutty(userID, userName) {
		bot.GetNutty(m)	
	}

	set, err := bot.SelectTopUsersByNuts()
	if err != nil {
		fmt.Printf("Error: %s", err)
		return
	}
	
	var res string
	for i, row := range set {
		res = res + fmt.Sprintf("(%v) %s = %v nuts ... ", i+1, row.UserName, row.Nuts)
	}
	bot.Message(res)
}

//func (bot *Bot) Trivia(m map[string]string, message string) {
//	if bot.triviaquestion.Question == "" {
//		return
//	}
//
//	userID, userName, err := getIdentifiers(m)
//	if err != nil {
//		fmt.Printf("Error: %s", err)
//		return
//	}
//
//	if !bot.isNutty(userID, userName) {
//		bot.GetNutty(m)
//	}
//
//	a := strings.Split(message, "!trivia ")
//	answer := a[1]
//	if strings.ToLower(bot.triviaquestion.Answer) != answer {
//		return
//	}
//
//	reward := 10.0
//	if err = bot.AddNuts(userID, reward); err != nil {
//		fmt.Printf("Trivia - Error: %s\n", err)
//		return
//	}
//	bot.Message(fmt.Sprintf("@%s has answered the trivia question correctly! You've been rewarded %v nuts!", userName, reward))
//	bot.triviaquestion = TriviaQuestion{}
//}

func (bot *Bot) Points(m map[string]string) {
	userID, userName, err := getIdentifiers(m)
	if err != nil {
		fmt.Printf("Error: %s", err)
		return
	}

	if !bot.isNutty(userID, userName) {
		bot.GetNutty(m)	
	}

	points, err := bot.SelectChatPoints(userID)
	if err != nil {
		yogiibot_error(userName, err)
		return
	}
	bot.Message(fmt.Sprintf("@%s - Chat Points = %v", userName, points))
}

func (bot *Bot) Nuts(m map[string]string) {
	userID, userName, err := getIdentifiers(m)
	if err != nil {
		fmt.Printf("Error: %s", err)
		return
	}

	if !bot.isNutty(userID, userName) {
		bot.GetNutty(m)
	}

	nuts, err := bot.SelectNuts(userID)
	if err != nil {
		yogiibot_error(userName, err)
		return
	}
	bot.Message(fmt.Sprintf("@%s - Nuts = %v", userName, nuts))
}

func (bot *Bot) Thanks(m map[string]string, message string) {
	userID, userName, err := getIdentifiers(m)
	if err != nil {
		fmt.Printf("\nThanks - Error: %s", err)
		return
	}
	
	reward := 1.00
	a := strings.Split(message, "!thanks ")
	referencedByUserName := a[1]

	referencedByUserID, err := bot.SelectUserID(referencedByUserName)
	if err != nil && err != sql.ErrNoRows {
		yogiibot_error(userName, err)
		return
	}

	switch {
	case !bot.isNutty(userID, userName):
		bot.GetNutty(m)	
	case !bot.isSubscriber(m):
		return
	case userName == referencedByUserName:
		fmt.Printf("\n%s- attempted to thank themself.", userName)
		return
	case !bot.isNutty(referencedByUserID, userName):
		fmt.Printf("\n%s - attempted to thank someone who hasn't ran !getnutty.", userName, referencedByUserName)
		return
	}

	ok, err := bot.ReferenceExists(userID)
	switch {
	case err != nil && err != sql.ErrNoRows:
		yogiibot_error(userName, err)
		return
	case ok:
		fmt.Printf("@%s you already thanked another user!", userName, referencedByUserName)
		return
	}

	if err := bot.CreateReference(userID, referencedByUserID); err != nil && err != sql.ErrNoRows {
		yogiibot_error(userName, err)
		return
	}
	if err := bot.AddNuts(referencedByUserID, reward); err != nil {
		yogiibot_error(userName, err)
		return
	}

	bot.Message(fmt.Sprintf("Thanks @%s, for recommending penutty_ to @%s! You've earned %v nut!", referencedByUserName, userName, reward))

}

func (bot *Bot) GetNutty(m map[string]string) {
	userID, userName, err := getIdentifiers(m)
	if err != nil {
		fmt.Printf("\nGetNutty - Error: %s", err)
		return
	}

	if bot.isNutty(userID, userName) {
		fmt.Printf("\n%s - Is already nutty", userName)
		return
	}
	if err := bot.CreateUser(userName, userID); err != nil {
		yogiibot_error(userName, err)
		return
	}
	bot.Message(fmt.Sprintf("Rufffff! Welcome @%s! Scroll down to the info section to see what commands you can use!", userName))

}

func (bot *Bot) FindYogi(m map[string]string, message string) {
	userID, userName, err := getIdentifiers(m)
	if err != nil {
		fmt.Printf("FindYogi - Error: %s", err)
		return
	}

	if !bot.isNutty(userID, userName) {
		bot.GetNutty(m)	
	}

	reward := 3.00
	a := strings.Split(message, "!findyogi ")
	hash := a[1]

	found, ok := bot.yogihashs[hash]
	if found || !ok {
		return
	}

	bot.yogihashs[hash] = true
	if err := bot.AddChatPoints(userID, reward); err != nil {
		fmt.Errorf("Error: %s\n", err)
		return
	}
	bot.Message(fmt.Sprintf("@%s found Yogi!!! You've been rewarded %v chat points.", userName, reward))
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
		bot.GetNutty(m)	
	}
	if err := bot.AddChatPoints(userID, reward); err != nil {
		yogiibot_error(userName, err)
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

func (bot *Bot) isNutty(userID int, username string) bool {
	ok, err := bot.UserIDExists(userID)
	if err != nil && err != sql.ErrNoRows {
		fmt.Printf("\nisNutty - Error: %s", err)
		yogiibot_error(username, err)
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
func (bot *Bot) isSubscriber(m map[string]string) bool {
	if m["subscriber"] != "1" {
		return false
	}
	return true
}

type betRound struct {
	open          bool
	startTime     time.Time
	betees        map[int]bool
	totalWinBets  float64 
	totalLoseBets float64 
	loseBets      []*bet
	winBets       []*bet
}

type bet struct {
	userID int
	amount float64 
}

func (bot *Bot) FortniteBet(m map[string]string) {
	if !bot.isMod(m) && !bot.isBroadcaster(m) {
		return
	}

	br := &betRound{
		open:          true,
		startTime:     time.Now(),
		betees:        make(map[int]bool),
		totalWinBets:  0,
		totalLoseBets: 0,
		loseBets:      make([]*bet, 0),
		winBets:       make([]*bet, 0),
	}
	bot.bet = br
	bot.Message("BETTING BEGINS")
}

func (bot *Bot) FortniteEndBet(m map[string]string) {
	if !bot.isMod(m) && !bot.isBroadcaster(m) {
		return
	}
	bot.bet.open = false
	bot.Message("BETTING ENDS")
}

func (bot *Bot) FortniteResolveBet(m map[string]string, message string) {
	if bot.bet == nil {
		return
	}
	if bot.bet.open {
		return
	}
	if !bot.isMod(m) && !bot.isBroadcaster(m) {
		return
	}
	if len(bot.bet.winBets) == 0 || len(bot.bet.loseBets) == 0 {
		return
	}

	a := strings.Split(message, "!fortniteresolvebet ")
	result := a[1]

	var profitors, debitors []*bet
	var totalDebits, totalProfits float64 
	if result == "win" {
		profitors = bot.bet.winBets
		debitors = bot.bet.loseBets
		totalDebits = bot.bet.totalLoseBets
		totalProfits = bot.bet.totalWinBets
	} else {
		profitors = bot.bet.loseBets
		debitors = bot.bet.winBets
		totalDebits = bot.bet.totalWinBets
		totalProfits = bot.bet.totalLoseBets
	}

	for _, b := range debitors {
		if err := bot.RemoveNuts(b.userID, b.amount); err != nil {
			fmt.Printf("FortniteResolveBet - RemoveNuts - Error: %s\n", err)
			return
		}
	}

	for _, b := range profitors {
		reward := float64(totalDebits) * (b.amount / float64(totalProfits))
		if err := bot.AddNuts(b.userID, reward); err != nil {
			fmt.Printf("FortniteResolveBet - AddNuts - Error: %s\n", err)
			return
		}
	}
	bot.bet.open = false

	if result == "win" {
		bot.Message("PENUTTY_ WON!")
	} else {
		bot.Message("PENUTTY LOST.")
	}
}

func (bot *Bot) FortniteCancelBet(m map[string]string) {
	if bot.bet == nil {
		return
	}
	if !bot.isMod(m) && !bot.isBroadcaster(m) {
		return
	}

	br := &betRound{
		open:          false,
		totalWinBets:  0,
		totalLoseBets: 0,
		loseBets:      make([]*bet, 1),
		winBets:       make([]*bet, 1),
	}
	bot.bet = br
	bot.Message("BET CANCELLED.")
}

func (bot *Bot) Win(m map[string]string, message string) {
	if bot.bet == nil {
		return
	}
	if !bot.bet.open {
		return
	}
	userID, userName, err := getIdentifiers(m)
	if err != nil {
		fmt.Printf("\nWin - Error: %s", err)
		return
	}

	if _, ok := bot.bet.betees[userID]; ok {
		return
	}

	a := strings.Split(message, "!win ")
	amount, err := strconv.ParseFloat(a[1], 64)
	if err != nil {
		fmt.Printf("\nWin - Error: %s", err)
	}

	nuts, err := bot.SelectNuts(userID)
	if err != nil {
		fmt.Printf("\nLose - Error: %s", err)
		return
	}
	if nuts < amount {
		return
	}

	b := &bet{userID, amount}
	bot.bet.winBets = append(bot.bet.winBets, b)
	bot.bet.totalWinBets += amount
	bot.bet.betees[userID] = true
	bot.Message(fmt.Sprintf("@%s bet %v nuts on penutty winning!", userName, amount))
}

func (bot *Bot) Lose(m map[string]string, message string) {
	if bot.bet == nil {
		return
	}
	if !bot.bet.open {
		return
	}
	userID, userName, err := getIdentifiers(m)
	if err != nil {
		fmt.Printf("\nLose - Error: %s", err)
		return
	}
	fmt.Printf("\n\nIn Lose\n\n")

	if _, ok := bot.bet.betees[userID]; ok {
		return
	}

	a := strings.Split(message, "!lose ")
	amount, err := strconv.ParseFloat(a[1], 64)
	if err != nil {
		fmt.Printf("\nLose - Error: %s", err)
		return
	}

	nuts, err := bot.SelectNuts(userID)
	if err != nil {
		fmt.Printf("\nLose - Error: %s", err)
		return
	}
	if nuts < amount {
		return
	}

	b := &bet{userID, amount}
	bot.bet.loseBets = append(bot.bet.loseBets, b)
	bot.bet.totalLoseBets += amount
	bot.bet.betees[userID] = true
	bot.Message(fmt.Sprintf("@%s bet %v nuts on penutty losing.", userName, amount))

}

//func (bot *Bot) Duel(m map[string]string) {
//	if !bot.isMod(m) && !bot.isBroadcaster(m) {
//		return
//	}
//	if bot.duelOpen {
//		return
//	}
//
//	bot.duel = make(map[string][]Vote)
//	bot.duelOpen = true
//	bot.duelStart = time.Now()
//	bot.Message("DUEL START")
//	bot.Message("!penutty - bet on penutty")
//	bot.Message("!opponent - bet on the opponent")
//}
//
//func (bot *Bot) DuelWinner(m map[string]string, message string) {
//	if !bot.duelOpen {
//		return
//	}
//	if !bot.isMod(m) && !bot.isBroadcaster(m) {
//		return
//	}
//
//	a := strings.Split(message, "!duelwinner ")
//	winner := a[1]
//	bot.duelOpen = false
//	duelLength := time.Since(bot.duelStart).Seconds()
//
//	var nuts, winners float64
//	for _, a := range bot.duel {
//		nuts += float64(len(a))
//	}
//	winners = float64(len(bot.duel[winner]))
//	minPoints := 1.00
//	maxPoints := float64(math.Ceil(nuts / winners))
//
//	for _, v := range bot.duel[winner] {
//		points := (v.dt.Sub(bot.duelStart).Seconds() / duelLength) * maxPoints
//		if points < minPoints {
//			points = minPoints
//		}
//		if err := bot.AddNuts(v.userID, points, points-1); err != nil {
//			fmt.Printf("Error: %s", err)
//			return
//		}
//	}
//
//	bot.duel = make(map[string][]Vote)
//	bot.votees = make(map[int]bool)
//	bot.Message(fmt.Sprintf("%s has won!", winner))
//}
//
//func (bot *Bot) DuelCancel(m map[string]string) {
//	if !bot.duelOpen {
//		return
//	}
//	if !bot.isMod(m) && !bot.isBroadcaster(m) {
//		return
//	}
//	for _, v := range bot.duel["penutty"] {
//		if err := bot.AddNuts(v.userID, 1.00, 0.00); err != nil {
//			fmt.Printf("Error: %s", err)
//			return
//		}
//	}
//	for _, v := range bot.duel["opponent"] {
//		if err := bot.AddNuts(v.userID, 1.00, 0.00); err != nil {
//			fmt.Printf("Error: %s", err)
//			return
//		}
//
//	}
//	bot.duel = make(map[string][]Vote)
//	bot.votees = make(map[int]bool)
//	bot.duelOpen = false
//	bot.Message("DUEL CANCELLED")
//}
//

