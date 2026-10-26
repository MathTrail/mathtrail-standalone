def folded(index, line):
    # Where a row or a column lands when the sheet is folded along the line
    # after that many rows or columns: the part past the line lies on the rest.
    if index < line:
        return index
    return 2 * line - 1 - index

def solve(options):
    # A 4 by 4 sheet folded in half one way and then the other; every cell of
    # the folded 2 by 2 square may be the one cut out.
    counts = set()
    for cut_row in range(2):
        for cut_column in range(2):
            missing = 0
            for row in range(4):
                for column in range(4):
                    if folded(row, 2) == cut_row and folded(column, 2) == cut_column:
                        missing += 1
            counts.add(missing)
    if len(counts) != 1:
        fail("the cuts leave %d answers" % len(counts))
    return match(options, list(counts)[0])
