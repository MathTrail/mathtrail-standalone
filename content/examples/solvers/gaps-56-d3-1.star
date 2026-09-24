ROWS = 4
COLUMNS = 6

def split_off(piece):
    # Break a single row, or else a single square, off the piece.
    rows, columns = piece
    if rows > 1:
        return [(1, columns), (rows - 1, columns)]
    return [(1, 1), (1, columns - 1)]

def halve(piece):
    # Break the piece as near its middle as the grooves allow, across its longer side.
    rows, columns = piece
    if rows >= columns:
        return [(rows // 2, columns), (rows - rows // 2, columns)]
    return [(rows, columns // 2), (rows, columns - columns // 2)]

def breaks(split):
    # Break one piece at a time until every piece is a single square.
    pieces = [(ROWS, COLUMNS)]  # each piece as its rows and columns
    count = 0
    while any([rows * columns > 1 for rows, columns in pieces]):
        whole = [piece for piece in pieces if piece[0] * piece[1] > 1][0]
        pieces.remove(whole)
        pieces += split(whole)
        count += 1
    return count

def solve(options):
    # Two very different orders of breaking have to need the same number of breaks.
    counts = set([breaks(split_off), breaks(halve)])
    if len(counts) != 1:
        fail("the orders need %d different numbers of breaks" % len(counts))
    return match(options, list(counts)[0])
