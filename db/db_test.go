package dba

import (
	"database/sql"
	"github.com/stretchr/testify/assert"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
	"os"
	"strconv"
	"testing"
)

const (
	uname0 = "name_0"
	uname1 = "name_1"

	uid0 = iota
	uid1

	nut0 = float64(iota)
	nut1

	item0 = int(iota)

	quote0 = "quote_0"
)

func TestMain(m *testing.M) {
	call := m.Run()

	os.Exit(call)
}

func Test_UserNameExists(t *testing.T) {
	type test struct {
		rowval   uint8
		expected bool
	}

	tests := []test{
		test{1, true},
		test{0, false},
	}

	db, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer db.Close()

	sql := `^SELECT CASE
			WHEN \[UserName\] IS NOT NULL THEN 1
			ELSE 0
			END
		  FROM \[info\]\.\[Users\]
		  WHERE \[UserName\] = \?$`

	c, err := NewClient()
	c.db = db
	assert.Nil(t, err)

	for i, te := range tests {
		t.Run("test_"+strconv.Itoa(i), func(t *testing.T) {
			mock.ExpectQuery(sql).WithArgs(uname0).WillReturnRows(sqlmock.NewRows([]string{"0"}).AddRow(te.rowval))
			res, err := c.UserNameExists(uname0)
			assert.Nil(t, err)
			assert.Equal(t, te.expected, res)

			assert.Nil(t, mock.ExpectationsWereMet())
		})
	}
}

func Test_UserIDExists(t *testing.T) {
	type test struct {
		rowval   uint8
		expected bool
	}

	tests := []test{
		test{1, true},
		test{0, false},
	}

	db, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer db.Close()

	c, err := NewClient()
	c.db = db
	assert.Nil(t, err)

	sql := `^SELECT CASE
			WHEN \[UserID\] IS NOT NULL THEN 1
			ELSE 0
			END
		  FROM \[info\]\.\[Users\]
		  WHERE \[UserID\] = \?$`

	for i, te := range tests {
		t.Run("test_"+strconv.Itoa(i), func(t *testing.T) {
			mock.ExpectQuery(sql).WithArgs(uid0).WillReturnRows(sqlmock.NewRows([]string{"0"}).AddRow(te.rowval))
			res, err := c.UserIDExists(uid0)
			assert.Nil(t, err)
			assert.Equal(t, te.expected, res)

			assert.Nil(t, mock.ExpectationsWereMet())
		})
	}
}

func Test_CreateUser(t *testing.T) {
	type test struct {
		expected     error
		affectedRows int64
	}

	tests := []test{
		test{nil, 1},
		test{ErrorNotOneRowAffected, 0},
		test{ErrorNotOneRowAffected, 2},
	}

	db, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer db.Close()

	c, err := NewClient()
	c.db = db
	assert.Nil(t, err)

	sql := `^INSERT INTO \[info\]\.\[Users\] \(\[Username\], \[UserID\]\)
		VALUES\(\?, \?\)$`

	for i, te := range tests {
		t.Run("test_"+strconv.Itoa(i), func(t *testing.T) {
			mock.ExpectExec(sql).WithArgs(uname0, uid0).WillReturnResult(sqlmock.NewResult(0, te.affectedRows))

			err := c.CreateUser(uname0, uid0)
			assert.Equal(t, te.expected, err)

			assert.Nil(t, mock.ExpectationsWereMet())
		})
	}
}

func Test_UpdateUserName(t *testing.T) {
	type test struct {
		expected     error
		affectedRows int64
	}

	tests := []test{
		test{nil, 1},
		test{ErrorNotOneRowAffected, 0},
		test{ErrorNotOneRowAffected, 2},
	}

	db, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer db.Close()

	c, err := NewClient()
	c.db = db
	assert.Nil(t, err)

	sql := `^UPDATE \[info\]\.\[Users\]
		SET \[UserName\] = \?
		WHERE \[UserID\] = \?$`

	for i, te := range tests {
		t.Run("test_"+strconv.Itoa(i), func(t *testing.T) {
			mock.ExpectExec(sql).WithArgs(uname0, uid0).WillReturnResult(sqlmock.NewResult(0, te.affectedRows))

			err := c.UpdateUserName(uid0, uname0)
			assert.Equal(t, te.expected, err)

			assert.Nil(t, mock.ExpectationsWereMet())
		})
	}

}

func Test_ReferenceExists(t *testing.T) {
	type test struct {
		rowval   uint8
		expected bool
	}

	tests := []test{
		test{1, true},
		test{0, false},
	}

	db, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer db.Close()

	c, err := NewClient()
	assert.Nil(t, err)
	c.db = db

	sql := `^SELECT CASE
			WHEN \[UserID\] IS NOT NULL THEN 1
			ELSE 0
			END
		  FROM \[info\]\.\[References\]
		  WHERE \[UserID\] = \?$`

	for i, te := range tests {
		t.Run("test_"+strconv.Itoa(i), func(t *testing.T) {
			mock.ExpectQuery(sql).WithArgs(uid0).WillReturnRows(sqlmock.NewRows([]string{"0"}).AddRow(te.rowval))
			res, err := c.ReferenceExists(uid0)
			assert.Nil(t, err)
			assert.Equal(t, te.expected, res)

			assert.Nil(t, mock.ExpectationsWereMet())
		})
	}

}

func Test_CreateReference(t *testing.T) {
	type test struct {
		expected     error
		affectedRows int64
	}

	tests := []test{
		test{nil, 1},
		test{ErrorNotOneRowAffected, 0},
		test{ErrorNotOneRowAffected, 2},
	}

	db, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer db.Close()

	c, err := NewClient()
	assert.Nil(t, err)
	c.db = db

	sql := `^INSERT INTO \[info\]\.\[References\] \(\[UserID\], \[ReferencedByUserID\]\)
		VALUES\(\?, \?\)$`

	for i, te := range tests {
		t.Run("test_"+strconv.Itoa(i), func(t *testing.T) {
			mock.ExpectExec(sql).WithArgs(uid0, uid1).WillReturnResult(sqlmock.NewResult(0, te.affectedRows))

			err := c.CreateReference(uid0, uid1)
			assert.Equal(t, te.expected, err)

			assert.Nil(t, mock.ExpectationsWereMet())
		})
	}
}

func Test_AddNuts(t *testing.T) {
	type test struct {
		expected     error
		affectedRows int64
	}

	tests := []test{
		test{nil, 1},
		test{ErrorNotOneRowAffected, 0},
		test{ErrorNotOneRowAffected, 2},
	}
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer db.Close()

	c, err := NewClient()
	assert.Nil(t, err)
	c.db = db

	sql := `^UPDATE \[info\]\.\[Users\]
		SET Nuts = Nuts \+ \?
		WHERE \[UserID\] = \?$`

	for i, te := range tests {
		t.Run("test_"+strconv.Itoa(i), func(t *testing.T) {
			mock.ExpectExec(sql).WithArgs(nut0, uid0).WillReturnResult(sqlmock.NewResult(0, te.affectedRows))

			err := c.AddNuts(uid0, nut0)
			assert.Equal(t, te.expected, err)

			assert.Nil(t, mock.ExpectationsWereMet())
		})
	}
}

func Test_RemoveNuts(t *testing.T) {
	type test struct {
		expected     error
		affectedRows int64
	}

	tests := []test{
		test{nil, 1},
		test{ErrorNotOneRowAffected, 0},
		test{ErrorNotOneRowAffected, 2},
	}
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer db.Close()

	c, err := NewClient()
	assert.Nil(t, err)
	c.db = db

	sql := `^UPDATE \[info\]\.\[Users\]
		SET Nuts = Nuts - \?
		WHERE \[UserID\] = \?$`

	for i, te := range tests {
		t.Run("test_"+strconv.Itoa(i), func(t *testing.T) {
			mock.ExpectExec(sql).WithArgs(nut0, uid0).WillReturnResult(sqlmock.NewResult(0, te.affectedRows))

			err := c.RemoveNuts(uid0, nut0)
			assert.Equal(t, te.expected, err)

			assert.Nil(t, mock.ExpectationsWereMet())
		})
	}

}

func Test_SelectNuts(t *testing.T) {
	type test struct {
		expectedErr  error
		expectedNuts float64
	}

	tests := []test{
		test{sql.ErrNoRows, 0},
		test{nil, nut1},
	}

	db, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer db.Close()

	c, err := NewClient()
	assert.Nil(t, err)
	c.db = db

	sqlstr := `^SELECT Nuts
		FROM \[info\]\.\[Users\] 
		WHERE \[UserID\] = \?$`

	mock.ExpectQuery(sqlstr).WithArgs(uid0).WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(sqlstr).WithArgs(uid0).WillReturnRows(sqlmock.NewRows([]string{"Nuts"}).AddRow(nut1))

	for i, te := range tests {
		t.Run("test_"+strconv.Itoa(i), func(t *testing.T) {

			nuts, err := c.SelectNuts(uid0)
			assert.Equal(t, te.expectedErr, err)
			assert.Equal(t, te.expectedNuts, nuts)

		})
	}
	assert.Nil(t, mock.ExpectationsWereMet())
}

func Test_SelectTopUsersByNuts(t *testing.T) {
	type test struct {
		expectedErr error
	}

	tests := []test{
		test{sql.ErrNoRows},
		test{nil},
	}

	db, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer db.Close()

	c, err := NewClient()
	assert.Nil(t, err)
	c.db = db

	sqlstr := `^SELECT TOP 5 \[UserName\], \[Nuts\]
		   FROM \[info\]\.\[Users\] 
		   WHERE \[UserName\] NOT IN\(\?,\?\)
		   ORDER BY \[Nuts\] DESC$`

	mock.ExpectQuery(sqlstr).WithArgs("penutty_", "").WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(sqlstr).WithArgs("penutty_", "").WillReturnRows(sqlmock.NewRows([]string{"UserName", "Nuts"}).AddRow(uname0, nut0))

	for i, te := range tests {
		t.Run("test_"+strconv.Itoa(i), func(t *testing.T) {

			_, err := c.SelectTopUsersByNuts()
			assert.Equal(t, te.expectedErr, err)
		})
	}
	assert.Nil(t, mock.ExpectationsWereMet())

}

func Test_SelectUserID(t *testing.T) {
	type test struct {
		expectedErr    error
		expectedUserId int
	}

	tests := []test{
		test{sql.ErrNoRows, 0},
		test{nil, uid1},
	}

	db, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer db.Close()

	c, err := NewClient()
	assert.Nil(t, err)
	c.db = db

	sqlstr := `^SELECT \[UserID\]
		    FROM \[info\]\.\[Users\]
		    WHERE \[UserName\] = \?`

	mock.ExpectQuery(sqlstr).WithArgs(uname1).WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(sqlstr).WithArgs(uname1).WillReturnRows(sqlmock.NewRows([]string{"UserID"}).AddRow(uid1))

	for i, te := range tests {
		t.Run("test_"+strconv.Itoa(i), func(t *testing.T) {

			userID, err := c.SelectUserID(uname1)
			assert.Equal(t, te.expectedErr, err)
			assert.Equal(t, te.expectedUserId, userID)
		})
	}
	assert.Nil(t, mock.ExpectationsWereMet())
}

func Test_SelectUserName(t *testing.T) {
	type test struct {
		expectedErr      error
		expectedUserName string
	}

	tests := []test{
		test{sql.ErrNoRows, ""},
		test{nil, uname1},
	}

	db, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer db.Close()

	c, err := NewClient()
	assert.Nil(t, err)
	c.db = db

	sqlstr := `^SELECT \[UserName\]
		    FROM \[info\]\.\[Users\]
		    WHERE \[UserID\] = \?`

	mock.ExpectQuery(sqlstr).WithArgs(uid1).WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(sqlstr).WithArgs(uid1).WillReturnRows(sqlmock.NewRows([]string{"UserName"}).AddRow(uname1))

	for i, te := range tests {
		t.Run("test_"+strconv.Itoa(i), func(t *testing.T) {

			userName, err := c.SelectUserName(uid1)
			assert.Equal(t, te.expectedErr, err)
			assert.Equal(t, te.expectedUserName, userName)
		})
	}
	assert.Nil(t, mock.ExpectationsWereMet())
}

func Test_InsertRedeem(t *testing.T) {
	type test struct {
		expectedErr error
	}

	tests := []test{
		test{nil},
		test{ErrorNotOneRowAffected},
		test{ErrorNotOneRowAffected},
	}

	db, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer db.Close()

	c, err := NewClient()
	assert.Nil(t, err)
	c.db = db

	sqlstr := `^INSERT INTO \[info\]\.\[Redems\] \(\[UserID\], \[NutCost\], \[ItemID\]\)
		   VALUES \( \?, \?, \?\)$`

	mock.ExpectExec(sqlstr).WithArgs(uid1, nut1, item0).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(sqlstr).WithArgs(uid1, nut1, item0).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(sqlstr).WithArgs(uid1, nut1, item0).WillReturnResult(sqlmock.NewResult(0, 2))

	for i, te := range tests {
		t.Run("test_"+strconv.Itoa(i), func(t *testing.T) {

			err := c.InsertRedeem(uid1, item0, nut1)
			assert.Equal(t, te.expectedErr, err)
		})
	}
	assert.Nil(t, mock.ExpectationsWereMet())
}

func Test_SelectSubStatus(t *testing.T) {
	type test struct {
		expectedErr       error
		expectedSubStatus bool
	}

	tests := []test{
		test{sql.ErrNoRows, false},
		test{nil, true},
		test{nil, false},
	}

	db, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer db.Close()

	c, err := NewClient()
	assert.Nil(t, err)
	c.db = db

	sqlstr := `^SELECT \[HasSubbed\]
		    FROM \[info\]\.\[Users\]
		    WHERE \[UserID\] = \?$`

	mock.ExpectQuery(sqlstr).WithArgs(uid1).WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(sqlstr).WithArgs(uid1).WillReturnRows(sqlmock.NewRows([]string{"HasSubbed"}).AddRow("1"))
	mock.ExpectQuery(sqlstr).WithArgs(uid1).WillReturnRows(sqlmock.NewRows([]string{"HasSubbed"}).AddRow("0"))

	for i, te := range tests {
		t.Run("test_"+strconv.Itoa(i), func(t *testing.T) {

			status, err := c.SelectSubStatus(uid1)
			assert.Equal(t, te.expectedErr, err)
			assert.Equal(t, te.expectedSubStatus, status)
		})
	}
	assert.Nil(t, mock.ExpectationsWereMet())
}

func Test_UpdateSubStatus(t *testing.T) {
	type test struct {
		expectedErr error
	}

	tests := []test{
		test{nil},
		test{ErrorNotOneRowAffected},
		test{ErrorNotOneRowAffected},
	}

	db, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer db.Close()

	c, err := NewClient()
	assert.Nil(t, err)
	c.db = db

	sqlstr := `^UPDATE \[info\]\.\[Users\]
		  SET \[HasSubbed\] = 1
		  WHERE \[UserID\] = \?$`

	mock.ExpectExec(sqlstr).WithArgs(uid1).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(sqlstr).WithArgs(uid1).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(sqlstr).WithArgs(uid1).WillReturnResult(sqlmock.NewResult(0, 2))

	for i, te := range tests {
		t.Run("test_"+strconv.Itoa(i), func(t *testing.T) {

			err := c.UpdateSubStatus(uid1)
			assert.Equal(t, te.expectedErr, err)
		})
	}
	assert.Nil(t, mock.ExpectationsWereMet())
}

func Test_UpdateQuote(t *testing.T) {
	type test struct {
		expectedErr error
	}

	tests := []test{
		test{nil},
		test{ErrorNotOneRowAffected},
		test{ErrorNotOneRowAffected},
	}

	db, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer db.Close()

	c, err := NewClient()
	assert.Nil(t, err)
	c.db = db

	sqlstr := `^UPDATE \[info\]\.\[Users\]
		  SET \[Quote\] = \? 
		  WHERE \[UserID\] = \?$`

	mock.ExpectExec(sqlstr).WithArgs(quote0, uid1).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(sqlstr).WithArgs(quote0, uid1).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(sqlstr).WithArgs(quote0, uid1).WillReturnResult(sqlmock.NewResult(0, 2))

	for i, te := range tests {
		t.Run("test_"+strconv.Itoa(i), func(t *testing.T) {

			err := c.UpdateQuote(uid1, quote0)
			assert.Equal(t, te.expectedErr, err)
		})
	}
	assert.Nil(t, mock.ExpectationsWereMet())
}

func Test_SelectQuote(t *testing.T) {
	type test struct {
		expectedErr   error
		expectedQuote string
	}

	tests := []test{
		test{sql.ErrNoRows, ""},
		test{nil, ""},
		test{nil, quote0},
	}

	db, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer db.Close()

	c, err := NewClient()
	assert.Nil(t, err)
	c.db = db

	sqlstr := `^SELECT \[Quote\]
		    FROM \[info\]\.\[Users\]
		    WHERE \[UserName\] = \?$`

	mock.ExpectQuery(sqlstr).WithArgs(uname0).WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(sqlstr).WithArgs(uname0).WillReturnRows(sqlmock.NewRows([]string{"Quote"}).AddRow(""))
	mock.ExpectQuery(sqlstr).WithArgs(uname0).WillReturnRows(sqlmock.NewRows([]string{"Quote"}).AddRow(quote0))

	for i, te := range tests {
		t.Run("test_"+strconv.Itoa(i), func(t *testing.T) {

			quote, err := c.SelectQuote(uname0)
			assert.Equal(t, te.expectedErr, err)
			assert.Equal(t, te.expectedQuote, quote)
		})
	}
	assert.Nil(t, mock.ExpectationsWereMet())

}
