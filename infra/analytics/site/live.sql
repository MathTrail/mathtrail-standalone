-- The live numbers of the page "Research" as the public views show them, in
-- the one value the site's snapshot is written from: the latest month counted
-- whole, its answers over every range of the chance promised, and the ranges
-- of chance and of the child's answers the views show of it, from the lowest.
-- A range of chance is ordered by the chance it promised on average, which
-- lies within it, and a range of answers by its first answer. Before a month
-- is counted whole there is no month, and a month of too few children shows
-- neither its answers nor a range. It reads the public views and nothing else.
WITH latest AS (
  SELECT MAX(month) AS month FROM `${public}.closed_months`
)
SELECT TO_JSON_STRING(STRUCT(
  (SELECT FORMAT_DATE('%Y-%m', month) FROM latest) AS month,
  (
    SELECT AS STRUCT shown.learners, shown.answers, shown.promised_mean, shown.correct_share
    FROM `${public}.chances_total` AS shown
    JOIN latest ON latest.month = shown.month
  ) AS total,
  ARRAY(
    SELECT AS STRUCT shown.chance, shown.learners, shown.answers, shown.promised_mean, shown.correct_share
    FROM `${public}.chances` AS shown
    JOIN latest ON latest.month = shown.month
    ORDER BY shown.promised_mean
  ) AS chances,
  ARRAY(
    SELECT AS STRUCT
      shown.answers_range, shown.learners, shown.answers, shown.came_true_less_promised, shown.standard_error
    FROM `${public}.kept_up` AS shown
    JOIN latest ON latest.month = shown.month
    ORDER BY SAFE_CAST(REGEXP_EXTRACT(shown.answers_range, r'^[0-9]+') AS INT64)
  ) AS kept_up
)) AS live
