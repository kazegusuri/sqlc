CREATE TABLE authors (
          id   BIGSERIAL PRIMARY KEY,
          name text      NOT NULL,
          bio  text
);

CREATE UNIQUE INDEX ON authors (name);

CREATE INDEX authors_bio_idx ON authors (bio);
