MATCHES = 5
BOXES = ["home team wins", "away team wins", "draw"]

def solve(options):
    # Every card: one box ticked for each match.
    return match(options, len(product(BOXES, repeat = MATCHES)))
