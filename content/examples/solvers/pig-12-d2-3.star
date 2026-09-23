def solve(options):
    items = 13
    boxes = 12
    # Too many ways to try one by one: the fullest box is smallest when the
    # items are spread as evenly as they go.
    for most in range(items + 1):
        if boxes * most >= items:
            return match(options, most)
    fail("no number of items in a box is ever enough")
