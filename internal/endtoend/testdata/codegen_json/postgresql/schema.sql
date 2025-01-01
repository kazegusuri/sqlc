-- author
CREATE TABLE authors (
          id   BIGSERIAL PRIMARY KEY, -- id
          name text      NOT NULL, -- name
          bio  text -- bio
);

-- status
CREATE TYPE status AS ENUM (
          'online', -- online
          'offline' -- offline
);
