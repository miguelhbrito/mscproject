ALTER TABLE deep_comment_enus DROP CONSTRAINT answer_id;

DROP TABLE IF EXISTS deep_comment_enus;

ALTER TABLE deep_answer_enus DROP CONSTRAINT question_id;

DROP TABLE IF EXISTS deep_answer_enus;

DROP TABLE IF EXISTS deep_question_enus;

DROP TABLE IF EXISTS deep_classified_posts_enus;