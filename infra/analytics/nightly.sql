-- The counts of the children, made again every night from the raw lines the
-- activity bucket keeps for 62 days, into tables that hold no name of any
-- child and are kept for years.
--
-- Every period whose first day the bucket still holds whole is counted again
-- from that day: a night that did not run is made up for by the next, and a
-- run made by hand counts what a night would. A period whose first day the
-- bucket no longer holds keeps the rows it was last counted to. A child is
-- counted under the name it has that month, and the children the load tool or
-- MCP Inspector handed a task to are not counted at all.

-- Yesterday: the last day whose lines are all in.
DECLARE counted_until DATE DEFAULT DATE_SUB(CURRENT_DATE('UTC'), INTERVAL 1 DAY);

-- The first day the bucket holds whole: 60 of its 62 days back, and never
-- before the day after the first line that names a child — a build older than
-- the counting names none, and the day the counting arrived on may have begun
-- without it.
DECLARE lines_from DATE DEFAULT (
  SELECT GREATEST(
    DATE_SUB(CURRENT_DATE('UTC'), INTERVAL 60 DAY),
    DATE_ADD(IFNULL(MIN(DATE(timestamp, 'UTC')), CURRENT_DATE('UTC')), INTERVAL 1 DAY))
  FROM `${logs}`
  WHERE JSON_VALUE(json_payload.learner) IS NOT NULL
);

-- The first week and the first month that begin on that day or after it: an
-- earlier one would be counted from part of its days.
DECLARE weeks_from DATE DEFAULT IF(
  DATE_TRUNC(lines_from, WEEK(MONDAY)) = lines_from,
  lines_from,
  DATE_ADD(DATE_TRUNC(lines_from, WEEK(MONDAY)), INTERVAL 1 WEEK));
DECLARE months_from DATE DEFAULT IF(
  DATE_TRUNC(lines_from, MONTH) = lines_from,
  lines_from,
  DATE_ADD(DATE_TRUNC(lines_from, MONTH), INTERVAL 1 MONTH));

-- The lines of those days, each once — the log may hand a line over twice —
-- with the fields the counts read. A number may come back from the log as 2.0.
CREATE TEMP TABLE counted_lines AS
SELECT
  DATE(timestamp, 'UTC') AS on_day,
  timestamp AS written_at,
  JSON_VALUE(json_payload.message) AS event,
  JSON_VALUE(json_payload.learner) AS learner,
  IFNULL(JSON_VALUE(json_payload.host), 'unknown') AS host,
  IFNULL(JSON_VALUE(json_payload.language), 'unknown') AS language,
  SAFE_CAST(SAFE_CAST(JSON_VALUE(json_payload.grade) AS FLOAT64) AS INT64) AS grade,
  IFNULL(JSON_VALUE(json_payload.cohort), 'unknown') AS cohort,
  IFNULL(JSON_VALUE(json_payload.country), 'unknown') AS country,
  IFNULL(JSON_VALUE(json_payload.region), 'unknown') AS region,
  IFNULL(JSON_VALUE(json_payload.signin_country), 'unknown') AS signin_country,
  IFNULL(JSON_VALUE(json_payload.topic), 'other') AS topic,
  IFNULL(JSON_VALUE(json_payload.trap), '') AS trap,
  IFNULL(JSON_VALUE(json_payload.correct) = 'true', FALSE) AS correct,
  IFNULL(JSON_VALUE(json_payload.hint_used) = 'true', FALSE) AS hinted,
  IFNULL(JSON_VALUE(json_payload.confused) = 'true', FALSE) AS dont_know,
  SAFE_CAST(SAFE_CAST(JSON_VALUE(json_payload.topics_mastered) AS FLOAT64) AS INT64) AS topics_mastered
FROM (
  SELECT
    timestamp,
    json_payload,
    ROW_NUMBER() OVER (PARTITION BY log_name, timestamp, insert_id ORDER BY insert_id) AS copy_number
  FROM `${logs}`
  WHERE timestamp >= TIMESTAMP(lines_from) AND timestamp < TIMESTAMP(CURRENT_DATE('UTC'))
)
WHERE copy_number = 1
  AND JSON_VALUE(json_payload.learner) IS NOT NULL
  AND JSON_VALUE(json_payload.message) IN (${events});

-- The children the load tool or MCP Inspector handed a task to are no children
-- a family has: nothing their names are on is counted.
DELETE FROM counted_lines
WHERE learner IN (
  SELECT learner FROM counted_lines
  WHERE event = 'task_accepted' AND host IN ('load', 'inspector')
);

-- One night's counts land whole or not at all.
BEGIN TRANSACTION;

-- Every day from the first the bucket holds whole to yesterday, a day with
-- nothing on it among them: a missing day is a night that never ran.
DELETE FROM `${impact}.daily` WHERE day >= lines_from;
INSERT INTO `${impact}.daily` (day, learners, learners_week, learners_month, tasks, answers, topics_won)
SELECT
  counted_day,
  (SELECT COUNT(DISTINCT learner) FROM counted_lines
    WHERE event = 'task_accepted' AND on_day = counted_day),
  IF(DATE_SUB(counted_day, INTERVAL 6 DAY) < lines_from, NULL,
    (SELECT COUNT(DISTINCT learner) FROM counted_lines
      WHERE event = 'task_accepted' AND on_day BETWEEN DATE_SUB(counted_day, INTERVAL 6 DAY) AND counted_day)),
  IF(DATE_TRUNC(counted_day, MONTH) < lines_from, NULL,
    (SELECT COUNT(DISTINCT learner) FROM counted_lines
      WHERE event = 'task_accepted' AND on_day BETWEEN DATE_TRUNC(counted_day, MONTH) AND counted_day)),
  (SELECT COUNT(*) FROM counted_lines WHERE event = 'task_accepted' AND on_day = counted_day),
  (SELECT COUNT(*) FROM counted_lines WHERE event = 'answer_recorded' AND on_day = counted_day),
  (SELECT COUNT(*) FROM counted_lines WHERE event = 'topic_mastered' AND on_day = counted_day)
FROM UNNEST(GENERATE_DATE_ARRAY(lines_from, counted_until)) AS counted_day;

-- The children of each week and of each month, told apart by one thing at a
-- time. A child is counted under the value of its latest task of the period,
-- so that the values of a dimension add up to all and a child who moved from
-- one host or grade to another is counted once.
DELETE FROM `${impact}.learners_weekly` WHERE week >= weeks_from;
INSERT INTO `${impact}.learners_weekly` (week, dimension, value, learners)
SELECT latest.period, told.dimension, told.value, COUNT(*)
FROM (
  SELECT
    DATE_TRUNC(on_day, WEEK(MONDAY)) AS period,
    host, language, grade, country, region, signin_country, cohort,
    ROW_NUMBER() OVER (PARTITION BY DATE_TRUNC(on_day, WEEK(MONDAY)), learner ORDER BY written_at DESC) AS recency
  FROM counted_lines
  WHERE event = 'task_accepted' AND on_day >= weeks_from
) AS latest,
UNNEST([
  STRUCT('all' AS dimension, 'all' AS value),
  STRUCT('host' AS dimension, latest.host AS value),
  STRUCT('language' AS dimension, latest.language AS value),
  STRUCT('grade' AS dimension, IFNULL(CAST(latest.grade AS STRING), 'unknown') AS value),
  STRUCT('country' AS dimension, latest.country AS value),
  STRUCT('region' AS dimension, latest.region AS value),
  STRUCT('signin_country' AS dimension, latest.signin_country AS value),
  STRUCT('cohort' AS dimension, latest.cohort AS value)
]) AS told
WHERE latest.recency = 1
GROUP BY latest.period, told.dimension, told.value;

DELETE FROM `${impact}.learners_monthly` WHERE month >= months_from;
INSERT INTO `${impact}.learners_monthly` (month, dimension, value, learners)
SELECT latest.period, told.dimension, told.value, COUNT(*)
FROM (
  SELECT
    DATE_TRUNC(on_day, MONTH) AS period,
    host, language, grade, country, region, signin_country, cohort,
    ROW_NUMBER() OVER (PARTITION BY DATE_TRUNC(on_day, MONTH), learner ORDER BY written_at DESC) AS recency
  FROM counted_lines
  WHERE event = 'task_accepted' AND on_day >= months_from
) AS latest,
UNNEST([
  STRUCT('all' AS dimension, 'all' AS value),
  STRUCT('host' AS dimension, latest.host AS value),
  STRUCT('language' AS dimension, latest.language AS value),
  STRUCT('grade' AS dimension, IFNULL(CAST(latest.grade AS STRING), 'unknown') AS value),
  STRUCT('country' AS dimension, latest.country AS value),
  STRUCT('region' AS dimension, latest.region AS value),
  STRUCT('signin_country' AS dimension, latest.signin_country AS value),
  STRUCT('cohort' AS dimension, latest.cohort AS value)
]) AS told
WHERE latest.recency = 1
GROUP BY latest.period, told.dimension, told.value;

-- How much each child of a month did: the tasks handed to it, and the days a
-- task was, each in ranges.
DELETE FROM `${impact}.dose_monthly` WHERE month >= months_from;
INSERT INTO `${impact}.dose_monthly` (month, measure, bucket, learners)
SELECT per_child.period, measured.measure, measured.bucket, COUNT(*)
FROM (
  SELECT DATE_TRUNC(on_day, MONTH) AS period, learner, COUNT(*) AS tasks, COUNT(DISTINCT on_day) AS active_days
  FROM counted_lines
  WHERE event = 'task_accepted' AND on_day >= months_from
  GROUP BY period, learner
) AS per_child,
UNNEST([
  STRUCT('tasks' AS measure, CASE
    WHEN per_child.tasks = 1 THEN '1'
    WHEN per_child.tasks <= 4 THEN '2-4'
    WHEN per_child.tasks <= 9 THEN '5-9'
    WHEN per_child.tasks <= 19 THEN '10-19'
    WHEN per_child.tasks <= 49 THEN '20-49'
    ELSE '50+' END AS bucket),
  STRUCT('active_days' AS measure, CASE
    WHEN per_child.active_days = 1 THEN '1'
    WHEN per_child.active_days <= 3 THEN '2-3'
    WHEN per_child.active_days <= 7 THEN '4-7'
    WHEN per_child.active_days <= 14 THEN '8-14'
    ELSE '15+' END AS bucket)
]) AS measured
GROUP BY per_child.period, measured.measure, measured.bucket;

-- How the children of a month answered, by the months since their profile was
-- made, and how many topics each had mastered at most.
DELETE FROM `${impact}.learning_monthly` WHERE month >= months_from;
INSERT INTO `${impact}.learning_monthly` (month, tenure, learners, answers, correct, hinted, dont_know, mastered)
SELECT period, tenure, COUNT(*), SUM(answers), SUM(correct), SUM(hinted), SUM(dont_know), SUM(mastered)
FROM (
  SELECT
    period,
    DATE_DIFF(period, SAFE.PARSE_DATE('%Y-%m', cohort), MONTH) AS tenure,
    answers, correct, hinted, dont_know, mastered
  FROM (
    SELECT
      DATE_TRUNC(on_day, MONTH) AS period,
      learner,
      MAX(cohort) AS cohort,
      COUNT(*) AS answers,
      COUNTIF(correct) AS correct,
      COUNTIF(hinted) AS hinted,
      COUNTIF(dont_know) AS dont_know,
      IFNULL(MAX(topics_mastered), 0) AS mastered
    FROM counted_lines
    WHERE event = 'answer_recorded' AND on_day >= months_from
    GROUP BY period, learner
  )
)
WHERE tenure >= 0
GROUP BY period, tenure;

-- Each topic of a month: who answered its tasks and how, and the times it was
-- won.
DELETE FROM `${impact}.topics_monthly` WHERE month >= months_from;
INSERT INTO `${impact}.topics_monthly` (month, topic, learners, answers, correct, dont_know, won, winners)
SELECT
  DATE_TRUNC(on_day, MONTH) AS period,
  topic,
  COUNT(DISTINCT IF(event = 'answer_recorded', learner, NULL)),
  COUNTIF(event = 'answer_recorded'),
  COUNTIF(event = 'answer_recorded' AND correct),
  COUNTIF(event = 'answer_recorded' AND dont_know),
  COUNTIF(event = 'topic_mastered'),
  COUNT(DISTINCT IF(event = 'topic_mastered', learner, NULL))
FROM counted_lines
WHERE event IN ('answer_recorded', 'topic_mastered') AND on_day >= months_from
GROUP BY period, topic;

-- The traps the wrong answers of a month fell into: by topic and grade, by
-- topic over every grade, and by grade over every topic, so that each row
-- counts its own children once.
DELETE FROM `${impact}.traps_monthly` WHERE month >= months_from;
INSERT INTO `${impact}.traps_monthly` (month, topic, grade, trap, answers, learners)
SELECT DATE_TRUNC(on_day, MONTH), topic, grade, trap, COUNT(*), COUNT(DISTINCT learner)
FROM counted_lines
WHERE event = 'answer_recorded' AND trap != '' AND on_day >= months_from
GROUP BY 1, 2, 3, 4
UNION ALL
SELECT DATE_TRUNC(on_day, MONTH), topic, CAST(NULL AS INT64), trap, COUNT(*), COUNT(DISTINCT learner)
FROM counted_lines
WHERE event = 'answer_recorded' AND trap != '' AND on_day >= months_from
GROUP BY 1, 2, 4
UNION ALL
SELECT DATE_TRUNC(on_day, MONTH), CAST(NULL AS STRING), grade, trap, COUNT(*), COUNT(DISTINCT learner)
FROM counted_lines
WHERE event = 'answer_recorded' AND trap != '' AND on_day >= months_from
GROUP BY 1, 3, 4;

COMMIT TRANSACTION;
