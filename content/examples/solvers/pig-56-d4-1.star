SEASONS = 4
GENDERS = 2  # boy or girl
WANTED = 3  # pupils of one kind

def solve(options):
    kinds = list(product(range(SEASONS), range(GENDERS)))  # a season together with boy or girl
    # Every group that is not yet enough, with fewer than WANTED pupils of
    # each kind, built up one kind at a time; the answer is the first size not
    # among them.
    failed = set([0])
    for _ in kinds:
        failed = set([size + count for size in failed for count in range(WANTED)])
    pupils = 0
    while pupils in failed:
        pupils += 1
    return match(options, pupils)
