-- How the answers of each month came out against the chance promised, by the
-- range of that chance, where ten children and thirty answers or more stand
-- behind a range. A child's answers fall in many ranges, so no range is folded
-- in with others: a range of fewer is not shown.
SELECT
  counted.month,
  counted.bucket AS chance,
  CAST(ROUND(counted.learners / 5) * 5 AS INT64) AS learners,
  CAST(ROUND(counted.answers / 5) * 5 AS INT64) AS answers,
  ROUND(SAFE_DIVIDE(counted.promised, 100 * counted.answers), 3) AS promised_mean,
  ROUND(SAFE_DIVIDE(counted.correct, counted.answers), 3) AS correct_share
FROM `${impact}.chance_monthly` AS counted
JOIN `${public}.closed_months` AS closed ON closed.month = counted.month
WHERE counted.bucket IS NOT NULL AND counted.learners >= 10 AND counted.answers >= 30
