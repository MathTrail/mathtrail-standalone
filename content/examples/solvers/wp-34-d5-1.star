def answers(state):
    # Each coin that may be either kind is two possible answers, and each coin
    # that can only be heavy, or only light, is one.
    either, heavy, light, real = state
    return 2 * either + heavy + light

def settles(state, weighings, before, coins):
    # Whether some weighing leaves every one of its three outcomes settled in
    # one weighing fewer.
    if answers(state) <= 1:
        return True
    outcomes = 1
    for w in range(weighings):
        outcomes *= 3
    if answers(state) > outcomes:
        return False
    either, heavy, light, real = state
    for e1 in range(either + 1):
        for e2 in range(either + 1 - e1):
            for h1 in range(heavy + 1):
                for h2 in range(heavy + 1 - h1):
                    for l1 in range(light + 1):
                        for l2 in range(light + 1 - l1):
                            left = e1 + h1 + l1
                            right = e2 + h2 + l2
                            # Coins known to be real make up the pan that has fewer.
                            if left + right == 0 or abs(left - right) > real:
                                continue
                            balance = (either - e1 - e2, heavy - h1 - h2, light - l1 - l2, real + left + right)
                            left_down = (0, e1 + h1, e2 + l2, coins - (e1 + h1 + e2 + l2))
                            right_down = (0, e2 + h2, e1 + l1, coins - (e2 + h2 + e1 + l1))
                            if before[balance] and before[left_down] and before[right_down]:
                                return True
    return False

def fake_weighings(coins):
    # A state counts the coins that may be heavy or light, those that can only
    # be heavy, those that can only be light, and those surely real. Starting
    # with every coin unknown, a weighing only ever leads to states of two
    # kinds — some unknown and the rest real, or none unknown and the rest
    # heavy, light or real — so those are every state there is to settle.
    states = [(either, 0, 0, coins - either) for either in range(coins + 1)]
    for heavy in range(coins + 1):
        for light in range(coins + 1 - heavy):
            if heavy + light > 0:
                states.append((0, heavy, light, coins - heavy - light))
    # Recursion is off, so which states can be settled is filled in from the
    # bottom: none left to find first, then one weighing more at a time.
    start = (coins, 0, 0, 0)
    settled = {state: answers(state) <= 1 for state in states}
    weighings = 0
    while not settled[start]:
        weighings += 1
        if weighings > coins:
            fail("no number of weighings is ever enough")
        settled = {state: settles(state, weighings, settled, coins) for state in states}
    return weighings

def solve(options):
    return match(options, fake_weighings(12))
