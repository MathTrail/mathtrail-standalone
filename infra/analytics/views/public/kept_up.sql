-- How far the answers of each month came out from the chance promised, by the
-- range of the child's answers they fall in, where ten children and thirty
-- answers or more stand behind a range: right less the chance, on average over
-- the answers, with its standard error counted by child from the sums the
-- night keeps, which name no child. A child's answers fall in many ranges, so
-- no range is folded in with others. The margins are in hundredths, and the
-- sum under the root is held at zero or more, where rounding may take an
-- error of nothing a hair below it.
WITH shown AS (
  SELECT
    counted.month, counted.bucket, counted.learners, counted.answers,
    counted.margins_squared, counted.answers_by_margins, counted.answers_squared,
    SAFE_DIVIDE(100 * counted.correct - counted.promised, counted.answers) AS mean_margin
  FROM `${impact}.kept_up_monthly` AS counted
  JOIN `${public}.closed_months` AS closed ON closed.month = counted.month
  WHERE counted.learners >= 10 AND counted.answers >= 30
)
SELECT
  month,
  bucket AS answers_range,
  CAST(ROUND(learners / 5) * 5 AS INT64) AS learners,
  CAST(ROUND(answers / 5) * 5 AS INT64) AS answers,
  ROUND(mean_margin / 100, 3) AS came_true_less_promised,
  ROUND(SAFE_DIVIDE(
    SQRT(SAFE_DIVIDE(learners, learners - 1) * GREATEST(
      margins_squared - 2 * mean_margin * answers_by_margins + mean_margin * mean_margin * answers_squared, 0)),
    100 * answers), 3) AS standard_error
FROM shown
