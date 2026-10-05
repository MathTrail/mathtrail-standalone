-- How the answers of each month came out against the chance promised, by the
-- range of that chance and over every range, an empty range: the chance
-- promised on average and the share right. Exact numbers, for the owner alone.
SELECT
  month,
  bucket AS chance,
  learners,
  answers,
  SAFE_DIVIDE(promised, 100 * answers) AS promised_mean,
  SAFE_DIVIDE(correct, answers) AS correct_share
FROM `${impact}.chance_monthly`
