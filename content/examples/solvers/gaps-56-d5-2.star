FIRST = (30, 6)  # minutes of a part, minutes of an advert between two parts
SECOND = (15, 3)

def ends(film, channel):
    # The film part by part, an advert between every two parts.
    part, advert = channel
    clock = 0
    shown = 0
    while shown < film:
        if shown > 0:
            clock += advert
        clock += part
        shown += part
    return clock

def sooner(minutes):
    return "%d minute%s sooner" % (minutes, "" if minutes == 1 else "s")

def solve(options):
    # Every length up to five hours that both channels cut into whole parts.
    gaps = set()
    for film in range(1, 301):
        if film % FIRST[0] == 0 and film % SECOND[0] == 0:
            gaps.add(ends(film, SECOND) - ends(film, FIRST))
    if not gaps:
        fail("no length up to five hours splits into whole parts on both channels")
    if len(gaps) != 1:
        return match(options, "It cannot be told without the length of the film")
    gap = list(gaps)[0]
    if gap > 0:
        return match(options, "The first, " + sooner(gap))
    if gap < 0:
        return match(options, "The second, " + sooner(-gap))
    return match(options, "Both end at the same moment")
