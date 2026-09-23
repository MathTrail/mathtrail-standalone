def ways_to_pay(amount, coins):
    found = 0
    for size in range(1, amount + 1):
        for chosen in combinations_with_replacement(coins, size):
            if sum(chosen) == amount:
                found = found + 1
    return found

def solve(options):
    return match(options, ways_to_pay(5, [1, 2]))
