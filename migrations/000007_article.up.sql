ALTER TABLE articles RENAME COLUMN created TO inserted_at;
ALTER TABLE articles ADD COLUMN published INTEGER DEFAULT (cast(strftime('%s','now') as int));
ALTER TABLE articles ADD COLUMN attributed_to VARCHAR(255);