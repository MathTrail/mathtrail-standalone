ROWS = 5
COLUMNS = 8

def split_off(piece):
    # Tear a single row, or else a single stamp, off the piece.
    rows, columns = piece
    if rows > 1:
        return [(1, columns), (rows - 1, columns)]
    return [(1, 1), (1, columns - 1)]

def halve(piece):
    # Tear the piece as near its middle as the lines of holes allow, across its longer side.
    rows, columns = piece
    if rows >= columns:
        return [(rows // 2, columns), (rows - rows // 2, columns)]
    return [(rows, columns // 2), (rows, columns - columns // 2)]

def tears(split):
    # Tear one piece at a time until every piece is a single stamp.
    pieces = [(ROWS, COLUMNS)]  # each piece as its rows and columns
    count = 0
    while any([rows * columns > 1 for rows, columns in pieces]):
        whole = [piece for piece in pieces if piece[0] * piece[1] > 1][0]
        pieces.remove(whole)
        pieces += split(whole)
        count += 1
    return count

def solve(options):
    # Two very different orders of tearing have to need the same number of tears.
    counts = set([tears(split_off), tears(halve)])
    if len(counts) != 1:
        fail("the orders need %d different numbers of tears" % len(counts))
    return match(options, list(counts)[0])
