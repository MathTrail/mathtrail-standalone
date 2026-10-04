# Five children stand in a queue, and four clues tell where: try every queue,
# keep the queues every clue allows, and read off who stands last. The
# example's note says that without the clue that Lena is not last the queue
# is not settled, so the clues are tried without it as well.
# Written from the template orders of logic.ordering.

CHILDREN = ["Lena", "Max", "Nina", "Oleg", "Pavel"]  # as the answer names them

def fits(place, lena_not_last):
    # Place 1 is the front of the queue. Nina stands third; Oleg stands right
    # behind Nina; Max stands somewhere in front of Lena; Lena is not last.
    return (place["Nina"] == 3 and place["Oleg"] == place["Nina"] + 1 and
            place["Max"] < place["Lena"] and (place["Lena"] != 5 or not lena_not_last))

def last(lena_not_last):
    found = set()
    for order in permutations(CHILDREN):
        place = {name: i + 1 for i, name in enumerate(order)}
        if fits(place, lena_not_last):
            found.add(order[4])  # who stands last
    return found

def solve(options):
    if len(last(False)) < 2:
        fail("the clues settle the queue without Lena's, and the note says they do not")
    answers = last(True)
    if len(answers) != 1:
        fail("the clues fit %d different answers" % len(answers))
    return match(options, list(answers)[0])
