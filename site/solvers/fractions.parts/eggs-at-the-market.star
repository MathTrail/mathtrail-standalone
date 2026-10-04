# Three buyers each take half of Granny's eggs and half an egg more, of what is
# left at that moment, nobody breaks an egg, and all the eggs are sold: try
# every number of eggs, take the shares in turn, and keep the numbers in which
# every share is a whole number of eggs and nothing is left. The steps work
# backwards through 3 and 1 eggs, so those are checked too.
# Written from the template what-is-left of fractions.parts.

BUYERS = 3
LARGEST = 1000  # the most eggs worth trying

def before_each(whole):
    # The eggs Granny has before each buyer and after the last, or None when
    # a buyer would have to break an egg.
    left = [whole]
    for _ in range(BUYERS):
        # Half of the eggs and half an egg more is (eggs + 1) / 2, a whole
        # number of eggs only when the eggs are odd.
        if (left[-1] + 1) % 2 != 0:
            return None
        left.append(left[-1] - (left[-1] + 1) // 2)
    return left

def sold_out(whole):
    left = before_each(whole)
    return left != None and left[-1] == 0

def solve(options):
    wholes = [whole for whole in range(1, LARGEST + 1) if sold_out(whole)]
    if len(wholes) != 1:
        fail("%d numbers of eggs are all sold, and the question has one" % len(wholes))
    if before_each(wholes[0]) != [wholes[0], 3, 1, 0]:
        fail("the eggs before each buyer are not the ones the steps work back through")
    return match(options, wholes[0])
