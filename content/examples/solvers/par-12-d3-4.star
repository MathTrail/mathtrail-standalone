def solve(options):
    payable = [2 * coins for coins in range(50)]
    fitting = ["%d cents" % amount for amount in [7, 11, 15, 16, 19] if amount in payable]
    if len(fitting) != 1:
        fail("%d of the amounts fit" % len(fitting))
    return match(options, fitting[0])
