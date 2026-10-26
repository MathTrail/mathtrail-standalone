BIRDS = [9, 6, 5]  # sparrows, robins and finches

def kept(stay):
    # At least 4 of one kind and 3 of another.
    for one in range(3):
        for other in range(3):
            if one != other and stay[one] >= 4 and stay[other] >= 3:
                return True
    return False

def solve(options):
    # Every way the birds could stay after some pictures; the photographer is
    # sure while every one of them is kept.
    total = sum(BIRDS)
    sure = None
    for pictures in range(total + 1):
        ways = [stay for stay in product(range(BIRDS[0] + 1), range(BIRDS[1] + 1), range(BIRDS[2] + 1))
                if sum(stay) == total - pictures]
        if not all([kept(stay) for stay in ways]):
            break
        sure = pictures
    if sure == None:
        fail("the photographer is never sure")
    return match(options, sure)
