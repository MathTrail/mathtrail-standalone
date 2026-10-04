-- How much each child of a month did, in ranges: the tasks handed to it, and
-- the days a task was. Exact numbers, for the owner alone.
SELECT month, measure, bucket, learners
FROM `${impact}.dose_monthly`
