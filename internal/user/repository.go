package user

import (
	"context"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/rfizzle/go-starter/internal/storage"
	"github.com/rfizzle/go-starter/internal/storage/postgres/db"
)

// psql is a squirrel builder pinned to PostgreSQL's $N placeholder style, used
// for the dynamic queries that sqlc's compile-time model cannot express.
var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

// Repository is the data-access boundary for the user domain.
//
// It owns the connection pool directly (the pattern chosen for M-5): sqlc talks
// through the db.DBTX interface via queries, and squirrel executes against the
// same db.DBTX via exec. Both are set to the pool by default; WithinTx swaps
// them for a pgx.Tx, so ByID/Create/List all become transactional without any
// code in this package caring which connection it holds.
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

// ByID returns the user with the given id, or storage.ErrNotFound if none
// exists. This is the sqlc path: a fixed, compile-time-checked query.
func (r *Repository) ByID(ctx context.Context, id string) (User, error) {
	uid, err := parseUUID(id)
	if err != nil {
		return User{}, fmt.Errorf("parse user id: %w", err)
	}

	row, err := r.queries.GetUserByID(ctx, uid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, storage.ErrNotFound
		}
		return User{}, fmt.Errorf("get user by id: %w", err)
	}

	return fromDB(row), nil
}

// ByEmail returns the user with the given email, or storage.ErrNotFound.
func (r *Repository) ByEmail(ctx context.Context, email string) (User, error) {
	row, err := r.queries.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, storage.ErrNotFound
		}
		return User{}, fmt.Errorf("get user by email: %w", err)
	}

	return fromDB(row), nil
}

// CreateParams is the input to Create. PasswordHash is expected to be an
// already-hashed credential; this layer never hashes.
type CreateParams struct {
	Email        string
	PasswordHash string
	FirstName    string
	LastName     string
}

// Create inserts a user and returns the stored row.
func (r *Repository) Create(ctx context.Context, params CreateParams) (User, error) {
	row, err := r.queries.CreateUser(ctx, db.CreateUserParams{
		Email:        params.Email,
		PasswordHash: params.PasswordHash,
		FirstName:    params.FirstName,
		LastName:     params.LastName,
	})
	if err != nil {
		return User{}, fmt.Errorf("create user: %w", err)
	}

	return fromDB(row), nil
}

// ListFilter narrows a List query. Every field is optional; zero values are
// omitted from the generated WHERE clause. This is what makes the query
// "dynamic" and why squirrel earns its place over sqlc here.
type ListFilter struct {
	// EmailLike, when set, matches emails with a case-insensitive ILIKE. The
	// value is used as an ILIKE pattern: '%' and '_' act as wildcards rather
	// than literal characters. The value is always bound as a parameter, so it
	// is safe from injection.
	EmailLike string
	// Limit caps the number of rows returned. Zero means no limit.
	Limit uint64
}

// List returns users matching filter, ordered by creation time. It is the
// squirrel path: the SQL is assembled at runtime because the set of predicates
// is not known at compile time, so sqlc cannot generate it.
func (r *Repository) List(ctx context.Context, filter ListFilter) ([]User, error) {
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
	var users []User
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
// This is the transaction-propagation pattern the review asked for: the pool and
// the tx are interchangeable because both satisfy db.DBTX.
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
