def circle_alternates(children):
    # Every row in which neighbours differ, grown one place at a time: a row
    # with two boys or two girls side by side is out, and stays out.
    rows = [["B"], ["G"]]
    for place in range(1, children):
        rows = [row + [side] for row in rows for side in ["B", "G"] if side != row[-1]]
    # Round a circle, the last child stands beside the first as well.
    return len([row for row in rows if row[-1] != row[0]]) > 0

def solve(options):
    fitting = [n for n in [7, 9, 11, 12, 13] if circle_alternates(n)]
    if len(fitting) != 1:
        fail("%d of the numbers fit" % len(fitting))
    return match(options, fitting[0])
