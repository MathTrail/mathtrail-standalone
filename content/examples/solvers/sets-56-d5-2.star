def solve(options):
    # Every number Ann writes, kept when Ben writes it as well.
    ann = range(5, 101, 6)
    ben = set(range(2, 101, 5))
    both = [n for n in ann if n in ben]
    return match(options, len(both))
