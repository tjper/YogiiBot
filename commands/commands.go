package command

import (
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Command struct {
	Author  *User
	Message string
}

func NewCommand(bot *Bot, line string) (*Command, error) {
	m, err := lineToMap(line)
	if err != nil {
		return nil, err
	}

	command := new(Command)
	command.Author, err = NewUser(bot, m)
	if err != nil {
		return nil, err
	}
	message := strings.Split(line, fmt.Sprintf("PRIVMSG %s :", bot.channel))
	command.Message = message[1]

	return command, nil
}

type User struct {
	Id            int
	Name          string
	IsSubscriber  bool
	IsMod         bool
	IsBroadcaster bool
}

var ErrorInvalidIdentifiers = errors.New("Invalid user identifiers.")

func NewUser(bot *Bot, m map[string]string) (u *User, err error) {

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
		registered, err := bot.dba.SelectSubStatus(u.Id)
		if err != nil {
			return u, err
		}
		if registered {
			return u, nil
		}
		if err := bot.dba.UpdateSubStatus(u.Id); err != nil {
			return u, err
		}
		if err := bot.dba.AddNuts(u.Id, 1.0); err != nil {
			return u, err
		}
	}

	return u, nil
}

var (
	nutsRE     = regexp.MustCompile(`^(!nuts)$`)
	pointsRE   = regexp.MustCompile(`^(!points)$`)
	thanksRE   = regexp.MustCompile(`^(\!thanks)(\s){1}([a-zA-Z0-9_]){4,25}$`)
	getnuttyRE = regexp.MustCompile(`^(\!getnutty)$`)

	findYogiRE = regexp.MustCompile(`^(\!findyogi)(\s){1}([a-zA-z]){5}$`)

	winRE                = regexp.MustCompile(`^(\!win)(\s){1}([0-9]){0,3}(\.)?([0-9]){0,2}$`)
	loseRE               = regexp.MustCompile(`^(\!lose)(\s){1}([0-9]){0,3}(\.)?([0-9]){0,2}$`)
	fortniteBetRE        = regexp.MustCompile(`^(\!fortnitebet)$`)
	fortniteEndBetRE     = regexp.MustCompile(`^(\!fortniteendbet)$`)
	fortniteResolveBetRE = regexp.MustCompile(`^(\!fortniteresolvebet)(\s){1}(win|lose){1}$`)
	fortniteCancelBetRE  = regexp.MustCompile(`^(\!fortnitecancelbet)$`)

	triviaRE = regexp.MustCompile(`^(\!trivia)(\s){1}(.)+$`)

	leaderboardRE = regexp.MustCompile(`^(\!leaderboard)$`)

	redeemduoRE = regexp.MustCompile(`^(\!redeem)(\s){1}(duo)$`)
	duoqueueRE  = regexp.MustCompile(`^(\!duoqueue)$`)
	duoremoveRE = regexp.MustCompile(`^(\!duoremove)$`)
	duochargeRE = regexp.MustCompile(`^(\!duocharge)$`)
	duoopenRE   = regexp.MustCompile(`^(\!duoopen)$`)
	duocloseRE  = regexp.MustCompile(`^(\!duoclose)$`)

	quoteRE    = regexp.MustCompile(`^(\!quote)(\s){1}("){1}(.){1,254}("){1}$`)
	getQuoteRE = regexp.MustCompile(`^(\!)([a-zA-Z0-9_]){4,25}$`)

	redeemvbucksRE = regexp.MustCompile(`^(\!redeem)(\s){1}(vbucks)$`)
)

func (c *Command) Exec(bot *Bot) {
	if !isNutty(bot, c.Author) {
		getNutty(bot, c.Author)
	}

	switch {
	case winRE.MatchString(c.Message):
		win(bot, c.Author, c.Message)
	case loseRE.MatchString(c.Message):
		lose(bot, c.Author, c.Message)
	case fortniteBetRE.MatchString(c.Message):
		fortniteBet(bot, c.Author)
	case fortniteEndBetRE.MatchString(c.Message):
		fortniteEndBet(bot, c.Author)
	case fortniteCancelBetRE.MatchString(c.Message):
		fortniteCancelBet(bot, c.Author)
	case fortniteResolveBetRE.MatchString(c.Message):
		fortniteResolveBet(bot, c.Author, c.Message)
	case pointsRE.MatchString(c.Message):
		points(bot, c.Author)
	case nutsRE.MatchString(c.Message):
		nuts(bot, c.Author)
	case thanksRE.MatchString(c.Message):
		thanks(bot, c.Author, c.Message)
	case findYogiRE.MatchString(c.Message):
		findYogi(bot, c.Author, c.Message)
	case leaderboardRE.MatchString(c.Message):
		leaderBoard(bot, c.Author)
	case redeemduoRE.MatchString(c.Message):
		redeemDuo(bot, c.Author)
	case duoqueueRE.MatchString(c.Message):
		duoQueue(bot)
	case duoremoveRE.MatchString(c.Message):
		duoRemove(bot, c.Author)
	case duochargeRE.MatchString(c.Message):
		duoCharge(bot, c.Author)
	case duoopenRE.MatchString(c.Message):
		duoOpen(bot, c.Author)
	case duocloseRE.MatchString(c.Message):
		duoClose(bot, c.Author)
	case redeemvbucksRE.MatchString(c.Message):
		redeemVBucks(bot, c.Author)
	case quoteRE.MatchString(c.Message):
		quote(bot, c.Author, c.Message)
	case getQuoteRE.MatchString(c.Message):
		getQuote(bot, c.Author, c.Message)
	default:
		dne(bot, c.Author)
	}
}

func getQuote(in dba.Selecter, bot *Bot, u *User, message string) {
	message = strings.ToLower(message)
	a := strings.Split(message, "!")
	author := a[1]

	quote, err := in.SelectQuote(author)
	if err != nil {
		fmt.Printf("Error - GetQuote: %s\n", err)
		return
	}

	bot.Message(fmt.Sprintf("\" %s \" - %s", quote, author))
	return
}

func quote(in dba.Updater, u *User, message string) {
	if !u.IsSubscriber {
		return
	}

	a := strings.Split(message, "!quote ")
	quote := a[1]
	quote = strings.Trim(quote, "\"")

	if err := in.UpdateQuote(u.Id, quote); err != nil {
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

func redeemDuo(in dba.Selecter, bot *Bot, u *User) {
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

	nuts, err := in.SelectNuts(u.Id)
	if err != nil {
		return
	}
	if nuts < DuoCost {
		return
	}
	bot.Message(fmt.Sprintf("@%s has redeemed DUOS!", u.Name))
	bot.duoqueue = append(bot.duoqueue, u)
}

func duoCharge(in dba.Redeemer, bot *Bot, u *User) {
	if !u.IsMod && !u.IsBroadcaster {
		return
	}
	if len(bot.duoqueue) <= 0 {
		fmt.Println("DuoQueue Empty")
		return
	}

	r := bot.duoqueue[0]

	if err := in.InsertRedeem(r.Id, TypeDuo, DuoCost); err != nil {
		return
	}
	if err := in.RemoveNuts(r.Id, DuoCost); err != nil {
		fmt.Printf("RemoveNuts - Error: %s\n", err)
		return
	}
	bot.Message(fmt.Sprintf("@%s has been charged for DUOS!", r.Name))
	bot.duoqueue = bot.duoqueue[1:]
}

func duoRemove(bot *Bot, u *User) {
	if !u.IsMod && !u.IsBroadcaster {
		return
	}
	if len(bot.duoqueue) <= 0 {
		fmt.Println("DuoQueue Empty")
		return
	}
	bot.duoqueue = bot.duoqueue[1:]
}

func duoQueue(bot *Bot) {
	var msg string
	for i, u := range bot.duoqueue {
		msg = msg + fmt.Sprintf("%v. %s   ", i+1, u.Name)
	}
	bot.Message(msg)
}

func duoOpen(bot *Bot, u *User) {
	if !u.IsMod && !u.IsBroadcaster {
		return
	}
	bot.duoopen = true
	bot.Message(fmt.Sprintf("DUOS is now open! type \"!redeem duo\" to play with penutty."))
}

func duoClose(bot *Bot, u *User) {
	if !u.IsMod && !u.IsBroadcaster {
		return
	}
	bot.duoopen = false
	bot.Message(fmt.Sprintf("DUOS is now closed."))
}

func redeemVBucks(in dba.SelectRedeemer, bot *Bot, u *User) {
	cost := 8.00
	vbucks := 2
	nuts, err := in.SelectNuts(u.Id)
	if err != nil {
		fmt.Printf("SelectNuts - Error: %s\n", err)
		return
	}
	if nuts < cost {
		return
	}
	if err := in.InsertRedeem(u.Id, vbucks, cost); err != nil {
		return
	}
	if err := in.RemoveNuts(u.Id, cost); err != nil {
		fmt.Printf("RemoveNuts - Error: %s\n", err)
		return
	}
	bot.Message(fmt.Sprintf("@%s has redeemed VBUCKS!", u.Name))
}

func leaderBoard(in dba.TopSelecter, bot *Bot, u *User) {

	set, err := in.SelectTopUsersByNuts()
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

func trivia(in dba.Adder, bot *Bot, u *User, message string) {
	if bot.triviaquestion.Question == "" {
		return
	}

	a := strings.Split(message, "!trivia ")
	answer := a[1]
	if strings.ToLower(bot.triviaquestion.Answer) != answer {
		return
	}

	reward := 10.0
	if err := in.AddNuts(u.Id, reward); err != nil {
		fmt.Printf("Trivia - Error: %s\n", err)
		return
	}
	bot.Message(fmt.Sprintf("@%s has answered the trivia question correctly! You've been rewarded %v nuts!", u.Name, reward))
	bot.triviaquestion = TriviaQuestion{}
}

func nuts(in dba.Selecter, bot *Bot, u *User) {
	nuts, err := in.SelectNuts(u.Id)
	if err != nil {
		return
	}
	bot.Message(fmt.Sprintf("@%s - Nuts = %v", u.Name, nuts))
}

func thanks(in dba.Thanker, bot *Bot, u *User, message string) {
	message = strings.ToLower(message)
	a := strings.Split(message, "!thanks ")
	referencedByUserName := a[1]

	referencedByUserID, err := in.SelectUserID(referencedByUserName)
	if err != nil && err != sql.ErrNoRows {
		return
	}

	nuts, err := in.SelectNuts(u.Id)
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
	case !isNutty(bot, &User{Id: referencedByUserID, Name: referencedByUserName}):
		fmt.Printf("\n%s - attempted to thank someone who hasn't ran !getnutty.", u.Name, referencedByUserName)
		return
	}

	ok, err := in.ReferenceExists(u.Id)
	switch {
	case err != nil && err != sql.ErrNoRows:
		return
	case ok:
		fmt.Printf("@%s you already thanked another user!", u.Name, referencedByUserName)
		return
	}

	if err := in.CreateReference(u.Id, referencedByUserID); err != nil && err != sql.ErrNoRows {
		return
	}
	reward := 0.25
	if err := in.AddNuts(referencedByUserID, reward); err != nil {
		return
	}

	bot.Message(fmt.Sprintf("Thanks @%s, for recommending penutty_ to @%s! You've earned %v nut!", referencedByUserName, u.Name, reward))

}

func getNutty(in dba.CreateUserer, bot *Bot, u *User) {
	if err := in.CreateUser(u.Name, u.Id); err != nil {
		return
	}
	bot.Message(fmt.Sprintf("/w %s Rufffff! Welcome to penutty's channel %s! type !how in the channel chat to see how to earn !nuts. Nuts can be redeemed for VBUCKS, playing duos with penutty, and more!", u.Name, u.Name))
}

func findYogi(in dba.Adder, bot *Bot, u *User, message string) {
	message = strings.ToLower(message)
	a := strings.Split(message, "!findyogi ")
	hash := a[1]

	found, ok := bot.yogihashs[hash]
	if found || !ok {
		return
	}

	bot.yogihashs[hash] = true
	reward := 0.1
	if err := in.AddNuts(u.Id, reward); err != nil {
		fmt.Errorf("Error: %s\n", err)
		return
	}
	bot.Message(fmt.Sprintf("@%s found Yogi!!! You've been rewarded %v chat points.", u.Name, reward))
}

func dne(in dba.Adder, bot *Bot, u *User) {
	lastMsg, ok := bot.lastMsg[u.Id]
	if ok && time.Since(lastMsg) <= 10*time.Minute {
		return
	}

	reward := 0.005
	if err := in.AddNuts(u.Id, reward); err != nil {
		return
	}

	bot.lastMsg[u.Id] = time.Now()
}

func isNutty(in dba.UserNameSelectUpdater, bot *Bot, u *User) bool {
	userName, err := in.SelectUserName(u.Id)
	if err == sql.ErrNoRows {
		return false
	}
	if err != nil {
		fmt.Printf("\nis Nutty - Error: %s", err)
	}
	if userName == u.Name {
		return true
	}
	if err = in.UpdateUserName(u.Id, u.Name); err != nil {
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

func fortniteBet(bot *Bot, u *User) {
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

func fortniteEndBet(bot *Bot, u *User) {
	if !u.IsMod && !u.IsBroadcaster {
		return
	}
	bot.bet.open = false
	bot.Message("BETTING ENDS")
}

func fortniteResolveBet(bot *Bot, u *User, message string) {
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
	message = strings.ToLower(message)
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
		if err := bot.dba.RemoveNuts(b.userID, b.amount); err != nil {
			fmt.Printf("FortniteResolveBet - RemoveNuts - Error: %s\n", err)
			return
		}
	}

	for _, b := range profitors {
		reward := totalDebits * (b.amount / totalProfits)
		if err := bot.dba.AddNuts(b.userID, reward); err != nil {
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

func fortniteCancelBet(bot *Bot, u *User) {
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

func win(bot *Bot, u *User, message string) {
	message = strings.ToLower(message)
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

	nuts, err := bot.dba.SelectNuts(u.Id)
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

func lose(bot *Bot, u *User, message string) {
	message = strings.ToLower(message)
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

	nuts, err := bot.dba.SelectNuts(u.Id)
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

var InvalidLineFormat = errors.New("Twitch IRC Line format is invalid.")

func lineToMap(line string) (map[string]string, error) {
	m := make(map[string]string)
	sets := strings.Split(line, ";")
	for _, v := range sets {
		pair := strings.Split(v, "=")
		if len(pair) != 2 {
			return m, InvalidLineFormat
		}
		if pair[0] == "@badges" {
			var err error
			m, err = badgesToMap(pair)
			if err != nil {
				fmt.Printf("err: %s\n", err)
			}
		} else {
			m[pair[0]] = pair[1]
		}
	}
	return m, nil
}

var InvalidBadgesFormat = errors.New("Twitch IRC Badge format is invalid.")

func badgesToMap(badges []string) (map[string]string, error) {
	m := make(map[string]string)
	if len(badges) != 2 {
		return m, InvalidBadgesFormat
	}
	bdgs := strings.Split(badges[1], ",")
	if len(bdgs) != 3 {
		return m, InvalidBadgesFormat
	}
	for _, b := range bdgs {
		set := strings.Split(b, "/")
		if len(set) != 2 {
			return m, InvalidBadgesFormat
		}
		m[set[0]] = set[1]
	}
	return m, nil
}
