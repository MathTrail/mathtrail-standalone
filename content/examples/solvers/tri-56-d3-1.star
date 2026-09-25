DEPENDS = "It depends on the number"  # as the option writes it

def result(number):
    # Add 5, double, take 4 away, halve, take the number away; the halving is of an even number.
    after = (number + 5) * 2 - 4
    return after // 2 - number

def solve(options):
    results = set([result(n) for n in range(-100, 101)])
    return match(options, list(results)[0] if len(results) == 1 else DEPENDS)
