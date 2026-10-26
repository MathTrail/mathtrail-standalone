def liars_can_score(answers):
    # Whether the liars' real goals can add up to 9 with each liar's goals
    # other than the number he said: the totals reachable player by player.
    reachable = set([0])
    for said in answers:
        reachable = set([total + goals for total in reachable for goals in range(0, 10) if goals != said and total + goals <= 9])
    return 9 in reachable

def solve(options):
    # Every set of answers: the truth-tellers' answers are their real goals
    # and add up to 12, the liars' answers are all false and their goals add
    # up to 9, Misha alone said four, and one, three and seven were all said.
    teams = set()
    for misha_lies in [True, False]:
        for truth in product([1, 3, 7], repeat=6 if misha_lies else 5):
            truth_answers = list(truth) + ([] if misha_lies else [4])
            if sum(truth_answers) != 12:
                continue
            for liars in product([1, 3, 7], repeat=5 if misha_lies else 6):
                liar_answers = list(liars) + ([4] if misha_lies else [])
                said = truth_answers + liar_answers
                if all([n in said for n in [1, 3, 7]]) and liars_can_score(liar_answers):
                    teams.add("liars" if misha_lies else "truth-tellers")
                    break
    if teams == set(["liars"]):
        return match(options, "Yes, he is a liar")
    if teams == set(["truth-tellers"]):
        return match(options, "No, he tells the truth")
    if teams:
        return match(options, "It cannot be told without knowing his team")
    fail("no match fits")
