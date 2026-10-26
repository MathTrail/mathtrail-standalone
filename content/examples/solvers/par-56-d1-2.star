BOYS = 7

MOST_GIRLS = 14  # the largest number the options name

def fits(circle):
    # Every child holds the hands of two boys or of two girls.
    size = len(circle)
    return all([circle[i - 1] == circle[(i + 1) % size] for i in range(size)])

def solve(options):
    # Every circle of the boys and from 1 to 14 girls, built child by child
    # round it. A child placed must be like the one two places back, since
    # the child between them holds the hands of both, so a row that breaks
    # that is dropped at once.
    girls = set()
    for count in range(1, MOST_GIRLS + 1):
        size = BOYS + count
        rows = [([], 0)]  # the children so far, and how many are boys
        while rows:
            row, boys = rows.pop()
            if len(row) == size:
                if boys == BOYS and fits(row):
                    girls.add(count)
                continue
            for child in ["boy", "girl"]:
                grown = row + [child]
                now = boys + (1 if child == "boy" else 0)
                if len(grown) >= 3 and grown[-1] != grown[-3]:
                    continue
                if now > BOYS or len(grown) - now > count:
                    continue
                rows.append((grown, now))
    if not girls:
        fail("no circle fits")
    if len(girls) != 1:
        return match(options, "It cannot be told")
    return match(options, list(girls)[0])
