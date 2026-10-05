-- How far the answers of each month came out from the chance promised, by the
-- range of the child's answers they fall in: right less the chance, on
-- average over the answers, with its standard error counted by child, as the
-- public view counts it. Exact numbers, for the owner alone.
WITH ranges AS (
  SELECT
    month, bucket, learners, answers, margins_squared, answers_by_margins, answers_squared,
    SAFE_DIVIDE(100 * correct - promised, answers) AS mean_margin
  FROM `${impact}.kept_up_monthly`
)
SELECT
  month,
  bucket AS answers_range,
  learners,
  answers,
  mean_margin / 100 AS came_true_less_promised,
  SAFE_DIVIDE(
    SQRT(SAFE_DIVIDE(learners, learners - 1) * GREATEST(
      margins_squared - 2 * mean_margin * answers_by_margins + mean_margin * mean_margin * answers_squared, 0)),
    100 * answers) AS standard_error
FROM ranges
