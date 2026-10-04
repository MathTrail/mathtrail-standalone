-- The months a public view shows: those every day of which was counted. A
-- month still under way, the month the counting began in, and a month some of
-- whose nights never ran are none of them.
SELECT DATE_TRUNC(day, MONTH) AS month
FROM `${impact}.daily`
GROUP BY month
HAVING COUNT(*) = EXTRACT(DAY FROM LAST_DAY(MIN(day), MONTH))
