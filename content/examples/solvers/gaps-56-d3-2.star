ROWS = 3
COLUMNS = 4

def solve(options):
    strips = set()
    for row in range(ROWS):
        for column in range(COLUMNS):
            # The four sides of a pane, named by the line they lie on, so a shared strip is kept once.
            strips.add(("across", row, column))  # above the pane
            strips.add(("across", row + 1, column))  # below it
            strips.add(("upright", row, column))  # on its left
            strips.add(("upright", row, column + 1))  # on its right
    return match(options, len(strips))
