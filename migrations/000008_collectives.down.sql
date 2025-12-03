ALTER TABLE files DROP COLUMN host;
ALTER TABLE articles DROP COLUMN author;
ALTER TABLE collectives RENAME TO instances;
ALTER TABLE instances RENAME COLUMN host TO hostname;