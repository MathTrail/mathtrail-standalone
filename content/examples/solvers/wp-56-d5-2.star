WEIGHTS = {"A": 1, "B": 5, "C": 2, "D": 6, "E": 3, "F": 4}  # grams of the ball at each corner, round the hexagon
SWAPS = [("A", "D"), ("B", "E"), ("C", "F")]  # the balls at opposite corners
CANDIDATES = [(["A", "E"], ["D"]), (["A"], ["D"]), (["B"], ["E"]), (["A", "B"], ["C", "D"])]  # the weighings the options name, by corner

def result(weighing, swap):
    # What the balance shows after the swap: 1 when the left pan sinks, -1
    # when the right one does, 0 when they balance.
    weights = dict(WEIGHTS)
    weights[swap[0]], weights[swap[1]] = WEIGHTS[swap[1]], WEIGHTS[swap[0]]
    difference = sum([weights[ball] for ball in weighing[0]]) - sum([weights[ball] for ball in weighing[1]])
    return 0 if difference == 0 else (1 if difference > 0 else -1)

def tells(weighing):
    # A weighing tells the swap when each swap tips it its own way.
    return len(set([result(weighing, swap) for swap in SWAPS])) == len(SWAPS)

def side(corners):
    return ("the ball at " if len(corners) == 1 else "the balls at ") + " and ".join(corners)

def written(weighing):
    text = side(weighing[0]) + " against " + side(weighing[1])
    return text[0].upper() + text[1:]

def solve(options):
    # A weighing that tells the swap is also what makes the last option, that
    # no weighing can, wrong.
    fitting = [weighing for weighing in CANDIDATES if tells(weighing)]
    if len(fitting) != 1:
        fail("%d of the weighings tell the swap" % len(fitting))
    return match(options, written(fitting[0]))
