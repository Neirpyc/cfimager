package database

import (
	"context"
	"database/sql"
)

type QueriesEmailValidations struct {
	GetId *sql.Stmt
}

func (c *cfimager) createEmailValidationTable() (err error) {
	if _, err = c.db.Exec(`create table if not exists cfimager.emailValidations(
	email VARCHAR(254) not null unique,
	id BIGINT unsigned auto_increment,
	hash BINARY(64) not null,
	salt BINARY(64) not null,
	primary key(email),
	unique key (id)
);
`); err != nil {
		return
	}
	return
}

func (c *cfimager) prepareEmailValidationQueries() (err error) {
	if c.Queries.EmailValidation.GetId, err = c.db.Prepare(`select id from cfimager.emailValidations where email = ?`); err != nil {
		return
	}
	return
}

func (c *cfimager) AddTempUser(email string, hash, salt []byte) (id uint64, err error) {
	ctx := context.Background()

	var tx *sql.Tx
	if tx, err = c.db.BeginTx(ctx, nil); err != nil {
		return
	}

	//check for sql injection
	if _, err = tx.ExecContext(ctx, `insert into cfimager.emailValidations (email, hash, salt) values (?, ?, ?)`, email, hash, salt); err != nil {
		_ = tx.Rollback()
		return
	}

	if err = tx.QueryRowContext(ctx, `select LAST_INSERT_ID()`).Scan(&id); err != nil {
		_ = tx.Rollback()
		return
	}

	if err = tx.Commit(); err != nil {
		_ = tx.Rollback()
		return
	}

	return
}

func (c *cfimager) ValidateUser(id uint64) (err error) {
	ctx := context.Background()

	var tx *sql.Tx
	if tx, err = c.db.BeginTx(ctx, nil); err != nil {
		return err
	}

	var email string
	var hash, salt []byte
	if err = tx.QueryRowContext(ctx, `select email, hash, salt from cfimager.emailValidations where id = ?`, id).Scan(&email, &hash, &salt); err != nil {
		_ = tx.Rollback()
		return
	}

	if _, err = tx.ExecContext(ctx, `delete from cfimager.emailValidations where id = ?`, id); err != nil {
		_ = tx.Rollback()
		return
	}

	if _, err = tx.ExecContext(ctx, `insert into cfimager.users (email, hash, salt) values (?, ?, ?)`, email, hash, salt); err != nil {
		_ = tx.Rollback()
		return
	}

	if err = tx.Commit(); err != nil {
		_ = tx.Rollback()
		return
	}

	return
}

func (c *cfimager) GetTempUserId(email string) (id uint64, err error) {
	err = c.Queries.EmailValidation.GetId.QueryRow(email).Scan(&id)
	return
}
