-- Each topic of each month that ten children or more answered. A child answers
-- many topics, so no topic is folded in with others: the topics of fewer than
-- ten are not shown, and their children cannot be worked out from the rest.
-- The times a topic was won are shown when ten children or more won it.
SELECT
  counted.month,
  counted.topic,
  CAST(ROUND(counted.learners / 5) * 5 AS INT64) AS learners,
  CAST(ROUND(counted.answers / 5) * 5 AS INT64) AS answers,
  ROUND(counted.correct / counted.answers, 3) AS correct_share,
  IF(counted.winners >= 10, CAST(ROUND(counted.won / 5) * 5 AS INT64), NULL) AS won
FROM `${impact}.topics_monthly` AS counted
JOIN `${public}.closed_months` AS closed ON closed.month = counted.month
WHERE counted.learners >= 10
