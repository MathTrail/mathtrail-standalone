def all_tails_after(moves):
    states = {(1, 1, 1): True}
    for move in range(moves):
        turned = {}
        for state in states:
            for coin in range(3):
                after = list(state)
                after[coin] = 1 - after[coin]
                turned[tuple(after)] = True
        states = turned
    return (0, 0, 0) in states

def solve(options):
    fitting = [n for n in [2, 4, 5, 6, 8] if all_tails_after(n)]
    if len(fitting) != 1:
        fail("%d of the numbers fit" % len(fitting))
    return match(options, fitting[0])
