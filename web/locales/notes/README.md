# Notes for whoever translates the card's words

`keys.json` has a note for every key of [`../en.json`](../en.json), the words every other dictionary is translated from, and a test holds its keys to English's. Nothing is built from this folder: a dictionary is a file of `../`, never one of this folder.

## What a note says

- **`reader`** — whom the text is written for: `child`, who answers the tasks on the card; `parent`, the adult who runs the lesson and types in the chat, a parent or a tutor; or `both`, where either reads it.
- **`where`** — where the text stands on the card and what it does there.
- **`length`** — how much room it has:
  - `label` — a word or a few, on a button, a tab, a chip, a badge or a heading, beside other things: no longer than the English needs, a single word where the English has one;
  - `line` — a single line, which should not wrap on a card 320 px wide;
  - `sentence` — a sentence or a few, which wrap as they need;
  - `heard` — said by a screen reader and never seen: as long as it needs to be clear.
- **`address`** — how the text speaks to its reader: `informal`, the "you" said to a child (ты, du, tú); `polite`, the "you" said to an adult one does not know (вы, Sie, usted); or `none`, where it speaks to nobody — a name, a heading, a statement.
- **`mind`** — anything else the text needs, where there is something.

## What every text needs

- **From the English, with the Russian beside it.** The Russian was written beside the English rather than translated from it, and shows the form of address and the tone each text takes.
- **Nothing shows whether the child is a boy or a girl.** No past tense, adjective or participle that agrees with the child, as Slavic languages and Lithuanian make them, and no second person that tells the two apart, as Hebrew, Arabic and Amharic do: the present tense, an impersonal form, a verbal noun or "let's" instead. The parent's gender stays out too, where a language's second person would show it.
- **The slots stay as they are.** `{count}`, `{grade}`, `{topic}` and the rest keep their names and their braces, wherever the sentence needs them; a number put into one is written in the language's own digits by the card.
- **The plural forms are the language's own.** A text that changes with a number, `{"one": "…", "other": "…"}` in English, is written for exactly the plural categories CLDR gives the language — `new Intl.PluralRules([tag, "en"]).resolvedOptions().pluralCategories` — and a test refuses any other set. A language CLDR has no plural rules for is counted as English is, `one` and `other`, which is what the platform answers with English after the language.
- **No character a reader cannot see**, such as a mark that turns the way text runs, a space of no width or a soft hyphen: the card sets the direction and the wrapping itself. The joiners a script writes inside its words are part of those words.
- **Ranks are numbers**, "Rank 3", never titles.
- **MathTrail is a name** and stays as it is written.
