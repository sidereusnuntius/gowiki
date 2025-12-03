-- This table has been repurposed to hold the data not of remote servers, but of collective actors, such as groups.
ALTER TABLE instances RENAME COLUMN hostname TO host;
ALTER TABLE instances RENAME TO collectives;
ALTER TABLE articles ADD COLUMN author VARCHAR(255);
ALTER TABLE files ADD COLUMN host VARCHAR(255);