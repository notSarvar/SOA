package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

type Postgres struct {
	pool   *pgxpool.Pool
	logger *logrus.Logger
}

func NewPostgres(host string, port int, user, password, dbname string, logger *logrus.Logger) (*Postgres, error) {
	connString := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	pool, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Postgres{
		pool:   pool,
		logger: logger,
	}, nil
}

func (p *Postgres) Close() {
	p.pool.Close()
}

// Методы для работы с событиями
func (p *Postgres) SavePromoView(ctx context.Context, promoID, userID string) error {
	query := `
		INSERT INTO promo_views (promo_id, user_id, viewed_at)
		VALUES ($1, $2, NOW())
	`

	_, err := p.pool.Exec(ctx, query, promoID, userID)
	if err != nil {
		return fmt.Errorf("failed to save promo view: %w", err)
	}

	return nil
}

func (p *Postgres) SavePromoClick(ctx context.Context, promoID, userID string) error {
	query := `
		INSERT INTO promo_clicks (promo_id, user_id, clicked_at)
		VALUES ($1, $2, NOW())
	`

	_, err := p.pool.Exec(ctx, query, promoID, userID)
	if err != nil {
		return fmt.Errorf("failed to save promo click: %w", err)
	}

	return nil
}

func (p *Postgres) SavePromoLike(ctx context.Context, promoID, userID string) error {
	query := `
		INSERT INTO promo_likes (promo_id, user_id, liked_at)
		VALUES ($1, $2, NOW())
	`

	_, err := p.pool.Exec(ctx, query, promoID, userID)
	if err != nil {
		return fmt.Errorf("failed to save promo like: %w", err)
	}

	return nil
}

func (p *Postgres) SavePromoComment(ctx context.Context, promoID, userID, commentText string) error {
	query := `
		INSERT INTO promo_comments (promo_id, user_id, comment_text, commented_at)
		VALUES ($1, $2, $3, NOW())
	`

	_, err := p.pool.Exec(ctx, query, promoID, userID, commentText)
	if err != nil {
		return fmt.Errorf("failed to save promo comment: %w", err)
	}

	return nil
}

func (p *Postgres) GetPromoComments(ctx context.Context, promoID string, page, pageSize int) ([]PromoComment, int, error) {
	var total int
	countQuery := `
		SELECT COUNT(*) FROM promo_comments WHERE promo_id = $1
	`
	err := p.pool.QueryRow(ctx, countQuery, promoID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count comments: %w", err)
	}

	query := `
		SELECT id, promo_id, user_id, comment_text, commented_at
		FROM promo_comments
		WHERE promo_id = $1
		ORDER BY commented_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := p.pool.Query(ctx, query, promoID, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query comments: %w", err)
	}
	defer rows.Close()

	var comments []PromoComment
	for rows.Next() {
		var comment PromoComment
		err := rows.Scan(&comment.ID, &comment.PromoID, &comment.UserID, &comment.CommentText, &comment.CommentedAt)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan comment: %w", err)
		}
		comments = append(comments, comment)
	}

	return comments, total, nil
}

type PromoComment struct {
	ID          string
	PromoID     string
	UserID      string
	CommentText string
	CommentedAt string
}
