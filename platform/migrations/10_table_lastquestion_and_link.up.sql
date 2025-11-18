CREATE TABLE IF NOT EXISTS "link_scraper" (
  "id"   varchar PRIMARY KEY,
	"link" varchar NOT NULL
);

CREATE TABLE IF NOT EXISTS "last_question" (
  "id" 			          varchar PRIMARY KEY,
  "question_number" 	int NOT NULL
);

INSERT INTO link_scraper (id, link) VALUES ('ha_enus','http://7eoz4h2nvw4zlr7gvlbutinqqpm546f5egswax54az6lt2u7e3t6d7yd.onion/index.php');

INSERT INTO last_question (id, question_number) VALUES ('ha_enus',1);

INSERT INTO link_scraper (id, link) VALUES ('ha_ptbr','http://xh6liiypqffzwnu5734ucwps37tn2g6npthvugz3gdoqpikujju525yd.onion/index.php');

INSERT INTO last_question (id, question_number) VALUES ('ha_ptbr',1);

INSERT INTO link_scraper (id, link) VALUES ('deep_ptbr','http://deeptyspkdq3nfvqvyzbkgwhtok4qoyhypsyiuo24wux4jnb6e3nyiqd.onion/index.php?qa=');

INSERT INTO last_question (id, question_number) VALUES ('deep_ptbr',1);

INSERT INTO link_scraper (id, link) VALUES ('deep_enus','http://deepa2kol4ur4wkzpmjf5rf7lvsflzisslnrnr2n7goaebav4j6w7zyd.onion/index.php?qa=');

INSERT INTO last_question (id, question_number) VALUES ('deep_enus',1);

INSERT INTO link_scraper (id, link) VALUES ('deep_eses','http://deepesio33g3zojyrfxnnfdefhuxlpftsdp5siddprkw2qw5adakluid.onion/index.php?qa=');

INSERT INTO last_question (id, question_number) VALUES ('deep_eses',1);

INSERT INTO link_scraper (id, link) VALUES ('raddle','http://c32zjeghcp5tj3kb72pltz56piei66drc63vkhn5yixiyk4cmerrjtid.onion/');

INSERT INTO last_question (id, question_number) VALUES ('raddle',1);