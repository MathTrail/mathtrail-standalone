def solve(options):
    names = ["Ben", "Ann", "Kim", "Dan"]
    found = set()
    for ben in range(5, 50):
        ages = {"Ben": ben, "Ann": ben + 2, "Kim": ben + 3}
        ages["Dan"] = ages["Kim"] - 1
        least = min([ages[name] for name in names])
        youngest = [name for name in names if ages[name] == least]
        if len(youngest) != 1:
            fail("two children are the youngest when Ben is %d" % ben)
        found.add(youngest[0])
    if len(found) != 1:
        fail("who is the youngest depends on how old Ben is")
    return match(options, list(found)[0])
