# 2026 × 2026 − 2027 × 2025: work it out the long way, the big numbers
# multiplied out in full, and not by the trick the steps use. The note tries
# the trick on small numbers, 5 × 5 − 6 × 4, so that is checked too.
# Written from the template long-way of arithmetic.tricks.

def solve(options):
    if 5 * 5 - 6 * 4 != 1:
        fail("the note's small numbers do not give 1")
    return match(options, 2026 * 2026 - 2027 * 2025)
