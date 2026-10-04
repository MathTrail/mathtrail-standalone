-- The traps the wrong answers of each month fell into, topic by topic over
-- every grade, where ten children or more fell into the trap.
SELECT
  counted.month,
  counted.topic,
  counted.trap,
  CAST(ROUND(counted.learners / 5) * 5 AS INT64) AS learners,
  CAST(ROUND(counted.answers / 5) * 5 AS INT64) AS answers
FROM `${impact}.traps_monthly` AS counted
JOIN `${public}.closed_months` AS closed ON closed.month = counted.month
WHERE counted.topic IS NOT NULL AND counted.grade IS NULL AND counted.learners >= 10
