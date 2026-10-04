-- The main numbers of the latest months in exact numbers, for the owner: every
-- month counted, the one under way among them.
SELECT
  FORMAT_DATE('%Y-%m', counted.month) AS month,
  counted.learners,
  counted.new_learners,
  us.learners AS in_us,
  signin.learners AS signed_in_from_us,
  counted.tasks,
  counted.answers,
  ROUND(counted.tasks_per_child_week, 1) AS tasks_per_child_week,
  ROUND(counted.active_days_per_child, 1) AS active_days_per_child,
  counted.topics_won
FROM `${private}.months` AS counted
LEFT JOIN `${private}.learners` AS us
  ON us.month = counted.month AND us.dimension = 'country' AND us.value = 'US'
LEFT JOIN `${private}.learners` AS signin
  ON signin.month = counted.month AND signin.dimension = 'signin_country' AND signin.value = 'US'
ORDER BY counted.month DESC
LIMIT ${months}
