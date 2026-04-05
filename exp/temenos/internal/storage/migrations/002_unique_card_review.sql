-- Add unique constraint on reviews(card_id)
-- This ensures each card has at most one active review at a time
DROP INDEX IF EXISTS idx_reviews_card_id;
CREATE UNIQUE INDEX IF NOT EXISTS uidx_reviews_card_id ON reviews(card_id);

-- Add index on reviews(session_id) for faster queries
CREATE INDEX IF NOT EXISTS idx_reviews_session_id ON reviews(session_id);
