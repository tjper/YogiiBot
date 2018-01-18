// db
package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	_ "github.com/denisenkom/go-mssqldb"
)

func (bot *Bot) readSettingsDB(channel string) bool {
	settings, err := ioutil.ReadFile("settings#" + channel + ".ini")
	bot.channel = "#" + channel
	if err != nil {
		fmt.Println("Unable to read SettingsDB from " + channel)
		return false
	}
	split1 := strings.Split(string(settings), "\n")
	for _, splitted1 := range split1 {
		split2 := strings.Split(splitted1, "|")
		if split2[0] == "nickname" {
			bot.nick = split2[1]
		}

	}
	return true
}

func (bot *Bot) writeSettingsDB() {
	dst, err := os.Create("settings" + bot.channel + ".ini")
	defer dst.Close()
	if err != nil {
		fmt.Println("Can't write to SettingsDB from " + bot.channel)
		return
	}
	fmt.Fprintf(dst, "nickname|"+bot.nick+"\n")
}

func (bot *Bot) OpenNuttyDB() {
	server := "nuttydb.database.windows.net"
	port := "1433"
	username := "yogiibot"
	pass := "tIrjONIN4gtKRaJ5SHtN"
	database := "NuttyDB"

	connstr := fmt.Sprintf("server=%s;user id=%s;password=%s;port=%s;database=%s", server, username, pass, port, database)

	db, err := sql.Open("mssql", connstr)
	if err != nil {
		fmt.Println("Can't connect to NuttyDB")
		return
	}
	bot.dbconn = db
}

func (bot *Bot) CloseNuttyDB() {
	bot.dbconn.Close()
}

func (bot *Bot) UserExists(name string) (ok bool) {
	query := `SELECT CASE
			  WHEN [UserName] IS NOT NULL THEN 1
			  ELSE 0
			 END
		  FROM [info].[Users]
		  WHERE [UserName] = ?`
	args := []interface{}{name}

	if err := bot.dbconn.QueryRow(query, args...).Scan(&ok); err != nil {
		fmt.Println("UserExists failed to check for user")
	}
	return
}

func (bot *Bot) CreateUser(name string) {
	query := `INSERT INTO [info].[Users] ([Username])
		  VALUES(?)`
	args := []interface{}{name}

	res, err := bot.dbconn.Exec(query, args...)
	if err != nil {
		fmt.Printf("Error: %s", err)
		return
	}
	cnt, err := res.RowsAffected()
	if err != nil {
		fmt.Printf("Error: %s", err)
		return
	}
	if cnt == 1 {
		message := fmt.Sprintf("Welcome to penutty_'s channel @%s!", name)
		bot.Message(message)
	}
}
