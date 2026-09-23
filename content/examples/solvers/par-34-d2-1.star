def solve(options):
    payable = [7 * price for price in range(1, 100, 2)]
    fitting = ["%d cents" % amount for amount in [42, 56, 63, 70, 84] if amount in payable]
    if len(fitting) != 1:
        fail("%d of the amounts fit" % len(fitting))
    return match(options, fitting[0])
