def solve(options):
    # Every split of the 30 pupils into the three pairs of clubs.
    drama = set()
    for chess_music in range(31):
        for chess_drama in range(31 - chess_music):
            music_drama = 30 - chess_music - chess_drama
            if chess_music + chess_drama == 20 and chess_music + music_drama == 18:
                drama.add(chess_drama + music_drama)
    if len(drama) != 1:
        fail("the class gives %d answers" % len(drama))
    return match(options, list(drama)[0])
