package database

import (
	"bytes"
	"database/sql"
	"errors"
	"time"
)

type QueriesAuthTokens struct {
	GetToken      *sql.Stmt
	CreateToken   *sql.Stmt
	DeleteExpired *sql.Stmt
	RevokeForUser *sql.Stmt
	DeleteToken   *sql.Stmt
}

func (c *cfimager) tokenCleanUp() {
	_, _ = c.Queries.AuthTokens.DeleteExpired.Exec(time.Now())
}

func (c *cfimager) createAuthTable() (err error) {
	if _, err = c.db.Exec(`	create table if not exists cfimager.authTokens	(
		token BINARY(64) not null,
		owner BIGINT not null,
	    expire BIGINT not null,
	    primary key (token, expire)
	)`); err != nil {
		return
	}
	return
}

func (c *cfimager) prepareAuthQueries() (err error) {
	if c.Queries.AuthTokens.CreateToken, err = c.db.Prepare(`insert into cfimager.authTokens (token, owner, expire) values (?, ?, ?)`); err != nil {
		return
	}
	if c.Queries.AuthTokens.GetToken, err = c.db.Prepare(`select token, expire from cfimager.authTokens where token = ?`); err != nil {
		return
	}
	if c.Queries.AuthTokens.DeleteExpired, err = c.db.Prepare(`delete from cfimager.authTokens where expire < ?`); err != nil {
		return
	}
	if c.Queries.AuthTokens.RevokeForUser, err = c.db.Prepare(`delete from cfimager.authTokens where owner = ?`); err != nil {
		return
	}
	if c.Queries.AuthTokens.DeleteToken, err = c.db.Prepare(`delete from cfimager.authTokens where token = ?`); err != nil {
		return
	}
	return
}

var invalidToken = "invalid token"

func (c *cfimager) RetrieveAuthToken(token []byte) error {
	var expected []byte
	var expire int64
	if err := c.Queries.AuthTokens.GetToken.QueryRow(token).Scan(&expected, &expire); err != nil {
		return err
	}
	if expire > time.Now().Unix() {
		return nil
	}
	if bytes.Equal(token, expected) {
		return nil
	}
	return errors.New(invalidToken)
}

func (c *cfimager) RegisterAuthToken(owner uint64, token []byte) (err error) {
	if _, err = c.Queries.AuthTokens.CreateToken.Exec(token, owner, time.Now().Unix()+c.refreshTokenExpire); err != nil {
		return
	}
	return
}

func (c *cfimager) RevokeUsersTokens(userId uint64) (err error) {
	_, err = c.Queries.AuthTokens.RevokeForUser.Exec(userId)
	return
}

func (c *cfimager) DeleteAuthToken(token []byte) (err error) {
	_, err = c.Queries.AuthTokens.DeleteToken.Exec(token)
	return
}
