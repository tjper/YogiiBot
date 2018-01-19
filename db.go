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

func (bot *Bot) UserExists(name string) (ok bool, err error) {
	query := `SELECT CASE
			  WHEN [UserName] IS NOT NULL THEN 1
			  ELSE 0
			 END
		  FROM [info].[Users]
		  WHERE [UserName] = ?`
	args := []interface{}{name}

	if err := bot.dbconn.QueryRow(query, args...).Scan(&ok); err != nil && err != sql.ErrNoRows {
		fmt.Printf("Error: %s", err)
	}
	return
}

func (bot *Bot) CreateUser(name string) (err error) {
	query := `INSERT INTO [info].[Users] ([Username])
		  VALUES(?)`
	args := []interface{}{name}

	if _, err = bot.dbconn.Exec(query, args...); err != nil {

		fmt.Printf("Error: %s", err)
		return
	}
	return
}

func (bot *Bot) CreateReference(username, referencedByUserName string) (err error) {
	query := `INSERT INTO [info].[References] ([UserName], [ReferencedByUserName])
		  VALUES(?, ?)`
	args := []interface{}{username, referencedByUserName}

	if _, err = bot.dbconn.Exec(query, args...); err != nil {
		fmt.Printf("Error: %s", err)
		return
	}
	return
}

func (bot *Bot) AddNuts(userName string, cnt int) (err error) {
	query := `UPDATE [info].[Users]
		  SET NutsAllTime = NutsAllTime + ?,
		      NutsCurrent = NutsCurrent + ?
		  WHERE [UserName] = ?`
	args := []interface{}{cnt, cnt, userName}

	if _, err = bot.dbconn.Exec(query, args...); err != nil {
		fmt.Printf("Error: %s", err)
		return
	}
	return
}

func (bot *Bot) SelectNuts(username string) (nuts string, err error) {
	var (
		nutsAllTime int
		nutsCurrent int
	)

	query := `SELECT NutsAllTime, 
			 NutsCurrent
		  FROM [info].[Users]
		  WHERE [UserName] = ?`
	args := []interface{}{username}
	if err = bot.dbconn.QueryRow(query, args...).Scan(&nutsAllTime, &nutsCurrent); err != nil {
		fmt.Printf("Error: %s", err)
		return
	}
	nuts = fmt.Sprintf("current = %v | all-time = %v", nutsCurrent, nutsAllTime)
	return
}
