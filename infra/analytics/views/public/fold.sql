-- The rule a public view of children told apart by one thing keeps to, so
-- that no group of fewer than ten children is shown and none is left out to be
-- worked out from the rest: the categories of fewer than ten go into one,
-- (others), and while (others) holds fewer than ten, the smallest of the rest
-- go in with them. A group of fewer than ten in all folds whole, and shows
-- nothing. It reads categories, each group's categories with their children,
-- and says what each category is shown as. It holds only where a child is in
-- one category of a group, or (others) would count one child more than once.
ranked AS (
  SELECT
    grp,
    value,
    ROW_NUMBER() OVER (PARTITION BY grp ORDER BY learners, value) AS place,
    SUM(learners) OVER (PARTITION BY grp ORDER BY learners, value ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW) AS running,
    COUNTIF(learners < 10) OVER (PARTITION BY grp) AS small
  FROM categories
),
cut AS (
  SELECT
    grp,
    IF(MAX(small) = 0, 0, IFNULL(MIN(IF(place >= small AND running >= 10, place, NULL)), MAX(place))) AS folded
  FROM ranked
  GROUP BY grp
),
labels AS (
  SELECT ranked.grp, ranked.value, IF(ranked.place <= cut.folded, '(others)', ranked.value) AS label
  FROM ranked
  JOIN cut ON cut.grp = ranked.grp
)
