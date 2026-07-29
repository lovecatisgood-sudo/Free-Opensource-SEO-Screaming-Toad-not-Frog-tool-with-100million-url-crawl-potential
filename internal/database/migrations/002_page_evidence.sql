ALTER TABLE page ADD COLUMN viewport TEXT;
ALTER TABLE page ADD COLUMN html_hash TEXT;
ALTER TABLE page ADD COLUMN x_robots_tag TEXT;
ALTER TABLE page ADD COLUMN social_json TEXT NOT NULL DEFAULT '{}' CHECK (json_valid(social_json));
ALTER TABLE image ADD COLUMN declared_width INTEGER;
ALTER TABLE image ADD COLUMN declared_height INTEGER;

