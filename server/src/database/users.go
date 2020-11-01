package database

import "database/sql"

type QueriesUsers struct {
	Create          *sql.Stmt
	Get             *sql.Stmt
	GetUsedCapacity *sql.Stmt
	ChangePassword  *sql.Stmt
	GetUserId       *sql.Stmt
}

func (c *cfimager) createUsersTable() (err error) {
	if _, err = c.db.Exec(`create table if not exists cfimager.users(
    email VARCHAR(254) character set utf8 not null,
    id BIGINT unsigned auto_increment,
    hash BINARY(64) not null,
    salt BINARY(64) not null,
    capacity BIGINT unsigned default 2097152,
    used BIGINT unsigned default 0,
	primary key(email),
    unique key (id)        
);`); err != nil {
		return err
	}
	return nil
}

func (c *cfimager) prepareUsersQueries() (err error) {
	if c.Queries.Users.Create, err = c.db.Prepare(`insert into cfimager.users (email, hash, salt) values (?, ?, ?)`); err != nil {
		return
	}
	if c.Queries.Users.Get, err = c.db.Prepare(`select id, hash, salt from cfimager.users where email = ?`); err != nil {
		return
	}
	if c.Queries.Users.GetUsedCapacity, err = c.db.Prepare(`select used, capacity from cfimager.users where id = ?`); err != nil {
		return
	}
	if c.Queries.Users.ChangePassword, err = c.db.Prepare(`update cfimager.users set hash = ?, salt = ? where id = ?`); err != nil {
		return
	}
	if c.Queries.Users.GetUserId, err = c.db.Prepare(`select id from cfimager.users where email = ?`); err != nil {
		return
	}
	return
}

func (c *cfimager) CreateUser(email string, hash, salt []byte) (err error) {
	if _, err = c.Queries.Users.Create.Exec(email, hash, salt); err != nil {
		return
	}
	return
}

func (c *cfimager) GetUserAuthData(email string, hash, salt *[]byte) (id uint64, err error) {
	if err = c.Queries.Users.Get.QueryRow(email).Scan(&id, hash, salt); err != nil {
		return 0, err
	}
	return id, nil
}

func (c *cfimager) GetUsedAndCapacity(userId uint64) (used, capacity uint64, err error) {
	if err = c.Queries.Users.GetUsedCapacity.QueryRow(userId).Scan(&used, &capacity); err != nil {
		return 0, 0, err
	}
	return
}

func (c *cfimager) ChangePassword(userId uint64, hash, salt []byte) (err error) {
	if _, err = c.Queries.Users.ChangePassword.Exec(hash, salt, userId); err != nil {
		return err
	}
	return
}
func (c *cfimager) GetId(email string) (id uint64, err error) {
	if err = c.Queries.Users.GetUserId.QueryRow(email).Scan(&id); err != nil {
		return 0, err
	}
	return
}
