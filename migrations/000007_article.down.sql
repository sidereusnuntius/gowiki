ALTER TABLE articles DROP COLUMN attributed_to;
ALTER TABLE articles DROP COLUMN published;
ALTER TABLE articles RENAME COLUMN inserted_at TO created;