# Masha thought of a number, doubled it, added 7, divided the result by 3 and
# got 9: try every whole number and keep the ones the steps of the story turn
# into 9. Where the story divides by 3, check that the third is whole with
# % 3 == 0 and take it with //.
# Written from the template number-thought-of of arithmetic.tricks.

RESULT = 9  # what Masha got at the end

def fits(n):
    added = n * 2 + 7
    return added % 3 == 0 and added // 3 == RESULT

def solve(options):
    fitting = [n for n in range(1001) if fits(n)]
    if len(fitting) != 1:
        fail("%d whole numbers fit the story" % len(fitting))
    return match(options, fitting[0])
