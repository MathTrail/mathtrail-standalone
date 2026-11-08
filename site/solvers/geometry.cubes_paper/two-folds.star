# A square sheet of 4 by 4 cells folded in half, left over right, then in half
# again, top over bottom, and punched once: count the layers on the cell the
# punch goes through, each a hole in the opened sheet.
# Written from the template fold-sheet of geometry.cubes_paper.

SIDE = 4  # the sheet is SIDE by SIDE cells
FOLDS = ["across", "down"]  # left over right at the middle, then top over bottom
PUNCH = (3, 3)  # the cell of the folded sheet the hole goes through

def solve(options):
    layers = {}
    for x in range(SIDE):
        for y in range(SIDE):
            layers[(x, y)] = 1
    half = SIDE // 2
    for fold in FOLDS:
        moved = {}
        for cell in layers:
            x, y = cell
            if fold == "across" and x < half:
                cell = (2 * half - 1 - x, y)
            elif fold == "down" and y < half:
                cell = (x, 2 * half - 1 - y)
            moved[cell] = moved.get(cell, 0) + layers[(x, y)]
        layers = moved
    return match(options, layers.get(PUNCH, 0))
