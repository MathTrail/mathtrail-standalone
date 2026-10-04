-- The children of each month, told apart by one thing at a time, each under
-- the value of its latest task of the month, and the share of the month's
-- children each value holds. Exact numbers, for the owner alone.
SELECT
  counted.month,
  counted.dimension,
  counted.value,
  counted.learners,
  SAFE_DIVIDE(counted.learners, everyone.learners) AS share
FROM `${impact}.learners_monthly` AS counted
JOIN `${impact}.learners_monthly` AS everyone
  ON everyone.month = counted.month AND everyone.dimension = 'all'
