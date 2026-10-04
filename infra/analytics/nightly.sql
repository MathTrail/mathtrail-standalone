-- The counts of the children, made again every night from the raw lines the
-- activity bucket keeps for 62 days, into tables that hold no name of any
-- child and are kept for years.
--
-- A child is counted under the name it has that month, and the children the
-- load tool or MCP Inspector handed a task to that month are not counted at
-- all; so a period is counted whole only while the bucket holds every month a
-- name in it was given for — what the exclusion has to read, from the first of
-- each. While it does, every night counts the period again: a night that did
-- not run is made up for by the next, and a run made by hand counts what a
-- night would. Once it does not, the period keeps the rows it was last counted
-- to, counted whole. A day or a week with no row yet is counted from what the
-- bucket holds, which, the first days of the counting aside, is all of it.

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

-- The children of each week and of each month, told apart by one thing at a
-- time. A child is counted under the value of its latest task of the period,
-- so that the values of a dimension add up to all and a child who moved from
-- one host or grade to another is counted once.
CREATE TEMP TABLE told_apart AS
SELECT latest.kind, latest.period, told.dimension, told.value
FROM (
  SELECT
    kind, period, host, language, grade, country, region, signin_country, cohort,
    ROW_NUMBER() OVER (PARTITION BY kind, period, learner ORDER BY written_at DESC) AS recency
  FROM (
    SELECT
      'week' AS kind, DATE_TRUNC(on_day, WEEK(MONDAY)) AS period,
      learner, written_at, host, language, grade, country, region, signin_country, cohort
    FROM counted_lines
    WHERE event = 'task_accepted'
    UNION ALL
    SELECT
      'month' AS kind, DATE_TRUNC(on_day, MONTH) AS period,
      learner, written_at, host, language, grade, country, region, signin_country, cohort
    FROM counted_lines
    WHERE event = 'task_accepted'
  )
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
WHERE latest.recency = 1;

-- One night's counts land whole or not at all.
BEGIN TRANSACTION;

-- Every day from the first the bucket holds whole to yesterday, a day with
-- nothing on it among them: a missing day is a night that never ran. A day is
-- counted again while the bucket holds the month of the first of the seven
-- days to it, which is every name of its week and of its month. Each day is
-- laid beside the children of the days that reach back to the start of its
-- week or of its month, whichever is earlier, and the rest is joined on the
-- day itself.
DELETE FROM `${impact}.daily`
WHERE day >= lines_from AND DATE_TRUNC(DATE_SUB(day, INTERVAL 6 DAY), MONTH) >= lines_from;
INSERT INTO `${impact}.daily` (day, learners, learners_week, learners_month, tasks, answers, topics_won)
WITH counted_days AS (
  SELECT counted_day
  FROM UNNEST(GENERATE_DATE_ARRAY(lines_from, counted_until)) AS counted_day
  WHERE counted_day NOT IN (SELECT day FROM `${impact}.daily`)
),
children_days AS (
  SELECT DISTINCT on_day, learner FROM counted_lines WHERE event = 'task_accepted'
),
reach AS (
  SELECT
    counted_days.counted_day,
    COUNT(DISTINCT IF(children_days.on_day = counted_days.counted_day, children_days.learner, NULL)) AS learners,
    COUNT(DISTINCT IF(children_days.on_day >= DATE_SUB(counted_days.counted_day, INTERVAL 6 DAY), children_days.learner, NULL)) AS learners_week,
    COUNT(DISTINCT IF(children_days.on_day >= DATE_TRUNC(counted_days.counted_day, MONTH), children_days.learner, NULL)) AS learners_month
  FROM counted_days
  CROSS JOIN children_days
  WHERE children_days.on_day <= counted_days.counted_day
    AND children_days.on_day >= LEAST(DATE_SUB(counted_days.counted_day, INTERVAL 6 DAY), DATE_TRUNC(counted_days.counted_day, MONTH))
  GROUP BY counted_days.counted_day
),
volume AS (
  SELECT
    on_day,
    COUNTIF(event = 'task_accepted') AS tasks,
    COUNTIF(event = 'answer_recorded') AS answers,
    COUNTIF(event = 'topic_mastered') AS topics_won
  FROM counted_lines
  GROUP BY on_day
)
SELECT
  counted_days.counted_day,
  IFNULL(reach.learners, 0),
  IF(DATE_SUB(counted_days.counted_day, INTERVAL 6 DAY) < lines_from, NULL, IFNULL(reach.learners_week, 0)),
  IF(DATE_TRUNC(counted_days.counted_day, MONTH) < lines_from, NULL, IFNULL(reach.learners_month, 0)),
  IFNULL(volume.tasks, 0),
  IFNULL(volume.answers, 0),
  IFNULL(volume.topics_won, 0)
FROM counted_days
LEFT JOIN reach ON reach.counted_day = counted_days.counted_day
LEFT JOIN volume ON volume.on_day = counted_days.counted_day;

-- A week is counted again while the bucket holds the month it begins in.
DELETE FROM `${impact}.learners_weekly` WHERE DATE_TRUNC(week, MONTH) >= lines_from;
INSERT INTO `${impact}.learners_weekly` (week, dimension, value, learners)
SELECT period, dimension, value, COUNT(*)
FROM told_apart
WHERE kind = 'week'
  AND period >= weeks_from
  AND period NOT IN (SELECT week FROM `${impact}.learners_weekly`)
GROUP BY period, dimension, value;

-- A month is counted only whole, and again while the bucket holds it.
DELETE FROM `${impact}.learners_monthly` WHERE month >= months_from;
INSERT INTO `${impact}.learners_monthly` (month, dimension, value, learners)
SELECT period, dimension, value, COUNT(*)
FROM told_apart
WHERE kind = 'month' AND period >= months_from
GROUP BY period, dimension, value;

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
-- made, and how many topics each had mastered at most. A child whose lines
-- name no month its profile was made in has no months of use to be put under.
DELETE FROM `${impact}.learning_monthly` WHERE month >= months_from;
INSERT INTO `${impact}.learning_monthly` (month, tenure, learners, answers, correct, hinted, dont_know, mastered)
SELECT period, tenure, COUNT(*), SUM(answers), SUM(correct), SUM(hinted), SUM(dont_know), SUM(mastered)
FROM (
  SELECT
    period,
    DATE_DIFF(period, made_in, MONTH) AS tenure,
    answers, correct, hinted, dont_know, mastered
  FROM (
    SELECT
      DATE_TRUNC(on_day, MONTH) AS period,
      learner,
      MIN(SAFE.PARSE_DATE('%Y-%m', cohort)) AS made_in,
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
-- counts its own children once. An empty topic or grade means every one; an
-- answer whose line names no grade is counted over every grade alone.
DELETE FROM `${impact}.traps_monthly` WHERE month >= months_from;
INSERT INTO `${impact}.traps_monthly` (month, topic, grade, trap, answers, learners)
SELECT DATE_TRUNC(on_day, MONTH), topic, grade, trap, COUNT(*), COUNT(DISTINCT learner)
FROM counted_lines
WHERE event = 'answer_recorded' AND trap != '' AND grade IS NOT NULL AND on_day >= months_from
GROUP BY 1, 2, 3, 4
UNION ALL
SELECT DATE_TRUNC(on_day, MONTH), topic, CAST(NULL AS INT64), trap, COUNT(*), COUNT(DISTINCT learner)
FROM counted_lines
WHERE event = 'answer_recorded' AND trap != '' AND on_day >= months_from
GROUP BY 1, 2, 4
UNION ALL
SELECT DATE_TRUNC(on_day, MONTH), CAST(NULL AS STRING), grade, trap, COUNT(*), COUNT(DISTINCT learner)
FROM counted_lines
WHERE event = 'answer_recorded' AND trap != '' AND grade IS NOT NULL AND on_day >= months_from
GROUP BY 1, 3, 4;

COMMIT TRANSACTION;
