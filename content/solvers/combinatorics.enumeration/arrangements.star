# For "in how many ways": make every arrangement with the helper that makes
# them, keep the ones the conditions allow, and count them. Orders are
# permutations, groups are combinations, independent choices are a product.
# From reference task enum-12-d4-4.

CHILDREN = ["Ann", "Ben", "Kim"]  # who sits on the chairs in a row

def allowed(row):
    # Ann does not sit next to Ben.
    place = {name: i for i, name in enumerate(row)}
    return abs(place["Ann"] - place["Ben"]) != 1

def solve(options):
    return match(options, len([row for row in permutations(CHILDREN) if allowed(row)]))
