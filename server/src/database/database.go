package database

import (
	"database/sql"
	"errors"
	_ "github.com/go-sql-driver/mysql"
	"time"
)

func init() {

}

func NewCFImager(url string, refreshTokenExpire time.Duration) (CFImager, error) {
	var err error
	var c cfimager
	stop := time.After(10 * time.Second)
	for {
		if c.db, err = sql.Open("mysql", url); err == nil {
			break
		}
		select {
		case <-stop:
			return nil, errors.New("database connection timeout")
		default:
			time.Sleep(1 * time.Millisecond)
			continue
		}
	}

	c.refreshTokenExpire = int64(refreshTokenExpire.Seconds())
	if _, err = c.db.Exec(`create database if not exists cfimager`); err != nil {
		return nil, err
	}
	if err = c.createTables(); err != nil {
		return nil, err
	}
	if err = c.prepareQueries(); err != nil {
		return nil, err
	}
	ticker := time.NewTicker(30 * time.Second)
	go func() {
		for {
			<-ticker.C
			c.tokenCleanUp()
		}
	}()
	return &c, nil
}

func (c *cfimager) createTables() (err error) {
	if err = c.createUsersTable(); err != nil {
		return
	}
	if err = c.createAuthTable(); err != nil {
		return
	}
	if err = c.createEmailValidationTable(); err != nil {
		return
	}
	if err = c.createFunctionsTable(); err != nil {
		return
	}
	return
}

func (c *cfimager) prepareQueries() (err error) {
	if err = c.prepareUsersQueries(); err != nil {
		return
	}
	if err = c.prepareAuthQueries(); err != nil {
		return
	}
	if err = c.prepareEmailValidationQueries(); err != nil {
		return
	}
	if err = c.prepareFunctionsQueries(); err != nil {
		return
	}
	return
}

type Queries struct {
	AuthTokens      QueriesAuthTokens
	EmailValidation QueriesEmailValidations
	Functions       QueriesFunctions
	Users           QueriesUsers
}

type cfimager struct {
	db                 *sql.DB
	Queries            Queries
	refreshTokenExpire int64
}

type CFImager interface {
	RetrieveAuthToken(token []byte) error
	RegisterAuthToken(owner uint64, token []byte) (err error)
	RevokeUsersTokens(userId uint64) (err error)
	AddTempUser(email string, hash, salt []byte) (id uint64, err error)
	ValidateUser(id uint64) (err error)
	GetTempUserId(email string) (id uint64, err error)
	SetSource(ownerId, funcId uint64, source string) (err error)
	SetCompiledAndError(ownerId, funcId uint64, wasm, js, wjs []byte, errStr string) (err error)
	GetWasm(ownerId, funcId uint64) (data []byte, err error)
	GetJS(ownerId, funcId uint64) (data []byte, err error)
	GetWJS(ownerId, funcId uint64) (data []byte, err error)
	GetSourceAndError(ownerId, funcId uint64) (source, errStr string, err error)
	SetName(ownerId, funcId uint64, name string) (err error)
	SetSourceAndName(ownerId, funcId uint64, source, name string) (err error)
	NewFunction(ownerId uint64, name string) (id uint64, err error)
	CreateUser(email string, hash, salt []byte) (err error)
	GetUserAuthData(email string, hash, salt *[]byte) (id uint64, err error)
	GetSource(ownerId, funcId uint64) (source string, err error)
	ListFunctions(ownerId uint64) (funcs Functions, err error)
	DeleteFunction(ownerId, funcId uint64) (err error)
	GetUsedAndCapacity(userId uint64) (used, capacity uint64, err error)
	ChangePassword(userId uint64, hash, salt []byte) (err error)
	GetId(email string) (id uint64, err error)
	DeleteAuthToken(token []byte) (err error)
}
