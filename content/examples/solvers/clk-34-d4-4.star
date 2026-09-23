def at(clock_time):
    parts = clock_time.split(":")
    return int(parts[0]) * 60 + int(parts[1])

def solve(options):
    for watch in range(at("8:00"), at("9:30")):
        believed = watch - 10  # he takes the watch to be ten minutes fast
        if believed == at("8:40"):
            real = watch + 10  # it is ten minutes slow instead
            return match(options, real + 20 - at("9:00"))
    fail("he never sets off")
