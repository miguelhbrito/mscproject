CREATE TABLE IF NOT EXISTS "deep_question_eses" (
  "id" SERIAL PRIMARY KEY,
  "title" varchar NOT NULL,
  "question" varchar NOT NULL,
  "category" varchar NOT NULL,
  "tags" text[],
  "votes" varchar NOT NULL,
  "author" varchar NOT NULL,
  "type" varchar NOT NULL,
  "created_at" timestamp NOT NULL,
  "points" varchar 
);

CREATE TABLE IF NOT EXISTS "deep_answer_eses" (
  "id" SERIAL PRIMARY KEY,
  "question_id" int,
  "answer_content" varchar NOT NULL,
  "votes" varchar NOT NULL,
  "author" varchar NOT NULL,
  "type" varchar NOT NULL,
  "created_at" timestamp NOT NULL,
  "points" varchar 
);

CREATE TABLE IF NOT EXISTS "deep_comment_eses" (
  "id" SERIAL PRIMARY KEY,
  "answer_id" int,
  "commentary" varchar NOT NULL,
  "author" varchar NOT NULL,
  "type" varchar NOT NULL,
  "created_at" timestamp NOT NULL
);

CREATE TABLE IF NOT EXISTS "deep_classified_posts_eses" (
  "id" int PRIMARY KEY,
  "category" varchar,
  "title" varchar,
  "content" varchar,
  "tags" varchar,
  "relevant" boolean,
  "classified_at" timestamp
);

ALTER TABLE deep_answer_eses ADD FOREIGN KEY (question_id) REFERENCES deep_question_eses (id) ON DELETE CASCADE;

ALTER TABLE deep_comment_eses ADD FOREIGN KEY (answer_id) REFERENCES deep_answer_eses (id) ON DELETE CASCADE;