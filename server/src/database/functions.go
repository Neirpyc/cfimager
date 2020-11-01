package database

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
)

type QueriesFunctions struct {
	SetSource           *sql.Stmt
	GetSource           *sql.Stmt
	SetCompiledAndError *sql.Stmt
	GetWasm             *sql.Stmt
	GetJs               *sql.Stmt
	GetWJS              *sql.Stmt
	GetSourceAndError   *sql.Stmt
	SetName             *sql.Stmt
	SetNameAndSource    *sql.Stmt
}

func (c *cfimager) createFunctionsTable() (err error) {
	if _, err = c.db.Exec(`create table if not exists cfimager.functions(
	ownerId BIGINT not null, 
	id BIGINT auto_increment,
	name VARCHAR(128) not null,
	source_gzip MEDIUMBLOB  default "",
	error_gzip MEDIUMBLOB default "",
	cfimager_wasm_gzip MEDIUMBLOB default "",
	cfimager_js_gzip MEDIUMBLOB default "",
	cfimager_worker_js_gzip MEDIUMBLOB default "",
	size BIGINT unsigned default 0,
	primary key (ownerId, name),
    unique key (id)
	);`); err != nil {
		return
	}
	return
}

func (c *cfimager) prepareFunctionsQueries() (err error) {
	if c.Queries.Functions.SetSource, err = c.db.Prepare(`update cfimager.functions set source_gzip = ? where ownerId = ? and id = ?`); err != nil {
		return
	}
	if c.Queries.Functions.GetSource, err = c.db.Prepare(`select source_gzip from cfimager.functions where ownerId = ? and id = ?`); err != nil {
		return
	}
	if c.Queries.Functions.SetCompiledAndError, err = c.db.Prepare(
		`update cfimager.functions set cfimager_wasm_gzip = ?, cfimager_js_gzip = ?, cfimager_worker_js_gzip = ?, error_gzip = ? where ownerId = ? and id = ?`,
	); err != nil {
		return
	}
	if c.Queries.Functions.GetWasm, err = c.db.Prepare(`select cfimager_wasm_gzip from cfimager.functions where ownerId = ? and id = ?`); err != nil {
		return
	}
	if c.Queries.Functions.GetJs, err = c.db.Prepare(`select cfimager_js_gzip from cfimager.functions where ownerId = ? and id = ?`); err != nil {
		return
	}
	if c.Queries.Functions.GetWJS, err = c.db.Prepare(`select cfimager_worker_js_gzip from cfimager.functions where ownerId = ? and id = ?`); err != nil {
		return
	}
	if c.Queries.Functions.GetSourceAndError, err = c.db.Prepare(`select source_gzip, error_gzip from cfimager.functions where ownerId = ? and id = ?`); err != nil {
		return
	}
	if c.Queries.Functions.SetName, err = c.db.Prepare(`update cfimager.functions set name = ? where ownerId = ? and id = ?`); err != nil {
		return
	}
	if c.Queries.Functions.SetNameAndSource, err = c.db.Prepare(`update cfimager.functions set name = ?, source_gzip = ? where ownerId = ? and id = ?`); err != nil {
		return
	}
	return
}

func (c *cfimager) SetSource(ownerId, funcId uint64, source string) (err error) {
	var sourceGzip bytes.Buffer

	ctx := context.Background()

	var tx *sql.Tx
	if tx, err = c.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
		ReadOnly:  false,
	}); err != nil {
		return err
	}

	var oldSize, oldTotalSize int
	if err = tx.QueryRowContext(
		ctx,
		`select OCTET_LENGTH(source_gzip), size from cfimager.functions where ownerId = ? and id = ?`,
		ownerId, funcId,
	).Scan(&oldSize, &oldTotalSize); err != nil {
		_ = tx.Rollback()
		return
	}

	var userSize, userCap int
	if err = tx.QueryRowContext(ctx, `select used, capacity from cfimager.users where id = ?`, ownerId).Scan(&userSize, &userCap); err != nil {
		_ = tx.Rollback()
		return
	}

	if err = zip(&sourceGzip, []byte(source)); err != nil {
		_ = tx.Rollback()
		return
	}
	sizeDiff := len(sourceGzip.Bytes()) - oldSize
	newSize := userSize + sizeDiff
	if newSize > userCap {
		_ = tx.Rollback()
		return
	}
	if _, err = tx.ExecContext(ctx, `update cfimager.users set used = ? where id = ?`, newSize, ownerId); err != nil {
		_ = tx.Rollback()
		return
	}
	if _, err = tx.ExecContext(
		ctx,
		`update cfimager.functions set source_gzip = ?, size = ? where ownerId = ? and id = ? `,
		sourceGzip.Bytes(),
		oldTotalSize+sizeDiff,
		ownerId,
		funcId,
	); err != nil {
		_ = tx.Rollback()
		return
	}

	if err = tx.Commit(); err != nil {
		_ = tx.Rollback()
		return
	}
	return err
}

func (c *cfimager) GetSource(ownerId, funcId uint64) (source string, err error) {
	var sourceBytes []byte
	if err = c.Queries.Functions.GetSource.QueryRow(ownerId, funcId).Scan(&sourceBytes); err != nil {
		return
	}
	if sourceBytes, err = unzip(bytes.NewBuffer(sourceBytes)); err != nil {
		return
	}
	return string(sourceBytes), nil
}

func (c *cfimager) SetCompiledAndError(ownerId, funcId uint64, wasm, js, wjs []byte, errStr string) (err error) {
	var wasmGzip, jsGzip, wjsGzip, errGzip bytes.Buffer

	ctx := context.Background()

	var tx *sql.Tx
	if tx, err = c.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
		ReadOnly:  false,
	}); err != nil {
		return err
	}

	var oldSizeWasm, oldSizeJs, oldSizeWjs, oldSizeErr, oldTotalSize int
	if err = tx.QueryRowContext(
		ctx,
		`select OCTET_LENGTH(cfimager_wasm_gzip), OCTET_LENGTH(cfimager_js_gzip), OCTET_LENGTH(cfimager_worker_js_gzip), OCTET_LENGTH(error_gzip), size from cfimager.functions where ownerId = ? and id = ?`,
		ownerId, funcId,
	).Scan(&oldSizeWasm, &oldSizeJs, &oldSizeWjs, &oldSizeErr, &oldTotalSize); err != nil {
		_ = tx.Rollback()
		return
	}

	var userSize, userCap int
	if err = tx.QueryRowContext(ctx, `select used, capacity from cfimager.users where id = ?`, ownerId).Scan(&userSize, &userCap); err != nil {
		_ = tx.Rollback()
		return
	}

	if err = zip(&wasmGzip, wasm); err != nil {
		return
	}
	if err = zip(&jsGzip, js); err != nil {
		return
	}
	if err = zip(&wjsGzip, wjs); err != nil {
		return
	}
	if err = zip(&errGzip, []byte(errStr)); err != nil {
		return
	}

	oldSum := oldSizeWasm + oldSizeJs + oldSizeWjs + oldSizeErr
	newSum := wasmGzip.Len() + jsGzip.Len() + wjsGzip.Len() + errGzip.Len()
	diff := newSum - oldSum
	newSize := oldTotalSize + diff

	newUserSize := userSize + diff
	if newUserSize > userCap {
		_ = tx.Rollback()
		return errors.New("ERR_NOT_ENOUGH_SPACE")
	}

	if _, err = tx.ExecContext(ctx, `update cfimager.users set used = ? where id = ?`, newUserSize, ownerId); err != nil {
		_ = tx.Rollback()
		return
	}

	if _, err = tx.ExecContext(
		ctx,
		`update cfimager.functions set cfimager_wasm_gzip = ?, cfimager_js_gzip = ?, cfimager_worker_js_gzip = ?, error_gzip = ?, size = ? where ownerId = ? and id = ?`,
		wasmGzip.Bytes(),
		jsGzip.Bytes(),
		wjsGzip.Bytes(),
		errGzip.Bytes(),
		newSize,
		ownerId,
		funcId,
	); err != nil {
		_ = tx.Rollback()
		return
	}

	if err = tx.Commit(); err != nil {
		_ = tx.Rollback()
		return
	}
	return err
}

func (c *cfimager) GetWasm(ownerId, funcId uint64) (data []byte, err error) {
	if err = c.Queries.Functions.GetWasm.QueryRow(ownerId, funcId).Scan(&data); err != nil {
		return
	}
	return unzip(bytes.NewBuffer(data))
}

func (c *cfimager) GetJS(ownerId, funcId uint64) (data []byte, err error) {
	if err = c.Queries.Functions.GetJs.QueryRow(ownerId, funcId).Scan(&data); err != nil {
		return
	}
	return unzip(bytes.NewBuffer(data))
}

func (c *cfimager) GetWJS(ownerId, funcId uint64) (data []byte, err error) {
	if err = c.Queries.Functions.GetWJS.QueryRow(ownerId, funcId).Scan(&data); err != nil {
		return
	}
	return unzip(bytes.NewBuffer(data))
}

func (c *cfimager) GetSourceAndError(ownerId, funcId uint64) (source, errStr string, err error) {
	var sourceBytes, errBytes []byte
	if err = c.Queries.Functions.GetSourceAndError.QueryRow(
		ownerId,
		funcId,
	).Scan(&sourceBytes, &errBytes); err != nil {
		return
	}
	if len(sourceBytes) != 0 {
		if sourceBytes, err = unzip(bytes.NewBuffer(sourceBytes)); err != nil {
			return
		}
	}
	if len(errBytes) != 0 {
		if errBytes, err = unzip(bytes.NewBuffer(errBytes)); err != nil {
			return
		}
	}
	return string(sourceBytes), string(errBytes), nil
}

func (c *cfimager) SetName(ownerId, funcId uint64, name string) (err error) {
	_, err = c.Queries.Functions.SetName.Exec(name, ownerId, funcId)
	return
}

func (c *cfimager) SetSourceAndName(ownerId, funcId uint64, source, name string) (err error) {
	ctx := context.Background()

	var tx *sql.Tx
	if tx, err = c.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelWriteCommitted,
		ReadOnly:  false,
	}); err != nil {
		return
	}

	if _, err := tx.ExecContext(ctx, `update cfimager.functions set name = ? where ownerId = ? and id = ?`, ownerId, funcId, name); err != nil {
		_ = tx.Rollback()
		return err
	}
	if err := c.SetSource(ownerId, funcId, source); err != nil {
		_ = tx.Rollback()
		return err
	}

	if err = tx.Commit(); err != nil {
		_ = tx.Rollback()
		return
	}
	return
}

func (c *cfimager) NewFunction(ownerId uint64, name string) (id uint64, err error) {
	ctx := context.Background()

	var tx *sql.Tx
	if tx, err = c.db.BeginTx(ctx, nil); err != nil {
		return
	}

	if _, err = tx.ExecContext(ctx, `insert into cfimager.functions (ownerId, name) values(?, ?)`, ownerId, name); err != nil {
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

type Functions struct {
	Functions []Function
	Total     uint64
	Capacity  uint64
	Low, High uint64
}

type Function struct {
	Id   uint64
	Name string
	Size uint64
}

func (c *cfimager) ListFunctions(ownerId uint64) (funcs Functions, err error) {
	var rows *sql.Rows

	ctx := context.Background()

	var tx *sql.Tx
	if tx, err = c.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
		ReadOnly:  false,
	}); err != nil {
		return
	}

	if rows, err = tx.QueryContext(ctx, `select name, id, size from cfimager.functions where ownerId = ?`, ownerId); err != nil {
		_ = tx.Rollback()
		return
	}
	var f Function
	for rows.Next() {
		if err = rows.Scan(&f.Name, &f.Id, &f.Size); err != nil {
			return
		}
		funcs.Functions = append(funcs.Functions, f)
	}
	if err = tx.QueryRowContext(ctx, `select used, capacity from cfimager.users where id = ?`, ownerId).Scan(&funcs.Total, &funcs.Capacity); err != nil {
		_ = tx.Rollback()
		return
	}
	funcs.Low = funcs.Capacity / 10
	funcs.High = funcs.Capacity - funcs.Capacity/4
	if err = tx.Commit(); err != nil {
		_ = tx.Rollback()
		return
	}
	return
}

func (c *cfimager) DeleteFunction(ownerId, funcId uint64) (err error) {

	ctx := context.Background()

	var tx *sql.Tx
	if tx, err = c.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
		ReadOnly:  false,
	}); err != nil {
		return err
	}

	var oldSize uint64
	if err = tx.QueryRowContext(ctx, `select size from cfimager.functions where ownerId = ? and id = ?`,
		ownerId, funcId).Scan(&oldSize); err != nil {
		_ = tx.Rollback()
		return
	}

	if _, err = tx.ExecContext(ctx, `delete from cfimager.functions where ownerId = ? and id = ?`,
		ownerId, funcId); err != nil {
		_ = tx.Rollback()
		return
	}

	var userSize uint64
	if err = tx.QueryRowContext(ctx, `select used from cfimager.users where id = ?`, ownerId).Scan(&userSize); err != nil {
		_ = tx.Rollback()
		return
	}

	newSize := userSize - oldSize

	if _, err = tx.ExecContext(ctx, `update cfimager.users set used = ? where id = ?`, newSize, ownerId); err != nil {
		_ = tx.Rollback()
		return
	}

	if err = tx.Commit(); err != nil {
		_ = tx.Rollback()
		return
	}
	return err
}
