def solve(options):
    coins = 8
    pieces = [5 * 3, 3 * 3]  # every loaf cut into 3 equal pieces
    # Three people eat equal shares of all the pieces.
    eaten = [e for e in range(0, 100) if e * 3 == pieces[0] + pieces[1]]
    if len(eaten) != 1:
        fail("the bread does not split into three equal shares")
    given = [have - eaten[0] for have in pieces]
    # Every split of the coins, kept when it pays the same for every piece given.
    first = [c for c in range(0, coins + 1) if c * given[1] == (coins - c) * given[0]]
    if len(first) != 1:
        fail("%d splits of the coins are fair" % len(first))
    return match(options, first[0])
