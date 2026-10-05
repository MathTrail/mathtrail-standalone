-- The counts of the children, made again every night from the raw lines the
-- activity bucket keeps for 62 days, into tables that hold no name of any
-- child and are kept for years.
--
-- A child is counted under the name it has that month, and the children the
-- load tool or MCP Inspector handed a task to that month are not counted at
-- all, so counting a period reads every line of every month a name in it was
-- given for. While the bucket still holds every line since the counting began,
-- every night counts every period again. Once it has begun to let lines go, a
-- period is counted again only while the bucket holds those months from their
-- first day, and then keeps the rows it was last counted to. Either way a
-- night that did not run is made up for by the next, and a run made by hand
-- counts what a night would; a day or a week with no row yet is counted from
-- what the bucket holds.

-- Yesterday: the last day whose lines are all in.
DECLARE counted_until DATE DEFAULT DATE_SUB(CURRENT_DATE('UTC'), INTERVAL 1 DAY);

-- The first day the bucket still holds whole.
DECLARE window_from DATE DEFAULT DATE_SUB(CURRENT_DATE('UTC'), INTERVAL ${window_days} DAY);

-- The day the counting began: the first day a night counted, or, until one
-- has, the day after the first line that names a child — a build older than
-- the counting names none, and the day the counting arrived on may have begun
-- without it. Empty while no line names a child.
DECLARE counting_began DATE DEFAULT IFNULL(
  (SELECT MIN(day) FROM `${impact}.daily`),
  (SELECT DATE_ADD(MIN(DATE(timestamp, 'UTC')), INTERVAL 1 DAY) FROM `${logs}`
    WHERE JSON_VALUE(json_payload.learner) IS NOT NULL));

-- The first day counted: the bucket's first whole day, or the counting's first
-- if that came later. A day after it with no line on it is a day nobody had a
-- task, and is counted as one.
DECLARE lines_from DATE DEFAULT GREATEST(window_from, IFNULL(counting_began, CURRENT_DATE('UTC')));

-- The first day a month must begin on for the bucket to hold all of it: none
-- while the bucket still holds every line since the counting began, when
-- there is nothing before that to hold.
DECLARE names_from DATE DEFAULT IF(
  IFNULL(counting_began, CURRENT_DATE('UTC')) >= window_from,
  DATE '1970-01-01',
  lines_from);

-- The first week and the first month that begin on the first day counted or
-- after it: an earlier one would be counted from part of its days.
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
  SAFE_CAST(SAFE_CAST(JSON_VALUE(json_payload.topics_mastered) AS FLOAT64) AS INT64) AS topics_mastered,
  SAFE_CAST(ROUND(SAFE_CAST(JSON_VALUE(json_payload.chance) AS FLOAT64) * 100) AS INT64) AS chance,
  IFNULL(JSON_VALUE(json_payload.tutor_mode), '') AS tutor_mode,
  IF(JSON_VALUE(json_payload.trial) IS NULL, 0, SAFE_CAST(SAFE_CAST(JSON_VALUE(json_payload.trial) AS FLOAT64) AS INT64)) AS trial,
  IFNULL(JSON_VALUE(json_payload.answers_bucket), '') AS answers_range
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

-- The answers weighed against the chance a task was handed out at, as the
-- report weighs them: to tasks the rule chose, after the trial series and
-- without the hint, each with its chance in hundredths and the range of the
-- child's answers it falls in. A line the report could not read weighs
-- nothing here either, and a chance out of nothing to a hundred hundredths,
-- which the report would put in its first range or its last, weighs nothing
-- rather than stopping the night.
CREATE TEMP TABLE weighed AS
SELECT DATE_TRUNC(on_day, MONTH) AS period, learner, chance, answers_range, correct
FROM counted_lines
WHERE event = 'answer_recorded'
  AND on_day >= months_from
  AND chance BETWEEN 0 AND 100
  AND tutor_mode = 'rule'
  AND trial = 0
  AND NOT hinted
  AND answers_range != '';

-- The children of each week and of each month, told apart by one thing at a
-- time. A child is counted under the value of its latest task of the period,
-- so that the values of a dimension add up to all and a child who moved from
-- one host or grade to another is counted once.
CREATE TEMP TABLE told_apart AS
SELECT latest.kind, latest.period, told.dimension, told.value
FROM (
  SELECT
    periods.kind, periods.period,
    tasks.host, tasks.language, tasks.grade, tasks.country, tasks.region, tasks.signin_country, tasks.cohort,
    ROW_NUMBER() OVER (PARTITION BY periods.kind, periods.period, tasks.learner ORDER BY tasks.written_at DESC) AS recency
  FROM counted_lines AS tasks,
  UNNEST([
    STRUCT('week' AS kind, DATE_TRUNC(tasks.on_day, WEEK(MONDAY)) AS period),
    STRUCT('month' AS kind, DATE_TRUNC(tasks.on_day, MONTH) AS period)
  ]) AS periods
  WHERE tasks.event = 'task_accepted'
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

-- Every day from the first counted to yesterday, a day with nothing on it
-- among them: a missing day is a night that never ran. A day is counted again
-- while the bucket holds the month of the first of the seven days to it, which
-- is every name of its week and of its month. Each day is laid beside the
-- children of the days that reach back to the start of its week or of its
-- month, whichever is earlier, and the rest is joined on the day itself.
DELETE FROM `${impact}.daily`
WHERE day >= lines_from AND DATE_TRUNC(DATE_SUB(day, INTERVAL 6 DAY), MONTH) >= names_from;
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
DELETE FROM `${impact}.learners_weekly` WHERE DATE_TRUNC(week, MONTH) >= names_from;
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

-- How the answers weighed came out against the chance promised: by the range
-- of that chance, in the report's ranges, and over every range, so that each
-- row counts its own children once. An empty range means every one.
DELETE FROM `${impact}.chance_monthly` WHERE month >= months_from;
INSERT INTO `${impact}.chance_monthly` (month, bucket, learners, answers, correct, promised)
SELECT
  period,
  CASE
    WHEN chance <= 49 THEN 'under 0.50'
    WHEN chance <= 59 THEN '0.50-0.59'
    WHEN chance <= 69 THEN '0.60-0.69'
    WHEN chance <= 77 THEN '0.70-0.77'
    WHEN chance <= 85 THEN '0.78-0.85'
    ELSE 'over 0.85' END,
  COUNT(DISTINCT learner), COUNT(*), COUNTIF(correct), SUM(chance)
FROM weighed
GROUP BY 1, 2
UNION ALL
SELECT period, CAST(NULL AS STRING), COUNT(DISTINCT learner), COUNT(*), COUNTIF(correct), SUM(chance)
FROM weighed
GROUP BY 1;

-- How far the answers weighed came out from the chance promised, by the range
-- of the child's answers they fall in: the sums a public view reads the mean
-- and its standard error by child from, with no child named. A child's margin
-- is its right answers times a hundred less its chances in hundredths.
DELETE FROM `${impact}.kept_up_monthly` WHERE month >= months_from;
INSERT INTO `${impact}.kept_up_monthly` (month, bucket, learners, answers, correct, promised, margins_squared, answers_by_margins, answers_squared)
SELECT
  period, answers_range, COUNT(*), SUM(answers), SUM(correct), SUM(promised),
  SUM(margin * margin), SUM(answers * margin), SUM(answers * answers)
FROM (
  SELECT
    period, answers_range, learner,
    COUNT(*) AS answers,
    COUNTIF(correct) AS correct,
    SUM(chance) AS promised,
    100 * COUNTIF(correct) - SUM(chance) AS margin
  FROM weighed
  GROUP BY period, answers_range, learner
)
GROUP BY period, answers_range;

COMMIT TRANSACTION;
