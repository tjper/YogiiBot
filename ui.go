package main

import (
	"log"
	"net/http"
)

func (bot *Bot) OpenUI() {
	http.Handle("/", http.FileServer(http.Dir("/home/james/go/src/github.com/penutty/YogiiBot/static")))
	log.Fatal(http.ListenAndServe(":8081", nil))
}
