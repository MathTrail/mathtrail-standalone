YEARS = 10  # ten years ago, and in ten years

def fits(son, mum):
    ago = son > YEARS and mum - YEARS == 4 * (son - YEARS)
    ahead = mum + YEARS == 2 * (son + YEARS)
    return ago and ahead

def solve(options):
    fitting = [son for son in range(120) for mum in range(120) if fits(son, mum)]
    if len(fitting) != 1:
        fail("%d ages fit the clues" % len(fitting))
    return match(options, fitting[0])
