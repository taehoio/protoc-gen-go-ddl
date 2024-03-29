// Code generated by protoc-gen-go-ddl. DO NOT EDIT.
// versions:
//  protoc-gen-go-ddl v0.0.1-alpha
//  protoc            (unknown)
// source: taehoio/ddl/services/test/v1/tmp.proto

package testv1

import (
	"context"
	"database/sql"
	"strings"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

//go:generate mockgen -package testv1 -self_package "github.com/taehoio/protoc-gen-go-ddl/gen/go/ddl/services/test/v1" -source ./tmp_dml_table_mysql.pb.go -destination ./tmp_dml_table_mysql_mock.pb.go -mock_names TableRecorder=MockTableRecorder "github.com/taehoio/protoc-gen-go-ddl/gen/go/ddl/services/test/v1" TableRecorder

const (
	tableInsertStmt = "INSERT INTO `table` " + `(
			id, created_at, updated_at, deleted_at, tmp
		) VALUES (
			?, ?, ?, ?, ?
		)
	`

	tableUpdateStmt = "UPDATE `table` SET" + `
			id = ?, created_at = ?, updated_at = ?, deleted_at = ?, tmp = ?
		WHERE
			id = ?
	`

	tableDeleteStmt = "UPDATE `table` SET" + `
            deleted_at = ?
        WHERE
            id = ?
    `
)

var (
	_ = timestamppb.Timestamp{}
)

type TableRecorder interface {
	Get(ctx context.Context, db *sql.DB, id int64) (*Table, error)
	List(ctx context.Context, db *sql.DB, paginationOpts ...PaginationOption) ([]*Table, error)
	FindByIDs(ctx context.Context, db *sql.DB, ids []int64) ([]*Table, error)
	Save(ctx context.Context, db *sql.DB, message *Table) error
	SaveTx(ctx context.Context, tx *sql.Tx, message *Table) error
	Delete(ctx context.Context, db *sql.DB, id int64) error
	DeleteTx(ctx context.Context, tx *sql.Tx, id int64) error
}

func (m *Table) Get(ctx context.Context, db *sql.DB, id int64) (*Table, error) {
	stmt, err := db.PrepareContext(ctx, "SELECT "+
		"id, created_at, updated_at, deleted_at, tmp"+
		" FROM `table` WHERE id = ? AND deleted_at IS NULL LIMIT 1")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var mm Table

	var createdAt sql.NullTime
	var updatedAt sql.NullTime
	var deletedAt sql.NullTime

	if err = stmt.QueryRowContext(ctx, id).Scan(
		&mm.Id,
		&createdAt,
		&updatedAt,
		&deletedAt,
		&mm.Tmp,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}

	if createdAt.Valid {
		mm.CreatedAt = timestamppb.New(createdAt.Time)
	}
	if updatedAt.Valid {
		mm.UpdatedAt = timestamppb.New(updatedAt.Time)
	}
	if deletedAt.Valid {
		mm.DeletedAt = timestamppb.New(deletedAt.Time)
	}

	return &mm, nil
}

func (m *Table) List(ctx context.Context, db *sql.DB, paginationOpts ...PaginationOption) ([]*Table, error) {
	q := "SELECT " +
		"id, created_at, updated_at, deleted_at, tmp" +
		" FROM `table` WHERE deleted_at IS NULL"

	var opt pagination
	for _, o := range paginationOpts {
		o(&opt)
	}
	if opt.LastID != nil {
		switch opt.Order {
		case OrderAscending:
			q += " AND id > ?"
		case OrderDescending:
			q += " AND id < ?"
		}
	}
	switch opt.Order {
	case OrderAscending:
		q += " ORDER BY id ASC"
	case OrderDescending:
		q += " ORDER BY id DESC"
	}
	if opt.Limit != nil {
		q += " LIMIT ?"
	}

	stmt, err := db.PrepareContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var args []interface{}
	if opt.LastID != nil {
		args = append(args, opt.LastID)
	}
	if opt.Limit != nil {
		args = append(args, *opt.Limit)
	}

	rows, err := stmt.QueryContext(ctx, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var arr []*Table

	for rows.Next() {
		var mm Table

		var createdAt sql.NullTime
		var updatedAt sql.NullTime
		var deletedAt sql.NullTime

		if err = rows.Scan(
			&mm.Id,
			&createdAt,
			&updatedAt,
			&deletedAt,
			&mm.Tmp,
		); err != nil {
			return nil, err
		}

		if createdAt.Valid {
			mm.CreatedAt = timestamppb.New(createdAt.Time)
		}
		if updatedAt.Valid {
			mm.UpdatedAt = timestamppb.New(updatedAt.Time)
		}
		if deletedAt.Valid {
			mm.DeletedAt = timestamppb.New(deletedAt.Time)
		}

		arr = append(arr, &mm)
	}

	return arr, nil
}

func (m *Table) FindByIDs(ctx context.Context, db *sql.DB, ids []int64) ([]*Table, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	q := "SELECT " +
		"id, created_at, updated_at, deleted_at, tmp" +
		" FROM `table` WHERE deleted_at IS NULL AND id IN ("
	for i := range ids {
		if i > 0 {
			q += ", "
		}
		q += "?"
	}
	q += ")"

	stmt, err := db.PrepareContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var args []interface{}
	for _, id := range ids {
		args = append(args, id)
	}

	rows, err := stmt.QueryContext(ctx, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var arr []*Table

	for rows.Next() {
		var mm Table

		var createdAt sql.NullTime
		var updatedAt sql.NullTime
		var deletedAt sql.NullTime

		if err = rows.Scan(
			&mm.Id,
			&createdAt,
			&updatedAt,
			&deletedAt,
			&mm.Tmp,
		); err != nil {
			return nil, err
		}

		if createdAt.Valid {
			mm.CreatedAt = timestamppb.New(createdAt.Time)
		}
		if updatedAt.Valid {
			mm.UpdatedAt = timestamppb.New(updatedAt.Time)
		}
		if deletedAt.Valid {
			mm.DeletedAt = timestamppb.New(deletedAt.Time)
		}

		arr = append(arr, &mm)
	}

	return arr, nil
}

func (m *Table) Save(ctx context.Context, db *sql.DB, message *Table) error {
	if message.Id == 0 {
		return ErrIDNotExist
	}

	shouldInsert := true
	mm, err := m.Get(ctx, db, message.Id)
	if err != nil && err != ErrNotFound {
		return err
	}
	if mm != nil {
		shouldInsert = false
	}

	if shouldInsert {
		if err := m.insert(ctx, db, message); err != nil {
			return err
		}
	} else {
		message.CreatedAt = mm.CreatedAt
		if err := m.update(ctx, db, message); err != nil {
			return err
		}
	}

	mm, err = m.Get(ctx, db, message.Id)
	if err != nil {
		return err
	}

	proto.Merge(message, mm)

	return nil
}

func (m *Table) insert(ctx context.Context, db *sql.DB, message *Table) error {
	currentAt := timestamppb.Now()

	_, err := db.ExecContext(
		ctx,
		tableInsertStmt,
		message.Id,
		currentAt.AsTime(),
		currentAt.AsTime(),
		nil,
		message.Tmp,
	)
	if err != nil {
		if strings.HasPrefix(err.Error(), "Error 1062: Duplicate entry") {
			return ErrDuplicateEntry
		}
		return err
	}

	return nil
}

func (m *Table) update(ctx context.Context, db *sql.DB, message *Table) error {
	currentAt := timestamppb.Now()

	_, err := db.ExecContext(
		ctx,
		tableUpdateStmt,
		message.Id,
		message.CreatedAt.AsTime(),
		currentAt.AsTime(),
		nil,
		message.Tmp,
		message.Id,
	)
	if err != nil {
		return err
	}

	return nil
}

func (m *Table) Delete(ctx context.Context, db *sql.DB, id int64) error {
	currentAt := timestamppb.Now()

	_, err := db.ExecContext(
		ctx,
		tableDeleteStmt,
		currentAt.AsTime(),
		id,
	)
	if err != nil {
		return err
	}

	return nil
}

func (m *Table) DeleteTx(ctx context.Context, tx *sql.Tx, id int64) error {
	currentAt := timestamppb.Now()

	_, err := tx.ExecContext(
		ctx,
		tableDeleteStmt,
		currentAt.AsTime(),
		id,
	)
	if err != nil {
		return err
	}

	return nil
}

func (m *Table) SaveTx(ctx context.Context, tx *sql.Tx, message *Table) error {
	if message.Id == 0 {
		return ErrIDNotExist
	}

	shouldInsert := true
	mm, err := m.getTx(ctx, tx, message.Id)
	if err != nil && err != ErrNotFound {
		return err
	}
	if mm != nil {
		shouldInsert = false
	}

	if shouldInsert {
		if err := m.insertTx(ctx, tx, message); err != nil {
			return err
		}
	} else {
		message.CreatedAt = mm.CreatedAt
		if err := m.updateTx(ctx, tx, message); err != nil {
			return err
		}
	}

	return nil
}

func (m *Table) insertTx(ctx context.Context, tx *sql.Tx, message *Table) error {
	currentAt := timestamppb.Now()

	_, err := tx.ExecContext(
		ctx,
		tableInsertStmt,
		message.Id,
		currentAt.AsTime(),
		currentAt.AsTime(),
		nil,
		message.Tmp,
	)
	if err != nil {
		if strings.HasPrefix(err.Error(), "Error 1062: Duplicate entry") {
			return ErrDuplicateEntry
		}
		return err
	}

	return nil
}

func (m *Table) updateTx(ctx context.Context, tx *sql.Tx, message *Table) error {
	currentAt := timestamppb.Now()

	_, err := tx.ExecContext(
		ctx,
		tableUpdateStmt,
		message.Id,
		message.CreatedAt.AsTime(),
		currentAt.AsTime(),
		nil,
		message.Tmp,
		message.Id,
	)
	if err != nil {
		return err
	}

	return nil
}

func (m *Table) getTx(ctx context.Context, tx *sql.Tx, id int64) (*Table, error) {
	stmt, err := tx.PrepareContext(ctx, "SELECT "+
		"id, created_at, updated_at, deleted_at, tmp"+
		" FROM `table` WHERE id = ? AND deleted_at IS NULL")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var mm Table

	var createdAt sql.NullTime
	var updatedAt sql.NullTime
	var deletedAt sql.NullTime

	if err = stmt.QueryRowContext(ctx, id).Scan(
		&mm.Id,
		&createdAt,
		&updatedAt,
		&deletedAt,
		&mm.Tmp,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}

	if createdAt.Valid {
		mm.CreatedAt = timestamppb.New(createdAt.Time)
	}
	if updatedAt.Valid {
		mm.UpdatedAt = timestamppb.New(updatedAt.Time)
	}
	if deletedAt.Valid {
		mm.DeletedAt = timestamppb.New(deletedAt.Time)
	}

	return &mm, nil
}
