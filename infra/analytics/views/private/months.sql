-- Each month counted, the one under way among them: the children, those whose
-- profile was made that month, what they did, and how often they came back.
-- Tasks a child a week are counted over the days of the month counted so far.
-- Exact numbers, for the owner alone.
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
  SELECT
    month,
    SUM(IF(dimension = 'all', learners, 0)) AS learners,
    SUM(IF(dimension = 'cohort' AND value = FORMAT_DATE('%Y-%m', month), learners, 0)) AS new_learners
  FROM `${impact}.learners_monthly`
  GROUP BY month
)
SELECT
  days.month,
  days.days_counted,
  children.learners,
  children.new_learners,
  days.tasks,
  days.answers,
  days.topics_won,
  days.active_days,
  SAFE_DIVIDE(days.tasks * 7, children.learners * days.days_counted) AS tasks_per_child_week,
  SAFE_DIVIDE(days.active_days, children.learners) AS active_days_per_child
FROM days
JOIN children ON children.month = days.month
