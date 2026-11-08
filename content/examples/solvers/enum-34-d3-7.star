def solve(options, picture):
    flavours = 0
    toppings = 0
    for pile in picture["piles"]:
        if pile.get("shape", "dot") == "dot":
            flavours += pile["count"]
        else:
            toppings += pile["count"]
    return match(options, flavours * (toppings + 1))
