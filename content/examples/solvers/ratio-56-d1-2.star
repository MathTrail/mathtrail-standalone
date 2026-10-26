def solve(options):
    # The road in km from Nina's ride, then every whole speed that takes Omar
    # along it in 20 minutes.
    nina_speed, nina_minutes, omar_minutes = 12, 30, 20
    if nina_speed * nina_minutes % 60 != 0:
        fail("the road is not a whole number of km")
    road = nina_speed * nina_minutes // 60
    speeds = [s for s in range(1, 201) if s * omar_minutes == road * 60]
    if len(speeds) != 1:
        fail("%d speeds fit" % len(speeds))
    return match(options, "%d km an hour" % speeds[0])
