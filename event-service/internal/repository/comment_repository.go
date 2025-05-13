package repository

import (
	"context"
	"database/sql"
	"time"

	"promoservice/proto/event"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type CommentRepository struct {
	db *sql.DB
}

func NewCommentRepository(db *sql.DB) *CommentRepository {
	return &CommentRepository{db: db}
}

type Comment struct {
	ID          uuid.UUID
	PromoID     uuid.UUID
	ClientID    uuid.UUID
	CommentText string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (r *CommentRepository) Create(ctx context.Context, promoID, clientID uuid.UUID, commentText string) (*Comment, error) {
	query := `
		INSERT INTO promo_comments (promo_id, client_id, comment_text)
		VALUES ($1, $2, $3)
		RETURNING id, promo_id, client_id, comment_text, created_at, updated_at
	`

	comment := &Comment{}
	err := r.db.QueryRowContext(ctx, query, promoID, clientID, commentText).Scan(
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

func (r *CommentRepository) GetByPromoID(ctx context.Context, promoID uuid.UUID, page, pageSize int32) ([]*Comment, int32, error) {
	// Получаем общее количество комментариев
	var totalCount int32
	countQuery := `SELECT COUNT(*) FROM promo_comments WHERE promo_id = $1`
	err := r.db.QueryRowContext(ctx, countQuery, promoID).Scan(&totalCount)
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

	var comments []*Comment
	for rows.Next() {
		comment := &Comment{}
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

	return comments, totalCount, nil
}

func (r *CommentRepository) ToProto(comment *Comment) *event.PromoComment {
	return &event.PromoComment{
		Id:          comment.ID.String(),
		ClientId:    comment.ClientID.String(),
		PromoId:     comment.PromoID.String(),
		CommentText: comment.CommentText,
		CommentedAt: timestamppb.New(comment.CreatedAt),
	}
}
