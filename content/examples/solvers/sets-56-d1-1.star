SORTS = [("cat", "Vaska"), ("cat", ""), ("dog", "Vaska"), ("dog", "")]

def solve(options):
    # Every yard of up to 3 animals of each sort: cats called Vaska, other
    # cats, other animals called Vaska and animals that are neither. If the
    # comparison came out differently in different yards, it would depend on
    # the yard.
    answers = set()
    for counts in product(range(4), repeat=4):
        animals = []
        for sort in range(4):
            for number in range(counts[sort]):
                animals.append((SORTS[sort][0], SORTS[sort][1], number))
        cats = [a for a in animals if a[0] == "cat"]
        vaskas = [a for a in animals if a[1] == "Vaska"]
        # The cats, leaving out those not called Vaska, and the animals called
        # Vaska, leaving out those that are not cats.
        cats_left = [a for a in cats if a[1] == "Vaska"]
        vaskas_left = [a for a in vaskas if a[0] == "cat"]
        if len(cats_left) > len(vaskas_left):
            answers.add("The cats")
        elif len(cats_left) < len(vaskas_left):
            answers.add("The animals called Vaska")
        else:
            answers.add("They are as many")
    if len(answers) != 1:
        return match(options, "It depends on how many cats there are")
    return match(options, list(answers)[0])
