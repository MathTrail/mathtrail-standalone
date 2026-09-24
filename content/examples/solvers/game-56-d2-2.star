def solve(options):
    goal = 22
    steps = [1, 2, 3]
    # wins[said] says whether the player about to speak after that number can
    # force a win; after 22 there is nothing left to say, and the game is lost.
    wins = [False] * (goal + 1)
    for said in range(goal - 1, -1, -1):
        wins[said] = any([not wins[said + step] for step in steps if said + step <= goal])
    first = [step for step in steps if not wins[step]]  # Kate starts from 0
    if len(first) != 1:
        fail("%d first numbers win, and the question asks for the one" % len(first))
    return match(options, first[0])
