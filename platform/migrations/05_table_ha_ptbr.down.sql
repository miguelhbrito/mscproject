ALTER TABLE ha_comment_ptbr DROP CONSTRAINT answer_id;

DROP TABLE IF EXISTS ha_comment_ptbr;

ALTER TABLE ha_answer_ptbr DROP CONSTRAINT question_id;

DROP TABLE IF EXISTS ha_answer_ptbr;

DROP TABLE IF EXISTS ha_question_ptbr;

DROP TABLE IF EXISTS ha_classified_posts_ptbr;