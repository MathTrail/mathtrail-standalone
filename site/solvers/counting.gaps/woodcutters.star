# Woodcutters made 9 cuts in some logs and got 14 pieces, every log cut at
# least once. Line the cuts up log after log: the logs split that line at some
# of the 8 places between neighbouring cuts, so trying every split tries every
# way the cuts could be shared. On each log, count the pieces as the gaps
# between neighbouring points, its two ends and its cuts. Each number of logs
# must give one count of pieces however the cuts are shared, as the note says,
# and exactly one number of logs must give 14.
# Written from the template neighbours of counting.gaps.

CUTS = 9
PIECES = 14

def pieces(cuts):
    # A log has its ends at 0 and at cuts + 1, and its cuts at 1 to cuts.
    return len([(point, point + 1) for point in range(0, cuts + 1)])

def solve(options):
    found = {}  # the pieces each number of logs gives
    for splits in product([False, True], repeat=CUTS - 1):
        shares = [1]  # the cuts of each log, the first cut on the first log
        for split in splits:
            if split:
                shares.append(1)
            else:
                shares[-1] += 1
        total = sum([pieces(cuts) for cuts in shares])
        if found.get(len(shares), total) != total:
            fail("%d logs give different numbers of pieces" % len(shares))
        found[len(shares)] = total
    fitting = [logs for logs in found if found[logs] == PIECES]
    if len(fitting) != 1:
        fail("%d numbers of logs give %d pieces" % (len(fitting), PIECES))
    return match(options, fitting[0])
