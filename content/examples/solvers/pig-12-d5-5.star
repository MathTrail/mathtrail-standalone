def solve(options):
    fitting = set()
    for red in range(5):
        basket = ["red"] * red + ["green"] * (4 - red)
        always = True
        for three in combinations(basket, 3):
            if "red" not in three or "green" not in three:
                always = False
                break
        if always:
            fitting.add(red)
    if len(fitting) != 1:
        fail("%d numbers of red apples fit the clue" % len(fitting))
    return match(options, list(fitting)[0])
