BOWLS = 20
KIKI_AND_LULU = 11
FEWEST = 2  # bowls each monkey ate at least

def solve(options):
    # Every way the bowls could have been eaten, at least two each, with Mo
    # eating more than each of the other three.
    pip = set()
    for kiki in range(FEWEST, KIKI_AND_LULU - FEWEST + 1):
        lulu = KIKI_AND_LULU - kiki
        for mo in range(FEWEST, BOWLS + 1):
            left = BOWLS - KIKI_AND_LULU - mo
            if left >= FEWEST and mo > kiki and mo > lulu and mo > left:
                pip.add(left)
    if not pip:
        fail("no way fits")
    if len(pip) != 1:
        return match(options, "It cannot be told")
    return match(options, list(pip)[0])
