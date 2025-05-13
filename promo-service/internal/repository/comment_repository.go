package repository

import (
	"context"
	"database/sql"

	"promoservice/promo-service/internal/models"

	"github.com/google/uuid"
)

type CommentRepository struct {
	db *sql.DB
}

func NewCommentRepository(db *sql.DB) *CommentRepository {
	return &CommentRepository{db: db}
}

func (r *CommentRepository) Create(ctx context.Context, comment *models.Comment) (*models.Comment, error) {
	query := `
		INSERT INTO promo_comments (id, promo_id, client_id, comment_text, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, promo_id, client_id, comment_text, created_at, updated_at
	`

	err := r.db.QueryRowContext(
		ctx,
		query,
		comment.ID,
		comment.PromoID,
		comment.ClientID,
		comment.CommentText,
		comment.CreatedAt,
		comment.UpdatedAt,
	).Scan(
		&comment.ID,
		&comment.PromoID,
		&comment.ClientID,
		&comment.CommentText,
		&comment.CreatedAt,
		&comment.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return comment, nil
}

func (r *CommentRepository) GetByPromoID(ctx context.Context, promoID uuid.UUID, page, pageSize int32) ([]*models.Comment, int64, error) {
	// Получаем общее количество комментариев
	var total int64
	err := r.db.QueryRowContext(
		ctx,
		"SELECT COUNT(*) FROM promo_comments WHERE promo_id = $1",
		promoID,
	).Scan(&total)

	if err != nil {
		return nil, 0, err
	}

	// Получаем комментарии с пагинацией
	query := `
		SELECT id, promo_id, client_id, comment_text, created_at, updated_at
		FROM promo_comments
		WHERE promo_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	offset := (page - 1) * pageSize
	rows, err := r.db.QueryContext(ctx, query, promoID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var comments []*models.Comment
	for rows.Next() {
		comment := &models.Comment{}
		err := rows.Scan(
			&comment.ID,
			&comment.PromoID,
			&comment.ClientID,
			&comment.CommentText,
			&comment.CreatedAt,
			&comment.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		comments = append(comments, comment)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, err
	}

	return comments, total, nil
}
