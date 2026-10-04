-- The main numbers of the latest months as the public views show them, for a
-- grant application: whole months of ten children or more, every number to the
-- nearest five, and none of a group of fewer than ten.
SELECT
  FORMAT_DATE('%Y-%m', counted.month) AS month,
  counted.learners,
  counted.new_learners,
  us.learners AS in_us,
  signin.learners AS signed_in_from_us,
  counted.tasks,
  counted.answers,
  counted.tasks_per_child_week,
  counted.active_days_per_child,
  counted.topics_won
FROM `${public}.months` AS counted
LEFT JOIN `${public}.learners_by_country` AS us
  ON us.month = counted.month AND us.value = 'US'
LEFT JOIN `${public}.learners_by_signin_country` AS signin
  ON signin.month = counted.month AND signin.value = 'US'
ORDER BY counted.month DESC
LIMIT ${months}
