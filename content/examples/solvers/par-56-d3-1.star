NUMBERS = list(range(1, 11))  # on the board at the start
STEPS = [23, 45, 9, 10]  # the numbers of steps the options name
NEVER = "It can never be done"  # as the option writes it

def sum_can_be_shared(steps):
    # Each step adds 1 to two of the numbers, so after `steps` steps they add
    # up to their first sum plus twice that; ten equal numbers add up to ten
    # times one of them.
    return (sum(NUMBERS) + 2 * steps) % len(NUMBERS) == 0

def solve(options):
    # Ten numbers that keep growing are too many positions to search one by
    # one, so the sum stands in for them, and it has to settle every option on
    # its own. Its remainder on dividing by ten comes round again within ten
    # steps, so ten steps show every remainder it can ever have.
    if [steps for steps in STEPS if sum_can_be_shared(steps)]:
        fail("the sum does not rule out every number of steps the options name")
    if any([sum_can_be_shared(steps) for steps in range(len(NUMBERS))]):
        fail("after some number of steps the sum could be shared equally")
    return match(options, NEVER)
