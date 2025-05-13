CREATE TABLE IF NOT EXISTS promo_comments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    promo_id UUID NOT NULL,
    client_id UUID NOT NULL,
    comment_text TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_promo_comments_promo_id ON promo_comments(promo_id);
CREATE INDEX IF NOT EXISTS idx_promo_comments_client_id ON promo_comments(client_id);
CREATE INDEX IF NOT EXISTS idx_promo_comments_created_at ON promo_comments(created_at); 