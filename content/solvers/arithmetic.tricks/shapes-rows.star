# For "each shape stands for a number": rows of shapes on the picture, the
# number beside a row the sum of its shapes, and the row asked about marked ?.
# Trying every shape from 0 to 20 runs out of steps at four shapes, so the
# rows are read first: a row with one kind of shape left unknown says what it
# is. Only where the rows say nothing more is a shape tried, from 0 to the most
# a row it stands in allows, and the rows read again. Every way the rows allow
# is followed, so a picture whose rows allow two answers is told apart.
# From reference task tri-34-d3-6.

def most(rows, shape):
    # The largest value a shape can take: a row's total shared by its copies.
    least = None
    for row in rows:
        copies = len([drawn for drawn in row["shapes"] if drawn == shape])
        if row["total"] != "?" and copies > 0:
            share = int(row["total"]) // copies
            if least == None or share < least:
                least = share
    return least

def settled(rows, known):
    # Every shape the rows give away, or None when they contradict what is known.
    known = dict(known)
    changed = True
    while changed:
        changed = False
        for row in rows:
            if row["total"] == "?":
                continue
            left = [drawn for drawn in row["shapes"] if drawn not in known]
            rest = int(row["total"]) - sum([known[drawn] for drawn in row["shapes"] if drawn in known])
            if not left:
                if rest != 0:
                    return None
            elif len(set(left)) == 1:
                if rest < 0 or rest % len(left) != 0:
                    return None
                known[left[0]] = rest // len(left)
                changed = True
    return known

def solve(options, picture):
    rows = picture["rows"]
    shapes = []
    for row in rows:
        for drawn in row["shapes"]:
            if drawn not in shapes:
                shapes.append(drawn)
    asked = [row for row in rows if row["total"] == "?"][0]
    answers = []
    ways = [{}]
    while ways:
        known = settled(rows, ways.pop())
        if known == None:
            continue
        unknown = [shape for shape in shapes if shape not in known]
        if not unknown:
            answer = sum([known[drawn] for drawn in asked["shapes"]])
            if answer not in answers:
                answers.append(answer)
            continue
        largest = most(rows, unknown[0])
        if largest == None:
            fail("a shape that stands in no row with a total can be any number")
        for value in range(largest + 1):
            tried = dict(known)
            tried[unknown[0]] = value
            ways.append(tried)
    if len(answers) != 1:
        return match(options, "It cannot be told")
    return match(options, answers[0])
