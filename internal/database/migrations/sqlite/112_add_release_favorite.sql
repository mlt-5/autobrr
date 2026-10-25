ALTER TABLE "release"
    ADD COLUMN favorite BOOLEAN DEFAULT FALSE NOT NULL;

CREATE INDEX release_favorite_index
    ON "release" (favorite);
