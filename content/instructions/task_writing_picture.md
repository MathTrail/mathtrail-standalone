## A task solved from the picture

This brief's `constraints` hold `facts_on_the_picture`: the task is solved from its picture. Its facts stand on the picture alone, and without the picture it cannot be solved. Three rules of "The task" and "The picture" turn round for it:

- **The picture carries every fact:** the walls of a maze, the lines of a figure, the numbers on tiles, the steps of a staircase, the posts of a fence. The question says what the picture shows, what the rules of the task are and what to find, and gives none of the picture's facts again: no number the picture shows and no count the child takes off it. A number in the question is a rule, such as the 7 that opposite faces of a die add up to, or the name of something drawn, such as face 3 of a net.
- **The picture is drawn whole:** a row the child counts is not cut short, and every cell, line and label the child works from is there. It still shows nothing the question asks for: never a path through the maze, the result of a fold or the step that solves the task. A picture may draw the right value only beside the values of at least three wrong options, as the numbers a tile holds or the totals of rows do, and a picture of `tiles` or `shapes` marks the place asked about with `?`.
- **The task is drawn, not described:** the wording leaves out what the picture holds, so a task the child could solve from its words alone is not one solved from the picture.

Draw it with a kind from `pictures`, as its `purpose` says: a figure on squared paper (`lattice`), identical tiles (`tiles`), rows of shapes with their totals (`shapes`), a staircase (`stairs`), or a kind you already know when it holds what the task needs, such as a row of posts or piles of counters. Give the members its `purpose` names that the task needs, each within `limits`; the example shows a few of them. Name the picture's labels in the question by their Latin capitals, as "The picture" says; its numbers need no naming.

**The solver reads the picture.** Define `solve(options, picture)`: `picture` is the description you wrote, as Starlark values. An object is a dictionary and a list of numbers or texts a tuple, so a cell `[2, 3]` reads as `(2, 3)` and can stand in a set; a number is a number and a label is a text, `int(label)` for its number. Take every fact from `picture`, never from numbers copied into the program: read a member as `picture["cells"]`, or as `picture.get("cells", ())` where it may be left out. A member that carries facts and is never read refuses the task, `solver_unread`. For a lattice, how many filled cells have a filled cell just to their right:

```python
def solve(options, picture):
    filled = set(picture.get("cells", ()))
    pairs = 0
    for x, y in filled:
        if (x + 1, y) in filled:
            pairs += 1
    return match(options, pairs)
```

`solution_picture` draws the task's picture again, of the same kind, with what the solution works out marked on it: the path, the sheet unfolded, the faces counted. In the self-check, ask whether the question repeats a fact of the picture or the picture lacks one the solution uses, and record it as `facts_in_words`.
