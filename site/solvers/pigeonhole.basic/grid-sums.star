# A 2 by 2 grid filled with 1, 2 and 3, and the sums of its 2 rows, 2 columns
# and 2 diagonals: try every way of filling the grid and look for one whose
# six sums are all different. The sums are the pigeons and the values a sum
# can take the holes, which the steps count as five, so that is checked too.
# Written from the template boxes of pigeonhole.basic.

VALUES = [1, 2, 3]
LINES = [
    (0, 1), (2, 3),  # the rows: the top two cells, the bottom two
    (0, 2), (1, 3),  # the columns
    (0, 3), (1, 2),  # the diagonals
]
HOLES = 5  # the values a sum can take, from 2 to 6

def sums(grid):
    return [grid[a] + grid[b] for a, b in LINES]

def solve(options):
    holes = set([a + b for a, b in product(VALUES, repeat=2)])
    if len(holes) != HOLES:
        fail("a sum takes %d values, and the steps say %d" % (len(holes), HOLES))
    all_different = 0
    for grid in product(VALUES, repeat=4):
        if len(set(sums(grid))) == len(LINES):
            all_different += 1
    return match(options, "Yes, always" if all_different == 0 else "No")
