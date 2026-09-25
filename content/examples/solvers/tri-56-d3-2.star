RESULT = 50  # what Anya got

def fits(number):
    # The number, half of it and 5 more.
    return number + number / 2 + 5 == RESULT

def solve(options):
    # Nothing says the number is whole, so halves and quarters are tried too;
    # each of them is exact as a fraction of a power of two.
    fitting = [n / 4 for n in range(4001) if fits(n / 4)]
    if len(fitting) != 1:
        fail("%d numbers fit the story" % len(fitting))
    return match(options, fitting[0])
