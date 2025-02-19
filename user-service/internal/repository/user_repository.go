package repository

import (
	"context"
	"database/sql"
	"time"

	"promoservice/user-service/internal/models"

	"github.com/google/uuid"
)

type postgresUserRepository struct {
	db *sql.DB
}

func NewPostgresUserRepository(db *sql.DB) UserRepository {
	return &postgresUserRepository{db: db}
}

func (r *postgresUserRepository) CreateUser(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (id, login, password_hash, email, first_name, last_name, phone, birth_date, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at, updated_at`

	return r.db.QueryRowContext(ctx, query,
		user.ID,
		user.Login,
		user.PasswordHash,
		user.Email,
		user.FirstName,
		user.LastName,
		user.Phone,
		user.BirthDate,
		user.CreatedAt,
		user.UpdatedAt,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
}

func (r *postgresUserRepository) GetUserByLogin(ctx context.Context, login string) (*models.User, error) {
	query := `
		SELECT id, login, password_hash, email, first_name, last_name, phone, birth_date, created_at, updated_at
		FROM users
		WHERE login = $1`

	user := &models.User{}
	err := r.db.QueryRowContext(ctx, query, login).Scan(
		&user.ID,
		&user.Login,
		&user.PasswordHash,
		&user.Email,
		&user.FirstName,
		&user.LastName,
		&user.Phone,
		&user.BirthDate,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *postgresUserRepository) GetUserByEmail(ctx context.Context, login string) (*models.User, error) {
	query := `
		SELECT id, login, password_hash, email, first_name, last_name, phone, birth_date, created_at, updated_at
		FROM users
		WHERE email = $1`

	user := &models.User{}
	err := r.db.QueryRowContext(ctx, query, login).Scan(
		&user.ID,
		&user.Login,
		&user.PasswordHash,
		&user.Email,
		&user.FirstName,
		&user.LastName,
		&user.Phone,
		&user.BirthDate,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *postgresUserRepository) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	userID, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}

	query := `
		SELECT id, login, password_hash, email, first_name, last_name, phone, birth_date, created_at, updated_at
		FROM users
		WHERE id = $1`

	user := &models.User{}
	err = r.db.QueryRowContext(ctx, query, userID).Scan(
		&user.ID,
		&user.Login,
		&user.PasswordHash,
		&user.Email,
		&user.FirstName,
		&user.LastName,
		&user.Phone,
		&user.BirthDate,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *postgresUserRepository) UpdateUser(ctx context.Context, user *models.User) error {
	query := `
		UPDATE users
		SET email = $1, first_name = $2, last_name = $3, phone = $4, birth_date = $5, updated_at = $6
		WHERE id = $7
		RETURNING updated_at`

	return r.db.QueryRowContext(ctx, query,
		user.Email,
		user.FirstName,
		user.LastName,
		user.Phone,
		user.BirthDate,
		time.Now(),
		user.ID,
	).Scan(&user.UpdatedAt)
}

func (r *postgresUserRepository) DeleteUser(ctx context.Context, id uuid.UUID) error {
	query := `
		DELETE FROM users
		WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, id)
	return err
}
