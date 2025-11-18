ALTER TABLE deep_comment_eses DROP CONSTRAINT answer_id;

DROP TABLE IF EXISTS deep_comment_eses;

ALTER TABLE deep_answer_eses DROP CONSTRAINT question_id;

DROP TABLE IF EXISTS deep_answer_eses;

DROP TABLE IF EXISTS deep_question_eses;

DROP TABLE IF EXISTS deep_classified_posts_eses;