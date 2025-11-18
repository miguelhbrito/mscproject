CREATE TABLE IF NOT EXISTS "raddle_post" (
  "id" SERIAL PRIMARY KEY,
  "title" varchar NOT NULL,
  "link" varchar,
  "post" varchar NOT NULL,
  "forum" varchar NOT NULL,
  "votes" varchar NOT NULL,
  "author" varchar NOT NULL,
  "created_at" timestamp NOT NULL
);

CREATE TABLE IF NOT EXISTS "raddle_commentary" (
  "id" SERIAL PRIMARY KEY,
  "post_id" int,
  "commentary" varchar NOT NULL,
  "author" varchar NOT NULL,
  "votes" varchar NOT NULL,
  "created_at" timestamp NOT NULL
);

ALTER TABLE raddle_commentary ADD FOREIGN KEY (post_id) REFERENCES raddle_post (id) ON DELETE CASCADE;