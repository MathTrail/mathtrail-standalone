def solve(options):
    limit = 1000
    # Shared among 9 as far as they go, the nuts left over equal each share.
    fits = [nuts for nuts in range(1, limit + 1) if nuts // 9 == nuts % 9]
    largest = max(fits)
    # A larger number could lie beyond where the search stops, so the answer is
    # trusted only when the search went on at least as far again past it.
    if largest * 2 > limit:
        fail("the largest number of nuts is too near where the search stops")
    return match(options, largest)
