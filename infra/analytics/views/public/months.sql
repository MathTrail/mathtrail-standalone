-- Each month with ten children or more: how many, how many began that month,
-- what they did and how often they came back. Every number is to the nearest
-- five and none of fewer than ten is shown; every mean is to one place. The
-- new children are those of the month's own cohort as the view by cohort shows
-- them, so that the two views cannot be set against each other to find a
-- smaller group.
WITH days AS (
  SELECT
    DATE_TRUNC(day, MONTH) AS month,
    COUNT(*) AS days_counted,
    SUM(learners) AS active_days,
    SUM(tasks) AS tasks,
    SUM(answers) AS answers,
    SUM(topics_won) AS topics_won
  FROM `${impact}.daily`
  GROUP BY month
),
children AS (
  SELECT month, learners
  FROM `${impact}.learners_monthly`
  WHERE dimension = 'all'
)
SELECT
  closed.month,
  CAST(ROUND(children.learners / 5) * 5 AS INT64) AS learners,
  cohorts.learners AS new_learners,
  IF(days.tasks >= 10, CAST(ROUND(days.tasks / 5) * 5 AS INT64), NULL) AS tasks,
  IF(days.answers >= 10, CAST(ROUND(days.answers / 5) * 5 AS INT64), NULL) AS answers,
  IF(days.topics_won >= 10, CAST(ROUND(days.topics_won / 5) * 5 AS INT64), NULL) AS topics_won,
  ROUND(SAFE_DIVIDE(days.tasks * 7, children.learners * days.days_counted), 1) AS tasks_per_child_week,
  ROUND(SAFE_DIVIDE(days.active_days, children.learners), 1) AS active_days_per_child
FROM `${public}.closed_months` AS closed
JOIN days ON days.month = closed.month
JOIN children ON children.month = closed.month
LEFT JOIN `${public}.learners_by_cohort` AS cohorts
  ON cohorts.month = closed.month AND cohorts.value = FORMAT_DATE('%Y-%m', closed.month)
WHERE children.learners >= 10
