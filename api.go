package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

func (bot *Bot) OpenAPI() {
	http.HandleFunc("/fortnitebet", bot.GetFortniteBet)
	http.HandleFunc("/triviaquestion", bot.GetTriviaQuestion)
	http.HandleFunc("/giftedsub", bot.GetGiftedSub)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func (bot *Bot) GetFortniteBet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusNotImplemented), http.StatusNotImplemented)
		return
	}
	w = setHeaders(w)

	if bot.bet == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if !bot.bet.open || (bot.bet.totalWinBets == 0 && bot.bet.totalLoseBets == 0) {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	type Body struct {
		StartTime time.Time
		Win       int
		Lose      int
	}
	b := &Body{
		StartTime: bot.bet.startTime,
		Win:       bot.bet.totalWinBets,
		Lose:      bot.bet.totalLoseBets,
	}

	if err := json.NewEncoder(w).Encode(b); err != nil {
		fmt.Printf("GetFortniteBet - Error: %s", err)
	}
}

func (bot *Bot) GetTriviaQuestion(w http.ResponseWriter, r *http.Request) {
	fmt.Println("In GetTriviaQuestion")
	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusNotImplemented), http.StatusNotImplemented)
		return
	}
	w = setHeaders(w)

	fmt.Printf("\nquestion = %s", bot.triviaquestion.Question)

	if bot.triviaquestion.Question == "" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	type Body struct {
		Question string
		Answers  []string
	}
	b := &Body{
		Question: bot.triviaquestion.Question,
		Answers:  append(bot.triviaquestion.IncorrectAnswers, bot.triviaquestion.Answer),
	}

	if err := json.NewEncoder(w).Encode(b); err != nil {
		fmt.Printf("GetTriviaQuestion - Error: %s", err)
	}
}

func (bot *Bot) GetGiftedSub(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusNotImplemented), http.StatusNotImplemented)
		return
	}
	w = setHeaders(w)

	if len(bot.giftedsubqueue) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	type Body struct {
		UserName string
	}
	b := &Body{
		UserName: bot.giftedsubqueue[0],
	}
	bot.giftedsubqueue = bot.giftedsubqueue[1:]

	if err := json.NewEncoder(w).Encode(b); err != nil {
		fmt.Printf("GetGiftedSub - Error: %s", err)
	}
}

func setHeaders(w http.ResponseWriter) http.ResponseWriter {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	return w
}
