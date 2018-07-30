package commands

import (
	"errors"
	yogiDB "github.com/penutty/YogiiBot/db"
	"github.com/stretchr/testify/assert"
	"os"
	"strconv"
	"testing"
)

const (
	id0 = 1
	id1 = 2

	name0 = "name_0"
	name1 = "name_1"

	nut0 = 10.00

	quote0 = "quote_0"

	intError    = 1000
	stringError = "error"
)

func TestMain(m *testing.M) {
	yogiDBClient = new(mockClient)
	call := m.Run()
	os.Exit(call)
}

func Test_NewUser(t *testing.T) {
	type test struct {
		m        map[string]string
		expected error
	}
	makeMap := func(id int, name string, broadcaster, mod, subscriber int) map[string]string {
		m := make(map[string]string)
		if id > 0 {
			m["user-id"] = strconv.Itoa(id)
		}
		if name != "" {
			m["display-name"] = name
		}
		m["broadcaster"] = strconv.Itoa(broadcaster)
		m["mod"] = strconv.Itoa(mod)
		m["subscriber"] = strconv.Itoa(subscriber)

		return m
	}
	tests := []test{
		test{makeMap(id0, name0, 0, 0, 0), nil},
		test{makeMap(id0, name0, 1, 1, 1), nil},
		test{makeMap(0, name0, 0, 0, 0), ErrInvalidIdentifiers},
		test{makeMap(id0, "", 0, 0, 0), ErrInvalidIdentifiers},
	}

	for i, te := range tests {
		t.Run("test_"+strconv.Itoa(i), func(t *testing.T) {
			_, err := NewUser(te.m)
			assert.Equal(t, te.expected, err)
		})
	}
}

var (
	linePass      = `@badges=broadcaster/1,subscriber/0,premium/1;color=#B42222;display-name=penutty;emotes=;id=209c8436-1962-4c13-bc42-bbe8b76a5ea7;mod=0;room-id=83245887;subscriber=1;tmi-sent-ts=1532972330183;turbo=0;user-id=83245887;user-type= :penutty!penutty@penutty.tmi.twitch.tv PRIVMSG #penutty :123`
	lineBadgeErr  = `@badges=broadcaster/1,subscriber/0;color=#B42222;display-name=penutty;emotes=;id=209c8436-1962-4c13-bc42-bbe8b76a5ea7;mod=0;room-id=83245887;subscriber=1;tmi-sent-ts=1532972330183;turbo=0;user-id=83245887;user-type= :penutty!penutty@penutty.tmi.twitch.tv PRIVMSG #penutty :123`
	lineBadgeErr2 = `@badges=broadcaster,subscriber/0,premium/1;color=#B42222;display-name=penutty;emotes=;id=209c8436-1962-4c13-bc42-bbe8b76a5ea7;mod=0;room-id=83245887;subscriber=1;tmi-sent-ts=1532972330183;turbo=0;user-id=83245887;user-type= :penutty!penutty@penutty.tmi.twitch.tv PRIVMSG #penutty :123`
	lineErr       = `@badges=broadcaster/1,subscriber/0,premium/1;color=#B42222;display-name=penutty;emotes=;id=209c8436-1962-4c13-bc42-bbe8b76a5ea7;mod;room-id=83245887;subscriber=1;tmi-sent-ts=1532972330183;turbo=0;user-id=83245887;user-type= :penutty!penutty@penutty.tmi.twitch.tv PRIVMSG #penutty :123`
)

func Test_NewCommand(t *testing.T) {
	type test struct {
		line     string
		expected error
	}

	tests := []test{
		test{linePass, nil},
		test{lineBadgeErr, ErrInvalidBadgesFormat},
		test{lineBadgeErr2, ErrInvalidBadgesFormat},
		test{lineErr, ErrInvalidLineFormat},
	}

	for i, te := range tests {
		t.Run("test_"+strconv.Itoa(i), func(t *testing.T) {
			_, err := NewCommand(te.line, "#penutty")
			assert.Equal(t, te.expected, err)
		})
	}

}

var lineBase = `@badges=broadcaster/1,subscriber/0,premium/1;color=#B42222;display-name=penutty;emotes=;id=209c8436-1962-4c13-bc42-bbe8b76a5ea7;mod=0;room-id=83245887;subscriber=1;tmi-sent-ts=1532972330183;turbo=0;user-id=83245887;user-type= :penutty!penutty@penutty.tmi.twitch.tv PRIVMSG #penutty :`

func Test_Exec(t *testing.T) {
	type test struct {
		line     string
		expected error
	}
	tests := []test{
		test{lineBase + "!nuts", nil},
		test{lineBase + "!win 10", nil},
		test{lineBase + "!lose 10", nil},
		test{lineBase + "!fortnitebet", nil},
		test{lineBase + "!fortniteendbet", nil},
		test{lineBase + "!fortnitecancelbet", nil},
		test{lineBase + "!fortniteresolvebet win", nil},
		test{lineBase + "!thanks user", nil},
		test{lineBase + "!findyogi sgdfb", nil},
		test{lineBase + "!leaderboard", nil},
		test{lineBase + "!redeem duo", nil},
		test{lineBase + "!duoqueue", nil},
		test{lineBase + "!duoremove", nil},
		test{lineBase + "!duocharge", nil},
		test{lineBase + "!duoopen", nil},
		test{lineBase + "!duoclose", nil},
		test{lineBase + "!redeem vbucks", nil},
		test{lineBase + "!quote \"a quote\"", nil},
		test{lineBase + "!penutty", nil},
		test{lineBase + "123", nil},
	}

	for i, te := range tests {
		t.Run("test_"+strconv.Itoa(i), func(t *testing.T) {
			c, err := NewCommand(te.line, "#penutty")
			assert.Nil(t, err)
			_, err = c.Exec()
			assert.Equal(t, te.expected, err)
		})
	}
}

func Test_redeemDuo(t *testing.T) {

}

var ErrMock = errors.New("mock Error")

type mockClient struct{}

func (c *mockClient) UserNameExists(u string) (bool, error) {
	if u == stringError {
		return false, ErrMock
	}
	return false, nil
}

func (c *mockClient) UserIDExists(id int) (bool, error) {
	if id == intError {
		return false, ErrMock
	}
	return false, nil
}

func (c *mockClient) CreateUser(u string, id int) error {
	if id == intError {
		return ErrMock
	}
	return nil
}

func (c *mockClient) UpdateUserName(id int, u string) error {
	if id == intError {
		return ErrMock
	}
	return nil
}

func (c *mockClient) ReferenceExists(id int) (bool, error) {
	if id == intError {
		return false, ErrMock
	}
	return true, nil
}

func (c *mockClient) CreateReference(id1 int, id2 int) error {
	if id1 == intError {
		return ErrMock
	}
	return nil
}

func (c *mockClient) AddNuts(id int, amount float64) error {
	if id == intError {
		return ErrMock
	}
	return nil
}

func (c *mockClient) RemoveNuts(id int, amount float64) error {
	if id == intError {
		return ErrMock
	}
	return nil
}

func (c *mockClient) SelectNuts(id int) (float64, error) {
	if id == intError {
		return 0.00, ErrMock
	}
	return nut0, nil
}

func (c *mockClient) SelectTopUsersByNuts() ([]yogiDB.UsersRow, error) {
	users := []yogiDB.UsersRow{
		yogiDB.UsersRow{id0, name0, nut0},
		yogiDB.UsersRow{id1, name1, nut0},
	}
	return users, nil
}

func (c *mockClient) SelectUserID(u string) (int, error) {
	if u == stringError {
		return 0, ErrMock
	}
	return id0, nil
}

func (c *mockClient) SelectUserName(id int) (string, error) {
	if id == intError {
		return "", ErrMock
	}
	return name0, nil
}

func (c *mockClient) InsertRedeem(item int, id int, amount float64) error {
	if id == intError {
		return ErrMock
	}
	return nil
}

func (c *mockClient) SelectSubStatus(id int) (bool, error) {
	if id == intError {
		return false, ErrMock
	}
	return true, nil
}

func (c *mockClient) UpdateSubStatus(id int) error {
	if id == intError {
		return ErrMock
	}
	return nil
}

func (c *mockClient) UpdateQuote(id int, quote string) error {
	if id == intError {
		return ErrMock
	}
	return nil
}

func (c *mockClient) SelectQuote(u string) (string, error) {
	if u == stringError {
		return "", ErrMock
	}
	return quote0, nil
}
