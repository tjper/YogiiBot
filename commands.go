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
	points   = regexp.MustCompile(`^(!points)$`)
	thanks   = regexp.MustCompile(`^(\!thanks)(\s){1}([a-zA-Z0-9_]){4,25}$`)
	getnutty = regexp.MustCompile(`^(\!getnutty)$`)

	findYogi = regexp.MustCompile(`^(\!findyogi)(\s){1}([a-zA-z]){5}$`)

	win                = regexp.MustCompile(`^(\!win)(\s){1}([0-9]){0,3}(\.)?([0-9]){0,2}$`)
	lose               = regexp.MustCompile(`^(\!lose)(\s){1}([0-9]){0,3}(\.)?([0-9]){0,2}$`)
	fortniteBet        = regexp.MustCompile(`^(\!fortnitebet)$`)
	fortniteEndBet     = regexp.MustCompile(`^(\!fortniteendbet)$`)
	fortniteResolveBet = regexp.MustCompile(`^(\!fortniteresolvebet)(\s){1}(win|lose){1}$`)
	fortniteCancelBet  = regexp.MustCompile(`^(\!fortnitecancelbet)$`)

	trivia = regexp.MustCompile(`^(\!trivia)(\s){1}(.)+$`)

	leaderboard = regexp.MustCompile(`^(\!leaderboard)$`)

	redeemduo = regexp.MustCompile(`^(\!redeem)(\s){1}(duo)$`)
	duoqueue  = regexp.MustCompile(`^(\!duoqueue)$`)
	duoremove = regexp.MustCompile(`^(\!duoremove)$`)
	duocharge = regexp.MustCompile(`^(\!duocharge)$`)
	duoopen   = regexp.MustCompile(`^(\!duoopen)$`)
	duoclose  = regexp.MustCompile(`^(\!duoclose)$`)

	quote    = regexp.MustCompile(`^(\!quote)(\s){1}("){1}(.){1,254}("){1}$`)
	getQuote = regexp.MustCompile(`^(\!)([a-zA-Z0-9_]){4,25}$`)

	redeemvbucks = regexp.MustCompile(`^(\!redeem)(\s){1}(vbucks)$`)
)

func (bot *Bot) CmdInterpreter(m map[string]string, message string) {
	u, err := bot.NewUser(m)
	if err != nil {
		fmt.Printf("Error - CmdInterpreter: %s", err)
	}
	if !bot.isNutty(u) {
		bot.GetNutty(u)
	}

	switch {
	case win.MatchString(message):
		bot.Win(u, message)
	case lose.MatchString(message):
		bot.Lose(u, message)
	case fortniteBet.MatchString(message):
		bot.FortniteBet(u)
	case fortniteEndBet.MatchString(message):
		bot.FortniteEndBet(u)
	case fortniteCancelBet.MatchString(message):
		bot.FortniteCancelBet(u)
	case fortniteResolveBet.MatchString(message):
		bot.FortniteResolveBet(u, message)
	case points.MatchString(message):
		bot.Points(u)
	case nuts.MatchString(message):
		bot.Nuts(u)
	case thanks.MatchString(message):
		bot.Thanks(u, message)
	case findYogi.MatchString(message):
		bot.FindYogi(u, message)
	case leaderboard.MatchString(message):
		bot.LeaderBoard(u)
	case redeemduo.MatchString(message):
		bot.RedeemDuo(u)
	case duoqueue.MatchString(message):
		bot.DuoQueue()
	case duoremove.MatchString(message):
		bot.DuoRemove(u)
	case duocharge.MatchString(message):
		bot.DuoCharge(u)
	case duoopen.MatchString(message):
		bot.DuoOpen(u)
	case duoclose.MatchString(message):
		bot.DuoClose(u)
	case redeemvbucks.MatchString(message):
		bot.RedeemVBucks(u)
	case quote.MatchString(message):
		bot.Quote(u, message)
	case getQuote.MatchString(message):
		bot.GetQuote(u, message)
	default:
		bot.Default(u)
	}
}

type User struct {
	Id            int
	Name          string
	IsSubscriber  bool
	IsMod         bool
	IsBroadcaster bool
}

var ErrorInvalidIdentifiers = errors.New("Invalid user identifiers.")

func (bot *Bot) NewUser(m map[string]string) (u *User, err error) {

	fmt.Printf("\n\n%v\n\n", m)
	u = new(User)
	tuserID, ok := m["user-id"]
	if !ok {
		return u, ErrorInvalidIdentifiers
	}
	u.Id, err = strconv.Atoi(tuserID)
	if err != nil {
		return u, err
	}

	u.Name, ok = m["display-name"]
	if !ok {
		return u, ErrorInvalidIdentifiers
	}

	u.IsSubscriber = false
	u.IsMod = false
	u.IsBroadcaster = false

	if m["mod"] == "1" {
		u.IsMod = true
	}
	if m["broadcaster"] == "1" {
		u.IsBroadcaster = true
	}
	if m["subscriber"] == "1" {
		u.IsSubscriber = true
	}

	if u.IsSubscriber {
		registered, err := bot.SelectSubStatus(u.Id)
		if err != nil {
			return u, err
		}
		if registered {
			return u, nil
		}
		if err := bot.UpdateSubStatus(u.Id); err != nil {
			return u, err
		}
		if err := bot.AddNuts(u.Id, 1.0); err != nil {
			return u, err
		}
	}

	return u, nil
}

func (bot *Bot) GetQuote(u *User, message string) {
	message  = strings.ToLower(message)
	a := strings.Split(message, "!")
	author := a[1]

	quote, err := bot.SelectQuote(author)
	if err != nil {
		fmt.Printf("Error - GetQuote: %s\n", err)
		return
	}

	bot.Message(fmt.Sprintf("\" %s \" - %s", quote, author))
	return
}

func (bot *Bot) Quote(u *User, message string) {
	if !u.IsSubscriber {
		return
	}

	a := strings.Split(message, "!quote ")
	quote := a[1]
	quote = strings.Trim(quote, "\"")

	if err := bot.UpdateQuote(u.Id, quote); err != nil {
		fmt.Printf("Error - Quote: %s\n", err)
		return
	}
	return
}

var (
	DuoCost       = 1.00
	TypeDuo       = 1
	DuoQueueLimit = 3
)

func (bot *Bot) RedeemDuo(u *User) {
	if !bot.duoopen {
		return
	}
	if len(bot.duoqueue) >= DuoQueueLimit {
		fmt.Print("DuoQueueLimit Reached.")
		return
	}
	for _, r := range bot.duoqueue {
		if r.Name == u.Name {
			return
		}
	}

	nuts, err := bot.SelectNuts(u.Id)
	if err != nil {
		return
	}
	if nuts < DuoCost {
		return
	}
	bot.Message(fmt.Sprintf("@%s has redeemed DUOS!", u.Name))
	bot.duoqueue = append(bot.duoqueue, u)
}

func (bot *Bot) DuoCharge(u *User) {
	if !u.IsMod && !u.IsBroadcaster {
		return
	}
	if len(bot.duoqueue) <= 0 {
		fmt.Println("DuoQueue Empty")
		return
	}

	r := bot.duoqueue[0]

	if err := bot.InsertRedeem(r.Id, TypeDuo, DuoCost); err != nil {
		return
	}
	if err := bot.RemoveNuts(r.Id, DuoCost); err != nil {
		fmt.Printf("RemoveNuts - Error: %s\n", err)
		return
	}
	bot.Message(fmt.Sprintf("@%s has been charged for DUOS!", r.Name))
	bot.duoqueue = bot.duoqueue[1:]
}

func (bot *Bot) DuoRemove(u *User) {
	if !u.IsMod && !u.IsBroadcaster {
		return
	}
	if len(bot.duoqueue) <= 0 {
		fmt.Println("DuoQueue Empty")
		return
	}
	bot.duoqueue = bot.duoqueue[1:]
}

func (bot *Bot) DuoQueue() {
	var msg string
	for i, u := range bot.duoqueue {
		msg = msg + fmt.Sprintf("%v. %s   ", i+1, u.Name)
	}
	bot.Message(msg)
}

func (bot *Bot) DuoOpen(u *User) {
	if !u.IsMod && !u.IsBroadcaster {
		return
	}
	bot.duoopen = true
	bot.Message(fmt.Sprintf("DUOS is now open! type \"!redeem duo\" to play with penutty."))
}

func (bot *Bot) DuoClose(u *User) {
	if !u.IsMod && !u.IsBroadcaster {
		return
	}
	bot.duoopen = false
	bot.Message(fmt.Sprintf("DUOS is now closed."))
}

func (bot *Bot) RedeemVBucks(u *User) {
	cost := 8.00
	vbucks := 2
	nuts, err := bot.SelectNuts(u.Id)
	if err != nil {
		fmt.Printf("SelectNuts - Error: %s\n", err)
		return
	}
	if nuts < cost {
		return
	}
	if err := bot.InsertRedeem(u.Id, vbucks, cost); err != nil {
		return
	}
	if err := bot.RemoveNuts(u.Id, cost); err != nil {
		fmt.Printf("RemoveNuts - Error: %s\n", err)
		return
	}
	bot.Message(fmt.Sprintf("@%s has redeemed VBUCKS!", u.Name))
}

func (bot *Bot) LeaderBoard(u *User) {

	set, err := bot.SelectTopUsersByNuts()
	if err != nil {
		fmt.Printf("Error - LeaderBoard: %s", err)
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

func (bot *Bot) Points(u *User) {
	points, err := bot.SelectChatPoints(u.Id)
	if err != nil {
		return
	}
	bot.Message(fmt.Sprintf("@%s - Chat Points = %v", u.Name, points))
}

func (bot *Bot) Nuts(u *User) {
	nuts, err := bot.SelectNuts(u.Id)
	if err != nil {
		return
	}
	bot.Message(fmt.Sprintf("@%s - Nuts = %v", u.Name, nuts))
}

func (bot *Bot) Thanks(u *User, message string) {
	message  = strings.ToLower(message)
	a := strings.Split(message, "!thanks ")
	referencedByUserName := a[1]

	referencedByUserID, err := bot.SelectUserID(referencedByUserName)
	if err != nil && err != sql.ErrNoRows {
		return
	}

	nuts, err := bot.SelectNuts(u.Id)
	if err != nil {
		return
	}

	switch {
	case nuts < 0.5:
		return
	case !u.IsSubscriber:
		return
	case u.Id == referencedByUserID:
		fmt.Printf("\n%s- attempted to thank themself.", u.Name)
		return
	case !bot.isNutty(&User{Id: referencedByUserID, Name: referencedByUserName}):
		fmt.Printf("\n%s - attempted to thank someone who hasn't ran !getnutty.", u.Name, referencedByUserName)
		return
	}

	ok, err := bot.ReferenceExists(u.Id)
	switch {
	case err != nil && err != sql.ErrNoRows:
		return
	case ok:
		fmt.Printf("@%s you already thanked another user!", u.Name, referencedByUserName)
		return
	}

	if err := bot.CreateReference(u.Id, referencedByUserID); err != nil && err != sql.ErrNoRows {
		return
	}
	reward := 0.25
	if err := bot.AddNuts(referencedByUserID, reward); err != nil {
		return
	}

	bot.Message(fmt.Sprintf("Thanks @%s, for recommending penutty_ to @%s! You've earned %v nut!", referencedByUserName, u.Name, reward))

}

func (bot *Bot) GetNutty(u *User) {
	if err := bot.CreateUser(u.Name, u.Id); err != nil {
		return
	}
	bot.Message(fmt.Sprintf("/w %s Rufffff! Welcome to penutty's channel %s! type !how in the channel chat to see how to earn !nuts. Nuts can be redeemed for VBUCKS, playing duos with penutty, and more!", u.Name, u.Name))
}

func (bot *Bot) FindYogi(u *User, message string) {
	message  = strings.ToLower(message)
	a := strings.Split(message, "!findyogi ")
	hash := a[1]

	found, ok := bot.yogihashs[hash]
	if found || !ok {
		return
	}

	bot.yogihashs[hash] = true
	reward := 0.1
	if err := bot.AddNuts(u.Id, reward); err != nil {
		fmt.Errorf("Error: %s\n", err)
		return
	}
	bot.Message(fmt.Sprintf("@%s found Yogi!!! You've been rewarded %v chat points.", u.Name, reward))
}

func (bot *Bot) Default(u *User) {
	lastMsg, ok := bot.lastMsg[u.Id]
	if ok && time.Since(lastMsg) <= 10*time.Minute {
		return
	}

	reward := 0.005
	if err := bot.AddNuts(u.Id, reward); err != nil {
		return
	}

	bot.lastMsg[u.Id] = time.Now()
}

func (bot *Bot) isNutty(u *User) bool {
	userName, err := bot.SelectUserName(u.Id)
	if err  == sql.ErrNoRows {
		return false
	}
	if err != nil {
		fmt.Printf("\nis Nutty - Error: %s", err)
	}
	if userName == u.Name {
		return true
	}
	if err = bot.UpdateUserName(u.Id, u.Name); err != nil {
		fmt.Printf("\nisNutty - update: %s", err)
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

func (bot *Bot) FortniteBet(u *User) {
	if !u.IsMod && !u.IsBroadcaster {
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

func (bot *Bot) FortniteEndBet(u *User) {
	if !u.IsMod && !u.IsBroadcaster {
		return
	}
	bot.bet.open = false
	bot.Message("BETTING ENDS")
}

func (bot *Bot) FortniteResolveBet(u *User, message string) {
	if bot.bet == nil {
		return
	}
	if bot.bet.open {
		return
	}
	if !u.IsMod && !u.IsBroadcaster {
		return
	}
	if len(bot.bet.winBets) == 0 || len(bot.bet.loseBets) == 0 {
		return
	}
	message  = strings.ToLower(message)
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
		reward := totalDebits * (b.amount / totalProfits)
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

func (bot *Bot) FortniteCancelBet(u *User) {
	if bot.bet == nil {
		return
	}
	if !u.IsMod && !u.IsBroadcaster {
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

func (bot *Bot) Win(u *User, message string) {
	message  = strings.ToLower(message)
	if bot.bet == nil {
		return
	}
	if !bot.bet.open {
		return
	}

	if _, ok := bot.bet.betees[u.Id]; ok {
		return
	}

	a := strings.Split(message, "!win ")
	amount, err := strconv.ParseFloat(a[1], 64)
	if err != nil {
		fmt.Printf("\nWin - Error: %s", err)
	}

	nuts, err := bot.SelectNuts(u.Id)
	if err != nil {
		fmt.Printf("\nLose - Error: %s", err)
		return
	}
	if nuts < amount {
		return
	}

	b := &bet{u.Id, amount}
	bot.bet.winBets = append(bot.bet.winBets, b)
	bot.bet.totalWinBets += amount
	bot.bet.betees[u.Id] = true
	bot.Message(fmt.Sprintf("@%s bet %v nuts on penutty winning!", u.Name, amount))
}

func (bot *Bot) Lose(u *User, message string) {
	message  = strings.ToLower(message)
	if bot.bet == nil {
		return
	}
	if !bot.bet.open {
		return
	}

	if _, ok := bot.bet.betees[u.Id]; ok {
		return
	}

	a := strings.Split(message, "!lose ")
	amount, err := strconv.ParseFloat(a[1], 64)
	if err != nil {
		fmt.Printf("\nLose - Error: %s", err)
		return
	}

	nuts, err := bot.SelectNuts(u.Id)
	if err != nil {
		fmt.Printf("\nLose - Error: %s", err)
		return
	}
	if nuts < amount {
		return
	}

	b := &bet{u.Id, amount}
	bot.bet.loseBets = append(bot.bet.loseBets, b)
	bot.bet.totalLoseBets += amount
	bot.bet.betees[u.Id] = true
	bot.Message(fmt.Sprintf("@%s bet %v nuts on penutty losing.", u.Name, amount))

}
