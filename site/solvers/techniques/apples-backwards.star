# A story told forwards whose end is known: try every whole number of apples
# at the start and keep the ones the story turns into the end. Where the story
# halves, the half must be whole: check it with % 2 == 0 and take it with //.
# Written from the template number-thought-of of arithmetic.tricks.

PETYA_MORE = 1  # the apple Petya eats after half of what he finds
LEFT = 2  # apples left at the end

def fits(apples):
    if apples % 2 != 0:
        return False
    after_tanya = apples - apples // 2  # Tanya eats half
    if after_tanya % 2 != 0:
        return False
    return after_tanya - after_tanya // 2 - PETYA_MORE == LEFT  # half of the rest, and one more

def solve(options):
    fitting = [n for n in range(1001) if fits(n)]
    if len(fitting) != 1:
        fail("%d whole numbers fit the story" % len(fitting))
    return match(options, fitting[0])
