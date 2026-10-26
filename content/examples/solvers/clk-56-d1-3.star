FIRST_TALE = "9:00"
TALE = 7  # minutes
SONG = 4
ASKED = "10:00"

def at(clock_time):
    parts = clock_time.split(":")
    return int(parts[0]) * 60 + int(parts[1])

def solve(options):
    # Tales and songs one after another, a song after every tale, up to the
    # first song that starts at 10:00 or later.
    now = at(FIRST_TALE)
    while now <= at(ASKED) + 24 * 60:
        song = now + TALE
        if song >= at(ASKED):
            return match(options, song - at(ASKED))
        now = song + SONG
    fail("the cat starts no song within a day of 10:00")
