PUPILS = 21

def solve(options):
    # From the most girls down, the first number for which the boys left give
    # every girl a different number of friends among them, 0 included.
    for girls in range(PUPILS, -1, -1):
        boys = PUPILS - girls
        if len(combinations(list(range(0, boys + 1)), girls)) > 0:
            return match(options, girls)
    fail("no number of girls works")
