# A square of 3 by 3 cells drawn with all its lines: try every square of every
# size on the grid's lines and count the ones that fit inside the figure.
# Written from the template squares-on-lines of counting.figures.

SIDE = 3  # the figure is a square of SIDE by SIDE cells

def solve(options):
    count = 0
    for size in range(1, SIDE + 1):
        for x in range(SIDE - size + 1):
            for y in range(SIDE - size + 1):
                count += 1
    return match(options, count)
