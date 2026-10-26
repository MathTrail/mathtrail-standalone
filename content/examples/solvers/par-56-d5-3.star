ANN = 15  # games played
BEN = 21

def solve(options):
    # Every game in turn: the winner stays and plays the one who rested, and
    # the loser rests. A state is who rests, who stays, who comes back, and
    # the games each has played; the fewest games of Cat among the states
    # that end with Ann on 15 and Ben on 21.
    states = set()
    for rests in ["Ann", "Ben", "Cat"]:
        players = [p for p in ["Ann", "Ben", "Cat"] if p != rests]
        for winner in players:
            loser = players[0] if winner == players[1] else players[1]
            states.add((loser, winner, rests, 1 if "Ann" in players else 0, 1 if "Ben" in players else 0, 1 if "Cat" in players else 0))
    fewest = None
    for game in range(1, 2 * (ANN + BEN)):
        for state in states:
            if state[3] == ANN and state[4] == BEN and (fewest == None or state[5] < fewest):
                fewest = state[5]
        later = set()
        for rests, stays, comes, ann, ben, cat in states:
            played = {"Ann": ann, "Ben": ben, "Cat": cat}
            played[stays] += 1
            played[comes] += 1
            if played["Ann"] > ANN or played["Ben"] > BEN:
                continue
            for winner in [stays, comes]:
                loser = comes if winner == stays else stays
                later.add((loser, winner, rests, played["Ann"], played["Ben"], played["Cat"]))
        states = later
    if fewest == None:
        fail("no order of games fits")
    return match(options, fewest)
