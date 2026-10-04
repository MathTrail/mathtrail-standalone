-- Each topic of each month: the children who answered its tasks, the shares of
-- right answers and of I don't know, and the times it was won. Exact numbers,
-- for the owner alone.
SELECT
  month,
  topic,
  learners,
  answers,
  SAFE_DIVIDE(correct, answers) AS correct_share,
  SAFE_DIVIDE(dont_know, answers) AS dont_know_share,
  won,
  winners
FROM `${impact}.topics_monthly`
