PIECES = [2, 4, 6, 8, 10]  # cm, what is left
BITE = 3  # cm eaten at each bite

def solve(options):
    # Undo the bites one at a time: each bite split a piece in two and ate the
    # 3 cm between them, so undoing it joins two pieces with those 3 cm, until
    # the lace is one piece again.
    pieces = list(PIECES)
    while len(pieces) > 1:
        first = pieces.pop()
        second = pieces.pop()
        pieces.append(first + BITE + second)
    return match(options, "%d cm" % pieces[0])
