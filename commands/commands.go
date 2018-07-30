package commands

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	yogiDB "github.com/penutty/YogiiBot/db"
	"html"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	loginfo  *log.Logger
	logerror *log.Logger

	yogiDBClient yogiDB.Clienter
)

func init() {
	Logger := func(logType string) *log.Logger {
		file := "/home/james/go/log/yogiibot_commands.txt"
		f, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			panic(err)
		}

		l := log.New(f, strings.ToUpper(logType)+": ", log.Ldate|log.Ltime|log.Lmicroseconds|log.LUTC|log.Lshortfile)
		return l
	}

	loginfo = Logger("info")
	logerror = Logger("error")
	loginfo.Println("In init()")

	var err error
	yogiDBClient, err = yogiDB.NewClient()
	if err != nil {
		logerror.Println(err)
		panic(err)
	}
}

type User struct {
	Id            int
	Name          string
	IsSubscriber  bool
	IsMod         bool
	IsBroadcaster bool
}

func NewUser(m map[string]string) (u *User, err error) {
	u = new(User)
	tuserID, ok := m["user-id"]
	if !ok {
		return u, ErrInvalidIdentifiers
	}
	u.Id, err = strconv.Atoi(tuserID)
	if err != nil {
		return u, err
	}

	u.Name, ok = m["display-name"]
	if !ok {
		return u, ErrInvalidIdentifiers
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
		registered, err := yogiDBClient.SelectSubStatus(u.Id)
		if err != nil {
			logerror.Println(err)
			return u, err
		}
		if registered {
			return u, nil
		}
		if err := yogiDBClient.UpdateSubStatus(u.Id); err != nil {
			logerror.Println(err)
			return u, err
		}
		if err := yogiDBClient.AddNuts(u.Id, 1.0); err != nil {
			logerror.Println(err)
			return u, err
		}
	}

	return u, nil
}

type Command struct {
	Author   *User
	Message  string
	Response string
	Error    error
}

func NewCommand(line string, channel string) (*Command, error) {
	m, err := lineToMap(line)
	if err != nil {
		logerror.Println(err)
		return nil, err
	}

	command := new(Command)
	command.Author, err = NewUser(m)
	if err != nil {
		logerror.Println(err)
		return nil, err
	}
	message := strings.Split(line, fmt.Sprintf("PRIVMSG %s :", channel))
	command.Message = message[1]

	return command, nil
}

var (
	nutsRE     = regexp.MustCompile(`^(!nuts)$`)
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

func (c *Command) Exec() (string, error) {
	nutty, err := isNutty(c.Author)
	if err != nil {
		return "", err
	}
	if !nutty {
		getNutty(c.Author)
	}

	var msg string
	switch {
	case winRE.MatchString(c.Message):
		msg, err = win(c.Author, c.Message)
	case loseRE.MatchString(c.Message):
		msg, err = lose(c.Author, c.Message)
	case fortniteBetRE.MatchString(c.Message):
		msg, err = fortniteBet(c.Author)
	case fortniteEndBetRE.MatchString(c.Message):
		msg, err = fortniteEndBet(c.Author)
	case fortniteCancelBetRE.MatchString(c.Message):
		msg, err = fortniteCancelBet(c.Author)
	case fortniteResolveBetRE.MatchString(c.Message):
		msg, err = fortniteResolveBet(c.Author, c.Message)
	case nutsRE.MatchString(c.Message):
		msg, err = nuts(c.Author)
	case thanksRE.MatchString(c.Message):
		msg, err = thanks(c.Author, c.Message)
	case findYogiRE.MatchString(c.Message):
		msg, err = findYogi(c.Author, c.Message)
	case leaderboardRE.MatchString(c.Message):
		msg, err = leaderBoard(c.Author)
	case redeemduoRE.MatchString(c.Message):
		msg, err = redeemDuo(c.Author)
	case duoqueueRE.MatchString(c.Message):
		msg = duoQueue()
	case duoremoveRE.MatchString(c.Message):
		err = duoRemove(c.Author)
	case duochargeRE.MatchString(c.Message):
		msg, err = duoCharge(c.Author)
	case duoopenRE.MatchString(c.Message):
		msg, err = duoOpen(c.Author)
	case duocloseRE.MatchString(c.Message):
		msg, err = duoClose(c.Author)
	case redeemvbucksRE.MatchString(c.Message):
		msg, err = redeemVBucks(c.Author)
	case quoteRE.MatchString(c.Message):
		err = quote(c.Author, c.Message)
	case getQuoteRE.MatchString(c.Message):
		msg, err = getQuote(c.Author, c.Message)
	default:
		err = dne(c.Author)
	}
	return msg, err
}

func getQuote(u *User, message string) (string, error) {
	message = strings.ToLower(message)
	a := strings.Split(message, "!")
	author := a[1]

	quote, err := yogiDBClient.SelectQuote(author)
	if err != nil {
		logerror.Println(err)
		return "", err
	}

	return fmt.Sprintf("\" %s \" - %s", quote, author), nil
}

func quote(u *User, message string) error {
	if !u.IsSubscriber {
		return nil
	}

	a := strings.Split(message, "!quote ")
	quote := a[1]
	quote = strings.Trim(quote, "\"")

	if err := yogiDBClient.UpdateQuote(u.Id, quote); err != nil {
		logerror.Println(err)
		return err
	}
	return nil
}

var (
	DuoCost       = 1.00
	TypeDuo       = 1
	DuoQueueLimit = 3
)

var (
	duoopen  bool
	duoqueue []*User
)

func redeemDuo(u *User) (string, error) {
	if !duoopen {
		return "", nil
	}
	if len(duoqueue) >= DuoQueueLimit {
		return "", nil
	}
	for _, r := range duoqueue {
		if r.Name == u.Name {
			return "", nil
		}
	}

	nuts, err := yogiDBClient.SelectNuts(u.Id)
	if err != nil {
		logerror.Println(err)
		return "", err
	}
	if nuts < DuoCost {
		return "", nil
	}
	duoqueue = append(duoqueue, u)

	return fmt.Sprintf("@%s has joined duos queue!", u.Name), nil
}

func duoCharge(u *User) (string, error) {
	if !u.IsMod && !u.IsBroadcaster {
		return "", nil
	}
	if len(duoqueue) <= 0 {
		return "", nil
	}

	r := duoqueue[0]

	if err := yogiDBClient.InsertRedeem(r.Id, TypeDuo, DuoCost); err != nil {
		logerror.Println(err)
		return "", err
	}
	if err := yogiDBClient.RemoveNuts(r.Id, DuoCost); err != nil {
		logerror.Println(err)
		return "", err
	}

	duoqueue = duoqueue[1:]
	return fmt.Sprintf("@%s has been charged for DUOS!", r.Name), nil
}

func duoRemove(u *User) error {
	if !u.IsMod && !u.IsBroadcaster {
		return nil
	}
	if len(duoqueue) <= 0 {
		return nil
	}
	duoqueue = duoqueue[1:]
	return nil
}

func duoQueue() string {
	var msg string
	for i, u := range duoqueue {
		msg = msg + fmt.Sprintf("%v. %s   ", i+1, u.Name)
	}
	return msg
}

func duoOpen(u *User) (string, error) {
	if !u.IsMod && !u.IsBroadcaster {
		return "", nil
	}
	duoopen = true
	return fmt.Sprintf("DUOS is now open! type \"!redeem duo\" to play with penutty."), nil
}

func duoClose(u *User) (string, error) {
	if !u.IsMod && !u.IsBroadcaster {
		return "", nil
	}
	duoopen = false
	duoqueue = make([]*User, 0)
	return fmt.Sprintf("DUOS is now closed."), nil
}

func redeemVBucks(u *User) (string, error) {
	cost := 8.00
	vbucks := 2
	nuts, err := yogiDBClient.SelectNuts(u.Id)
	if err != nil {
		logerror.Println(err)
		return "", err
	}
	if nuts < cost {
		return "", nil
	}
	if err := yogiDBClient.InsertRedeem(u.Id, vbucks, cost); err != nil {
		logerror.Println(err)
		return "", err
	}
	if err := yogiDBClient.RemoveNuts(u.Id, cost); err != nil {
		logerror.Println(err)
		return "", err
	}
	return fmt.Sprintf("@%s has redeemed 1000 VBUCKS!", u.Name), nil
}

func leaderBoard(u *User) (string, error) {
	set, err := yogiDBClient.SelectTopUsersByNuts()
	if err != nil {
		logerror.Println(err)
		return "", err
	}

	var res string
	for i, row := range set {
		res = res + fmt.Sprintf("(%v) %s = %v nuts ... ", i+1, row.UserName, row.Nuts)
	}
	return res, nil
}

func nuts(u *User) (string, error) {
	nuts, err := yogiDBClient.SelectNuts(u.Id)
	if err != nil {
		logerror.Println(err)
		return "", err
	}
	return fmt.Sprintf("@%s - Nuts = %v", u.Name, nuts), nil
}

func thanks(u *User, message string) (string, error) {
	message = strings.ToLower(message)
	a := strings.Split(message, "!thanks ")
	referencedByUserName := a[1]

	referencedByUserID, err := yogiDBClient.SelectUserID(referencedByUserName)
	if err != nil && err != sql.ErrNoRows {
		logerror.Println(err)
		return "", err
	}

	nuts, err := yogiDBClient.SelectNuts(u.Id)
	if err != nil {
		logerror.Println(err)
		return "", err
	}

	nutty, err := isNutty(&User{Id: referencedByUserID, Name: referencedByUserName})
	if err != nil {
		logerror.Println(err)
		return "", err
	}
	switch {
	case nuts < 0.5:
		return "", nil
	case u.Id == referencedByUserID:
		return "", nil
	case !nutty:
		return "", nil
	}

	ok, err := yogiDBClient.ReferenceExists(u.Id)
	switch {
	case err != nil && err != sql.ErrNoRows:
		logerror.Println(err)
		return "", err
	case ok:
		return "", nil
	}

	if err := yogiDBClient.CreateReference(u.Id, referencedByUserID); err != nil && err != sql.ErrNoRows {
		logerror.Println(err)
		return "", err
	}
	reward := 0.25
	if err := yogiDBClient.AddNuts(referencedByUserID, reward); err != nil {
		logerror.Println(err)
		return "", err
	}

	return fmt.Sprintf("Thanks @%s, for recommending penutty_ to @%s! You've earned %v nut!", referencedByUserName, u.Name, reward), nil
}

func getNutty(u *User) error {
	if err := yogiDBClient.CreateUser(u.Name, u.Id); err != nil {
		logerror.Println(err)
		return err
	}
	return nil
}

var msgs = make(map[int]time.Time)

func dne(u *User) error {
	lastMsg, ok := msgs[u.Id]
	if ok && time.Since(lastMsg) <= 10*time.Minute {
		return nil
	}

	reward := 0.005
	if err := yogiDBClient.AddNuts(u.Id, reward); err != nil {
		logerror.Println(err)
		return err
	}
	msgs[u.Id] = time.Now()
	return nil
}

func isNutty(u *User) (bool, error) {
	userName, err := yogiDBClient.SelectUserName(u.Id)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		logerror.Println(err)
		return false, err
	}
	if userName == u.Name {
		return true, nil
	}
	if err = yogiDBClient.UpdateUserName(u.Id, u.Name); err != nil {
		logerror.Println(err)
		return true, err
	}
	return true, nil
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

func NewBetRound(open bool) *betRound {
	return &betRound{
		open:          open,
		startTime:     time.Now(),
		betees:        make(map[int]bool),
		totalWinBets:  0,
		totalLoseBets: 0,
		loseBets:      make([]*bet, 0),
		winBets:       make([]*bet, 0),
	}
}

type bet struct {
	userID int
	amount float64
}

var br = NewBetRound(false)

func fortniteBet(u *User) (string, error) {
	if !u.IsMod && !u.IsBroadcaster {
		return "", nil
	}

	br = NewBetRound(true)
	return "BETTING BEGINS", nil
}

func fortniteEndBet(u *User) (string, error) {
	if !u.IsMod && !u.IsBroadcaster {
		return "", nil
	}
	br.open = false
	return "BETTING ENDS", nil
}

func fortniteResolveBet(u *User, message string) (string, error) {
	if br == nil {
		return "", nil
	}
	if br.open {
		return "", nil
	}
	if !u.IsMod && !u.IsBroadcaster {
		return "", nil
	}
	if len(br.winBets) == 0 || len(br.loseBets) == 0 {
		return "", nil
	}
	message = strings.ToLower(message)
	a := strings.Split(message, "!fortniteresolvebet ")
	result := a[1]

	var profitors, debitors []*bet
	var totalDebits, totalProfits float64
	if result == "win" {
		profitors = br.winBets
		debitors = br.loseBets
		totalDebits = br.totalLoseBets
		totalProfits = br.totalWinBets
	} else {
		profitors = br.loseBets
		debitors = br.winBets
		totalDebits = br.totalWinBets
		totalProfits = br.totalLoseBets
	}

	for _, b := range debitors {
		if err := yogiDBClient.RemoveNuts(b.userID, b.amount); err != nil {
			logerror.Println(err)
			return "", err
		}
	}

	for _, b := range profitors {
		reward := totalDebits * (b.amount / totalProfits)
		if err := yogiDBClient.AddNuts(b.userID, reward); err != nil {
			logerror.Println(err)
			return "", err
		}
	}
	br.open = false
	br = NewBetRound(false)
	if result == "win" {
		return "PENUTTY_ WON!", nil
	}
	return "PENUTTY LOST.", nil
}

func fortniteCancelBet(u *User) (string, error) {
	if br == nil {
		return "", nil
	}
	if !u.IsMod && !u.IsBroadcaster {
		return "", nil
	}
	br = NewBetRound(false)
	return "BET CANCELLED.", nil
}

func win(u *User, message string) (string, error) {
	message = strings.ToLower(message)
	if br == nil {
		return "", nil
	}
	if !br.open {
		return "", nil
	}

	if _, ok := br.betees[u.Id]; ok {
		return "", nil
	}

	a := strings.Split(message, "!win ")
	amount, err := strconv.ParseFloat(a[1], 64)
	if err != nil {
		logerror.Println(err)
		return "", err
	}

	nuts, err := yogiDBClient.SelectNuts(u.Id)
	if err != nil {
		logerror.Println(err)
		return "", err
	}
	if nuts < amount {
		return "", nil
	}

	b := &bet{u.Id, amount}
	br.winBets = append(br.winBets, b)
	br.totalWinBets += amount
	br.betees[u.Id] = true
	return fmt.Sprintf("@%s bet %v nuts on penutty winning!", u.Name, amount), nil
}

func lose(u *User, message string) (string, error) {
	message = strings.ToLower(message)
	if br == nil {
		return "", nil
	}
	if !br.open {
		return "", nil
	}

	if _, ok := br.betees[u.Id]; ok {
		return "", nil
	}

	a := strings.Split(message, "!lose ")
	amount, err := strconv.ParseFloat(a[1], 64)
	if err != nil {
		logerror.Println(err)
		return "", err
	}

	nuts, err := yogiDBClient.SelectNuts(u.Id)
	if err != nil {
		logerror.Println(err)
		return "", err
	}
	if nuts < amount {
		return "", nil
	}

	b := &bet{u.Id, amount}
	br.loseBets = append(br.loseBets, b)
	br.totalLoseBets += amount
	br.betees[u.Id] = true
	return fmt.Sprintf("@%s bet %v nuts on penutty losing.", u.Name, amount), nil

}

var yogihash map[string]bool

func NewWildYogi(w io.Writer) {
	var wg sync.WaitGroup
	for {
		wg.Add(1)
		go func() {
			defer wg.Done()
			n := rand.Int() % 150
			time.Sleep(time.Duration(n) * time.Minute)
			r := randomString(5)
			yogihash[r] = false
			fmt.Fprintf(w, "A wild yogi has appeared! Who will catch him first? Type !findyogi %s", r)
		}()
		wg.Wait()
	}
}

func findYogi(u *User, message string) (string, error) {
	message = strings.ToLower(message)
	a := strings.Split(message, "!findyogi ")
	hash := a[1]

	found, ok := yogihash[hash]
	if found || !ok {
		return "", nil
	}

	yogihash[hash] = true
	reward := 0.1
	if err := yogiDBClient.AddNuts(u.Id, reward); err != nil {
		logerror.Println(err)
		return "", err
	}
	return fmt.Sprintf("@%s found Yogi!!! You've been rewarded %v chat points.", u.Name, reward), nil
}

type TriviaQuestion struct {
	Question string
	Answer   string
}

var triviaQuestion *TriviaQuestion

func NewTriviaQuestion(w io.Writer) {
	var wg sync.WaitGroup
	for {
		wg.Add(1)
		go func() {
			defer wg.Done()
			n := (rand.Int() % 20) + 10
			time.Sleep(time.Duration(n) * time.Minute)
			r, err := http.Get("https://opentdb.com/api.php?amount=1&category=15&type=multiple")
			if err != nil {
				logerror.Println(err)
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
				logerror.Println(err)
				return
			}
			triviaQuestion = &TriviaQuestion{
				Question: html.UnescapeString(b.Results[0].Question),
				Answer:   html.UnescapeString(b.Results[0].Correct_Answer),
			}
			fmt.Fprint(w, "%s", triviaQuestion.Question)
		}()
		wg.Wait()
	}
}

func trivia(u *User, message string) (string, error) {
	if triviaQuestion.Question == "" {
		return "", nil
	}

	a := strings.Split(message, "!trivia ")
	answer := a[1]
	if strings.ToLower(triviaQuestion.Answer) != answer {
		return "", nil
	}

	reward := 10.0
	if err := yogiDBClient.AddNuts(u.Id, reward); err != nil {
		logerror.Println(err)
		return "", err
	}
	triviaQuestion = &TriviaQuestion{}
	return fmt.Sprintf("@%s has answered the trivia question correctly! You've been rewarded %v nuts!", u.Name, reward), nil
}

// UTILITY FUNCTIONS
var (
	ErrInvalidIdentifiers  = errors.New("Invalid user identifiers.")
	ErrInvalidBadgesFormat = errors.New("Twitch IRC Badge format is invalid.")
	ErrInvalidLineFormat   = errors.New("Twitch IRC Line format is invalid.")
)

func lineToMap(line string) (map[string]string, error) {
	m := make(map[string]string)
	sets := strings.Split(line, ";")
	for _, v := range sets {
		pair := strings.Split(v, "=")
		if len(pair) != 2 {
			logerror.Println(ErrInvalidLineFormat)
			return m, ErrInvalidLineFormat
		}
		if pair[0] == "@badges" {
			var err error
			m, err = badgesToMap(pair[1])
			if err != nil {
				logerror.Println(err)
				return m, err
			}
		} else {
			m[pair[0]] = pair[1]
		}
	}
	return m, nil
}

func badgesToMap(badges string) (map[string]string, error) {
	m := make(map[string]string)
	bdgs := strings.Split(badges, ",")
	if len(bdgs) != 3 {
		logerror.Println(ErrInvalidBadgesFormat)
		return m, ErrInvalidBadgesFormat
	}
	for _, b := range bdgs {
		set := strings.Split(b, "/")
		if len(set) != 2 {
			logerror.Println(ErrInvalidBadgesFormat)
			return m, ErrInvalidBadgesFormat
		}
		m[set[0]] = set[1]
	}
	return m, nil
}

const letterBytes = "abcdefghijklmnopqrstuvwxyz"

func randomString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
