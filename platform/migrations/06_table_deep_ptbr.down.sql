ALTER TABLE deep_comment_ptbr DROP CONSTRAINT answer_id;

DROP TABLE IF EXISTS deep_comment_ptbr;

ALTER TABLE deep_answer_ptbr DROP CONSTRAINT question_id;

DROP TABLE IF EXISTS deep_answer_ptbr;

DROP TABLE IF EXISTS deep_question_ptbr;

DROP TABLE IF EXISTS deep_classified_posts_ptbr;