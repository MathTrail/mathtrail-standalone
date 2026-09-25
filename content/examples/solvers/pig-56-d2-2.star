COLOURS = 3
SIZES = 2
WANTED = 2  # socks alike in colour and in size

def solve(options):
    kinds = list(product(range(COLOURS), range(SIZES)))  # a colour together with a size
    # Every handful that is not yet enough, with fewer than WANTED socks of
    # each kind, built up one kind at a time; there are many socks of each
    # kind, so any such handful can come up. The answer is the first size not
    # among them.
    failed = set([0])
    for _ in kinds:
        failed = set([size + count for size in failed for count in range(WANTED)])
    socks = 0
    while socks in failed:
        socks += 1
    return match(options, socks)
