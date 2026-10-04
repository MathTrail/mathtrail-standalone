-- The children of each week, told apart by one thing at a time, each under the
-- value of its latest task of the week. A week that spans the turn of a month
-- counts a child twice. Exact numbers, for the owner alone.
SELECT week, dimension, value, learners
FROM `${impact}.learners_weekly`
