package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"promoservice/promo-service/internal/models"
)

type PromoRepository interface {
	CreatePromo(ctx context.Context, promo *models.Promo) error
	GetPromoByID(ctx context.Context, id string) (*models.Promo, error)
	GetPromoByCode(ctx context.Context, code string) (*models.Promo, error)
	UpdatePromo(ctx context.Context, promo *models.Promo) error
	DeletePromo(ctx context.Context, id string) error
	ListPromos(ctx context.Context, creatorID string, page, pageSize int, onlyActive bool) ([]*models.Promo, int, error)
}

type postgresPromoRepository struct {
	db *sql.DB
}

func NewPostgresPromoRepository(db *sql.DB) PromoRepository {
	return &postgresPromoRepository{db: db}
}

func (r *postgresPromoRepository) CreatePromo(ctx context.Context, promo *models.Promo) error {
	query := `
		INSERT INTO promos (
			id, name, description, creator_id, discount_amount, code,
			created_at, updated_at, is_active, valid_until, max_uses, current_uses
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
		)
		RETURNING created_at, updated_at`

	return r.db.QueryRowContext(ctx, query,
		promo.ID, promo.Name, promo.Description, promo.CreatorID,
		promo.DiscountAmount, promo.Code, promo.CreatedAt, promo.UpdatedAt,
		promo.IsActive, promo.ValidUntil, promo.MaxUses, promo.CurrentUses,
	).Scan(&promo.CreatedAt, &promo.UpdatedAt)
}

func (r *postgresPromoRepository) GetPromoByID(ctx context.Context, id string) (*models.Promo, error) {
	query := `
		SELECT id, name, description, creator_id, discount_amount, code,
			created_at, updated_at, is_active, valid_until, max_uses, current_uses
		FROM promos
		WHERE id = $1`

	var promo models.Promo
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&promo.ID, &promo.Name, &promo.Description, &promo.CreatorID,
		&promo.DiscountAmount, &promo.Code, &promo.CreatedAt, &promo.UpdatedAt,
		&promo.IsActive, &promo.ValidUntil, &promo.MaxUses, &promo.CurrentUses,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("promo not found")
	}
	if err != nil {
		return nil, err
	}

	return &promo, nil
}

func (r *postgresPromoRepository) GetPromoByCode(ctx context.Context, code string) (*models.Promo, error) {
	query := `
		SELECT id, name, description, creator_id, discount_amount, code,
			created_at, updated_at, is_active, valid_until, max_uses, current_uses
		FROM promos
		WHERE code = $1`

	var promo models.Promo
	err := r.db.QueryRowContext(ctx, query, code).Scan(
		&promo.ID, &promo.Name, &promo.Description, &promo.CreatorID,
		&promo.DiscountAmount, &promo.Code, &promo.CreatedAt, &promo.UpdatedAt,
		&promo.IsActive, &promo.ValidUntil, &promo.MaxUses, &promo.CurrentUses,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("promo not found")
	}
	if err != nil {
		return nil, err
	}

	return &promo, nil
}

func (r *postgresPromoRepository) UpdatePromo(ctx context.Context, promo *models.Promo) error {
	query := `
		UPDATE promos
		SET name = $1, description = $2, discount_amount = $3, code = $4,
			updated_at = $5, is_active = $6, valid_until = $7, max_uses = $8
		WHERE id = $9
		RETURNING updated_at`

	return r.db.QueryRowContext(ctx, query,
		promo.Name, promo.Description, promo.DiscountAmount, promo.Code,
		time.Now(), promo.IsActive, promo.ValidUntil, promo.MaxUses,
		promo.ID,
	).Scan(&promo.UpdatedAt)
}

func (r *postgresPromoRepository) DeletePromo(ctx context.Context, id string) error {
	query := `DELETE FROM promos WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("promo not found")
	}

	return nil
}

func (r *postgresPromoRepository) ListPromos(ctx context.Context, creatorID string, page, pageSize int, onlyActive bool) ([]*models.Promo, int, error) {
	offset := (page - 1) * pageSize

	var creatorIDParam interface{} = nil
	if creatorID != "" {
		creatorIDParam = creatorID
	}

	// Получаем общее количество записей
	countQuery := `
		SELECT COUNT(*)
		FROM promos
		WHERE ($1::uuid IS NULL OR creator_id = $1::uuid)
		AND ($2 = false OR is_active = true)
	`

	var total int
	err := r.db.QueryRowContext(ctx, countQuery, creatorIDParam, onlyActive).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Получаем записи с пагинацией
	query := `
		SELECT id, name, description, creator_id, discount_amount, code,
			created_at, updated_at, is_active, valid_until, max_uses, current_uses
		FROM promos
		WHERE ($1::uuid IS NULL OR creator_id = $1::uuid)
		AND ($2 = false OR is_active = true)
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4
	`

	rows, err := r.db.QueryContext(ctx, query, creatorIDParam, onlyActive, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var promos []*models.Promo
	for rows.Next() {
		var promo models.Promo
		err := rows.Scan(
			&promo.ID, &promo.Name, &promo.Description, &promo.CreatorID,
			&promo.DiscountAmount, &promo.Code, &promo.CreatedAt, &promo.UpdatedAt,
			&promo.IsActive, &promo.ValidUntil, &promo.MaxUses, &promo.CurrentUses,
		)
		if err != nil {
			return nil, 0, err
		}
		promos = append(promos, &promo)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, err
	}

	return promos, total, nil
}
