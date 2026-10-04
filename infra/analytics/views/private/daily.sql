-- Every day counted: the children a task was handed to that day, over the
-- seven days to it and over its month so far, and the tasks, the answers and
-- the topics won. Exact numbers, for the owner alone.
SELECT day, learners, learners_week, learners_month, tasks, answers, topics_won
FROM `${impact}.daily`
