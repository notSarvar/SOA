CREATE TABLE IF NOT EXISTS promo_views (
    id SERIAL PRIMARY KEY,
    promo_id VARCHAR(36) NOT NULL,
    user_id VARCHAR(36) NOT NULL,
    viewed_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS promo_clicks (
    id SERIAL PRIMARY KEY,
    promo_id VARCHAR(36) NOT NULL,
    user_id VARCHAR(36) NOT NULL,
    clicked_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS promo_likes (
    id SERIAL PRIMARY KEY,
    promo_id VARCHAR(36) NOT NULL,
    user_id VARCHAR(36) NOT NULL,
    liked_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS promo_comments (
    id SERIAL PRIMARY KEY,
    promo_id VARCHAR(36) NOT NULL,
    user_id VARCHAR(36) NOT NULL,
    comment_text TEXT NOT NULL,
    commented_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS user_activities (
    id SERIAL PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    activity_type VARCHAR(50) NOT NULL,
    activity_details JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Индексы
CREATE INDEX IF NOT EXISTS idx_promo_views_promo_id ON promo_views(promo_id);
CREATE INDEX IF NOT EXISTS idx_promo_views_user_id ON promo_views(user_id);
CREATE INDEX IF NOT EXISTS idx_promo_views_viewed_at ON promo_views(viewed_at);

CREATE INDEX IF NOT EXISTS idx_promo_clicks_promo_id ON promo_clicks(promo_id);
CREATE INDEX IF NOT EXISTS idx_promo_clicks_user_id ON promo_clicks(user_id);
CREATE INDEX IF NOT EXISTS idx_promo_clicks_clicked_at ON promo_clicks(clicked_at);

CREATE INDEX IF NOT EXISTS idx_promo_likes_promo_id ON promo_likes(promo_id);
CREATE INDEX IF NOT EXISTS idx_promo_likes_user_id ON promo_likes(user_id);
CREATE INDEX IF NOT EXISTS idx_promo_likes_liked_at ON promo_likes(liked_at);

CREATE INDEX IF NOT EXISTS idx_promo_comments_promo_id ON promo_comments(promo_id);
CREATE INDEX IF NOT EXISTS idx_promo_comments_user_id ON promo_comments(user_id);
CREATE INDEX IF NOT EXISTS idx_promo_comments_commented_at ON promo_comments(commented_at);

CREATE INDEX IF NOT EXISTS idx_user_activities_user_id ON user_activities(user_id);
CREATE INDEX IF NOT EXISTS idx_user_activities_activity_type ON user_activities(activity_type);
CREATE INDEX IF NOT EXISTS idx_user_activities_created_at ON user_activities(created_at); 