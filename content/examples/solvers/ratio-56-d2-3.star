def solve(options):
    spotted = 20
    # 3 striped for every 5 spotted, and every fish is one or the other.
    striped = [s for s in range(0, 1001) if s * 5 == spotted * 3]
    if len(striped) != 1:
        fail("%d numbers of striped fish keep the ratio" % len(striped))
    return match(options, spotted + striped[0])
