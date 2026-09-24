ROWS = ["A", "B", "C", "D"]  # from the top; columns 1 to 4 from the left
BLOCKED = "B2"

def squares(steps):
    # The squares a route passes through, starting at A1.
    row, column = 0, 1
    passed = [ROWS[row] + str(column)]
    for step in steps:
        if step == "right":
            column += 1
        else:
            row += 1
        passed.append(ROWS[row] + str(column))
    return passed

def solve(options):
    routes = []
    for steps in product(["right", "down"], repeat=6):
        # Three steps right and three down lead from A1 to D4.
        if len([step for step in steps if step == "right"]) == 3 and BLOCKED not in squares(steps):
            routes.append(steps)
    return match(options, len(routes))
