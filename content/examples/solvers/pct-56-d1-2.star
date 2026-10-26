def solve(options):
    # Every whole height in cm that is 150% more than the old 40 cm.
    old, percent = 40, 150
    heights = [h for h in range(0, 1001) if (h - old) * 100 == old * percent]
    if len(heights) != 1:
        fail("%d heights fit" % len(heights))
    return match(options, "%d cm" % heights[0])
