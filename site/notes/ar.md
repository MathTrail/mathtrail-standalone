# Arabic (`ar`)

The Arabic half of [the brief](README.md): the rules the Arabic texts follow on top of it, and the words they use. The card's Arabic dictionary, [`web/locales/ar.json`](../../web/locales/ar.json), comes first wherever it has a word; this file adds only what the site needs beyond it. Arabic is the site's first language that runs right to left, so this file also says how a sentence is worded for that.

## Register and variant

- **One Arabic, the card's: Modern Standard Arabic.** No word of one country's speech. Where school words differ between the Mashriq and the Maghreb, the site takes the card's word, and failing that the one of the Mashriq's school books (Egypt, the Gulf, the Levant, Iraq), which most readers learned with: «مسألة», «الصف», the letters «أ، ب، ج» for labels, the unknown «س». The months are those Intl writes for `ar`: «يناير … ديسمبر», neither «كانون الثاني» nor «جانفي».
- **The reader is addressed in the plural, «أنتم», never in the singular:** «اطلبوا مسألة في الدردشة», «طفلكم», «يمكنكم». Where a sentence needs no "you", the card's impersonal turns come first: «يمكن طلب مسألة في الدردشة.» (`waiting.ask_in_chat`), «يكفي طلب ملف شخصي في MathTrail في هذه الدردشة…» (`first_run.start`), «ستوضح الدردشة ما يجب تغييره.» (`profile.not_saved_detail`). A few keys of the card write the singular «ـك», which shows no gender only because short vowels go unwritten: «بياناتك» (`data.label`), «Google Drive الخاص بك» (`data.export_file`), «مرحبا بك في MathTrail» (`first_run.label`). The site copies them only where it quotes the card.
- **The child is never "you" either.** Arabic's second person tells a boy from a girl in every imperative, verb and adjective: «اختر/اختاري», «تعرف/تعرفين», «مستعد/مستعدة». The card speaks to the child with no "you" at all: a question, «ما الإجابة الصحيحة؟» (`task.pick`); "let's", «هيا نتمنى أمنية للغد» (`waiting.limit`), «لنجرب مرة أخرى» (`chat.not_sent`); a verbal noun, «تجربة كل مسألة أولًا من دون تلميح.» (`review.step_unaided`). That covers the quotes under «ما يمكن قوله», the questions of a tip, Pólya's questions and the parent's lines on "Why".
- **The patterns, once for all:**
  - "Ask for a task": a heading or a step's title is a verbal noun, «طلب مسألة»; a sentence speaks to the reader, «اطلبوا مسألة في الدردشة», or as the card does, «يمكن طلب مسألة في الدردشة».
  - "Tap the bulb": to the reader, «انقروا على المصباح»; of the child, a verbal noun, «النقر على المصباح يفتح تلميحًا».
  - "Write in the chat", "Press **Add**": «اكتبوا في الدردشة», «اضغطوا **Add**».
  - "Say to the child: …": «قولوا لطفلكم: «…»», and the quote itself speaks with "we" or with no "you": «لنرسم ثلاث محطات. كم فاصلًا بينها؟».
  - A question to the child: «أي معلومة لم نستعملها بعد؟», «هل يوجد أحد أبعد على الخط؟», «وماذا لو كان الحظ في أسوأ حالاته؟».
  - The child's own words: «أخطأتُ مرة أخرى…», «لماذا أحتاج إلى هذه المسائل؟ لم يطلبها مني أحد.».
  - "Your child": «طفلكم», never the subject of a verb: "your child taps" is «والنقر على البطاقة لطفلكم».
  - A tip for home: the title a verbal noun, «ترتيب العائلة في صف»; the text in the plural, «رتّبوا أفراد العائلة حسب الطول…», and what the child does there a verbal noun after «على»: «وعلى الطفل قلبها: …», never «أن يقلبها».
- **What a parent types to the chat** speaks to the chat in the singular, as people write to a chat: «أعطني مسألة في موضوع «الفرسان والكذابون»», «مسألة جديدة، من فضلك». The gender of «أعطني» is the chat's, a program's, and says nothing of the parent.
- **No gender shows, the child's or the parent's.** Arabic shows it far more than Spanish: in every verb of the second and the third person («يحلّ/تحلّ», «أخطأ/أخطأت»), in the pronouns («ـه/ـها»), in adjectives and participles («متأكد/متأكدة», «مخطئ/مخطئة»), in nouns of people («طفل/طفلة», «ابن/ابنة»). Only the first person shows none: «أخطأتُ», «انتهيت», «نحلّ». So:
  - "your child" is «طفلكم», "the child" «الطفل»: the card's generic noun for any child, as in «حسب تقدّم الطفل» (`topic_choice.coach_note`) and «على مستوى الطفل» (`review.step_practice`). But the child, the reader's or any one child, is never the subject of a verb, never stands behind a pronoun and takes no adjective; the card never writes «يحلّ الطفل» or «تقدّمه». Instead:
    - a verbal noun, the child after it: "once your child answers" «بعد إجابة طفلكم»; "until your child commits to one" «حتى يستقر اختيار طفلكم على إحداها»; "while your child solves" «في أثناء حلّ طفلكم للمسائل»; "how the child reasoned" «طريقة تفكير الطفل»;
    - the passive: "the answer is recorded with a tap" «تُسجَّل الإجابة بنقرة»; "the child counts the stops" «تُعدّ المحطات بدل الفواصل»;
    - a phrase with a preposition: "your child taps" «والنقر على البطاقة لطفلكم»; "your child needs neither" «وطفلكم في غنى عن كليهما»; "the child asks you questions" «والأسئلة على الطفل»;
    - a noun: "what your child already knows" «معارف طفلكم الحالية»; "how they are progressing" «مسار تقدّم الطفل»; "where your child went wrong" «موضع خطأ طفلكم».
  - The child's own words are in the first person, whose verbs show no gender: «أخطأتُ مرة أخرى…», «انتهيت! الإجابة خمسة.». No adjective about themself: «أنا على يقين», not «أنا متأكد».
  - A child's question to the parent asks about the words, not the person: «هل هذا الكلام كذب؟», not «هل أنت كذاب؟», which shows the parent's gender twice.
  - Children in general are «الأطفال», as the card's «للأطفال» (`first_run.lead`), and in the plural they may be the subject of a verb: «أين يخطئ الأطفال». Counted on "Research", «طفل» and «أطفال». A group inside a task is the task's own, in the generic masculine plural school books write: «24 طفلًا», «150 تلميذًا». The rule is about the reader's child.
  - "The parent" is the card's «ولي الأمر» (`profile.label`, `progress.grades_note`), "parents" «أولياء الأمور», "a parent or a tutor" the card's «أحد الوالدين أو معلم» (`first_run.lead`), "the adult" the card's «شخص بالغ», «الشخص البالغ». These generic role nouns may take a verb in the third person on the technical pages, «يسجّل الشخص البالغ الدخول مرة واحدة», though a verbal noun reads better where it fits.
  - "Together" is «معًا», which shows no gender: «وتحققوا منها معًا».
  - The people of the examples have a gender, as in the English: «تسقي كوثر الأزهار كل 3 أيام». The islanders are letters, in the generic masculine: «أ كذاب», «ب فارس». A knight with a name, in "Techniques", may be a girl: «أسماء فارسة».
- **An example's question is a task for the child**, as a task on the card is, so it asks with no "you": «كم عمودًا هناك؟», «أي القوارب فاز؟». An instruction inside it is passive or a verbal noun: «تُجمع الأعداد في كل صف…», «والمطلوب إيجاده دون ضرب الأعداد الكبيرة». The steps of a solution speak in the first person plural of Arabic school solutions, «نرسم…، نلاحظ أن…، إذن…», which is the reader and the child together and shows no gender. The notes speak to the reader.
- **Names that take a verb:** MathTrail, Claude, ChatGPT and Drive are masculine, as the card writes «يقدم MathTrail» (`first_run.lead`) and «يختارها MathTrail» (`review.step_practice`). Companies — Google, OpenAI, Anthropic, ALT EdTech — are feminine, as their own Arabic help writes «تجري OpenAI اختبارات» and «تُوقِف Google…».
- **What a topic teaches**, `hero.teaches`, goes on from its label «ماذا يعلّم» with a verbal noun, and a second verbal noun joins with «و»: «ربط المعلومات في صورة واحدة، وقراءة المقارنة في الاتجاهين». The main move, `hero.method`, is a verbal noun too: «رسم خط ووضع الجميع عليه».
- **Sentences every topic's page shares** read the same on every page: `method.lead` ends «تتدرّج الأمثلة أدناه من الأبسط إلى مستوى الأولمبياد.», and `description` is «ما هي مسائل …، وكيف تُحلّ، وأين يخطئ الأطفال عادةً، وكيف يمكن المساعدة في البيت.». «أدناه» is the card's word (`task.another_coming`).
- **The card's words, exactly.** A topic's name written out in a sentence is the card's. It goes without quotes where the sentence keeps its letters: «مسائل الترتيب», «موضوع حصر الحالات», «مبدأ برج الحمام». It goes in «» where Arabic grammar would change them: «الفرسان والكذابون», after a preposition or as the second term of an idafa, stays as the card writes it, inside guillemets, as the card's own «في موضوع «{topic}»» (`topic_choice.ask`): «مسائل «الفرسان والكذابون»». Never a paraphrase, though the English sometimes paraphrases: "putting things in order" is «الترتيب».

## Typography

- **Quotes:** «…», as the card; a quote inside a quote “…”. The English “…” become «…». No space inside the guillemets.
- **Punctuation, as the card writes it:** the Arabic comma «،», semicolon «؛» and question mark «؟»; the colon, the full stop and the exclamation mark are the ASCII «:», «.» and «!». No space before any mark and one after it, as in all 29 colons and 69 commas of the card's texts. The ellipsis is one character, «…». The Latin comma «,» never stands in Arabic text; it appears only inside a number, where Intl groups thousands.
- **«و» and «ف» join the word after them**, a Latin one too, with no space: «وClaude», «فالمسألة». Between two letters that name something, «و» stands apart instead (see "Letters as labels").
- **A sentence that ends with a quote** takes its full stop after the closing guillemet, even when the quote ends in «؟» or «!»: «قولوا: «هيا نحل مسألة من MathTrail».». A `say:` that is the quote alone takes a full stop after it when it ends in a statement, «ثلثا ماذا؟ لنُشِر إليه على الشريط».، and nothing after «؟» or «!»: «كم فاصلًا بينها؟».
- **Diacritics** only where the card writes them: the tanween of the accusative, before the alif («أولًا», «دائمًا», «12 مترًا»); a shadda where it prevents a misreading («تقدّم», «المطوّر»); a damma on a passive verb that would otherwise read as active («تُسجَّل», «يُحسب», «لم يُحفظ»), as the card writes «يُستخدم», «يُحتاج», «تُقسَّم». The site leans on the passive more than the card does, so it marks the passive every time. No other vowels, and no tatweel «ـ».
- **Dashes:** the English spaced dash becomes «،», «:» or two sentences. Where a dash stays, it is the card's spaced em dash « — », as in «تعذر الإرسال — لنجرب مرة أخرى» (`chat.not_sent`). Ranges take the en dash with no spaces, as the card's «الصف {from}–{to}» (`progress.grades`): «1–6», «70–85%».
- **Ordinals:** words outside the examples, agreeing with the noun: «المرة السابعة», «القفزة الخامسة». In an example — its question, steps, note and answer — an ordinal the English writes in digits stays in digits, after «رقم»: «الحلوى رقم 9», «العلم رقم 23», «المربع رقم 5». In a drawing the number stands alone: «21، 22، 23». The badge "2nd" is «الثاني»: its answer, "The second player.", has no number, and the build lets a badge without a number through.
- **Dates** use digits and the month's name, as Intl writes them for `ar`: «3 مايو», «10 أكتوبر 2026». "The 3rd" is «اليوم 3»; "the 2nd, 9th, 16th, 23rd and 30th" «أيام 2 و9 و16 و23 و30». The months: «يناير، فبراير، مارس، أبريل، مايو، يونيو، يوليو، أغسطس، سبتمبر، أكتوبر، نوفمبر، ديسمبر». The weekdays: «الاثنين، الثلاثاء، الأربعاء، الخميس، الجمعة، السبت، الأحد». Arabic has no short weekday (Intl's "short" is the full name), so a badge or a drawing writes the whole name: "Wed" is «الأربعاء», "Mon" «الاثنين».
- **Centuries** in words, never in Roman numerals: «القرن الخامس», «أواخر القرن التاسع عشر», «من القرن التاسع إلى التاسع عشر». Decades: «خمسينيات القرن العشرين». Years in digits, never grouped: «1894».
- **Times** as the English writes them: «4:25», «15:10», «0:00», the bait's «03:59», with no «ص» or «م». A length of time is «4 س 30 د» in a drawing (Intl's short units for `ar`) and «4 ساعات و30 دقيقة» in prose.
- **Numbers outside the maths** as `new Intl.NumberFormat("ar")` writes them in Node 24 (ICU 78, CLDR 48): Latin digits, «,» between thousands from four digits on, «.» before decimals: «12,450», the bait's coins, «0.2». So the formula reads «P = 0.2 + 0.8·σ(θ + δ − β)», as the English. Years are never grouped. The site's locale is plain `ar`: `ar-SA` or `ar-EG` would write Arabic-Indic digits, which the site never shows.
- **Numbers in the maths** — the examples, the tables of a topic's page, and the sums of its cards and tips — keep the English digits and no thousands separator, as the Russian and the Spanish do: «2026 × 2026», «1111», the arithmetic page's «10 و100 و1000» and its table's «1900», the percentages page's tip «1200». The decimal point stays a point, as both the English and Intl write it: «0.5 + ½ = 1».
- **A noun counted by a number in digits** takes the form Arabic gives that number, as the card's plurals do, and above 100 the last two digits decide, as in CLDR's categories: after 1 the singular («1 مسألة»); after 2 the dual, «2 مرتان» (`progress.times`), and «2 مرتين» after a preposition or as an object; after 3 to 10, and 103 to 110, the plural («3 أعمدة», «6 سنوات»); after 11 to 99, and 111 to 199, the singular in the accusative («12 مترًا», «21 عمودًا», «14 قطعة»); after 100, 101, 102, 200 the singular («100 متر»). A number the English writes in words is written in words, with Arabic's reversed gender for 3 to 10: «ثلاث محطات», «ثلاثة أعمدة».
- **The counted nouns of "Research"**, `research.*` in the site's dictionary, follow the number the page writes before them, and each one's dual is written for the place its slot stands in. `research.tasks`, `research.checks`, `research.topics`, `research.traps` and `research.pages` stand alone or as subjects, so their dual ends in «ان»: «فحصان», «صفحتان», «مسألتان مرجعيتان». `research.passed`, `research.places`, `research.trial`, `research.answers`, `research.given` and `research.children` stand as objects or after a preposition, so their dual ends in «ين»: «فحصين», «إجابتين», «طفلين»; the page's sentences put their slots there: «بعد أن تجتاز {count} {noun}» (`hero.lead`), «بمقدار {count} {noun}» (`theses.twice.text`), «سلسلة تجريبية من {trial}» (`theses.reach.text`, `live.ready.weighed`), «خطأ التقدير بعد {after}» (`model.rows.error_static.name`), «إذا ضمّت {children} أو أكثر، و{answers} أو أكثر» (`live.rule`), «{month}: شمل العدّ {children} و{answers}…» (`live.ready.month`).
- **Percent** is written as the English and Intl write it, the digits and then «%», with no space: «70–85%», «20%». Never «٪», and never a sign typed before the number. After an Arabic word the page draws the sign to the left of the digits, where Arabic readers expect it; at the start of a line or after a Latin word it lands on the right. So a percentage comes after an Arabic word wherever the English allows: «بنسبة 70–85%», «خصم 20%». The same holds on the page of percentages, words and drawings alike.
- **Money** appears only in the make-believe app on "Why" (`app.bait.hat`, `app.cards.free.price`): US dollars, «9.99 دولار» for the hat and «0 دولار» on the card of the free app. The region pays in many currencies, and the App Store charges in US dollars where it has no local currency: in Bahrain, Jordan, Kuwait, Lebanon and Oman among others, as Patreon's list of the App Store's currencies for in-app purchases gives them (https://support.patreon.com/hc/en-us/articles/24965190839437, seen through a search: the page refuses fetchers). One currency every reader knows suits a page read from Morocco to Iraq better than one market's riyal, pound or dirham. Intl's symbol form, `Intl.NumberFormat("ar", {style: "currency", currency: "USD"})`, writes the digits 9.99 and «US$» with a mark that turns the direction (U+200F) before them and a no-break space between, and a text never types such a mark; without it, the page draws «$US 9.99». Its name form, `currencyDisplay: "name"`, writes «9.99 دولار أمريكي», with Latin digits and no mark at all. The bait app writes it as a price tag does, «9.99 دولار», with a plain space.
- **Units** in a drawing are Intl's short units for `ar`: «م» for metres, «كغم», «لتر», «د» for minutes, «س» for hours, and «أيام» for the calendar's "d": «12 م», «200 كغم», «5 لتر», «40 د», «4 س», «7 أيام». In prose the word, counted: «12 مترًا», «5 لترات», «40 دقيقة». On "Service", «KB» and «px» stay as they are.
- **Signs:** × for multiplication; − (U+2212) for minus, never a hyphen; · in the formula. Division as each English page writes it: ÷ on the home page, "Why", "Techniques" and "Research"; : on the topic pages that use it (gaps, calendar, percentages, ratios, parts, arithmetic). Fractions as the English: «3/4», «½». An unknown is an Arabic letter, as school books write it: "x" is «س», «يأخذ ليث س، فتأخذ رنا 5 − س». An expression is typed in the English order, «12 ÷ 3 = 4»: the page shows it right to left, as the Mashriq's school books write sums, and nobody reverses it by hand. «>» and «<» are typed as the English, and the browser mirrors them.
- **Arrows:** in Arabic text an arrow that leads from one thing to the next points left, «←», as the card's «{before} ← {after}» (`result.rating_change`) and OpenAI's Arabic help, «الإعدادات ← التطبيقات». "More →" is «المزيد ←»; "fails → never reaches the child" «تفشل ← لا تصل إلى الطفل»; "Check: 12 → 6 → 6 − 3 − 1 = 2" «التحقق: 12 ← 6 ← 6 − 3 − 1 = 2». Between two of Claude's English labels it stays «→», since the pair reads left to right: «Customize → Connectors». The chat's "+ → Connectors" is «**+** ← **Connectors**»: "+" is no letter, so that path reads from the right. In a drawing, an arrow that points at something of the drawing keeps the English direction, «→», «↓», since the drawings run left to right in every language; inside an Arabic label, between two Arabic phrases, it is «←». "Open full screen ↗" is «فتح بملء الشاشة ↖», mirrored as right-to-left pages mirror the arrow of a link that leaves the page. The ferry's badge "→" stays: it means the right bank.
- **Capitals:** Arabic has none; Latin names keep theirs.
- **The card questions on the home page and on "Why"** write no Latin letter at all, as the English writes none there: the picture check reads a lone Latin capital in them as a label of the picture.
- **YAML:** a text that holds `: ` or begins with `*` is quoted, as is one that begins with `{` (a slot that opens the text) or with any of `[`, `]`, `&`, `!`, `|`, `>`, `%`, `@`, `` ` ``. «، ؛ ؟ « » need no quotes.
- **Titles of works** go in «» where the English writes them plain: «مدرسة أثينا», «الأصول», «الحساب الذهني». The emphasis of a text is the English one's, so *How to Solve It* is *البحث عن الحل*, in asterisks and without guillemets.
- **Right to left.** Every page runs right to left (the build sets `dir="rtl"` for `ar`), so a paragraph is right to left even when it opens with «MathTrail» or a number, and Latin words — MathTrail, Claude, ChatGPT, Google Drive, a file's name, code — and numbers sit in the sentence as they are. No mark that turns the direction is ever typed (U+200E, U+200F, U+061C, U+202A…): Intl puts them into `ar` percentages and prices, and a text never copies them. What keeps a sentence in order:
  - A Latin word before a sentence mark is safe: «…إلى Claude.», «…في Claude؟», «…من MathTrail».». So is a Latin phrase in guillemets: «Use caution before running this prompt».
  - **Two Latin pieces never meet across spaces and marks alone.** A Latin word and the next Latin word, letter, code or number glue into one left-to-right piece and carry the Arabic marks between them inside it. «يقول A: «B كذاب».» draws its opening guillemet between A and B; «Claude، ChatGPT» pulls the Arabic comma into English; «commit {commit} · {date}» puts the day's number beside the hash, away from its month; «Elo فوق IRT، 11 مرحلة» reads "IRT، 11". An Arabic word between them keeps the pieces apart: «يقول أ: «ب كذاب».» (hence Arabic letters for the islanders), «Claude وChatGPT» («و» joined to the Latin word, as OpenAI's Arabic help writes «ChatGPT وCodex»), «commit {commit}، بتاريخ {date}», «Elo فوق IRT، و11 مرحلة».
  - Before a Latin name stands a preposition of its own — «في», «على», «إلى», «من», «عن», «مع» — never «بـ», «لـ» or «كـ» joined to it: «إضافة إلى Claude», «الصفحة الرئيسية لموقع MathTrail», as the card writes «في MathTrail» and «على Google Drive».
  - A label of Claude's that opens with a sign, "+ Add", shows its "+" to the right of "Add". That is the order an Arabic reader reads it in, so it is written as Claude writes it, «**+ Add**».
  - A slot is a word like any other: «الصفوف {range}» draws the range after the word. A slot that holds Latin text, a name or a code, follows the rule of two Latin pieces.
  - Parentheses are typed as in English, «(Radesky وآخرون، 2022)», and the browser turns them.
- **Letters as labels** — the islanders, the children of enumeration, the boxes of ratios, the speakers on "Why" — are Arabic, «أ، ب، ج» in the abjad order the Mashriq's school books label with. Arabic letters written side by side join into one shape («أب» reads as a word), so in a sentence the letters of one case stand apart, a space between them: «أ ب ج، أ ج ب، ب أ ج»; "Yes: AB and BA are two lines" is «نعم: «أ ب» و«ب أ» صفّان مختلفان». Between two single letters «و» stands apart too, as geometry books write «النقطتان أ و ب»: «أ و ب كلاهما كذاب»; "Write K and L in pencil" is «كتابة ف و ك بقلم الرصاص».
- **Drawings** keep their left-to-right layout and take Arabic labels, written as natural Arabic, with the numbers where the English has them: «6 محطات، 3 دقائق بين كل محطتين», «40 د». Such a label shows in order only once the drawings' code gives each label a direction of its own (see the last section); the words do not bend around the code. A label that names things of the drawing from left to right, such as the flags' "21st, 22nd, 23rd", writes numbers and signs alone, «21، 22، 23», so that its order stays the drawing's. Left and right in a text that points at a drawing stay left and right: «الأبكر على اليسار».

## Platforms' labels

### Claude

Claude's interface does not speak Arabic. Its help lists the interface's languages as English, French, German, Hindi, Indonesian, Italian, Japanese, Korean, Portuguese (Brazilian), Spanish (Latin America) and Spanish (Spain) (C, updated on 6 August 2026); the help center's own language picker offers twelve languages and no Arabic, and its `/ar/` address serves the English page. So Claude's labels stay English, as the Russian keeps them, in bold inside Arabic sentences that speak in the plural: «اضغطوا **+ Add** واختاروا **Add custom connector**».

- [A] "Get started with custom connectors using remote MCP", https://support.claude.com/en/articles/11175166-getting-started-with-custom-connectors-using-remote-mcp (https://support.claude.com/ar/articles/11175166-getting-started-with-custom-connectors-using-remote-mcp serves the same English page)
- [B] "Use connectors to extend Claude's capabilities", https://support.claude.com/en/articles/11176164-use-connectors-to-extend-claude-s-capabilities
- [C] "How to use Claude in your preferred language", https://support.claude.com/en/articles/10769299-how-to-use-claude-in-your-preferred-language

| English on the site | Arabic | Where the help says it |
|---|---|---|
| Customize → Connectors | Customize → Connectors | A: "Navigate to Customize > Connectors". The `connectors` key and the link in help; the site's arrow, which stays «→» between two English labels |
| + / + Add | «اضغطوا **+ Add**» | A: "Click "+ Add" and then "Add custom connector"" |
| Add custom connector | **Add custom connector** | A |
| Name it MathTrail | «سمّوه MathTrail» | A: "Enter a name for the connector"; the field's label is not quoted |
| the address | «العنوان»; once, «عنوان خادم MCP البعيد» | A: "your connector's remote MCP server URL" |
| Continue, and Continue again | «اضغطوا **Continue**، ثم **Continue** مرة أخرى» | A quotes "Continue" once, after the URL; the second press is the same button |
| the sign-in settings Claude detects | «إعدادات تسجيل الدخول التي يكتشفها Claude لا تحتاج إلى تغيير» | A: "Review the authentication settings Claude detected"; described in Arabic, since the panel is English |
| Add | **Add** | A: "Finish setting up your connector by clicking "Add"" |
| Connect | **Connect** | B: "Click "Connect" or "Install"" |
| + → Connectors, in a chat | «**+** ← **Connectors**» | A and B: "the "+" button on the lower left of your chat interface, then "Connectors"" |
| the free plan | «الخطة المجانية»; the plan's name, where it is named, Free | A: "Free users are limited to one custom connector" |
| Always allow | **Always allow** | A, its section on security; not quoted on the site yet |
| "Use caution before running this prompt" | quoted in English: «Use caution before running this prompt» | not in Claude's help in any language found; the interface is English |
| Claude asks whether to allow MathTrail to save | described: «يسأل Claude قبل أن يحفظ MathTrail الملف الشخصي» | the question is not quoted in Claude's help; only the button "Always allow" is |

### ChatGPT

OpenAI's help speaks Arabic, though each of its Arabic pages says it was machine-translated («هذه الصفحة مترجمة باستخدام الترجمة الآلية»). help.openai.com refuses fetchers; the pages below were read in a browser on 11 October 2026.

| English on the site | Arabic | Source |
|---|---|---|
| developer mode | **«وضع المطوّر»** | https://help.openai.com/ar/articles/12584461-developer-mode-and-mcp-apps-in-chatgpt — «وضع المطور وتطبيقات MCP في ChatGPT». The body writes «وضع المطوّر», the title and the questions «وضع المطور», without the shadda. The path to turn it on: «الإعدادات ← التطبيقات ← الإعدادات المتقدمة ← وضع المطور»; to add an app, «التطبيقات ← إنشاء», then «انقر على إنشاء» |
| ChatGPT's directory | «دليل ChatGPT»; its name now «دليل المكوّنات الإضافية» | https://help.openai.com/ar/articles/20001256-plugins-in-chatgpt — «المكوّنات الإضافية في ChatGPT»: «يتوفر دليل المكوّنات الإضافية عبر خطط ChatGPT». The app directory became the plugin directory on 9 July 2026 (see [the Spanish notes](es.md)); the generic «دليل ChatGPT» stays true |
| apps · connected apps | «التطبيقات» · «التطبيقات المتصلة» | https://help.openai.com/ar/articles/11487775-connected-apps-in-chatgpt — «التطبيقات المتصلة في ChatGPT» |
| Plus, Pro, Business, Enterprise, Education | the same | names of plans |

The help writes its paths with «←», as the site does, and speaks in the masculine singular («انتقل», «انقر»); the site wraps the labels in the plural: «فعّلوا **وضع المطوّر** من **الإعدادات ← التطبيقات ← الإعدادات المتقدمة**، ثم أضيفوا تطبيقًا من **التطبيقات ← إنشاء**».

### Google

| English on the site | Arabic | Source |
|---|---|---|
| Google Drive | Google Drive, in Latin letters; My Drive is «ملفاتي» | https://support.google.com/drive/answer/2409045?hl=ar («على الجانب الأيمن، انقر على ملفاتي») |
| Drive's bin | «المهملات» in Drive; the button «إفراغ المهملات»; files go for good after 30 days, «يتم حذف الملفات التي تنقلها إلى المهملات نهائيًا بعد 30 يومًا». The card writes «سلة المهملات في Drive» (`data.delete_how`) | https://support.google.com/drive/answer/2375102?hl=ar, «حذف الملفات في Google Drive» |
| version history | «النُسخ السابقة» of the file; for a file such as the profile, the menu «إدارة النُسخ» and «الاحتفاظ للأبد». Google Docs calls its own history «سجلّ التعديلات», which the profile's file does not have | https://support.google.com/drive/answer/2409045?hl=ar, «الاطّلاع على النشاط ونُسخ الملفات»; https://support.google.com/docs/answer/190843?hl=ar |
| linked apps, third-party access | «صفحة التطبيقات المرتبطة في حسابك على Google» (myaccount.google.com/linkedapps); the section «إذن الوصول إلى حسابك على Google»; then «الاطّلاع على التفاصيل», «إزالة إذن الوصول» and «تأكيد», which an Arabic sentence joins with «←». The card's older words, «التطبيقات التابعة لجهات خارجية التي لها حق الوصول» (`data.access_how`), are not used on the site | https://support.google.com/accounts/answer/13533235?hl=ar, «إدارة عمليات الربط بين حسابك على Google وتطبيقات أنشأها مطوّرون آخرون». "Delete all connections" is not on that page; the site needs only «إزالة إذن الوصول» |
| the box for Google Drive on the consent screen | the box of Google Drive; Google words the permission «الاطّلاع على ملفات Google Drive المُستخدَمة في هذا التطبيق وتعديلها وإنشاؤها وحذفها» | as Google lists it in Arabic in its Workspace Marketplace: https://workspace.google.com/marketplace/app/docs_sheets_editor_for_drive/703230099925?hl=ar. Not seen on the live consent screen, and the English never quotes it |
| Sign in with Google | «تسجيل الدخول باستخدام حساب Google», the feature's and the button's name | https://developers.google.com/identity/branding-guidelines?hl=ar |
| Google account | «حساب Google» (card); Google's help writes «حسابك على Google» | |
| Enter, the key | «مفتاح Enter» | https://support.google.com/mail/answer/6594?hl=ar («اضغط على مفتاح Enter») |

MathTrail's own sign-in pages speak English, so "Tick the box for Google Drive" and **Back to Google** stay English.

## Glossary

"card" marks the card's own word, from `web/locales/ar.json`.

### Product and lesson

| English | Arabic | Note |
|---|---|---|
| task | مسألة | card, `task.label`; plural «مسائل» |
| problem (olympiad) | مسألة | the same word as task |
| school sum | تمرين مدرسي | "A school sum" is «تمرين مدرسي» |
| topic | موضوع | card; plural «مواضيع», as the card's «المواضيع» (`progress.topics`) |
| group of topics | مجموعة | `group.*` exactly |
| trap | الفخ | card, `result.trap`; plural «فخاخ» |
| the catch (not a trap) | الصعوبة | never «فخ» outside the card's sense: "the catch is that an hour has 60 minutes" is «والصعوبة أن الساعة 60 دقيقة» |
| technique | أسلوب | plural «أساليب» |
| the main move (`topic.method`) | الفكرة الأساسية | |
| hint | تلميح | card; never for "clue" |
| clue, a given of a task | معلومة | plural «معلومات»; "the givens" «المعطيات». Keeps «تلميح» for the card's hint: «كل معلومة سهم» |
| option | خيار | plural «خيارات»; the card lists them under «الإجابات» (`result.answers`). "Five options, A to E" is «خمسة خيارات، من A إلى E»: the card's letters are Latin |
| answer | الإجابة | card |
| right, wrong answer | إجابة صحيحة، إجابة خاطئة | card: «صحيح!», «ليس تماما» |
| solution | الحل | card |
| step | خطوة | |
| question, the text of a task | نص المسألة | card: «أخذ عدد من نص المسألة كإجابة» |
| question, what is asked | السؤال، المطلوب | card: «الإجابة عن سؤال آخر» |
| picture of a task | الرسم | «صورة» for a photograph |
| label of a picture | تسمية | |
| grade | الصف | card, `profile.grade` |
| Grade 3 | الصف 3 | card, `child.grade` |
| grades 1 to 6 | الصفوف من 1 إلى 6 | card, `first_run.lead` |
| grades 3–4, two slots | الصفوف {from}–{to} | card, `progress.grades_ranks` |
| level | المستوى | card: «المستوى العام» |
| difficulty | الصعوبة، مستوى الصعوبة | card, `waiting.step.topic` |
| rank | المرحلة | card, «المرحلة {rank}»: a number, never a title |
| rating | التصنيف | card: «التصنيف في هذا الموضوع», «التصنيف العام» |
| progress | التقدّم | card, «الملف الشخصي والتقدم» |
| mastered · mastery | تم إتقانه · الإتقان | card, `progress.mastered` |
| review, the card's section | المراجعة | card, `review.title` |
| how the answer went | مراجعة الإجابة | card, `result.label` |
| trial series | السلسلة التجريبية | card |
| skipped | تم التخطي | card |
| profile | الملف الشخصي | card. "The profile's file" is «ملف الملف الشخصي», or «الملف» once the profile is named |
| pseudonym | الاسم المستعار | card |
| interests | الاهتمامات | card |
| lesson | الدرس | card: «لغة الدروس» |
| coach | المدرب | card, `topic_choice.coach_short`, «مدرب الأولمبياد»; the Coach product too |
| companion, in the coach | الرفيق | a character, not the child |
| parent | ولي الأمر؛ أولياء الأمور؛ «أحد الوالدين» in a sentence | card |
| tutor | معلم | card, `first_run.lead` |
| adult | شخص بالغ، الشخص البالغ | card |
| child, your child | الطفل؛ طفلكم | see "No gender shows" |
| children | الأطفال | card; counted on "Research", «طفل» and «أطفال» |
| pupils, in a task | تلاميذ | the task's own group, in the generic masculine plural |
| class, of pupils | الفصل | never «الصف», which is the grade |
| chat | الدردشة | card |
| card | البطاقة | card |
| connector, custom connector | موصّل، موصّل مخصص | as OpenAI's Arabic help writes «موصّلات MCP مخصصة»; Claude's own labels stay English |
| app | تطبيق | OpenAI's Arabic help |
| sign in, sign-in | تسجيل الدخول | card: «أسجّل الدخول» (`profile.signin_country_off`); Google |
| account | حساب | card: «حساب Google» |
| Drive file, the profile's file | ملف على Google Drive، ملف الملف الشخصي | card, `data.export_file` |
| folder | مجلد | card, `data.export_where` |
| bin | المهملات | Google; the card writes «سلة المهملات في Drive» |
| daily limit | الحد اليومي | card |
| country · state, in the United States | البلد · الولاية | card, `profile.country`, `profile.region` |
| Another task | مسألة أخرى | the card's button, quoted exactly |
| Make a wish for tomorrow | هيا نتمنى أمنية للغد | card, quoted exactly |
| Profile & progress | الملف الشخصي والتقدم | card |
| Edit | تعديل | card |
| the bulb, the hint's button | زر المصباح | |
| the two arrows | زر السهمين | |
| tap | النقر؛ نقرة | "A tap records the answer" is «تُسجَّل الإجابة بنقرة» |
| press, click | الضغط: «اضغطوا» | |
| Enter, the key | مفتاح Enter | as Google's Arabic help names it |
| free, of price | مجاني | the chip `hero.free` |
| Free, the plan | الخطة المجانية؛ Free | the plan's name where it is named |
| open source | مفتوح المصدر | |
| code, of the project | الشيفرة، الشيفرة المصدرية | "View the code on GitHub" is «عرض الشيفرة على GitHub» |
| in development | قيد التطوير | |
| beta | Beta | |
| soon | قريبًا | |
| olympiad maths | رياضيات الأولمبياد | card: «مدرب الأولمبياد» |
| olympiad-style task | مسألة بأسلوب الأولمبياد | card, `first_run.lead` |
| maths circle | نادي الرياضيات | the word Arab parents know from school; «حلقة» reads as a study circle |
| school | المدرسة | card |
| school mark | الدرجة المدرسية | card: «درجة مدرسية» |
| homework | الواجبات المدرسية | |

### The site's sections and headings

| English | Arabic | Note |
|---|---|---|
| Topics (`nav.topics`) | المواضيع | 8 letters; card, `progress.topics` |
| Techniques (`nav.techniques`) | الأساليب | 8 |
| Why (`nav.why`) | لماذا | 5 |
| Coach (`nav.coach`) | المدرب | 6, beside the tag «Beta»; card |
| Research (`nav.research`) | الأبحاث | 7. Not «البحث», which a menu reads as "Search". The page's title is «الأبحاث» too, as a page's title is its menu entry in every language, and the footer shows it |
| Service (`nav.service`) | الخدمة | 6 |
| About (`nav.about`) | من نحن | 6 with its space, the usual Arabic, and the page's title too |
| Add to Claude (`nav.add`) | إضافة إلى Claude | a verbal noun, as the card's buttons «إضافة», «حفظ», «تعديل» |
| Technical pages (`nav.technical`) | الصفحات التقنية | said by a screen reader |
| Pages (`nav.pages`) · Menu · Language | صفحات الموقع · القائمة · اللغة | |
| Help · Privacy policy · Terms of use | المساعدة · سياسة الخصوصية · شروط الاستخدام | |
| How it is solved | طريقة الحل | |
| Where children go wrong | أين يخطئ الأطفال | |
| What to say | ما يمكن قوله | to the child; no "you" |
| How to help at home | المساعدة في البيت | |
| Learn first | قبل هذا الموضوع | |
| Opens the way to | يمهّد الطريق إلى | |
| Builds on · Opens | يعتمد على · يقود إلى | the topics page |
| Ask for a task on this topic | اطلبوا مسألة في هذا الموضوع | |
| Show · Hide the solution | عرض الحل · إخفاء الحل | card: «إخفاء التلميح» |
| Example 2 | مثال 2 | |
| Why it matters | لماذا هذا مهم | |
| How to solve: … | طريقة الحل: … | followed by a verbal noun: «طريقة الحل: رسم الخط» |
| The rules of … | قواعد … | |
| Check: | التحقق: | |
| When it helps: | متى يفيد: | the technique, «أسلوب», is masculine |
| In the topics: | في المواضيع: | |
| You are here · On this page · All topics | أنتم هنا · في هذه الصفحة · كل المواضيع | |
| More → | المزيد ← | |
| Honestly | بصراحة | |
| See a lesson | مشاهدة درس | |
| For technical readers | للجمهور التقني | «للقرّاء» would be a masculine plural |
| You, the parent's line on "Why" · C, the child's | أنتم · ط | «ط: طفلكم» |
| The adult · The model, in the chat's illustration | الشخص البالغ · النموذج | |
| Step 1, of the topics' map | الخطوة 1 | |
| Foundations · Rules · Hard problems | الأسس · القواعد · المسائل الصعبة | |
| Free · Open source, the chips | مجاني · مفتوح المصدر | |
| Copy · Copied | نسخ · تم النسخ | as the card's «تم الحفظ» |
| Task · Answer · Sign-in, the scenarios on "Service" | المسألة · الإجابة · الدخول | short: the tabs do not wrap at 320 px |

The menu fits with room to spare. Measured in Chromium at the menu's 15 px, the seven entries take 261 px together against 331 for the English, 382 for the Spanish and 389 for the Russian, and «إضافة إلى Claude» takes 98 px against 92 for "Add to Claude". The pinned Playwright image has no Arabic font but FreeSerif, which is narrow; the fonts of phones and computers draw Arabic wider, by perhaps a third, which still leaves the row below the Russian one. `just site-fold` measures it properly.

### Maths and olympiad terms

| English | Arabic | Note |
|---|---|---|
| olympiad | الأولمبياد | card; masculine, as «الأولمبياد الدولي للرياضيات» |
| reference task | مسألة مرجعية | |
| solver | برنامج الحل | the first time «برنامج يحل المسألة» |
| brute force | تجربة كل الحالات | «القوة الغاشمة» on the technical pages |
| worst case | أسوأ حالة | card, `trap.best_case_not_worst` |
| good, bad luck | الحظ الجيد، الحظ السيئ | card: «أسوأ ما قد يحدث من حظ» |
| to be sure, for certain | للتأكد، بالتأكيد | never «متأكد» |
| pigeonhole principle | مبدأ برج الحمام | card, and Arabic Wikipedia's title |
| pigeons, holes | الحمام، الأعشاش | "Pigeons and pigeonholes" is «الحمام والأعشاش» |
| parity · even · odd | الزوجية · زوجي · فردي | card |
| alternation, taking turns | التناوب | card |
| remainder · quotient | الباقي · ناتج القسمة | card |
| divisibility | قابلية القسمة | card |
| divides by 9 | يقبل القسمة على 9 | «القسمة على», throughout |
| divisibility rules | قواعد قابلية القسمة | card, `skill.divisibility_rules` |
| digit · two-digit number | رقم · عدد مكوّن من رقمين | «رقم» is a digit and «عدد» a number; never swapped |
| tens, units digit | رقم العشرات، رقم الآحاد | |
| enumeration | حصر الحالات | card |
| case · list | حالة · قائمة | |
| systematic listing | إنشاء قائمة منظمة | the name Saudi school books give the strategy, «خطة حل المسألة: إنشاء قائمة منظمة» |
| tree, the diagram | مخطط شجري | |
| working backwards | الحل عكسيًّا | the school books' name; "Work from the end" is «البدء من النهاية» |
| the assumption method | طريقة الافتراض | |
| proof by contradiction | برهان الخُلف | Arabic Wikipedia's title for it |
| suppose · assumption · contradiction | لنفترض أن… · افتراض · تناقض | |
| pattern | نمط | the school books' «البحث عن نمط» |
| symmetric strategy | استراتيجية التماثل | |
| winning strategy | استراتيجية رابحة | card |
| winning, losing position | وضع رابح، وضع خاسر | |
| move, in a game | نقلة | card, `trap.first_move_assumed` |
| the first, the second player | اللاعب الأول، اللاعب الثاني | |
| invariant | لامتغيّر | Arabic Wikipedia |
| knights and liars | الفرسان والكذابون | card |
| knight · liar | فارس · كذاب | «ف» and «ك» in drawings; a girl is «فارسة», «كذابة» |
| island · islander | جزيرة · ساكن الجزيرة | plural «سكان الجزيرة» |
| statement | قول | card, `trap.trusted_statement` |
| ordering | الترتيب | card |
| number line | خط الأعداد | card; Wikipedia's «مستقيم الأعداد» is not used |
| grid · squared paper | شبكة · ورق المربعات | card |
| cell · side of a cell | مربع · ضلع المربع | card: «شبكة من المربعات». A cell of a table is «خانة» |
| area · perimeter | المساحة · المحيط | card |
| ratio · in the ratio 3 to 4 | النسبة · بنسبة 3 إلى 4 | card: «النسب والتقسيم» |
| sharing · share | التقسيم · الحصة | card |
| percentage · percent | النسبة المئوية · بالمئة | card |
| fraction · the whole · the part | الكسر · الكل · الجزء | card |
| a half, a third, a quarter | النصف، الثلث، الربع | |
| gaps and boundaries | الفواصل والحدود | card |
| gap, between posts · post · fence | فاصل · عمود · سياج | card, `advice.off_by_one`: «الأعمدة نقاطًا والفواصل بينها» |
| off by one | خطأ بواحد | card |
| clock face · hands · the whole hour | وجه الساعة · العقربان · الساعة التامة | card: «وجه ساعة» |
| floor 1, of a house | الطابق 1 | never «الطابق الأرضي», which would change the count |
| calendar and age | التقويم والأعمار | card |
| pan balance · weighing · fake coin | ميزان ذو كفتين · وزنة · قطعة نقدية مزيفة | card: «ميزان ذو كفتين». «وزنة», plural «وزنات», is one use of the balance |
| jug · pour | إبريق · سكب | card: «الوزن وسكب السوائل» |
| state, of the jugs | حالة | |
| overlapping groups | المجموعات المتداخلة | card |
| arithmetic with a trick | حساب فيه خدعة | card |
| I thought of a number | فكّرتُ في عدد | the first person shows no gender |
| carrying | الحمل: «من غير حمل» | Gulf school books say «إعادة التجميع» |
| round number | عدد مدوَّر | 10, 100, 1000 |
| match, the stick | عود ثقاب | |
| a special case | حالة خاصة | |
| coin, as the money of a task | قطعة نقدية | "400 coins" is «400 قطعة نقدية» |
| x, the unknown | س | |

### The technical pages

| English | Arabic | Note |
|---|---|---|
| token · access token · refresh token | الرمز المميز · رمز الدخول · رمز إعادة التحميل | as Google's Arabic OAuth docs |
| authorization code | رمز التفويض | Google |
| seal · sealed · sealing | الختم · مختوم · الختم | |
| encrypted | مشفّر | |
| state · stateless | الحالة · بلا حالة | |
| MCP server · tool · tool call | خادم MCP · أداة · استدعاء أداة | |
| chat host | مضيف الدردشة | |
| consent screen | شاشة الموافقة | Google's docs speak of «موافقة المستخدم» |
| scope | النطاق | Google: «نطاقات OAuth»; «نطاق drive.file» |
| client · client metadata document | العميل · مستند البيانات الوصفية للعميل | |
| renewal · revocation | التجديد · الإبطال | |
| pace · ceiling | الوتيرة · السقف | |
| target, of the profile's size | الهدف | |
| log · logging | السجل · التسجيل | «Cloud Logging» stays |
| instance · revision · image | مثيل · نسخة · صورة الحاوية | |
| key rotation · ciphertext · nonce | تدوير المفاتيح · النص المشفر · nonce | |
| fingerprint · near-duplicate | بصمة · شبه مكرر | |
| readability · self-check | سهولة القراءة · التحقق الذاتي | |
| brief, for the model | التكليف | |
| package · schema · template | الحزمة · المخطط · القالب | |
| deterministic · the rule | حتمي · القاعدة | |
| cookie · cache | ملف تعريف الارتباط · ذاكرة التخزين المؤقت | Google's Arabic: «ملفات تعريف الارتباط» |
| graceful shutdown | إيقاف منظّم | |
| lifetime, of a token | مدة الصلاحية | |
| Workload Identity Federation | Workload Identity Federation | Google's docs of it have no Arabic page; the name stays |
| student model | نموذج المتعلّم | |
| chance, of a right answer | احتمال الإجابة الصحيحة | |
| corridor · interval · goal | الممر · الفترة · الهدف | "the corridor aimed at" is «الممر المستهدف» |
| the ceiling, an estimate that knows where the child stands | السقف | |
| bench · simulated children | منصة المحاكاة · أطفال افتراضيون | |
| a child who learns, who stays put | مع التعلّم، بلا تقدّم | no verb with the child: "a child who learns · a share of tasks" is «مع التعلّم · نسبة من المسائل» |
| lag · logit · points | التأخر · لوجيت · نقاط | |
| mastery declared falsely | إتقان أُعلن خطأً | |
| live data, live numbers | البيانات الفعلية، الأرقام الفعلية | they come monthly, not in real time |
| range, of the chance promised | الشريحة | |
| standard error · baseline · seed | الخطأ المعياري · خط الأساس · البذرة | |
| Reached · On the edge · Not reached | تحقّق · على الحد · لم يتحقق | |
| paper · public domain | الورقة البحثية · الملك العام | |
| commit · release | commit · الإصدار | «commit» is the borrowed term; keep an Arabic word between it and a date |
| endpoint · claim, of a JWT | نقطة النهاية · مطالبة | |
| request, the model's · store · router · issuer | الطلب (card) · المخزن · الموجّه · المُصدِر | |
| deploy · build · client secret · sandboxed | النشر · البناء · سر العميل · معزول | «سر العميل» as Google's Arabic docs |
| in the open, unsealed · timeout | مكشوف · المهلة | |
| digest · runtime · PKCE verifier · state machine | الملخص · بيئة التشغيل · أداة التحقق PKCE · آلة الحالات | |
| scenario · sketch · pinned · sliding, of a lifetime | سيناريو · رسم تقريبي · مثبّت · منزلق | |
| canvas · tablet | لوحة الرسم · جهاز لوحي | the coach |
| GitHub issue · computer · mouse | بلاغ على GitHub · حاسوب · فأرة | "open an issue on GitHub" is «فتح بلاغ (issue) على GitHub» the first time |
| the profile's notes | الملاحظات | «درجة» is a school mark |
| whisker, of an error bar · dashed | خط الخطأ · متقطع | |
| item response theory · psychometrics | نظرية الاستجابة للفقرة (IRT) · القياس النفسي | Arabic Wikipedia's titles |
| Elo | Elo | not translated; Arabic Wikipedia writes «نظام تصنيف إيلو» |
| KB · px | KB · px | as the English |

## Pólya's four steps

The Arabic edition is «البحث عن الحل: الأسلوب الرياضي من زاوية جديدة», translated by Ahmad Salim Saidan («أحمد سليم سعيدان») and reviewed by Wasfi Hijab («وصفي حجاب»), published in Beirut by «دار مكتبة الحياة» with «مؤسسة فرانكلين للطباعة والنشر»: 1960 in the catalogue of Beirut Arab University (http://librarycatalog.bau.edu.lb/cgi-bin/koha/opac-detail.pl?biblionumber=23637); archive.org's scan of the second edition is dated 1965 (https://archive.org/details/14740928). It is also Arabic Wikipedia's title for the book, which lists the four phases as «فهم المسألة»، «ابتكار الخطة»، «تنفيذ الخطة»، «المراجعة» (https://ar.wikipedia.org/wiki/البحث_عن_الحل). The scan bears them out: its sections are headed «فهم المسألة» (§7), «ابتكار الخطة» (§9) and «تنفيذ الخطة» (§10); its summary list heads the last two «تنفيذ الخطة» and «المراجعة»; the translator's preface sums the four up as «فهم المسألة، ورسم الخطة وتنفيذها، ثم مراجعة الحل». A newer translation, «كيف تحل مسألة: منحى جديد في الطريقة الرياضية» by Mu'in Rumiyya («معين رومية»), exists (https://www.goodreads.com/book/show/45310692, no publisher or year given); the site names the older, standard one.

The page "Techniques" names its steps after Saidan's, as verbal nouns:

| English | Arabic |
|---|---|
| *How to Solve It* (1945) | *البحث عن الحل* (1945) |
| Understand the problem | فهم المسألة |
| Make a plan | ابتكار خطة |
| Carry it out | تنفيذ الخطة |
| Look back | مراجعة الحل |

"Look back" takes Saidan's own «مراجعة الحل» rather than «المراجعة», which is the card's name for its review (`review.title`). A technique's name is a verbal noun, as these steps are, and takes the school books' name where they have one: «رسم صورة» (Draw the problem), «إنشاء جدول» (Make a table), «البدء بأعداد صغيرة» (Start with small numbers; aka «حالة خاصة»), «إنشاء قائمة منظمة» (List systematically), «البدء من النهاية» (Work from the end; aka «الحل عكسيًّا»), «افتراض أن الجميع من نوع واحد» (Suppose they are all the same; aka «طريقة الافتراض»), «البحث عن نمط» (Look for a pattern), «افتراض العكس» (Suppose the opposite; aka «برهان الخُلف»), «مراقبة الزوجية» (Watch the parity), «الحمام والأعشاش» (Pigeons and pigeonholes; aka «مبدأ برج الحمام»), «محاكاة نقلات الخصم كالمرآة» (Mirror the moves; aka «استراتيجية التماثل»). The Saudi school books' names: https://sahl.io/sa/lesson/484/ («خطة حل المسألة: إنشاء قائمة منظمة»), https://sahl.io/sa/lesson/284/ («البحث عن نمط»), https://sahl.io/sa/lesson/695/ («الحل عكسيًّا»).

The steps' questions speak to the child with "we" and no "you", in Saidan's words «المعطيات» and «المطلوب»: «لنُعِد صياغة المسألة بكلماتنا. ما المعطيات؟ وما المطلوب؟». The coach's shorter line, "Understand → make a plan → solve → check", is «الفهم ← وضع خطة ← الحل ← التحقق», and the picture's "understand, plan, solve, check" on "Why" is «الفهم والتخطيط والحل والتحقق».

## People's names

"Initial" marks a name whose first letter a drawing or a badge shows; within one drawing no two letters repeat, and none of one drawing differ only by their dots (ب ت ث ن, ج ح خ, ف ق).

| English | Arabic | Where | Initial, badge |
|---|---|---|---|
| Galt | غالت | home, the cards' pseudonym: «اختيرت لغالت», «مهارات غالت» | transliterated, as the Russian «Галт»; an invented name shows no gender. Never the subject of a verb |
| Comet | نيزك | "Why" and "Topics", the cards' pseudonym | a meteor, as the Russian «Комета»; not «مذنّب», which unvocalized reads as «مُذنِب», a sinner. No personal name, so no gender |
| C, "your child" on "Why" | ط | `parent.explain.child`: «ط: طفلكم» | the speaker's letter, for «طفل» |
| Ann | أمل | ordering, parts, ratios, clocks, calendar | «أ» |
| Ben | باسم | ordering, parts, ratios, parity | «ب» |
| Kate | كوثر | ordering, calendar | «ك»: `art.kate` is «ك» |
| Dan | داود | ordering's line | name in a drawing |
| Lena | ليلى | ordering's queue, gaps, ratios | «ل» |
| Max | مازن | ordering's queue, ratios | «م» |
| Nina | ندى | ordering's queue | «ن» |
| Oleg | عمر | ordering's queue | «ع» |
| Pavel | طارق | ordering's queue | «ط», and the badge «ط» |
| Ada | أروى | ordering's runners | «أ» |
| Bob | بدر | ordering's runners | «ب», as Ben's «باسم» on the same page, in another drawing |
| Cid | جاد | ordering's runners | «ج» |
| Dot | دانة | ordering's runners | «د» |
| Eli | سامي | ordering's runners | «س» |
| Fay | فرح | ordering's runners | «ف», and the badge «ف» |
| Anya | أسماء | "Techniques": the pets and the knights | «أ». A girl and a knight: «أسماء فارسة» |
| Borya | بلال | "Techniques": the pets and the knights | «ب». «بلال كذاب» |
| Vera | ريم | "Techniques", calendar | name in drawings |
| Tanya | سلمى | "Techniques", working backwards | |
| Petya | يوسف | "Techniques", working backwards | |
| Lea | لينا | enumeration | «ل» |
| Tom | تامر | enumeration, clocks, parts | «ت»; the handshake drawing shows «ل» and «ت» alone |
| Masha | مريم | gaps, arithmetic | «م» |
| Sasha | سارة | pigeonhole, the sweets | a girl: the English gives no gender, but an Arabic verb must |
| Gleb | خالد | pigeonhole, the socks | |
| Kim | رنا | winning strategy | name in drawings: «رنا −3» |
| Leo | ليث | winning strategy | name in drawings: «ليث −4» |
| Grandma, Granny | الجدة | clocks, parts | in drawings: «الجدة», «ساعة الجدة» |
| Mum, Dad | الأم، الأب | calendar, ordering's tip, "About" | in drawings: «الأم», «الأب» |
| islanders A, B, C | أ، ب، ج | knights | letters |
| knight K, liar L | ف، ك | knights: `art.knight`, `art.liar`, the rules' marks (`basis.cards.*.mark`), "K, L, K, L…" «ف، ك، ف، ك…», the tip "Write K and L"; `thumb.*` | the first letters of the card's «الفرسان والكذابون»; neither is an islander's |
| children A, B, C · boxes A, B | أ، ب، ج · أ، ب | enumeration · ratios | letters, apart: «أ ب ج» |
| the cook (she) · the baker (he) | طاهية · خبّاز | weighing · parts | |
| woodcutters · a farmer | حطّابون · مزارع | gaps · percentages | |

**The islanders are «أ، ب، ج», not A, B, C**, on the knights page in words and drawings alike. The page names them in nearly every sentence, beside colons and guillemets: a Latin capital there glues to the next one and pulls the marks between them into a left-to-right piece, as «يقول A: «B كذاب».» shows, while «يقول أ: «ب كذاب».» reads in order. Arabic olympiad and school books label persons and points «أ، ب، ج» in the abjad order, so the letters read as labels, not as words. They do not clash with the knight's «ف» and the liar's «ك». The drawings still run left to right, each letter its own element, so their order does not change; the card's pictures keep their Latin capitals, which the site does not touch. The children of enumeration and the boxes of ratios take Arabic letters too, so that the site labels the same way everywhere.

The people of history keep their Arabic names, as Arabic Wikipedia writes them where it has them: «إقليدس», «الملك بطليموس», «برقلس», «الأصول», «أرخميدس», «رفائيل», «مدرسة أثينا», «كونفوشيوس», «الفنون الست», «أريابهاتا», «بونه», «تقي الدين» and his observatory «مرصد إسطنبول», «إسطنبول», «لوكا بتشولي», «ياكوبو دي بارباري», «الخوارزمي», «بغداد», «الدروس التسع لفن الرياضيات», «أرسطو», «الإسكندر الأكبر», «نيكولاي بوغدانوف-بيلسكي» (Wikipedia writes «بوكدانوف بيلسكي»), «الحساب الذهني. في مدرسة س. أ. راتشينسكي الشعبية», «سانت لويس», «طوكيو», «ضريح كونّو هاتشيمان-غو», «سانغاكو», «لينينغراد», «موسكو», «المجر», «الولايات المتحدة», «الاتحاد السوفيتي», «جورج بوليا». The book "whose title gave us the word "algebra"" is the one whose title gave the world's languages «الجبر»: the Arabic says so, «الكتاب الذي أخذت لغات العالم من عنوانه كلمة «الجبر»», rather than that it gave Arabic a word. A photograph's credit, «Yu. D. Korolev» or «Momotarou2012», stays as its source writes it.

## Grades

`topics.grades` is **«الصفوف {range}»**: «الصفوف 1–2», «الصفوف 3–4», «الصفوف 5–6», «الصفوف 1–6»; so `topic.example` reads «مثال 2 · الصفوف 3–4», the label on "Techniques" «مسألة · الصفوف 2–3», and "Research" writes «الصفوف {first}–{last}».

- It is the card's own word for grades taken together, «مسائل الصفوف {from}–{to}» (`progress.grades_ranks`) and «في الصفوف من 1 إلى 6» (`first_run.lead`), and it reads right with every range, two grades or six.
- On the page it reads «الصفوف» and then the range, right to left, «1–6» drawn as Arabic readers read a range; no word follows the slot, so nothing can glue to it.
- The plural needs no case: «الصفوف» stays as it is after «من», after a verb or alone on a chip.

The alternative is «الصف {range}», the card's chip for a span of grades, «الصف {from}–{to}» (`progress.grades`): shorter, and the same words as that chip, but a singular for a span reads as a label rather than as Arabic. A text with two slots writes the card's form, «الصفوف {from}–{to}»; the home page's lead, which has the one slot, says «للصفوف {range}».

## Books on "Research"

| Key | The English page | Arabic | How sure |
|---|---|---|---|
| perelman.firsthundred | For Young Mathematicians. The First Hundred Puzzles | the English title | no Arabic edition found |
| perelman.secondhundred | The Second Hundred Puzzles | the English title | no Arabic edition found |
| perelman.living | Living Mathematics | «الرياضيات المسلية» (Mir, Moscow; translated by Ibrahim Mahmoud Shousha, «إبراهيم محمود شوشة») | sure it is this book: its first chapter, «إفطار مع الألغاز», opens with the squirrel in the meadow, «السنجاب في المرج», as «Живая математика» does (scan: https://archive.org/details/oldsovietbooks_gmail_2). The year 1986, which Arabic Wikipedia gives, is not checked; that article also names the Russian original wrongly as «Занимательная математика» |
| perelman.arithmetic | Entertaining Arithmetic | the English title | no Arabic edition found |
| perelman.quick | Quick Reckoning | the English title | no Arabic edition found |
| ignatyev.wit | In the Realm of Ingenuity, or Arithmetic for Everyone | the English title | no Arabic edition found |
| dudeney.amusements | Amusements in Mathematics | the English title | no Arabic edition found |
| dudeney.canterbury | The Canterbury Puzzles | the English title | no Arabic edition found; Arabic Wikipedia has an article, «ألغاز كانتربري», which names no edition |
| loyd.cyclopedia | Cyclopedia of Puzzles | the English title | no Arabic edition found |
| carroll.tangled | A Tangled Tale | the English title | no Arabic edition found; Arabic Wikipedia's article on Carroll calls it «حكاية متشابكة», with no edition |
| carroll.pillow | Pillow Problems | the English title | no Arabic edition found |

An Arabic title goes in «»; an English one stays plain, as the Spanish page writes them.

The authors' names, as Arabic Wikipedia writes them: «ياكوف بيرلمان», «يميليان إغناتيف» (no article; written as Arabic writes Russian names), «هنري إ. دوديني», «سام لويد», «لويس كارول». `sources.died`, "† {year}", is «ت {year}»: Arabic marks a death year with «ت», for «توفي», where the English uses a dagger, and the letter shows no gender.

## Search words

Gathered on 2026-10-12 from Google's suggestions for Egypt, Saudi Arabia, Morocco and Jordan, and the titles of the first results; they show that a phrase is typed, not how often, and [the brief](README.md#search-words) says how a text uses them. The suggestions are the same in every country, so these words are the whole Arab world's.

**Who searches, and how.** Google is 94–98 % of searches, Bing 2–4 %. Parents type a grade in words, never in digits: «الصف الثالث الابتدائي», «ثالث ابتدائي», in Jordan «الصف الثالث الأساسي», in Morocco «المستوى الثالث ابتدائي». The word most of them share is «ابتدائي». `topics.grades`, «الصفوف {range}», is the noun of Egypt, the Gulf and the Levant, and a title keeps its digits through the slot; a description can name the stage in words.

**What parents type**, strongest first:

1. «ألغاز رياضيات للأطفال», with «مع الحل» or a grade;
2. «أسئلة ذكاء رياضيات», with a grade or «للأطفال»;
3. «مسائل ذكاء للأطفال», the nearest to the site's «مسائل»;
4. «مسائل رياضيات» with «ابتدائي» or a grade, which leans to homework;
5. «مسائل أولمبياد رياضيات», whose suggestions are mostly for older pupils;
6. «ألغاز منطقية»;
7. «استراتيجيات حل المسائل» and «خطة حل المسألة»;
8. «التفكير المنطقي» and «تنمية التفكير المنطقي».

**The head's title.**

- A topic's page: **«{topic}: مسائل أولمبياد رياضيات للصفوف {range}»**. Without the article, «مسائل أولمبياد رياضيات» reads as olympiad problems in general, where «أولمبياد الرياضيات» reads as the named school contest; «رياضيات» says what a topic's name such as «الساعات» does not; and «للصفوف {range}» is `topics.grades` with «لـ», as the home page's lead writes it. The longest, «ألعاب باستراتيجية رابحة: مسائل أولمبياد رياضيات للصفوف 5–6», is 58 units. The other way, «{topic}: مسائل بأسلوب الأولمبياد للصفوف {range}», 59 at most, is the card's own wording without «رياضيات».
- Any other page: **«{page} — MathTrail»**, the card's dash; the longest, «استراتيجيات حل المسائل في رياضيات الأولمبياد — MathTrail», is 56.

**Topics.** ⚑ marks a card's name nobody types.

| Topic | The card | What parents type |
|---|---|---|
| ordering | «الترتيب» ⚑ | «ألغاز منطقية»; «الترتيب» is the lesson on ascending and descending order |
| knights and liars | «الفرسان والكذابون» ⚑ | «لغز الصادق والكاذب», «لغز الحارسين» |
| enumeration | «حصر الحالات» ⚑ | «مبدأ العد», «الرسم الشجري», «عدد الطرق الممكنة»; the card's name is a school's census forms |
| gaps and boundaries | «الفواصل والحدود» ⚑ | nothing found |
| clocks | «الساعات» | «مسائل الوقت», «الزمن المنقضي» |
| calendar and age | «التقويم والأعمار» ⚑ | «مسائل الأعمار» |
| pigeonhole principle | «مبدأ برج الحمام» | the same |
| parity and alternation | «الزوجية والتناوب» ⚑ | «الأعداد الزوجية والفردية»; «الزوجية» alone is marriage |
| arithmetic with a trick | «حساب فيه خدعة» ⚑ | «خدع رياضية», «حيل رياضية» |
| weighing and pouring | «الوزن وسكب السوائل» ⚑ | «لغز الميزان», «لغز القطعة النقدية المزيفة» |
| parts and shares | «الأجزاء والحصص» ⚑ | «مسائل الكسور» |
| percentages | «النسب المئوية» | the same, «مسائل النسبة المئوية» |
| ratios and sharing | «النسب والتقسيم» ⚑ | «النسبة والتناسب», «التقسيم التناسبي» |
| figures on a grid | «أشكال على ورق المربعات» | «المساحة والمحيط»; «ورق المربعات» is drawing |
| divisibility and remainders | «قابلية القسمة والبواقي» | the same; «باقي القسمة» |
| overlapping groups | «المجموعات المتداخلة» ⚑ | «مخطط فن» |
| games with a winning strategy | «ألعاب باستراتيجية رابحة» ⚑ | nothing that is maths: «ألعاب رياضيات» are games for screens |

**Descriptions** can carry «مسائل رياضيات», «ألغاز رياضيات للأطفال», the stage in words, «المرحلة الابتدائية», on "Techniques" «استراتيجيات حل المسائل», and on "Why" «التفكير المنطقي», which the page's own sources back. «مع الحل» fits a topic's page and "Techniques" only; the home page's honest word is «شرح الخطأ».

**Not to use.**

- Contests and tests: «كانجارو», «موهبة», «مسابقة الأولمبياد العلمية», the school olympiads of Palestine, Algeria and Tunisia, «القدرات».
- Homework and what the site does not give: «pdf», «مذكرة», «الترم», «حل كتاب», «امتحان».
- «للموهوبين», which sorts children, and «اختبار ذكاء» or «ينمّي الذكاء», a test and a promise.
- Words that search reads as something else: «الغاز» without its hamza, gas; «ألعاب رياضية», sports; «الصادق والكاذب» without «لغز»; «أشكال فن».

**For the author.** Parents type «استراتيجيات حل المسائل», where the glossary calls a technique «أسلوب», which is never suggested; a description can say both. Parents type «فكر في رقم», and the glossary keeps «عدد» for a number.

## Places for the author's reading

- The plural «أنتم» for the reader throughout, where the card mostly writes the impersonal «يمكن…» and in a few keys the singular «ـك» (`data.label`, `data.export_*`, `first_run.label`, `first_run.fill`); `topic.path` is «أنتم هنا» and `topic.ask` «اطلبوا مسألة في هذا الموضوع».
- The child never the subject of a verb, though the generic «الطفل» with a masculine verb is ordinary Arabic: it costs a style heavy with verbal nouns and passives, and it is what the card does.
- **The drawings.** They are laid out left to right (`dir="ltr"` on each drawing), and on a page written right to left every text in them takes the direction its first letter sets, so that «6 محطات، 3 دقائق بين كل محطتين» and «40 د» read as they are written while a box of numbers and signs stays left to right. In a sentence, the build keeps a sum, a time, a citation or a piece of code left to right, «12 ÷ 3 = 4», «/mcp», and leaves a word, a number and a range, «3–4», to the sentence's direction. To be read in the preview, page by page.
- The question mark a drawing writes where it asks is the site's word `sign.question`, «؟»; enumeration's lost case is the page's word `lost`, «ب ج؟», its two letters apart so that they do not join into a word. A card's picture keeps the Latin «?», as the card does.
- Arabic letters for the islanders, the children and the boxes, against Latin A, B, C, which would keep the code's "BC?" working but break the knights' sentences.
- `topic.child`, «الطفل»: the generic noun, masculine in form. The alternative is «محاولة الطفل», "the child's try", which is wider than the 64-pixel column.
- `nav.research`, «الأبحاث», plural, against «البحث», which reads as "Search", and «الدراسة», "the study", the Spanish choice.
- The pseudonyms «غالت», which can also be read as a rare verb, «غالَتْ», and «نيزك», a meteor rather than a comet, chosen over «مذنّب».
- The money: US dollars written «9.99 دولار» and «0 دولار», shortened from Intl's «دولار أمريكي»; the alternatives are Intl's full name form, or a large market's own currency, the riyal «9.99 ريال سعودي» or the pound «9.99 جنيه مصري».
- The bait app's cheer, "You're a champion!!!", is «أداء الأبطال!!!», since «أنت بطل» would make the child a boy; its other lines lose their "you" the same way: «03:59 حتى المكافأة», «هيا نفتح الصندوق!», «تم حل 1 من 1. مذهل!».
- «نادي الرياضيات» for "maths circle", against «حلقة الرياضيات», nearer the tradition the page tells of.
- Sasha as a girl, «سارة», and Kim as a girl, «رنا»: the English gives neither a gender, and Arabic verbs need one.
- Percentages after an Arabic word wherever the English allows, so that the sign always lands on the left.
- Claude's labels in English: Claude's interface has no Arabic. Its warning over a prompt is quoted in English, and its question about saving is described.
- OpenAI's Arabic help is machine-translated and writes «وضع المطور» both with the shadda and without; the site takes «وضع المطوّر».
- The consent screen's wording of the Drive permission comes from Google's Marketplace, not from the screen itself.
- Saidan's «ابتكار خطة» and «مراجعة الحل» for Pólya's steps, and «البحث عن الحل» as the book's title rather than the newer «كيف تحل مسألة».
- The one Arabic edition found, «الرياضيات المسلية», for "Living Mathematics"; every other book keeps its English title.
- «ت {year}» for the dagger before a death year.
- The duals of the counted nouns on "Research", «ان» for some and «ين» for others, and the sentences worded to give each slot its case (see "The counted nouns of "Research""). The forms after 1 and after 0 follow the card's «1 مرة» and «0 مرة» (`progress.times`) rather than «مرة واحدة».
