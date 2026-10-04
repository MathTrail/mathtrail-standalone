-- The children of each month told apart by ${dimension}, by the rule of the
-- public views: no group of fewer than ten, the others folded into (others),
-- each number to the nearest five.
WITH categories AS (
  SELECT counted.month AS grp, counted.value, counted.learners
  FROM `${impact}.learners_monthly` AS counted
  JOIN `${public}.closed_months` AS closed ON closed.month = counted.month
  WHERE counted.dimension = '${dimension}'
),
${fold}
SELECT
  categories.grp AS month,
  labels.label AS value,
  CAST(ROUND(SUM(categories.learners) / 5) * 5 AS INT64) AS learners
FROM categories
JOIN labels ON labels.grp = categories.grp AND labels.value = categories.value
GROUP BY categories.grp, labels.label
HAVING SUM(categories.learners) >= 10
