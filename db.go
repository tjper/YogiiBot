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
type UsersRow struct {
	UserID int
	UserName string
	Nuts float64
}

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

func (bot *Bot) UserNameExists(name string) (ok bool, err error) {
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

func (bot *Bot) UserIDExists(userID int) (ok bool, err error) {
	query := `SELECT CASE
			  WHEN [UserID] IS NOT NULL THEN 1
			  ELSE 0
			 END
		  FROM [info].[Users]
		  WHERE [UserID] = ?`
	args := []interface{}{userID}

	if err := bot.dbconn.QueryRow(query, args...).Scan(&ok); err != nil && err != sql.ErrNoRows {
		fmt.Printf("Error: %s", err)
	}
	return
}

func (bot *Bot) CreateUser(name string, userID int) (err error) {
	query := `INSERT INTO [info].[Users] ([Username], [UserID])
		  VALUES(?, ?)`
	args := []interface{}{name, userID}

	if _, err = bot.dbconn.Exec(query, args...); err != nil {
		fmt.Printf("Error: %s", err)
		return
	}
	return
}

func (bot *Bot) ReferenceExists(userID int) (ok bool, err error) {
	query := `SELECT CASE
			  WHEN [UserID] IS NOT NULL THEN 1
			  ELSE 0
			 END
		  FROM [info].[References]
		  WHERE [UserID] = ?`
	args := []interface{}{userID}
	if err := bot.dbconn.QueryRow(query, args...).Scan(&ok); err != nil && err != sql.ErrNoRows {
		fmt.Printf("Error: %s", err)
	}
	return
}

func (bot *Bot) CreateReference(userID, referencedByUserID int) (err error) {
	query := `INSERT INTO [info].[References] ([UserID], [ReferencedByUserID])
		  VALUES(?, ?)`
	args := []interface{}{userID, referencedByUserID}

	if _, err = bot.dbconn.Exec(query, args...); err != nil {
		fmt.Printf("Error: %s", err)
		return
	}
	return
}

func (bot *Bot) AddChatPoints(userID int, cnt float64) (err error) {
	query := `UPDATE [info].[Users]
		  SET ChatPoints = ChatPoints + ?
		  WHERE [UserID] = ?`
	args := []interface{}{cnt, userID}

	if _, err = bot.dbconn.Exec(query, args...); err != nil {
		fmt.Printf("Error: %s", err)
		return
	}
	return
}

func (bot *Bot) AddNuts(userID int, cnt float64) (err error) {
	query := `UPDATE [info].[Users]
		  SET Nuts = Nuts + ?
		  WHERE [UserID] = ?`
	args := []interface{}{cnt, userID}

	if _, err = bot.dbconn.Exec(query, args...); err != nil {
		fmt.Printf("Error: %s", err)
		return
	}
	return
}

func (bot *Bot) RemoveNuts(userID int, cnt float64) (err error) {
	query := `UPDATE [info].[Users]
		  SET Nuts = Nuts - ?
		  WHERE [UserID] = ?`
	args := []interface{}{cnt, userID}

	if _, err = bot.dbconn.Exec(query, args...); err != nil {
		fmt.Printf("Error: %s", err)
		return
	}
	return
}

func (bot *Bot) SelectChatPoints(userID int) (points float64, err error) {
	query := `SELECT ChatPoints 
		  FROM [info].[Users]
		  WHERE [UserID] = ?`
	args := []interface{}{userID}
	if err = bot.dbconn.QueryRow(query, args...).Scan(&points); err != nil {
		fmt.Printf("Error: %s", err)
		return
	}
	return
}

func (bot *Bot) SelectNuts(userID int) (nuts float64, err error) {
	query := `SELECT Nuts
		  FROM [info].[Users]
		  WHERE [UserID] = ?`
	args := []interface{}{userID}
	if err = bot.dbconn.QueryRow(query, args...).Scan(&nuts); err != nil {
		fmt.Printf("Error: %s", err)
		return
	}
	return
}

func (bot *Bot) SelectTopUsersByNuts() (set []UsersRow, err error) {
	query := `SELECT TOP 5 [UserName], [Nuts]
		  FROM [info].[Users]
		  ORDER BY [Nuts] DESC`
	rows, err := bot.dbconn.Query(query)
	if err != nil {
		fmt.Printf("Error: %s", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var (
			userName string
			nuts float64
		)
		if err = rows.Scan(&userName, &nuts); err != nil {
			fmt.Printf("Error: %s", err)
			return
		}
		set = append(set, UsersRow{ UserName: userName, Nuts: nuts })
	}
	if err = rows.Err(); err != nil {
		fmt.Printf("Error: %s", err)
		return
	}
	return
}

func (bot *Bot) SelectUserID(username string) (userID int, err error) {
	query := `SELECT [UserID]
		  FROM [info].[Users]
		  WHERE [UserName] = ?`
	args := []interface{}{username}
	if err = bot.dbconn.QueryRow(query, args...).Scan(&userID); err != nil && err != sql.ErrNoRows {
		fmt.Printf("Error: %s", err)
		return
	}
	return
}

func (bot *Bot) SelectUserName(userID int) (username string, err error) {
	query := `SELECT [UserName]
		  FROM [info].[Users]
		  WHERE [UserID] = ?`
	args := []interface{}{userID}
	if err = bot.dbconn.QueryRow(query, args...).Scan(&userID); err != nil && err != sql.ErrNoRows {
		fmt.Printf("Error : %s", err)
		return
	}
	return
}

func (bot *Bot) InsertRedeem(userID int, itemID int, cost float64) (err error) {
	query := `INSERT INTO [info].[Redems] ([UserID], [NutCost], [ItemID])
		  VALUES ( ?, ?, ?)`
	args := []interface{}{userID, cost, itemID}
	if _, err = bot.dbconn.Exec(query, args...); err != nil {
		fmt.Printf("Error: %s", err)
		return
	}
	return
}

func (bot *Bot) SelectSubStatus(userID int) (status bool, err error) {
	query := `SELECT [HasSubbed]
		  FROM [info].[Users]
		  WHERE [UserID] = ?`
	args := []interface{}{userID}
	if err = bot.dbconn.QueryRow(query, args...).Scan(&status); err != nil {
		fmt.Printf("Error: %s", err)
		return
	}
	return
}

func (bot *Bot) UpdateSubStatus(userID int) (err error) {
	query := `UPDATE [info].[Users]
		  SET [HasSubbed] = 1
		  WHERE [UserID] = ?`
	args := []interface{}{userID}
	if _, err = bot.dbconn.Exec(query, args...); err != nil {
		fmt.Printf("Error: %s", err)
		return
	}
	return
}
