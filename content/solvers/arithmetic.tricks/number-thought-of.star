# For "I thought of a number": try every whole number and keep the ones the
# steps of the story turn into the result. Keep to whole numbers: where the
# story halves, check that the half is whole with % 2 == 0 and take it with //.
# From reference task tri-12-d5-4.

TO_BEN = 1  # sweets given to Ben after half went to Ann
LEFT = 4  # sweets left at the end

def fits(sweets):
    return sweets % 2 == 0 and sweets - sweets // 2 - TO_BEN == LEFT

def solve(options):
    fitting = [n for n in range(1001) if fits(n)]
    if len(fitting) != 1:
        fail("%d whole numbers fit the story" % len(fitting))
    return match(options, fitting[0])
