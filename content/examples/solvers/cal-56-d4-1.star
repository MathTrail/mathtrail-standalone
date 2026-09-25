NOW = 73  # the four ages together now
BEFORE = 54  # the four ages together five years ago
YEARS = 5
GROWN_UP = 18  # a parent is at least this old

def then(age):
    # Somebody not yet born five years ago adds nothing to the sum then.
    return age - YEARS if age >= YEARS else 0

def solve(options):
    younger = set()
    for young in range(60):
        for old in range(young, 60):
            parents = NOW - young - old  # the two parents together
            if parents >= 2 * GROWN_UP and parents - 2 * YEARS + then(old) + then(young) == BEFORE:
                younger.add(young)
    if len(younger) != 1:
        fail("%d ages of the younger child fit" % len(younger))
    return match(options, list(younger)[0])
