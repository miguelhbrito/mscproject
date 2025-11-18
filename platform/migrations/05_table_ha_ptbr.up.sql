CREATE TABLE IF NOT EXISTS "ha_question_ptbr" (
  "id" SERIAL PRIMARY KEY,
  "title" varchar NOT NULL,
  "question" varchar NOT NULL,
  "category" varchar NOT NULL,
  "tags" text[],
  "up_vote" varchar,
  "down_vote" varchar,
  "author" varchar NOT NULL,
  "type" varchar NOT NULL,
  "created_at" timestamp NOT NULL,
  "points" varchar
);

CREATE TABLE IF NOT EXISTS "ha_answer_ptbr" (
  "id" SERIAL PRIMARY KEY,
  "question_id" int,
  "answer_content" varchar NOT NULL,
  "up_vote" varchar,
  "down_vote" varchar,
  "author" varchar NOT NULL,
  "type" varchar NOT NULL,
  "created_at" timestamp NOT NULL,
  "points" varchar
);

CREATE TABLE IF NOT EXISTS "ha_comment_ptbr" (
  "id" SERIAL PRIMARY KEY,
  "answer_id" int,
  "commentary" varchar NOT NULL,
  "author" varchar NOT NULL,
  "type" varchar NOT NULL,
  "created_at" timestamp NOT NULL
);

CREATE TABLE IF NOT EXISTS"ha_classified_posts_ptbr" (
  "id" int PRIMARY KEY,
  "category" varchar,
  "title" varchar,
  "content" varchar,
  "tags" varchar,
  "relevant" boolean,
  "classified_at" timestamp
);

ALTER TABLE ha_answer_ptbr ADD FOREIGN KEY (question_id) REFERENCES ha_question_ptbr (id) ON DELETE CASCADE;

ALTER TABLE ha_comment_ptbr ADD FOREIGN KEY (answer_id) REFERENCES ha_answer_ptbr (id) ON DELETE CASCADE;