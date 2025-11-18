ALTER TABLE ha_comment_english DROP CONSTRAINT answer_id;

DROP TABLE IF EXISTS ha_comment_english;

ALTER TABLE ha_answer_english DROP CONSTRAINT question_id;

DROP TABLE IF EXISTS ha_answer_english;

DROP TABLE IF EXISTS ha_question_english;

DROP TABLE IF EXISTS ha_classified_posts_english;