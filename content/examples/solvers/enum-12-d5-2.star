def solve(options):
    shirts = ["red", "white", "blue"]
    trousers = ["green", "black", "grey"]
    allowed = [1 for shirt, pair in product(shirts, trousers) if not (shirt == "red" and pair == "green")]
    return match(options, len(allowed))
