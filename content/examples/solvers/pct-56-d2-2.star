def solve(options):
    parties = []
    for guests in range(1, 1001):
        # 60% and 30% of the guests have to be whole numbers of guests.
        if guests * 60 % 100 != 0 or guests * 30 % 100 != 0:
            continue
        vanilla = guests * 60 // 100
        chocolate = guests * 30 // 100
        if guests - vanilla - chocolate == 3:
            parties.append(guests)
    if len(parties) != 1:
        fail("%d parties fit" % len(parties))
    return match(options, parties[0])
