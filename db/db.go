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
	loginfo *log.Logger
	logerr  *log.Logger
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

type DbRunner interface {
	QueryRow(string, ...interface{}) *sql.Row
	Query(string, ...interface{}) (*sql.Rows, error)
	Exec(string, ...interface{}) (sql.Result, error)
}

type DatabaseAccess struct {
	dbconn *sql.DB
}

type UsersRow struct {
	UserID   int
	UserName string
	Nuts     float64
}

func NewDatabaseAccess() (*DatabaseAccess, error) {
	server := "nuttydb.database.windows.net"
	port := "1433"
	username := "yogiibot"
	pass := "tIrjONIN4gtKRaJ5SHtN"
	database := "NuttyDB"

	connstr := fmt.Sprintf("server=%s;user id=%s;password=%s;port=%s;database=%s", server, username, pass, port, database)

	db, err := sql.Open("mssql", connstr)
	if err != nil {
		logerror.Println(err)
		return nil, err
	}
	dba := new(DatabaseAccess)
	dba.dbconn = db
	return dba, nil
}

func CloseDatabaseAccess(dba *DatabaseAccess) {
	dba.dbconn.Close()
}

func UserNameExists(db DbRunner, name string) (ok bool, err error) {
	query := `SELECT CASE
			  WHEN [UserName] IS NOT NULL THEN 1
			  ELSE 0
			 END
		  FROM [info].[Users]
		  WHERE [UserName] = ?`
	args := []interface{}{name}

	if err := db.QueryRow(query, args...).Scan(&ok); err != nil && err != sql.ErrNoRows {
		logerror.Println(err)
		return
	}
	return
}

func UserIDExists(db DbRunner, userID int) (ok bool, err error) {
	query := `SELECT CASE
			  WHEN [UserID] IS NOT NULL THEN 1
			  ELSE 0
			 END
		  FROM [info].[Users]
		  WHERE [UserID] = ?`
	args := []interface{}{userID}

	if err := db.QueryRow(query, args...).Scan(&ok); err != nil && err != sql.ErrNoRows {
		logerror.Println(err)
		return
	}
	return
}

var ErrorNotOneRowAffected = errors.New("sql.Result indicates 0 rows affected by query.")

func CreateUser(db DbRunner, name string, userID int) (err error) {
	query := `INSERT INTO [info].[Users] ([Username], [UserID])
		  VALUES(?, ?)`
	args := []interface{}{name, userID}

	var res sql.Result
	if res, err = db.Exec(query, args...); err != nil {
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

func UpdateUserName(db DbRunner, userID int, userName string) (err error) {
	query := `UPDATE [info].[Users]
		  SET [UserName] = ?
		  WHERE [UserID] = ?`
	args := []interface{}{userName, userID}

	var res sql.Result
	if res, err = db.Exec(query, args...); err != nil {
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

func ReferenceExists(db DbRunner, userID int) (ok bool, err error) {
	query := `SELECT CASE
			  WHEN [UserID] IS NOT NULL THEN 1
			  ELSE 0
			 END
		  FROM [info].[References]
		  WHERE [UserID] = ?`
	args := []interface{}{userID}
	if err := db.QueryRow(query, args...).Scan(&ok); err != nil && err != sql.ErrNoRows {
		logerror.Println(err)
		return
	}
	return
}

func CreateReference(db DbRunner, userID, referencedByUserID int) (err error) {
	query := `INSERT INTO [info].[References] ([UserID], [ReferencedByUserID])
		  VALUES(?, ?)`
	args := []interface{}{userID, referencedByUserID}

	var res sql.Result
	if res, err = db.Exec(query, args...); err != nil {
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

func AddNuts(db DbRunner, userID int, cnt float64) (err error) {
	query := `UPDATE [info].[Users]
		  SET Nuts = Nuts + ?
		  WHERE [UserID] = ?`
	args := []interface{}{cnt, userID}

	var res sql.Result
	if res, err = db.Exec(query, args...); err != nil {
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

func RemoveNuts(db DbRunner, userID int, cnt float64) (err error) {
	query := `UPDATE [info].[Users]
		  SET Nuts = Nuts - ?
		  WHERE [UserID] = ?`
	args := []interface{}{cnt, userID}

	var res sql.Result
	if res, err = db.Exec(query, args...); err != nil {
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

func SelectNuts(db DbRunner, userID int) (nuts float64, err error) {
	query := `SELECT Nuts
		  FROM [info].[Users]
		  WHERE [UserID] = ?`
	args := []interface{}{userID}
	if err = db.QueryRow(query, args...).Scan(&nuts); err != nil {
		logerror.Println(err)
		return
	}
	return
}

func SelectTopUsersByNuts(db DbRunner) (set []UsersRow, err error) {
	query := `SELECT TOP 5 [UserName], [Nuts]
		  FROM [info].[Users]
		  WHERE [UserName] NOT IN(?,?)
		  ORDER BY [Nuts] DESC`
	args := []interface{}{"penutty_", ""}
	rows, err := db.Query(query, args...)
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

func SelectUserID(db DbRunner, username string) (userID int, err error) {
	query := `SELECT [UserID]
		  FROM [info].[Users]
		  WHERE [UserName] = ?`
	args := []interface{}{username}
	if err = db.QueryRow(query, args...).Scan(&userID); err != nil && err != sql.ErrNoRows {
		logerror.Println(err)
		return
	}
	return
}

func SelectUserName(db DbRunner, userID int) (username string, err error) {
	query := `SELECT [UserName]
		  FROM [info].[Users]
		  WHERE [UserID] = ?`
	args := []interface{}{userID}
	if err = db.QueryRow(query, args...).Scan(&username); err != nil && err != sql.ErrNoRows {
		logerror.Println(err)
		return
	}
	return
}

func InsertRedeem(db DbRunner, userID int, itemID int, cost float64) (err error) {
	query := `INSERT INTO [info].[Redems] ([UserID], [NutCost], [ItemID])
		  VALUES ( ?, ?, ?)`
	args := []interface{}{userID, cost, itemID}

	var res sql.Result
	if res, err = db.Exec(query, args...); err != nil {
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

func SelectSubStatus(db DbRunner, userID int) (status bool, err error) {
	query := `SELECT [HasSubbed]
		  FROM [info].[Users]
		  WHERE [UserID] = ?`
	args := []interface{}{userID}
	if err = db.QueryRow(query, args...).Scan(&status); err != nil {
		logerror.Println(err)
		return
	}
	return
}

func UpdateSubStatus(db DbRunner, userID int) (err error) {
	query := `UPDATE [info].[Users]
		  SET [HasSubbed] = 1
		  WHERE [UserID] = ?`
	args := []interface{}{userID}

	var res sql.Result
	if res, err = db.Exec(query, args...); err != nil {
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

func UpdateQuote(db DbRunner, userID int, quote string) (err error) {
	query := `UPDATE [info].[Users]
		  SET [Quote] = ?
		  WHERE [UserID] = ?`
	args := []interface{}{quote, userID}

	var res sql.Result
	if res, err = db.Exec(query, args...); err != nil {
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

func SelectQuote(db DbRunner, userName string) (quote string, err error) {
	query := `SELECT [Quote]
		  FROM [info].[Users]
		  WHERE [UserName] = ?`
	args := []interface{}{userName}
	if err = db.QueryRow(query, args...).Scan(&quote); err != nil {
		logerror.Println(err)
		return
	}
	return
}
