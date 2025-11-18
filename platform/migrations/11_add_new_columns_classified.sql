ALTER TABLE deep_classified_posts_enus
ADD COLUMN probability DOUBLE PRECISION null,
ADD COLUMN range integer null;

ALTER TABLE deep_classified_posts_eses
ADD COLUMN probability DOUBLE PRECISION null,
ADD COLUMN range integer null;

ALTER TABLE deep_classified_posts_ptbr
ADD COLUMN probability DOUBLE PRECISION null,
ADD COLUMN range integer null;

ALTER TABLE ha_classified_posts_english
ADD COLUMN probability DOUBLE PRECISION null,
ADD COLUMN range integer null;

ALTER TABLE ha_classified_posts_ptbr
ADD COLUMN probability DOUBLE PRECISION null,
ADD COLUMN range integer null;