def solve(options):
    children = 20
    both = []
    for dog_only, cat_only, dog_and_cat in product(range(children + 1), repeat=3):
        no_pet = children - dog_only - cat_only - dog_and_cat
        if no_pet >= 0 and dog_only + dog_and_cat == 12 and cat_only + dog_and_cat == 10:
            both.append(dog_and_cat)
    return match(options, min(both))
