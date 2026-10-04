# The number 38*6 with a digit in place of the star is to divide by 9 with
# nothing left over: try every digit and test the number it makes. The note
# says the digits for 3 are 1, 4 and 7, so that is checked too.
# Written from the template search of number.divisibility.

def number(digit):
    return 3806 + 10 * digit  # 38*6 with the digit in the tens

def found(by):
    return [digit for digit in range(0, 10) if number(digit) % by == 0]

def solve(options):
    if found(3) != [1, 4, 7]:
        fail("the digits for 3 are %r, and the note says 1, 4 and 7" % found(3))
    digits = found(9)
    if len(digits) != 1:
        fail("%d digits make a number that divides by 9" % len(digits))
    return match(options, digits[0])
