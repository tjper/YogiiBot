package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

func (bot *Bot) OpenAPI() {
	http.HandleFunc("/duel", bot.HandleDuel)
	http.HandleFunc("/fortnitebet", bot.GetFortniteBet)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func (bot *Bot) GetFortniteBet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusNotImplemented), http.StatusNotImplemented)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	type Body struct {
		StartTime time.Time
		Win       int
		Lose      int
	}

	if bot.bet == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if !bot.bet.open || (bot.bet.totalWinBets == 0 && bot.bet.totalLoseBets == 0) {
		w.WriteHeader(http.StatusNoContent)
		return
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

func (bot *Bot) HandleDuel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, http.StatusText(http.StatusNotImplemented), http.StatusNotImplemented)
		return
	}

	type Body struct {
		Command string
		Noun    string
	}
	b := new(Body)
	if err := json.NewDecoder(r.Body).Decode(b); err != nil {
		fmt.Printf("err = %s\n", err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	m := map[string]string{"mod": "1"}
	switch b.Command {
	case "!duel":
		bot.Duel(m)
	case "!duelcancel":
		bot.DuelCancel(m)
	case "!duelwinner":
		bot.DuelWinner(m, fmt.Sprintf("%s %s", b.Command, b.Noun))
	default:
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}
}
