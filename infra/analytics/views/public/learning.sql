-- How the children of each month answered, by the months since their profile
-- was made, by the rule of the public views: a child is in one group, so the
-- groups of fewer than ten fold into (others), and every share and mean stands
-- on ten children or more.
WITH categories AS (
  SELECT
    counted.month AS grp,
    CAST(counted.tenure AS STRING) AS value,
    counted.learners, counted.answers, counted.correct, counted.hinted, counted.dont_know, counted.mastered
  FROM `${impact}.learning_monthly` AS counted
  JOIN `${public}.closed_months` AS closed ON closed.month = counted.month
),
${fold}
SELECT
  categories.grp AS month,
  labels.label AS tenure,
  CAST(ROUND(SUM(categories.learners) / 5) * 5 AS INT64) AS learners,
  CAST(ROUND(SUM(categories.answers) / 5) * 5 AS INT64) AS answers,
  ROUND(SUM(categories.correct) / SUM(categories.answers), 3) AS correct_share,
  ROUND(SUM(categories.hinted) / SUM(categories.answers), 3) AS hint_share,
  ROUND(SUM(categories.dont_know) / SUM(categories.answers), 3) AS dont_know_share,
  ROUND(SUM(categories.mastered) / SUM(categories.learners), 1) AS topics_mastered_mean
FROM categories
JOIN labels ON labels.grp = categories.grp AND labels.value = categories.value
GROUP BY categories.grp, labels.label
HAVING SUM(categories.learners) >= 10
