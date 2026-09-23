def ordinal(number):
    return {1: "1st", 2: "2nd", 3: "3rd"}.get(number, "%dth" % number)

def on_after(presses):
    on = False
    for press in range(presses):
        on = not on
    return on

def solve(options):
    fitting = [ordinal(n) for n in [2, 4, 6, 8, 9] if on_after(n)]
    if len(fitting) != 1:
        fail("%d of the presses fit" % len(fitting))
    return match(options, fitting[0])
