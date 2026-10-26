BEN = (2013, 11, 20)  # born
KIM = (2014, 3, 5)
ON = (2026, 6, 1)  # the date the question asks about

def age(born, on):
    # The birthdays that have come by that date, one year at a time.
    years = 0
    for year in range(born[0] + 1, on[0] + 1):
        if (year, born[1], born[2]) <= on:
            years += 1
    return years

def solve(options):
    ben, kim = age(BEN, ON), age(KIM, ON)
    if ben == kim:
        return match(options, "Both are %d" % ben)
    return match(options, "Ben is %d and Kim is %d" % (ben, kim))
