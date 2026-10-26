ONE_LINE = ["A", "B", "C", "D"]
OTHER_LINE = ["E", "F", "G"]  # parallel to the first

def solve(options):
    # Three marked points make a triangle unless all three lie on one line.
    triangles = [corners for corners in combinations(ONE_LINE + OTHER_LINE, 3)
                 if not all([p in ONE_LINE for p in corners]) and not all([p in OTHER_LINE for p in corners])]
    return match(options, len(triangles))
