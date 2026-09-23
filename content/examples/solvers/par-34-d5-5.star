def solve(options):
    # Every break turns one piece into two, so each one adds exactly one piece
    # whichever piece is broken and wherever. A bar of 24 squares therefore
    # takes 23 breaks in every game, and the number of breaks alone decides who
    # makes the last one. The prototype sampled two hundred games to show it;
    # the invariant is the proof itself.
    squares = 4 * 6
    breaks = squares - 1
    return match(options, "Anya always wins" if breaks % 2 == 1 else "Borya always wins")
