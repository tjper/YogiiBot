package main

import (
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"strconv"
	"strings"
	"testing"
)

func Test_Exec(t *testing.T) {

}

func Test_authHandler(t *testing.T) {
	a := new(app)
	a.c = new(MockUserClient)

	testVars := []*RequestCodePair{
		&RequestCodePair{httptest.NewRequest(http.MethodPost, AuthEndpoint, NewAuthBody(tUser, tPassword)), http.StatusOK},
		&RequestCodePair{httptest.NewRequest(http.MethodPost, AuthEndpoint, strings.NewReader("")), http.StatusBadRequest},
		&RequestCodePair{httptest.NewRequest(http.MethodPost, AuthEndpoint, NewAuthBody("fail", tPassword)), http.StatusBadRequest},
		&RequestCodePair{httptest.NewRequest(http.MethodPost, AuthEndpoint, NewAuthBody(tUser, "fail")), http.StatusBadRequest},
		&RequestCodePair{httptest.NewRequest("METHOD_DNE", AuthEndpoint, NewAuthBody(tUser, tPassword)), http.StatusNotImplemented},
	}

	for i, v := range testVars {
		rec := httptest.NewRecorder()
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			a.authHandler(rec, v.req)
			assert.Equal(t, v.code, rec.Code)
		})
	}
}
