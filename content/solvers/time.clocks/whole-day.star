# For "how many times a day" a digital clock shows something: go through all
# 1440 minutes of the day and test what the clock shows. The four digits are
# worked out with // and %, so no text has to be taken apart.
# From reference task clk-34-d4-5.

def face(now):
    # The four digits of the clock, as in 09:05.
    hours, minutes = now // 60, now % 60
    return [hours // 10, hours % 10, minutes // 10, minutes % 10]

def shows_it(digits):
    return len(set(digits)) == 1  # four equal digits, like 11:11

def solve(options):
    return match(options, len([now for now in range(24 * 60) if shows_it(face(now))]))
