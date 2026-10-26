NUMBERS = [1, 2, 3, 4, 5]

def picks_up(left, even):
    # A plan: a card of the wanted kind, even or odd, while there is one,
    # and otherwise whichever card is left first.
    wanted = [n for n in left if (n % 2 == 0) == even]
    gone = wanted[0] if wanted else left[0]
    return [n for n in left if n != gone]

def ann_plan_wins(even):
    # Ann's plan against every card Ben may pick up: Ben picks up the first
    # and the third card, Ann the second and the fourth.
    for ben_first in NUMBERS:
        left = picks_up([n for n in NUMBERS if n != ben_first], even)
        for ben_second in left:
            if picks_up([n for n in left if n != ben_second], even)[0] % 2 == 0:
                return False
    return True

def ben_plan_wins(even):
    # Ben's plan against every card Ann may pick up.
    opening = picks_up(NUMBERS, even)
    for ann_first in opening:
        left = picks_up([n for n in opening if n != ann_first], even)
        for ann_second in left:
            if [n for n in left if n != ann_second][0] % 2 == 1:
                return False
    return True

def solve(options):
    # Every set of cards still on the table, as bits, smaller sets first: a
    # set of one card is Ann's when its number is odd; from a larger one Ben
    # moves when an odd count is left and Ann when an even one is.
    wins = {}
    for mask in range(1, 32):
        left = [i for i in range(5) if mask & (1 << i)]
        if len(left) == 1:
            wins[mask] = NUMBERS[left[0]] % 2 == 1
        elif len(left) % 2 == 1:
            wins[mask] = all([wins[mask - (1 << i)] for i in left])
        else:
            wins[mask] = any([wins[mask - (1 << i)] for i in left])
    # Each answer, held to the search: with no chance and no draw, one of
    # the two can always make sure of winning, so luck decides nothing.
    claims = {
        "Ann, by picking up an even card each time she can": ann_plan_wins(True),
        "Ben, by picking up an odd card each time": ben_plan_wins(False),
        "Ben, since he goes first": not wins[31],
        "Ann, by picking up an odd card each time": ann_plan_wins(False),
        "Nobody can be sure: it is a matter of luck": False,
    }
    true = [claim for claim, holds in claims.items() if holds]
    if len(true) != 1:
        fail("%d of the answers hold" % len(true))
    return match(options, true[0])
