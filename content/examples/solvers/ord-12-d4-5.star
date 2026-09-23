def solve(options):
    names = ["Ann", "Ben", "Kim"]
    found = set()
    for ben in range(50):
        stamps = {"Ann": ben + 2, "Ben": ben, "Kim": ben + 3}
        most = max([stamps[name] for name in names])
        leaders = [name for name in names if stamps[name] == most]
        if len(leaders) != 1:
            fail("two children have the most stamps when Ben has %d" % ben)
        found.add(leaders[0])
    if len(found) != 1:
        fail("who has the most depends on how many Ben has")
    return match(options, list(found)[0])
