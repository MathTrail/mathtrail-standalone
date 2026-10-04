-- How the children of each month answered, by the months since their profile
-- was made: the shares of right answers, of answers after the hint and of I
-- don't know, and the topics a child had mastered, on average. Exact numbers,
-- for the owner alone.
SELECT
  month,
  tenure,
  learners,
  answers,
  SAFE_DIVIDE(correct, answers) AS correct_share,
  SAFE_DIVIDE(hinted, answers) AS hint_share,
  SAFE_DIVIDE(dont_know, answers) AS dont_know_share,
  SAFE_DIVIDE(mastered, learners) AS topics_mastered_mean
FROM `${impact}.learning_monthly`
