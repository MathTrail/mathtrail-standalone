COLOURS = ["red", "white", "blue"]
BOARDS = 4

def allowed(fence):
    # Boards next to each other differ, and the first board is not red.
    neighbours_differ = all([fence[i] != fence[i + 1] for i in range(len(fence) - 1)])
    return neighbours_differ and fence[0] != "red"

def solve(options):
    return match(options, len([fence for fence in product(COLOURS, repeat=BOARDS) if allowed(fence)]))
