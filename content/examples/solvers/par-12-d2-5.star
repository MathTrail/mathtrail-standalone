def solve(options):
    colours = ["red", "white"]
    fitting = [n for n in [8, 10, 11, 12, 14] if colours[(n - 1) % 2] == "red"]
    if len(fitting) != 1:
        fail("%d of the numbers fit" % len(fitting))
    return match(options, fitting[0])
