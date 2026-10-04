-- The traps the wrong answers of each month fell into: by topic and grade, by
-- topic over every grade — with the share of the topic's answers each trap
-- caught — and by grade over every topic. Exact numbers, for the owner alone.
SELECT
  trapped.month,
  trapped.topic,
  trapped.grade,
  trapped.trap,
  trapped.answers,
  trapped.learners,
  IF(trapped.grade IS NULL, SAFE_DIVIDE(trapped.answers, topics.answers), NULL) AS share_of_topic_answers
FROM `${impact}.traps_monthly` AS trapped
LEFT JOIN `${impact}.topics_monthly` AS topics
  ON topics.month = trapped.month AND topics.topic = trapped.topic
