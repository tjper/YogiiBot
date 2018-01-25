package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func (bot *Bot) OpenAPI() {
	http.HandleFunc("/duel", bot.HandleDuel)

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func (bot *Bot) HandleDuel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		return
	}

	type Body struct {
		command string
		noun    string
	}
	b := new(Body)
	if err := json.NewDecoder(r.Body).Decode(b); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}

	m := map[string]string{"mod": "1"}
	switch b.command {
	case "!duel":
		bot.Duel(m)
	case "!duelcancel":
		bot.DuelCancel(m)
	case "!duelwinner":
		bot.DuelWinner(m, fmt.Sprintf("%s %s", b.command, b.noun))
	default:
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}
}
