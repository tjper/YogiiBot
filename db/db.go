package dba

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/denisenkom/go-mssqldb"
)

var (
	loginfo  *log.Logger
	logerror *log.Logger
)

func init() {
	Logger := func(logType string) *log.Logger {
		file := "/home/james/go/log/yogiibot_db.txt"
		f, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			panic(err)
		}

		l := log.New(f, strings.ToUpper(logType)+": ", log.Ldate|log.Ltime|log.Lmicroseconds|log.LUTC|log.Lshortfile)
		return l
	}

	loginfo = Logger("info")
	logerror = Logger("error")
}

type Clienter interface {
	UserNameExists(string) (bool, error)
	UserIDExists(int) (bool, error)
	CreateUser(string, int) error
	UpdateUserName(int, string) error
	ReferenceExists(int) (bool, error)
	CreateReference(int, int) error
	AddNuts(int, float64) error
	RemoveNuts(int, float64) error
	SelectNuts(int) (float64, error)
	SelectTopUsersByNuts() ([]UsersRow, error)
	SelectUserID(string) (int, error)
	SelectUserName(int) (string, error)
	InsertRedeem(int, int, float64) error
	SelectSubStatus(int) (bool, error)
	UpdateSubStatus(int) error
	UpdateQuote(int, string) error
	SelectQuote(string) (string, error)
}

type DbRunner interface {
	QueryRow(string, ...interface{}) *sql.Row
	Query(string, ...interface{}) (*sql.Rows, error)
	Exec(string, ...interface{}) (sql.Result, error)
}

type UsersRow struct {
	UserID   int
	UserName string
	Nuts     float64
}

type Client struct {
	db DbRunner
}

func NewClient() (*Client, error) {

	server := "an Microsoft Azure Db server"
	port := "1433"
	username := "yogiibot"
	pass := "fakepass"
	database := "NuttyDB"

	connstr := fmt.Sprintf("server=%s;user id=%s;password=%s;port=%s;database=%s", server, username, pass, port, database)

	db, err := sql.Open("mssql", connstr)
	if err != nil {
		logerror.Println(err)
		return nil, err
	}
	c := new(Client)
	c.db = db
	return c, nil
}

func (c *Client) UserNameExists(name string) (ok bool, err error) {
	query := `SELECT CASE
			  WHEN [UserName] IS NOT NULL THEN 1
			  ELSE 0
			 END
		  FROM [info].[Users]
		  WHERE [UserName] = ?`
	args := []interface{}{name}

	if err = c.db.QueryRow(query, args...).Scan(&ok); err != nil && err != sql.ErrNoRows {
		logerror.Println(err)
		return
	}
	return
}

func (c *Client) UserIDExists(userID int) (ok bool, err error) {
	query := `SELECT CASE
			  WHEN [UserID] IS NOT NULL THEN 1
			  ELSE 0
			 END
		  FROM [info].[Users]
		  WHERE [UserID] = ?`
	args := []interface{}{userID}

	if err = c.db.QueryRow(query, args...).Scan(&ok); err != nil && err != sql.ErrNoRows {
		logerror.Println(err)
		return
	}
	return
}

var ErrorNotOneRowAffected = errors.New("sql.Result indicates 0 rows affected by query.")

func (c *Client) CreateUser(name string, userID int) (err error) {
	query := `INSERT INTO [info].[Users] ([Username], [UserID])
		  VALUES(?, ?)`
	args := []interface{}{name, userID}

	var res sql.Result
	if res, err = c.db.Exec(query, args...); err != nil {
		logerror.Println(err)
		return
	}

	var ra int64
	if ra, err = res.RowsAffected(); err != nil {
		logerror.Println(err)
		return
	}
	if ra != 1 {
		logerror.Println(ErrorNotOneRowAffected)
		return ErrorNotOneRowAffected
	}
	return
}

func (c *Client) UpdateUserName(userID int, userName string) (err error) {
	query := `UPDATE [info].[Users]
		  SET [UserName] = ?
		  WHERE [UserID] = ?`
	args := []interface{}{userName, userID}

	var res sql.Result
	if res, err = c.db.Exec(query, args...); err != nil {
		logerror.Println(err)
		return
	}

	var ra int64
	if ra, err = res.RowsAffected(); err != nil {
		logerror.Println(err)
		return
	}
	if ra != 1 {
		logerror.Println(ErrorNotOneRowAffected)
		return ErrorNotOneRowAffected
	}
	return
}

func (c *Client) ReferenceExists(userID int) (ok bool, err error) {
	query := `SELECT CASE
			  WHEN [UserID] IS NOT NULL THEN 1
			  ELSE 0
			 END
		  FROM [info].[References]
		  WHERE [UserID] = ?`
	args := []interface{}{userID}
	if err = c.db.QueryRow(query, args...).Scan(&ok); err != nil && err != sql.ErrNoRows {
		logerror.Println(err)
		return
	}
	return
}

func (c *Client) CreateReference(userID, referencedByUserID int) (err error) {
	query := `INSERT INTO [info].[References] ([UserID], [ReferencedByUserID])
		  VALUES(?, ?)`
	args := []interface{}{userID, referencedByUserID}

	var res sql.Result
	if res, err = c.db.Exec(query, args...); err != nil {
		logerror.Println(err)
		return
	}

	var ra int64
	if ra, err = res.RowsAffected(); err != nil {
		logerror.Println(err)
		return
	}
	if ra != 1 {
		logerror.Println(ErrorNotOneRowAffected)
		return ErrorNotOneRowAffected
	}
	return
}

func (c *Client) AddNuts(userID int, cnt float64) (err error) {
	query := `UPDATE [info].[Users]
		  SET Nuts = Nuts + ?
		  WHERE [UserID] = ?`
	args := []interface{}{cnt, userID}

	var res sql.Result
	if res, err = c.db.Exec(query, args...); err != nil {
		logerror.Println(err)
		return
	}

	var ra int64
	if ra, err = res.RowsAffected(); err != nil {
		logerror.Println(err)
		return
	}
	if ra != 1 {
		logerror.Println(ErrorNotOneRowAffected)
		return ErrorNotOneRowAffected
	}
	return
}

func (c *Client) RemoveNuts(userID int, cnt float64) (err error) {
	query := `UPDATE [info].[Users]
		  SET Nuts = Nuts - ?
		  WHERE [UserID] = ?`
	args := []interface{}{cnt, userID}

	var res sql.Result
	if res, err = c.db.Exec(query, args...); err != nil {
		logerror.Println(err)
		return
	}

	var ra int64
	if ra, err = res.RowsAffected(); err != nil {
		logerror.Println(err)
		return
	}
	if ra != 1 {
		logerror.Println(ErrorNotOneRowAffected)
		return ErrorNotOneRowAffected
	}
	return

}

func (c *Client) SelectNuts(userID int) (nuts float64, err error) {
	query := `SELECT Nuts
		  FROM [info].[Users]
		  WHERE [UserID] = ?`
	args := []interface{}{userID}
	if err = c.db.QueryRow(query, args...).Scan(&nuts); err != nil {
		logerror.Println(err)
		return
	}
	return
}

func (c *Client) SelectTopUsersByNuts() (set []UsersRow, err error) {
	query := `SELECT TOP 5 [UserName], [Nuts]
		  FROM [info].[Users]
		  WHERE [UserName] NOT IN(?,?)
		  ORDER BY [Nuts] DESC`
	args := []interface{}{"penutty_", ""}
	rows, err := c.db.Query(query, args...)
	if err != nil {
		logerror.Println(err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var (
			userName string
			nuts     float64
		)
		if err = rows.Scan(&userName, &nuts); err != nil {
			logerror.Println(err)
			return
		}
		set = append(set, UsersRow{UserName: userName, Nuts: nuts})
	}
	if err = rows.Err(); err != nil {
		logerror.Println(err)
		return
	}
	return
}

func (c *Client) SelectUserID(username string) (userID int, err error) {
	query := `SELECT [UserID]
		  FROM [info].[Users]
		  WHERE [UserName] = ?`
	args := []interface{}{username}
	if err = c.db.QueryRow(query, args...).Scan(&userID); err != nil && err != sql.ErrNoRows {
		logerror.Println(err)
		return
	}
	return
}

func (c *Client) SelectUserName(userID int) (username string, err error) {
	query := `SELECT [UserName]
		  FROM [info].[Users]
		  WHERE [UserID] = ?`
	args := []interface{}{userID}
	if err = c.db.QueryRow(query, args...).Scan(&username); err != nil && err != sql.ErrNoRows {
		logerror.Println(err)
		return
	}
	return
}

func (c *Client) InsertRedeem(userID int, itemID int, cost float64) (err error) {
	query := `INSERT INTO [info].[Redems] ([UserID], [NutCost], [ItemID])
		  VALUES ( ?, ?, ?)`
	args := []interface{}{userID, cost, itemID}

	var res sql.Result
	if res, err = c.db.Exec(query, args...); err != nil {
		logerror.Println(err)
		return
	}

	var ra int64
	if ra, err = res.RowsAffected(); err != nil {
		logerror.Println(err)
		return
	}
	if ra != 1 {
		logerror.Println(ErrorNotOneRowAffected)
		return ErrorNotOneRowAffected
	}
	return
}

func (c *Client) SelectSubStatus(userID int) (status bool, err error) {
	query := `SELECT [HasSubbed]
		  FROM [info].[Users]
		  WHERE [UserID] = ?`
	args := []interface{}{userID}
	if err = c.db.QueryRow(query, args...).Scan(&status); err != nil {
		logerror.Println(err)
		return
	}
	return
}

func (c *Client) UpdateSubStatus(userID int) (err error) {
	query := `UPDATE [info].[Users]
		  SET [HasSubbed] = 1
		  WHERE [UserID] = ?`
	args := []interface{}{userID}

	var res sql.Result
	if res, err = c.db.Exec(query, args...); err != nil {
		logerror.Println(err)
		return
	}

	var ra int64
	if ra, err = res.RowsAffected(); err != nil {
		logerror.Println(err)
		return
	}
	if ra != 1 {
		logerror.Println(ErrorNotOneRowAffected)
		return ErrorNotOneRowAffected
	}
	return
}

func (c *Client) UpdateQuote(userID int, quote string) (err error) {
	query := `UPDATE [info].[Users]
		  SET [Quote] = ?
		  WHERE [UserID] = ?`
	args := []interface{}{quote, userID}

	var res sql.Result
	if res, err = c.db.Exec(query, args...); err != nil {
		logerror.Println(err)
		return
	}

	var ra int64
	if ra, err = res.RowsAffected(); err != nil {
		logerror.Println(err)
		return
	}
	if ra != 1 {
		logerror.Println(ErrorNotOneRowAffected)
		return ErrorNotOneRowAffected
	}
	return
}

func (c *Client) SelectQuote(userName string) (quote string, err error) {
	query := `SELECT [Quote]
		  FROM [info].[Users]
		  WHERE [UserName] = ?`
	args := []interface{}{userName}
	if err = c.db.QueryRow(query, args...).Scan(&quote); err != nil {
		logerror.Println(err)
		return
	}
	return
}
