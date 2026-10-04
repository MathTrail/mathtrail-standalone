-- The days no night counted, from the first day counted to yesterday: the
-- nightly query tells its failures to nobody, and a missing day is one.
SELECT FORMAT_DATE('%Y-%m-%d', missing) AS day
FROM UNNEST(GENERATE_DATE_ARRAY(
  (SELECT MIN(day) FROM `${impact}.daily`),
  DATE_SUB(CURRENT_DATE('UTC'), INTERVAL 1 DAY))) AS missing
WHERE missing NOT IN (SELECT day FROM `${impact}.daily`)
ORDER BY missing
