# For "what is ...": work it out the long way, term by term, and not by the
# trick the task is about. The trick is what the child is to find; a solver
# that uses it only repeats your reasoning instead of checking it.
# From reference task tri-12-d5-5.

FIRST = 100  # 100 - 99 + 98 - 97 + ... + 2 - 1: even terms added, odd ones taken away

def solve(options):
    return match(options, sum([n if n % 2 == 0 else -n for n in range(FIRST, 0, -1)]))
