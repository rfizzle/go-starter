package user

import (
	"context"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/rfizzle/go-starter/internal/domain"
	"github.com/rfizzle/go-starter/internal/platform/database/postgres/db"
)

// psql is a squirrel builder pinned to PostgreSQL's $N placeholder style, used
// for the dynamic queries that sqlc's compile-time model cannot express.
var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

// Repository is the data-access boundary for the user domain.
//
// sqlc talks through the db.DBTX interface via queries, and squirrel executes
// against the same db.DBTX via exec. Both use the caller-owned pool by default;
// WithinTx swaps them for one pgx.Tx so every repository method participates in
// the same transaction.
type Repository struct {
	pool    *pgxpool.Pool
	exec    db.DBTX
	queries *db.Queries
}

// NewRepository builds a Repository over an existing pgx pool. The pool is owned
// by the application and closed by it; the repository does not close it.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{
		pool:    pool,
		exec:    pool,
		queries: db.New(pool),
	}
}

// ByID returns the user with the given id, or domain.ErrNotFound if none
// exists. This is the sqlc path: a fixed, compile-time-checked query.
func (r *Repository) ByID(ctx context.Context, id string) (domain.User, error) {
	uid, err := parseUUID(id)
	if err != nil {
		return domain.User{}, fmt.Errorf("parse user id: %w", err)
	}

	row, err := r.queries.GetUserByID(ctx, uid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, domain.ErrNotFound
		}
		return domain.User{}, fmt.Errorf("get user by id: %w", err)
	}

	return fromDB(row), nil
}

// ByEmail returns the user with the given email, or domain.ErrNotFound.
func (r *Repository) ByEmail(ctx context.Context, email string) (domain.User, error) {
	row, err := r.queries.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, domain.ErrNotFound
		}
		return domain.User{}, fmt.Errorf("get user by email: %w", err)
	}

	return fromDB(row), nil
}

// CredentialByEmail returns the authentication-only projection of the user
// with the given email, or domain.ErrNotFound if none exists. It is the only
// Repository method that returns a password hash; callers outside
// authentication must use ByEmail instead.
func (r *Repository) CredentialByEmail(ctx context.Context, email string) (domain.Credential, error) {
	row, err := r.queries.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Credential{}, domain.ErrNotFound
		}
		return domain.Credential{}, fmt.Errorf("get credential by email: %w", err)
	}

	return credentialFromDB(row), nil
}

// Create inserts a user and returns the stored row.
func (r *Repository) Create(ctx context.Context, params domain.CreateUserParams) (domain.User, error) {
	row, err := r.queries.CreateUser(ctx, db.CreateUserParams{
		Email:        params.Email,
		PasswordHash: params.PasswordHash,
		FirstName:    params.FirstName,
		LastName:     params.LastName,
	})
	if err != nil {
		return domain.User{}, fmt.Errorf("create user: %w", err)
	}

	return fromDB(row), nil
}

// List returns users matching filter, ordered by creation time. It is the
// squirrel path: the SQL is assembled at runtime because the set of predicates
// is not known at compile time, so sqlc cannot generate it.
func (r *Repository) List(ctx context.Context, filter domain.UserFilter) ([]domain.User, error) {
	// password_hash is intentionally omitted: List has no consumer for it, so
	// there is no reason to read credential hashes across the wire.
	builder := psql.
		Select("id", "email", "first_name", "last_name", "created_at", "updated_at").
		From("users").
		OrderBy("created_at DESC")

	if filter.EmailLike != "" {
		builder = builder.Where(sq.ILike{"email": "%" + filter.EmailLike + "%"})
	}
	if filter.Limit > 0 {
		builder = builder.Limit(filter.Limit)
	}

	sql, args, err := builder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build list query: %w", err)
	}

	rows, err := r.exec.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	// Scan into a db.User (minus the omitted password_hash) so the pgtype ->
	// domain mapping stays centralized in fromDB.
	var users []domain.User
	for rows.Next() {
		var row db.User
		if scanErr := rows.Scan(
			&row.ID,
			&row.Email,
			&row.FirstName,
			&row.LastName,
			&row.CreatedAt,
			&row.UpdatedAt,
		); scanErr != nil {
			return nil, fmt.Errorf("scan user: %w", scanErr)
		}
		users = append(users, fromDB(row))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate users: %w", err)
	}

	return users, nil
}

// WithinTx runs fn against a transaction-scoped Repository. The nested
// repository shares this one's pool but routes every query through a pgx.Tx, so
// the same ByID/Create/List code composes atomically. The transaction commits
// when fn returns nil and rolls back otherwise.
//
// The pool and transaction are interchangeable because both satisfy db.DBTX.
func (r *Repository) WithinTx(ctx context.Context, fn func(txRepo *Repository) error) (err error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	// Roll back on any early return; the commit below makes rollback a no-op.
	defer func() {
		if rbErr := tx.Rollback(ctx); rbErr != nil && !errors.Is(rbErr, pgx.ErrTxClosed) {
			err = errors.Join(err, fmt.Errorf("rollback tx: %w", rbErr))
		}
	}()

	txRepo := &Repository{
		pool:    r.pool,
		exec:    tx,
		queries: r.queries.WithTx(tx),
	}

	if err = fn(txRepo); err != nil {
		return err
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}
