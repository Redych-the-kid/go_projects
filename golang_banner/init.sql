CREATE TABLE banners (
    id SERIAL PRIMARY KEY,
    tag_ids INTEGER[],
    feature_id INTEGER,
    content JSONB,
    is_active BOOLEAN,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    updated_at TIMESTAMP NOT NULL DEFAULT now()
);

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = now();
  RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_features_updated_at
BEFORE UPDATE ON features
FOR EACH ROW
EXECUTE PROCEDURE update_updated_at_column();