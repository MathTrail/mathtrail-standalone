# A bike costs 400 coins, its price goes up by 25% and then the new price goes
# down by 20%: make the changes in turn, each of the price at that moment, in
# hundredths. The note says the fall taken of the old price would give 420,
# so that is checked too.
# Written from the template changes of percent.basic.

START = 400  # the price before the changes
CHANGES = [25, -20]  # each change in per cent, in order: positive for a rise, negative for a fall

def after(price):
    # The price after every change, or None when a change is not a whole
    # number of coins.
    for change in CHANGES:
        if price * change % 100 != 0:
            return None
        price += price * change // 100
    return price

def solve(options):
    risen = START + START * CHANGES[0] // 100
    if risen + START * CHANGES[1] // 100 != 420:
        fail("the fall taken of the old price does not give 420")
    price = after(START)
    if price == None:
        fail("a change of %d coins is not a whole number of coins" % START)
    return match(options, price)
