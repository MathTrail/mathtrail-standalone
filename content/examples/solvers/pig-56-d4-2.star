DIVISOR = 3
GROUP = 3  # numbers whose sum must divide by DIVISOR

def always_found(count):
    # Only the remainders on dividing by DIVISOR decide whether a sum divides,
    # and numbers can leave any remainders, so every way `count` numbers can
    # leave them is tried.
    for remainders in product(range(DIVISOR), repeat=count):
        if not any([sum(group) % DIVISOR == 0 for group in combinations(remainders, GROUP)]):
            return False
    return True

def solve(options):
    count = GROUP
    while not always_found(count):
        count += 1
    return match(options, count)
