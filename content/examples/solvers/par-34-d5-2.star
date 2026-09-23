def possible(pupils):
    # Each friendship is counted by both friends, so the friend counts add up
    # to an even number, which an odd number of odd counts never does.
    if pupils % 2 == 1:
        return False
    # For an even number, pairing the pupils off gives each of them one friend.
    friends = {pupil: [pupil ^ 1] for pupil in range(pupils)}
    return all([len(friends[pupil]) % 2 == 1 for pupil in range(pupils)])

def solve(options):
    fitting = [n for n in [21, 23, 25, 26, 27] if possible(n)]
    if len(fitting) != 1:
        fail("%d of the numbers fit" % len(fitting))
    return match(options, fitting[0])
