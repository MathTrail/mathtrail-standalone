-- Each of the 31 days to yesterday, in UTC, with what the night counted of it
-- and what it cost the project as Cloud Billing's standard export has it so
-- far: charged before credits, the credits, which are below nought, and the
-- cost after them. A day of the export is the UTC day its usage began, as the
-- counts' day is, rather than the Pacific day of the billing console, and the
-- export may still add to a day after it. A day the night or the export holds
-- nothing of is empty. The flags are what the page filters on, so that it holds
-- no date of its own: yesterday; the days of yesterday's month, which on the
-- first of a month is the month just ended; and how many days back a day is.
-- The export's table is named after the billing account, so it is read through
-- a wildcard that matches the standard export alone, and no account is named.
-- Exact numbers, for the owner alone.
WITH bounds AS (
  SELECT
    DATE_SUB(CURRENT_DATE('UTC'), INTERVAL 31 DAY) AS first_day,
    DATE_SUB(CURRENT_DATE('UTC'), INTERVAL 1 DAY) AS yesterday
),
days AS (
  SELECT day
  FROM bounds, UNNEST(GENERATE_DATE_ARRAY(bounds.first_day, bounds.yesterday)) AS day
),
costs AS (
  SELECT
    DATE(billed.usage_start_time, 'UTC') AS day,
    SUM(billed.cost) AS charged,
    SUM(IFNULL((SELECT SUM(credit.amount) FROM UNNEST(billed.credits) AS credit), 0)) AS credits
  FROM `${billing}.gcp_billing_export_v1_*` AS billed, bounds
  WHERE billed.project.id = '${project_id}'
    AND billed.usage_start_time >= TIMESTAMP(bounds.first_day)
  GROUP BY day
)
SELECT
  days.day,
  counted.learners,
  counted.learners_week,
  counted.learners_month,
  counted.tasks,
  counted.answers,
  counted.topics_won,
  costs.charged,
  costs.credits,
  costs.charged + costs.credits AS cost,
  days.day = bounds.yesterday AS yesterday,
  DATE_TRUNC(days.day, MONTH) = DATE_TRUNC(bounds.yesterday, MONTH) AS this_month,
  DATE_DIFF(bounds.yesterday, days.day, DAY) + 1 AS days_ago
FROM days
CROSS JOIN bounds
LEFT JOIN `${impact}.daily` AS counted ON counted.day = days.day
LEFT JOIN costs ON costs.day = days.day
