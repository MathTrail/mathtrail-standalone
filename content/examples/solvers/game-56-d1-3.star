def solve(options):
    # Every game: each move rubs out two of the numbers left and writes their
    # sum, until one number is left. The boards are kept as sorted tuples.
    boards = set([(1, 2, 3, 4, 5, 6)])
    for _ in range(5):
        after = set()
        for board in boards:
            for i, j in combinations(range(len(board)), 2):
                left = [board[k] for k in range(len(board)) if k != i and k != j]
                after.add(tuple(sorted(left + [board[i] + board[j]])))
        boards = after
    last = set([board[0] % 2 for board in boards])
    if last == set([1]):
        return match(options, "Ben, whatever they do")
    fail("the number left is not always odd")
