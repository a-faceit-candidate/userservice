package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/a-faceit-candidate/userservice/internal/log"
	"github.com/a-faceit-candidate/userservice/internal/model"
	"github.com/go-sql-driver/mysql"
	"github.com/huandu/go-sqlbuilder"
)

const table = "user"

const (
	mysqlDuplicateEntryErrorCode = 1062
)

// MysqlRepository provides the mysql repository implementation
type MysqlRepository struct {
	db *sql.DB
}

func NewMysqlRepository(db *sql.DB) *MysqlRepository {
	return &MysqlRepository{
		db: db,
	}
}

func (r *MysqlRepository) Create(ctx context.Context, user *model.User) error {
	query, args := sqlStruct.InsertInto(table, userToSQL(user)).Build()
	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		if mysqlErr := (&mysql.MySQLError{}); errors.As(err, &mysqlErr) && mysqlErr.Number == mysqlDuplicateEntryErrorCode {
			return ErrConflict
		}
		return err
	}

	return nil
}

func (r *MysqlRepository) Update(ctx context.Context, user *model.User, prevUpdatedAt time.Time) error {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted, // READ_COMMITTED is the default one for mysql, although it doesn't make much difference in our usecase
		ReadOnly:  false,
	})
	if err != nil {
		return fmt.Errorf("can't start mysql transaction: %w", err)
	}
	// rollback just in case we didn't commit
	defer rollbackTx(ctx, tx)

	query := fmt.Sprintf("SELECT created_at, updated_at FROM %s WHERE id = ? FOR UPDATE", table)
	args := []interface{}{user.ID}

	var dbCreatedAt, dbUpdatedAt time.Time
	err = r.db.QueryRowContext(ctx, query, args...).Scan(&dbCreatedAt, &dbUpdatedAt)
	if err == sql.ErrNoRows {
		return ErrNotFound
	} else if err != nil {
		return fmt.Errorf("can't query for update: %w", err)
	}

	if dbCreatedAt != user.CreatedAt || dbUpdatedAt != prevUpdatedAt {
		return ErrConflict
	}

	sb := sqlStruct.Update(table, userToSQL(user))
	query, args = sb.Where(sb.Equal("id", user.ID)).Build()

	if _, err := r.db.ExecContext(ctx, query, args...); err != nil {
		return fmt.Errorf("can't update: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("can't commit update: %w", err)
	}
	return nil
}

func rollbackTx(ctx context.Context, tx *sql.Tx) {
	if err := tx.Rollback(); err != sql.ErrTxDone && err != nil {
		log.For(ctx).Warningf("Couldn't rollback transaction: %s", err)
	}
}

func (r *MysqlRepository) Get(ctx context.Context, id string) (*model.User, error) {
	sb := sqlStruct.SelectFrom(table)
	query, args := sb.Where(sb.Equal("id", id)).Build()

	u := new(sqlUser)
	err := r.db.QueryRowContext(ctx, query, args...).Scan(sqlStruct.Addr(u)...)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, fmt.Errorf("can't select: %w", err)
	}

	return sqlToUser(u), nil
}

func (r *MysqlRepository) Delete(ctx context.Context, id string) error {
	sb := sqlStruct.DeleteFrom(table)
	query, args := sb.Where(sb.Equal("id", id)).Build()

	res, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("can't delete: %w", err)
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("can't determine rows affected: %w", err)
	}

	if affected == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *MysqlRepository) ListAll(ctx context.Context) ([]*model.User, error) {
	sb := sqlStruct.SelectFrom(table)
	sb = sb.OrderBy("id").Asc()
	return r.list(ctx, sb)
}

func (r *MysqlRepository) ListCountry(ctx context.Context, countryCode string) ([]*model.User, error) {
	sb := sqlStruct.SelectFrom(table)
	sb = sb.Where(sb.Equal("country", countryCode))
	sb = sb.OrderBy("id").Asc()
	return r.list(ctx, sb)
}

func (r *MysqlRepository) list(ctx context.Context, builder *sqlbuilder.SelectBuilder) ([]*model.User, error) {
	query, args := builder.Build()

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("can't query rows: %w", err)
	}
	defer rows.Close()

	u := new(sqlUser)
	addrs := sqlStruct.Addr(u)
	var users []*model.User
	for rows.Next() {
		if err := rows.Scan(addrs...); err != nil {
			return nil, fmt.Errorf("can't scan row: %w", err)
		}
		users = append(users, sqlToUser(u))
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("couldn't get all rows: %w", err)
	}
	return users, nil
}

var sqlStruct = sqlbuilder.NewStruct(new(sqlUser))

type sqlUser struct {
	ID        string    `db:"id"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
	Name      string    `db:"name"`
	Email     string    `db:"email"`
	Country   string    `db:"country"`
}

func userToSQL(u *model.User) *sqlUser {
	return &sqlUser{
		ID:        u.ID,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
		Name:      u.Name,
		Email:     u.Email,
		Country:   u.Country,
	}
}

func sqlToUser(sq *sqlUser) *model.User {
	return &model.User{
		ID:        sq.ID,
		CreatedAt: sq.CreatedAt,
		UpdatedAt: sq.UpdatedAt,
		Name:      sq.Name,
		Email:     sq.Email,
		Country:   sq.Country,
	}
}
