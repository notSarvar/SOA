CREATE TABLE IF NOT EXISTS promo_comments (
    id UUID PRIMARY KEY,
    promo_id UUID NOT NULL REFERENCES promos(id) ON DELETE CASCADE,
    client_id UUID NOT NULL,
    comment_text TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_promo_comments_promo_id ON promo_comments(promo_id);
CREATE INDEX IF NOT EXISTS idx_promo_comments_client_id ON promo_comments(client_id);
CREATE INDEX IF NOT EXISTS idx_promo_comments_created_at ON promo_comments(created_at); 