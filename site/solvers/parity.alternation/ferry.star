# A ferry starts at the left bank and crosses the river 7 times: after each
# crossing, make every bank it can be at, which is the other one. The steps
# say that after an odd number of crossings it is at the right bank and after
# an even number at the left, so that is checked crossing by crossing.
# Written from the template reachable of parity.alternation.

CROSSINGS = 7
BANKS = ["left bank", "right bank"]  # as the answer names them

def solve(options):
    banks = set([BANKS[0]])
    for crossing in range(1, CROSSINGS + 1):
        banks = set([BANKS[1 - BANKS.index(bank)] for bank in banks])
        said = BANKS[1] if crossing % 2 == 1 else BANKS[0]
        if len(banks) != 1 or said not in banks:
            fail("after %d crossings the ferry is not where odd and even say" % crossing)
    return match(options, list(banks)[0])
