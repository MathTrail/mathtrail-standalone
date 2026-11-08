# A triangle with two lines from its top corner down to its base: the base is
# cut at 4 points, its two ends among them, and every triangle of the drawing
# has its top at the corner and its base between two of those points. Try
# every pair of points of the base.
# Written from the template triangles-on-lines of counting.figures.

POINTS = 4  # the points the lines from the top meet the base at, its ends among them

def solve(options):
    count = 0
    for left in range(POINTS):
        for right in range(left + 1, POINTS):
            count += 1
    return match(options, count)
