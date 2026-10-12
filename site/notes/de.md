# German (`de`)

The German half of [the brief](README.md): the rules the German texts follow on top of it, and the words they use. The card's German dictionary, [`web/locales/de.json`](../../web/locales/de.json), comes first wherever it has a word; this file adds only what the site needs beyond it.

## Register and variant

- **One German, the card's.** Standard German as Germany and Austria write it, with «ß», as the card writes «Schließen» and «Größten»; no Swiss «ss» and no Swiss guillemets. Where the variants split, the word Germany uses: «Januar», «Samstag». There is no separate Austrian or Swiss text.
- **The reader is «Sie»**, as the card speaks to the parent: «Bitten Sie im Chat um eine Aufgabe» (`waiting.ask_in_chat`), «Löschen Sie die Datei und leeren Sie den Papierkorb in Google Drive» (`data.delete_how`). «Sie», «Ihr» and «Ihnen» take capitals; a plural reader is «Sie» too.
- **The child is «du»**, as the card speaks to the child: «Wähle eine Antwort» (`task.pick`), «Wünsch dir etwas für morgen» (`waiting.limit`). «du», «dein» and «dir» stay lower case. That covers the quotes under «Was Sie sagen können», the questions of a tip, Pólya's questions and the child's own lines.
- **What a parent types to the chat is «du»** too, as people write to a chat: «Gib mir eine Aufgabe zum Thema „Ritter und Lügner“», «Eine neue Aufgabe, bitte», «Lass uns eine MathTrail-Aufgabe machen.»
- **MathTrail, Claude and ChatGPT take no article**, as the card writes «MathTrail gibt einem Kind …» and «MathTrail selbst speichert nichts». Where English would put a pronoun for MathTrail, the name comes again, as in the card's «Danach kann MathTrail auf nichts mehr zugreifen» (`data.access_how`), or «der Dienst». Google Drive is «Ihr Google Drive», as the card writes it.
- **No gender shows, the child's or the parent's.** German verbs never show it, and neither do adjectives and participles after «sein» or «werden»: «Ihr Kind ist fertig», «wenn es festhängt», «um sicher zu sein», «allein» are all safe, and so are «zusammen» and «gemeinsam» for the reader and the child together, which the Spanish had to work around. What shows gender in German is nouns of persons («Schüler», «Gewinner», «Lügner», «Champion»), adjectives and ordinals made into nouns («der Kleine», «eine Neue», «Erster», «Erste»), and the pronouns «er», «sie», «sein», «ihr» for a person. So:
  - "your child" is «Ihr Kind»; after that «es» and «sein», which agree with the neuter «das Kind», as the card writes «was es in der Schule noch nicht hatte» (`first_run.fill`), or the noun again. Never «er», «sie» or «er oder sie».
  - No noun of persons for the child, and no ordinal made into one: «wer löst», not «der Löser»; «an erster Stelle», not «Erster» or «Erste». In the make-believe app on "Why" the cheer is «Du bist ein Star!!!»: «Star» is said of anyone, while «Champion» would make the child a boy.
  - Children in general are «Kinder», as the card says «Kind»; counted on "Research" too, since «Kind» is neuter and no count of children shows a gender. The Spanish detour to «estudiantes» is not needed.
  - **A group inside a task** is «Kinder» where the English says children or pupils («29 Kinder fahren auf einen Ausflug», «An einer Schule lernen 300 Kinder»), and «Jungen» and «Mädchen» where the task counts them apart. Players are «wer anfängt», «wer als Zweites zieht», «das Gegenüber» and «wer am Zug ist», never «der erste Spieler» or «der Gegner». An unnamed person takes the gender the English gives («eine Köchin», «ein Bäcker»); where the English gives none, «eine Person» or «jemand»: the buyers of the eggs are «die erste Person», «die zweite», «die dritte». Other groups are written as German textbooks write them, an epicene word first («Gäste», «Leute»); «Holzfäller», and the islanders, who are letters, take the generic masculine: «A ist ein Lügner». The rule about gender is about the reader and the reader's child.
  - "The parent" in the third person is the card's «Elternteil» and «Eltern», and «für Eltern» on a label (`profile.label`). The one who signs in, on the technical pages, is «die erwachsene Person» (card `first_run.lead`); the tutor is «die Lehrkraft» (card), teachers «Lehrkräfte». Ownership reads best as the family's: «das Google Drive der Familie», «der Chat der Familie». Readers are «Fachleute» or «technisch Interessierte»; developers «Entwicklerinnen und Entwickler», as Google's German help names them.
  - The people of the examples have a gender, as in the English: «Dora kam nicht als Erste ins Ziel».
- **An example's question is a task for the child**, as a task on the card is, so it speaks to the child with «du»: «Finde es heraus, ohne die großen Zahlen zu multiplizieren», «Wie viele …?». The words around it speak to the reader. A step the English writes as an order to the reader is an infinitive, as the card's advice is written (`advice.off_by_one`: «Vor der Antwort eine kurze Skizze zeichnen …»): «Die erste Angabe umdrehen: …», «Die 6 Haltestellen als Punkte zeichnen.» A tip for home speaks «Sie»: «Hängen Sie einen Kalender auf …».
- **What a topic teaches**, `hero.teaches`, goes on from its label «Was man dabei lernt», so it is an infinitive group with no «zu», its first word capitalised, as German writes learning goals: «Hinweise zu einem Bild verbinden und einen Vergleich in beide Richtungen lesen», «Schritt für Schritt schließen und sich selbst prüfen». `hero.method`, after the label «Der Kniff», is an infinitive group the same way: «Eine Linie zeichnen und alle darauf einordnen».
- **Sentences every topic's page shares** read the same on every page: `method.lead` ends «Die Beispiele unten reichen von ganz einfach bis Olympiade-Niveau.», and `description` ends «…, wie man solche Aufgaben löst, wo Kinder Fehler machen und wie Sie zu Hause helfen können.» A description opens with the topic's name: «Worum es bei „Reihenfolgen“ geht, wie man solche Aufgaben löst, …», and keeps every fact the English one adds to it: «Worum es bei „Figuren auf Karopapier“ geht, also um Flächeninhalt und Umfang, …».
- **The card's words, exactly.** A topic's name written out in a sentence is the card's, in „…“ with its capital and uninflected, as the card writes „{topic}“ in `topic_choice.ask`: «zum Thema „Systematisches Probieren“», «Aufgaben wie „Ritter und Lügner“». Unquoted it stands only where it is an ordinary noun the sentence needs no other form of: «das Schubfachprinzip», «die Prozentrechnung». Never a paraphrase, though the English sometimes paraphrases ("putting things in order"), and never an inflected name: «beim systematischen Probieren» is a paraphrase.

## Typography

- **Quotes:** „…“, the opening mark U+201E and the closing one U+201C, as the card writes «„Profil & Fortschritt“»; a quote inside a quote ‚…‘ (U+201A, U+2018). The English “…” become „…“: the German closing mark is the English opening one.
- **A sentence that ends with a quote** keeps the quote's own end mark inside it and adds no period after a ? or !: «Fragen Sie: „Sechs wer?“». A quoted statement ends with its period inside the quote: «„Zeig es auf dem Streifen.“». A `say:` that is the quote alone is written so: «„Welchen Hinweis hast du noch nicht benutzt?“».
- **No space before** `:` `?` `!` `;`, as the card writes «Thema:» and «Welches Thema als Nächstes?». After a colon a whole sentence opens with a capital, as the card writes «Nur eine Angabe: Eine Änderung wirkt sich auf keine Wertung aus.» (`profile.grade_note`); a word or a list after it starts lower case unless it is a noun.
- **Ellipsis:** one character, «…», set close to the word before it, as the card writes «Wird geprüft…» and «Thema wird gespeichert…»; in a sum it stands spaced, as the English writes it: «1 + 2 + … + 10».
- **Dashes:** the English spaced dash becomes the German spaced en dash « – » (U+2013), as the card writes «Nicht gesendet – erneut versuchen», or a colon. Ranges take the en dash with no spaces: «1–6», «70–85 %», «9.–19. Jahrhundert»; in prose, «bis»: «von 1 bis 6». The hyphen joins compounds with a name, a letter or a digit: «Mathe-Olympiade», «MathTrail-Profil», «Google-Konto», «MCP-Server», «5-Liter-Kanne».
- **Ordinals:** a digit and a period: «3.», «21.», «2.». In an example — its question, steps, note and answer — an ordinal the English writes in digits stays in digits: «der 9. Bonbon», «das 5. Feld», «nach dem 5. Sprung»; one the English writes in words stays in words: «der siebte Wurf». Elsewhere, words up to twelve («das dritte Mal»), then digits: «die 23. Flagge»; but an ordinal the English writes in digits in a topic's tip or card keeps its digits, as the numbers of the maths do, so that one sentence never mixes the two: «Wer macht den 10. Zug? Und wer den 15.?». In a drawing, digits: «21., 22., 23.» for the flags; the badge "2nd" is «2.». An example's answer keeps in digits every number its solver proves.
- **Dates** take an ordinal with its period, as `Intl.DateTimeFormat("de")` writes them: «3. Mai», «vom 3. bis 10. Mai», «am 25.», «am 2., 9., 16., 23. und 30.». Weekdays and months are nouns and take capitals: «Montag», «Mai». A weekday in a drawing or a badge is short, as Intl writes it and the site's own drawings of weekdays already do, with no period: «Mo», «Di», «Mi». The date on the legal texts reads «Stand: 9. Oktober 2026.».
- **Centuries** take Arabic ordinals, as German writes them, never Roman numerals: «im 5. Jahrhundert», «9.–19. Jahrhundert», «Ende des 19. Jahrhunderts»; decades are «die 1950er-Jahre».
- **Times** as the English writes them: «4:25», «15:10», «0:00», the bait's «03:59». No «Uhr» follows a time in an example: «um 4:25 Uhr» would put the swimming lesson before dawn. «Uhr» stays for "o'clock": «3 Uhr und 5 Uhr». A length of time is «4 h 30 min» in a drawing, the unit symbols German primary schools write, and «4 Stunden 30 Minuten» in prose.
- **Numbers outside the maths** as `Intl.NumberFormat("de")` writes them: «12.450», the bait's coins, since German groups from four digits; «9,99»; «0,2», so the formula reads «P = 0,2 + 0,8·σ(θ + δ − β)»; «19 Millionen». Years are never grouped: «2026». A version keeps its form: «MCP 2026-07-28».
- **Numbers in the maths** — the examples, the tables of a topic's page, and the sums of its cards and tips — keep the English digits and no thousands separator, as German school maths writes a four-digit number: «2026 × 2026», «1111», the arithmetic page's «10, 100 oder 1000» and its table's «1900», the percentages page's tip «von 1200». A decimal takes the comma, as the Russian and the Spanish write it: «0,5 + ½ = 1», «1,5», «3,5».
- **Percent** takes a space before the sign, as DIN 5008 and Duden ask and as Intl writes the card's own percentages: «70–85 %». That space is the no-break space U+00A0 Intl writes, so the number never wraps away from its sign. Inside an example, and everywhere on the page of percentages, whose drawings write «20%» beside the words, it is written as the English writes it: «20%».
- **Money** appears only in the make-believe app on "Why": «9,99 €» for the hat and «0 €» for MathTrail, the symbol after the number and the no-break space U+00A0 between them, as `new Intl.NumberFormat("de", {style: "currency", currency: "EUR"})` writes euros.
- **Units** are symbols after a space, with no point: «12 m», «200 kg», «5 l», «40 min», «4 h», and «7 d», the calendar drawings' day, DIN 1301's symbol; Intl's «7 Tg.» is wider. In prose, «Meter», «Liter», «Tage».
- **Signs:** × for multiplication, as the English and the site's drawings write it, though German schools write «·»; − (U+2212) for minus, never a hyphen; · in the formula. Division as each English page writes it: ÷ on the home page, "Why", "Techniques" and "Research"; : on the topic pages that use it (gaps, calendar, percentages, ratios, parts, arithmetic). German schools write «:», but the card's solution picture on the home page and on "Why" is drawn from data every language shares and writes «12 ÷ 3 + 1 = 5», so a «12 : 3» beside it would give one division two signs. Fractions as the English: «3/4», «½».
- **Capitals:** every noun takes a capital, and titles and headings are otherwise in sentence case: «So helfen Sie zu Hause», «Die Regeln der Linie». A named contest keeps its capitals: «Internationale Mathematik-Olympiade», «Mathematik-Olympiade»; «Olympiade» alone is a noun like any other.
- **The card questions on the home page and on "Why"** write no lone Latin capital inside a sentence that the English does not: the picture check reads one there as a label of the picture. German has no one-letter words, and its nouns are whole words, so the rule forbids only a letter written alone, such as «M» for «Meter».
- **Apostrophes** are rare in German: «Galts Fähigkeiten» and «Omas Uhr» take none. Only a name ending in -s, -x or -z takes one in the genitive («Max’ Pilze»), and a phrase without it reads better: «die Pilze von Max».
- **YAML:** a text that holds `: ` or begins with `*` is quoted, and since a German colon is often followed by a whole sentence, quoting is frequent; „ and “ need no quotes. The first comment of each file says "in German" where the English says "in English".
- **Titles of works** — a book, a painting, a treatise — go in „…“ where the English writes them plain, as the Russian does: „Die Schule von Athen“, „Elemente“, „Neun Kapitel der Rechenkunst“. The emphasis of a text is the English one's, so Pólya's book keeps its `*…*`.

## Platforms' labels

### Claude

Claude's interface speaks German, one of its display languages [D], and its German help speaks «Sie», as the site does: so the site quotes the German labels, not the English ones the Russian keeps. Both the help and the interface mix «Konnektoren», «Connectoren» and «Connectors». The site's own words say «Konnektor», as the interface names its settings section, and every quoted label is copied as the interface shows it, or as the help quotes it where the interface was not seen, so «Benutzerdefinierten Connector hinzufügen» keeps its «Connector».

- [A] «Erste Schritte mit benutzerdefinierten Connectoren mit Remote MCP», https://support.claude.com/de/articles/11175166-getting-started-with-custom-connectors-using-remote-mcp
- [B] «Verwenden Sie Konnektoren, um die Fähigkeiten von Claude zu erweitern», https://support.claude.com/de/articles/11176164-use-connectors-to-extend-claude-s-capabilities
- [C] Claude's settings in German, described from a screenshot in a German guide of May 2026: a section «Konnektoren», the banner «Konnektoren wurden nach Anpassen verschoben», the buttons «Verbinden» and «Benutzerdefinierten Connector hinzufügen», https://okf.das-programm.io/kb/mcp-claude-verbinden
- [D] «How to use Claude in your preferred language», which lists German among the display languages, https://support.claude.com/en/articles/10769299-how-to-use-claude-in-your-preferred-language

| English on the site | German | Where the help says it |
|---|---|---|
| Customize → Connectors | Anpassen → Konnektoren | B: «Navigieren Sie zu Anpassen > Konnektoren»; C names the section «Konnektoren» and the page «Anpassen». A writes «Anpassen > Connectoren» once and «Anpassen > Connectors» once. The site keeps its own arrow. |
| + | + Hinzufügen | A: «Klicken Sie auf „+ Hinzufügen“ und dann auf „Benutzerdefinierten Connector hinzufügen“» |
| Add custom connector | Benutzerdefinierten Connector hinzufügen | A and C |
| Name it MathTrail | Geben Sie als Namen MathTrail ein | A: «Geben Sie einen Namen für den Connector ein»; the field's label is not quoted |
| the address | die Remote-MCP-Server-URL; on the site, «die Adresse» | A: «Fügen Sie die Remote-MCP-Server-URL Ihres Connectors hinzu» |
| Continue, and Continue again | Weiter, und noch einmal Weiter | A quotes «Weiter» after the URL and again after the authentication settings |
| the sign-in settings Claude detects | die Einstellungen zur Authentifizierung, die Claude erkannt hat | A: «Überprüfen Sie die Authentifizierungseinstellungen, die Claude erkannt hat»; the site parts the compound, whose 31 letters make the home page scroll sideways on a phone 320 px wide |
| Add | Hinzufügen | A: «Schließen Sie die Konfiguration Ihres Connectors ab, indem Sie auf „Hinzufügen“ klicken» |
| Connect | Verbinden | B: «Klicken Sie auf „Verbinden“ oder „Installieren“»; C |
| + → Connectors, in a chat | + → Konnektoren | B, its steps from a chat: «Bewegen Sie den Mauszeiger über „Konnektoren“»; A and B's section on using a service write «Connectors» there. The chat's menu itself was not seen. |
| the free plan | der Free-Plan | A: «Free-Benutzer sind auf einen benutzerdefinierten Connector beschränkt»; «Benutzer der Pläne Free, Pro, Max, Team und Enterprise» |
| Always allow | Immer zulassen | A, its section on security; B, on actions |
| Sign in now · Sign in when needed · No sign in | Jetzt anmelden · not in the German help | A quotes «Jetzt anmelden» under «Authentifizierung» and stops there; not on the site yet |
| Use Claude's published identity · Register automatically · Use your own OAuth client | Claudeidentität verwenden (empfohlen) · not in the German help | A, under «OAuth-Client», quotes the first only; not on the site yet |
| "Use caution before running this prompt" | not found in Claude's German help, nor quoted in German anywhere on the web | the text describes it: «unter einem Hinweis, der zur Vorsicht mahnt» |
| Claude asks whether to allow MathTrail to save | the question is not quoted in Claude's German help; only the button «Immer zulassen» is | the text describes it: «Claude fragt, ob MathTrail das Profil speichern darf» |

### ChatGPT

| English on the site | German | Source |
|---|---|---|
| developer mode | **der Entwicklermodus** | https://help.openai.com/de-de/articles/12584461-entwicklermodus-und-mcp-apps-in-chatgpt-beta — «Entwicklermodus und MCP-Apps in ChatGPT»; the switch is under «Einstellungen → Apps → Erweiterte Einstellungen», and an app is made under «Apps → Erstellen» |
| ChatGPT's directory | das Verzeichnis von ChatGPT | OpenAI's German help now calls it the «Plugin-Verzeichnis», reached through «Plugins» in the sidebar: https://help.openai.com/de-de/articles/20001256 («Plugins in ChatGPT»). The generic phrase stays true. |
| plan, of ChatGPT | Tarif | OpenAI's German help: «deinem ChatGPT-Tarif», «tarifübergreifend» (https://help.openai.com/de-de/articles/11487775-apps-in-chatgpt); Claude's says «Plan» |
| Plus, Pro, Business, Enterprise, Education | the same | names of plans |

OpenAI marks its German articles as machine-translated, and they speak «du»; the site wraps their labels in «Sie»: «Aktivieren Sie den Entwicklermodus unter …». help.openai.com refuses fetchers; these labels were read through a text-only reader of the same addresses.

### Google

Google's German help speaks «Sie», as the site does.

| English on the site | German | Source |
|---|---|---|
| Google Drive | Google Drive; «in Ihrem Google Drive», as the card writes it; My Drive is «Meine Ablage» | https://support.google.com/drive/answer/2409045?hl=de |
| Drive's bin | der Papierkorb in Google Drive (card); the button «Papierkorb leeren»; files go for good after 30 days | https://support.google.com/drive/answer/2375102?hl=de, «Dateien in Google Drive löschen» |
| version history | der Versionsverlauf; for a file such as the profile, the menu «Versionen verwalten» and «Für immer behalten» | https://support.google.com/drive/answer/2409045?hl=de, «Aktivitäten und Dateiversionen prüfen»; its summary once says «Niemals löschen», its steps «Für immer behalten» |
| linked apps, third-party access | die Seite mit den verknüpften Apps in Ihrem Google-Konto (myaccount.google.com/linkedapps); the filter «Zugriff auf Ihr Google-Konto»; «Details ansehen» → «Zugriff entfernen» → «Bestätigen» | https://support.google.com/accounts/answer/13533235?hl=de, «Verknüpfungen zwischen Ihrem Google-Konto und Apps von anderen Entwicklerinnen und Entwicklern verwalten». The card's «Verbindungen zu Drittanbieter-Apps» is Google's wording too: https://support.google.com/accounts/answer/3466521?hl=de. "Delete all connections" is in neither article, nor in the English one. |
| the box for Google Drive on the consent screen | das Kästchen für Google Drive; ticking it is «ein Häkchen setzen», as the card's «das Häkchen entfernen»; Google words the permission «Nur die Google Drive-Dateien abrufen, bearbeiten, erstellen und löschen, die Sie mit dieser App verwenden» | as Google lists it in German in its Workspace Marketplace: https://workspace.google.com/marketplace/app/docs_sheets_editor_for_drive/703230099925?hl=de. Not seen on the live consent screen, and the English never quotes it. |
| Sign in with Google | in prose, «mit Google anmelden»; Google's own name for the button and the feature is «Über Google anmelden» | https://developers.google.com/identity/siwg/features?hl=de |
| consent screen | Zustimmungsbildschirm; «OAuth-Zustimmungsbildschirm» in Google's docs | https://developers.google.com/identity/protocols/oauth2?hl=de |
| Enter, the key | Eingabetaste | https://support.google.com/chrome/answer/157179?hl=de |
| Google account | Google-Konto | card and Google |

MathTrail's own sign-in pages speak English, so „Tick the box for Google Drive“ and **Back to Google** stay English.

## Glossary

"card" marks the card's own word, from `web/locales/de.json`.

### Product and lesson

| English | German | Note |
|---|---|---|
| task | Aufgabe | card, `task.label` |
| problem (olympiad) | Aufgabe | the same word as task: German maths calls a problem «Aufgabe» |
| puzzle | Rätsel, Knobelaufgabe | Perelman's German editions say «Denkaufgabe» |
| school sum | Rechenaufgabe | "A school sum" is «Eine Aufgabe aus der Schule»; never «Schulaufgabe», which in Bavaria is a written test |
| topic | Thema | card |
| group of topics | Gruppe | card, `group.*` exactly |
| trap | Falle | card, «Die Falle» |
| the catch (not a trap) | der Haken, die Schwierigkeit | never «Falle» outside the card's sense |
| technique | Technik | «Problemlösetechniken» |
| the main move (`topic.method`) | der Kniff | |
| hint | Tipp | card; never for "clue" |
| clue, a given of a task | Hinweis, Angabe | keeps «Tipp» for the card's hint: «Jeder Hinweis ist ein Pfeil» |
| option | Antwortmöglichkeit, Antwort | the card lists the options under «Antworten» |
| answer | Antwort | card |
| right, wrong answer | richtige, falsche Antwort | card: «Richtig!», «Nicht ganz» |
| solution | Lösung; the card's section «Lösungsweg» | card |
| step | Schritt | |
| question, the text of a task | Aufgabe, Aufgabentext | card: «Eine Zahl aus der Aufgabe als Antwort genommen» |
| question, what is asked | Frage | card: «Eine andere Frage beantwortet» |
| picture of a task | Bild | card `picture.*`; «Zeichnung» for the site's drawings, «Foto» for a photograph |
| label of a picture | Beschriftung | |
| grade | Klasse | card `profile.grade` |
| Grade 3 | Klasse 3 | card `child.grade` |
| grades 1 to 6 | die Klassen 1 bis 6 | card `first_run.lead` |
| grades 3–4, two slots | Klassen {from}–{to} | card `progress.grades_ranks` |
| level | Niveau, Stufe | |
| difficulty | Schwierigkeit | card |
| rank | Rang | card, «Rang {rank}»: a number, never a title |
| rating | Wertung | card: «Wertung in diesem Thema», «Gesamtwertung» |
| progress | Fortschritt | card, «Profil und Fortschritt» |
| mastered | gemeistert | card |
| review, the card's section | Auswertung | card |
| how the answer went | die Besprechung der Antwort | card `result.label` |
| trial series | Proberunde | card |
| skipped | übersprungen | card |
| profile | Profil | card |
| pseudonym | Pseudonym | card |
| interests | Interessen | card |
| skills | Fähigkeiten | card `problem.not_in_catalog` |
| lesson | Lektion | card «Sprache der Lektionen», quoted as it is; the card also says «Übungen» in `progress.next_chosen` |
| coach | Trainer | card, `topic_choice.coach_short`; the Coach product too |
| companion, in the coach | Begleitfigur | a character the child picks: «die Begleitfigur», never «der Begleiter» |
| parent | Elternteil; Eltern; «für Eltern» on a label | card |
| tutor | Lehrkraft | card `first_run.lead` |
| adult | erwachsene Person | card |
| child, your child | das Kind; Ihr Kind | card; then «es», «sein» |
| children | Kinder | the plural of the card's «Kind» |
| chat | der Chat | card |
| card | Karte | card |
| connector, custom connector | Konnektor; benutzerdefinierter Connector | as Claude's German interface writes each |
| app | App | card |
| sign in | sich anmelden | card |
| sign-in | Anmeldung | |
| account | Konto | card, «Google-Konto» |
| Drive file, the profile's file | Datei in Google Drive, Profildatei | card |
| folder | Ordner | card |
| bin | Papierkorb | card and Google |
| daily limit | Tageslimit | card |
| Another task | Neue Aufgabe | card's button, quoted exactly; not «Andere Aufgabe» |
| Make a wish for tomorrow | Wünsch dir etwas für morgen | card, quoted exactly |
| Profile & progress | Profil & Fortschritt | card |
| Edit | Ändern | card |
| the bulb, the hint's button | die Glühbirne | |
| the two arrows | die zwei Pfeile | |
| tap | tippen | |
| press, click | klicken; «drücken» for a key | as Claude's help: «Klicken Sie auf …» |
| Enter, the key | Eingabetaste | as Google's German help names it |
| free, of price | kostenlos | the chip `hero.free` |
| Free, the plan | Free | «der Free-Plan»: a plan's name |
| plan | Plan for Claude, Tarif for ChatGPT | as each one's German help says it |
| open source | Open Source | the chip `hero.open` |
| in development | In Entwicklung | |
| beta | Beta | |
| soon | demnächst | |
| olympiad maths | Olympiade-Mathematik | |
| olympiad-style task | Aufgabe im Stil der Mathe-Olympiade | card `first_run.lead`; «Olympiade-Aufgabe» for short, as the card's `progress.grades_note` |
| maths circle | Mathezirkel | «Mathe-AG» where a school's own club is meant |
| school | Schule | card |
| school mark | Note, Schulnote | card; so the profile's notes are «Notizen» |
| homework | Hausaufgaben | |

### The site's sections and headings

| English | German | Note |
|---|---|---|
| Topics (`nav.topics`) | Themen | 6 characters, 85 px as an entry; card `progress.topics` |
| Techniques (`nav.techniques`) | Techniken | 9, 100 px |
| Why (`nav.why`) | Warum | 5, 77 px; «Wozu», 69 px, is a later cut |
| Coach (`nav.coach`) | Trainer | 7, beside the tag «Beta»; card `topic_choice.coach_short` |
| Research (`nav.research`) | Studie | 6, 66 px in the technical pages' capsule: «Forschung», 96 px, took the row past the Russian. The page's title is «Studie» too, as a page whose English title is its menu entry keeps the two the same, and the footer shows it |
| Service (`nav.service`) | Dienst | 6, 66 px: German's word for a running service, as in «Google-Dienste»; «Service», 73 px, reads as customer service in a German menu |
| About (`nav.about`) | Team | 4, 65 px, and the page's title too: «Über uns», 93 px, took the row past the Russian |
| Add to Claude (`nav.add`) | Zu Claude hinzufügen | 20 characters, 197 px as the header's button: the form of «Zu Chrome hinzufügen» in Google's German help (https://support.google.com/chrome_webstore/answer/2664769?hl=de), with Claude's own verb |
| Technical pages (`nav.technical`) | Technische Seiten | said by a screen reader |
| Pages · Menu · Language (`nav.pages`, `nav.menu`, `nav.language`) | Seiten der Website · Menü · Sprache | said by a screen reader |
| Help · Privacy policy · Terms of use | Hilfe · Datenschutzerklärung · Nutzungsbedingungen | |
| How it is solved (`topic.solving`) | Lösungsweg | and every `method.title` reads «Lösungsweg: …» |
| Where children go wrong | Wo Kinder Fehler machen | |
| What to say | Was Sie sagen können | the reader's words to the child |
| How to help at home | So helfen Sie zu Hause | |
| Learn first | Vorher lernen | |
| Opens the way to | Führt weiter zu | |
| Builds on · Opens | Setzt voraus · Führt weiter zu | the topics page; «baut auf … auf» cannot take a list after a colon |
| Ask for a task on this topic | Bitten Sie um eine Aufgabe zu diesem Thema | as the card's `waiting.ask_in_chat` |
| Show · Hide the solution | Lösung zeigen · Lösung ausblenden | card: «Tipp ausblenden» |
| Example 2 | Beispiel 2 | |
| The main move · What it teaches | Der Kniff · Was man dabei lernt | the topic's first screen |
| Why it matters | Warum das wichtig ist | |
| How to solve: … | Lösungsweg: … | |
| The rules of … | Die Regeln … | «Die Regeln der Linie», «Die Regeln der Insel» |
| Check: | Probe: | the word German schools check a result with |
| When it helps: | Wann es hilft: | |
| In the topics: | In den Themen: | |
| You are here · On this page · All topics | Sie sind hier · Auf dieser Seite · Alle Themen | |
| More → | Mehr dazu → | |
| Soon | Demnächst | |
| Honestly | Ehrlich gesagt | |
| See a lesson | Eine Lektion ansehen | |
| See the topics → · See the coach → | Zu den Themen → · Zum Trainer → | |
| For technical readers | Für Fachleute | no gendered «Leser» |
| You, the parent's line on "Why" · C, the child's | Sie · K | «K steht für Ihr Kind» |
| The adult · The model, in the chat's illustration | Erwachsene Person · Modell | |
| Your chat | Ihr Chat | |
| Step 1, of the topics' map | Stufe 1 | |
| Foundations · Rules · Hard problems | Grundlagen · Regeln · Schwere Aufgaben | |
| First · Topic · Then, the map's key | Vorher · Thema · Danach | |

The menu's row was measured in the site's own font and sizes (Onest, 15 px, weight 500; the header's button as the stylesheet draws it), in Chromium and WebKit from the pinned Playwright image, beside the Spanish row: a language's narrowest fit is its row's width plus a constant, which gives back the Spanish 1157 and 1197 px exactly. With the plain German words, «Forschung», «Über uns» and «Dienst», the row stays whole only down to about 1200 px in Chromium and 1182 px in WebKit (1207 and 1190 px with «Service»): wider than the Russian's 1163 and 1149 px, so wider than every language before. One of the two cuts alone is not enough (about 1170 px either way). With «Studie» and «Team» `just site-fold` measured the row holding down to 1142 px in Chromium and 1127 px in WebKit, narrower than the Spanish row as it ships, 1157 and 1143 px, and the menu still folds below 1230 px.

Should the row ever need more room, the next cuts, in this order: «Warum» → «Wozu» (8 px), «Team» → «Wir» (13 px), and «Zu Claude hinzufügen» → «In Claude einbinden» (16 px), which changes the verb and comes last. «Trainer» and «Themen» are the card's words and stay.

### Maths and olympiad terms

| English | German | Note |
|---|---|---|
| olympiad | Olympiade | the German contest is «die Mathematik-Olympiade», «Mathe-Olympiade» for short (card) |
| International Mathematical Olympiad | Internationale Mathematik-Olympiade (IMO) | |
| reference task | Referenzaufgabe | |
| solver | Lösungsprogramm; «Löser» in a label or a drawing | first time, «ein Programm, das die Aufgabe löst» |
| brute force | Durchprobieren aller Fälle | «Brute Force» on the technical pages |
| worst case | ungünstigster Fall | card, `trap.best_case_not_worst` |
| good, bad luck | Glück, Pech | the worst luck is «das größte Pech» |
| to be sure, for certain | sicher, mit Sicherheit | «um sicher zu sein» shows no gender |
| pigeonhole principle | Schubfachprinzip | card |
| pigeons, holes | Tauben, Käfige | the drawings draw barred cages; "Pigeons and pigeonholes" is «Tauben und Käfige» |
| parity · even · odd | Parität · gerade · ungerade | card |
| alternation, taking turns | abwechselnde Muster, sich abwechseln | card |
| remainder · quotient | Rest · Quotient | card |
| divisibility | Teilbarkeit | card |
| divides by 9 | ist durch 9 teilbar | «teilen durch», throughout |
| divisibility rules | Teilbarkeitsregeln | card |
| digit · two-digit number | Ziffer · zweistellige Zahl | |
| tens, units digit | Zehnerziffer, Einerziffer | «Zehner», «Einer», as German schools say |
| enumeration | systematisches Probieren; «Abzählen» in the group's name | card: «Systematisches Probieren», «Zählen und Abzählen» |
| case · list | Fall · Liste | |
| systematic listing | systematisch auflisten | the technique "List systematically" |
| tree, the diagram | Baumdiagramm | |
| working backwards | Rückwärtsarbeiten, vom Ende her rechnen | Pólya's German edition: «Rückwärts arbeiten» |
| the assumption method | Rechnen mit einer Annahme | German school maths has no fixed name for it |
| proof by contradiction | Widerspruchsbeweis | Pólya's German edition: «indirekter Beweis» |
| suppose · assumption · contradiction | angenommen, … · Annahme · Widerspruch | |
| pattern | Muster | |
| symmetric strategy | Symmetriestrategie | |
| winning strategy | Gewinnstrategie | card |
| winning, losing position | Gewinnstellung, Verluststellung | |
| move, in a game | Zug | card |
| the first, the second player · the other player · the player to move | wer anfängt, wer als Zweites zieht · das Gegenüber · wer am Zug ist | no «Spieler»: "The second player." is «Wer als Zweites zieht.» |
| invariant | Invariante | |
| knights and liars | Ritter und Lügner | card |
| knight · liar | Ritter · Lügner | R and L in drawings. Duden gives «Ritterin» as the holder of an order, so a named knight is a boy |
| island · islander | Insel · Inselbewohner | the islanders are letters, in the generic masculine |
| statement | Aussage | card |
| ordering | Reihenfolgen | card |
| taller, shorter · the tallest | größer, kleiner · am größten | the drawings' «am größten», «am kleinsten» show no gender |
| number line | Zahlenstrahl | card |
| grid · squared paper | Kästchengitter · Karopapier | card |
| cell · side of a cell | Kästchen · Kästchenseite | |
| area · perimeter | Flächeninhalt · Umfang | card |
| ratio · in the ratio 3 to 4 | Verhältnis · im Verhältnis 3 zu 4 | card |
| sharing · share | Verteilung · Anteil | card |
| percentage · percent | Prozentsatz · Prozent | card |
| what a percentage is taken of | Grundwert | card `trap.percent_wrong_base`, the German school word |
| fraction · the whole · the part | Bruch · das Ganze · der Teil | card |
| a half, a third, a quarter | die Hälfte, ein Drittel, ein Viertel | |
| gaps and boundaries | Pfosten und Zwischenräume | card |
| gap, between posts · post · fence | Zwischenraum · Pfosten · Zaun | card |
| off by one | um eins daneben | card: «um eins falsch gezählt» |
| clock face · hands · the whole hour | Zifferblatt · Zeiger · die volle Stunde | card: «Ein Zifferblatt» |
| floor 1, of a house | Etage 1 | the number after the noun, as the English writes it; never «Erdgeschoss» or «1. Stock», which count from another floor |
| flight of stairs · step | Treppe · Stufe | |
| calendar and age | Kalender und Alter | card |
| pan balance · weighing · fake coin | Balkenwaage · Wägung · falsche Münze | card |
| pan · weights | Waagschale · Gewichte | |
| jug · pour | Kanne · umfüllen | card: «Gefäße zum Umfüllen» |
| state, of the jugs | Zustand | |
| overlapping groups | Überlappende Gruppen | card |
| arithmetic with a trick | Knifflige Rechenaufgaben | card |
| I thought of a number | Ich denke mir eine Zahl | |
| carrying | Übertrag: «ohne Übertrag» | |
| round number | runde Zahl | |
| match, the stick | Streichholz | |
| counter, on a board | Spielstein | card `picture.piles` |
| coins, a task's money | Münzen | |
| a special case | ein Spezialfall | |

### The technical pages

| English | German | Note |
|---|---|---|
| the service | der Dienst | as German privacy policies and Google's German help say it |
| token · access token · refresh token | Token · Zugriffstoken · Aktualisierungstoken | as Google's German docs |
| authorization code | Autorisierungscode | |
| seal · sealed · sealing | versiegeln · versiegelt · Versiegelung | |
| encrypted | verschlüsselt | |
| state · stateless | Zustand · zustandslos | |
| MCP server · tool · tool call | MCP-Server · Tool · Tool-Aufruf | as Claude's German help |
| chat host | Chat-Host | |
| consent screen | Zustimmungsbildschirm | Google |
| scope | Berechtigung | «die Berechtigung drive.file»; Google's docs say «Bereich» |
| client · client metadata document | Client · Client-Metadatendokument | |
| renewal · revocation | Erneuerung · Widerruf | |
| pace · ceiling | Tempo · Obergrenze | |
| Sign-in, a scenario's tab on "Service" · target, of the profile's size | Anmeldung · Ziel | «Login» for the tab, should «Anmeldung» wrap at 320 px |
| log · logging | Protokoll · Protokollierung | «Cloud Logging» stays |
| instance · revision · image | Instanz · Revision · Image | |
| key rotation · ciphertext · nonce | Schlüsselrotation · Geheimtext · Nonce | |
| fingerprint · near-duplicate | Fingerabdruck · Beinahe-Duplikat | |
| readability · self-check | Lesbarkeit · Selbstprüfung | |
| brief, for the model | Auftrag | |
| package · schema · template | Paket · Schema · Vorlage | |
| deterministic · the rule | deterministisch · die Regel | |
| binary | Binary: «ein Go-Binary» | |
| cookie · cache | Cookie · Cache | |
| graceful shutdown | geordnetes Herunterfahren | |
| lifetime, of a token | Gültigkeitsdauer | |
| Workload Identity Federation | Workload Identity-Föderation | as Google's German docs |
| student model | Lernermodell | the term of German research on tutoring systems; a compound, so its generic masculine stands |
| chance, of a right answer | Lösungswahrscheinlichkeit | the term of German psychometrics |
| corridor · interval · goal | Korridor · Intervall · Ziel | |
| the ceiling, an estimate that knows the child · oracle | die Obergrenze · Orakel | |
| bench · simulated children | Prüfstand · simulierte Kinder | |
| a child who learns, who stays put | ein Kind, das dazulernt; eines, das stehen bleibt | |
| lag · logit · points | Rückstand · Logit · Punkte | |
| mastery declared falsely | fälschlich als gemeistert erklärt | card: «Gemeistert» |
| live data, live numbers | Echtdaten, echte Zahlen | they come monthly, not in real time |
| range, of the chance promised | Bereich | |
| standard error · baseline · seed | Standardfehler · Vergleichswert · Seed | |
| Reached · On the edge · Not reached | Erreicht · Grenzwertig · Nicht erreicht | |
| paper · public domain | Studie · gemeinfrei | the page is «Studie», and so is its paper: «Die ganze Studie · PDF» |
| copyright | Urheberrecht | |
| counting name, of the logs | Zählname | the privacy policy's own term |
| commit · release | Commit · Release | borrowed terms |
| endpoint · claim, of a JWT | Endpunkt · Claim | |
| request, the model's · store · router · issuer | Anfrage · Speicher · Router · Aussteller | |
| deploy · build · client secret · sandboxed | Deployment · Build · Client-Secret · in einer Sandbox | |
| in the open, unsealed · timeout | im Klartext · Zeitlimit | |
| digest · runtime · PKCE verifier · state machine | Digest · Laufzeitumgebung · PKCE-Verifier · Zustandsautomat | |
| scenario · sketch · pinned · sliding, of a lifetime | Szenario · Skizze · festgelegt · gleitend | |
| canvas · tablet | Zeichenfläche · Tablet | the coach |
| GitHub issue · computer · mouse | Issue · Computer · Maus | |
| the profile's notes | Notizen | «Note» is a school mark |
| whisker, of an error bar · dashed | Fehlerbalken · gestrichelt | |
| item response theory · psychometrics | Item-Response-Theorie (IRT) · Psychometrie | |

## Pólya's four steps

The German edition is «Schule des Denkens. Vom Lösen mathematischer Probleme» (Francke, Bern, 1949, Sammlung Dalp 36, translated by Elisabeth Behnke; the 4th edition, Francke, Tübingen und Basel, 1995, which some catalogues date 1994, adds a preface by Peter Roquette). Its chapter «Vier Phasen der Arbeit» names the phases «Verstehen der Aufgabe», «Ausdenken eines Planes», «Ausführen des Planes» and «Rückschau», as the contents of the 4th edition show in the Deutsche Nationalbibliothek's scan: https://d-nb.info/94245054x/04. Its little dictionary of heuristics has «Rückwärts arbeiten», «Spezialisierung», «Symmetrie» and «Reductio ad absurdum und indirekter Beweis», and asks «Hast du alle Daten benutzt?»: the book speaks «du» to whoever solves, as the site's steps speak to the child. The page "Techniques" names its steps after the phases:

| English | German |
|---|---|
| *How to Solve It* (1945) | *Schule des Denkens* (1945) |
| Understand the problem | Die Aufgabe verstehen |
| Make a plan | Einen Plan ausdenken |
| Carry it out | Den Plan ausführen |
| Look back | Rückschau halten |

A technique's name is an infinitive group, as Pólya's steps are: «Die Aufgabe zeichnen», «Eine Tabelle anlegen», «Mit kleinen Zahlen beginnen», «Systematisch auflisten», «Vom Ende her rechnen», «Annehmen, alle seien gleich», «Ein Muster suchen», «Das Gegenteil annehmen», «Auf die Parität achten», «Züge spiegeln»; "Pigeons and pigeonholes" is «Tauben und Käfige». The names after "aka" are «Ein Spezialfall», «Rückwärtsarbeiten», «Rechnen mit einer Annahme», «Widerspruchsbeweis», «Das Schubfachprinzip» and «Eine Symmetriestrategie».

The steps' questions speak to the child: «Erzähl die Aufgabe mit deinen eigenen Worten nach. Was weißt du? Was ist gesucht?», with «gesucht», the word German maths lessons name the unknown by. The coach's shorter line, "Understand → make a plan → solve → check", is «Verstehen → Planen → Lösen → Prüfen», and the picture's "understand, plan, solve, check" on "Why" is «Verstehen, Planen, Lösen, Prüfen»: the steps' names, nouns in German, with «Prüfen», the card's verb for its checks.

## People's names

"Initial" marks a name whose first letter a drawing or a badge shows; within one drawing no two letters repeat.

| English | German | Where | Initial, badge |
|---|---|---|---|
| Galt | Galt | home, the cards' pseudonym: «Ausgewählt für Galt», «Galts Fähigkeiten» | kept, as the Russian and the Spanish keep it; an invented name shows no gender |
| Comet | Komet | "Why" and "Topics", the cards' pseudonym | as «Комета» and «Cometa»; a name, so never «der Komet» |
| C, "your child" on "Why" | K | `parent.explain.child`: «K steht für Ihr Kind» | the speaker's letter, K for «Kind» |
| Ann | Anna | ordering, parts, ratios, clocks, calendar | A, kept |
| Ben | Ben | ordering, parts, ratios, parity | B, kept: among the commonest boys' names in Germany |
| Kate | Klara | ordering, calendar | K, kept: German spells the name with a K, so `art.kate` stays «K» |
| Dan | David | ordering's line | name in a drawing |
| Lena | Lena | ordering's queue, gaps, ratios | L, kept |
| Max | Max | ordering's queue, ratios | M, kept |
| Nina | Nina | ordering's queue | N, kept |
| Oleg | Ole | ordering's queue | O, kept |
| Pavel | Paul | ordering's queue | P, kept, and the badge «P» |
| Ada | Alma | ordering's runners | A, kept |
| Bob | Bruno | ordering's runners | B, kept; not Ben, who is on the same page |
| Cid | Carl | ordering's runners | C, kept; «Carl», the spelling with C |
| Dot | Dora | ordering's runners | D, kept; a girl, as the English «her» |
| Eli | Emil | ordering's runners | E, kept |
| Fay | Frida | ordering's runners | F, kept, and the badge «F» |
| Anya | Anton | "Techniques": the pets and the knights | A, kept. A boy, since a named knight is «Ritter» |
| Borya | Bea | "Techniques": the pets and the knights | B, kept. A girl, so that the pair stays a boy and a girl: «Bea ist eine Lügnerin». The step that supposes her a knight says «Angenommen, Bea sagt die Wahrheit» |
| Vera | Vera | "Techniques", calendar | name in drawings |
| Tanya | Tilda | "Techniques", working backwards | |
| Petya | Moritz | "Techniques", working backwards | |
| Lea | Lea | enumeration | L, kept |
| Tom | Tom | enumeration, clocks, parts | T, kept |
| Masha | Mia | gaps, arithmetic | M, kept |
| Sasha | Mika | pigeonhole, the sweets | used for boys and girls alike, as Sasha is; the English gives no gender, and neither does the German |
| Gleb | Jonas | pigeonhole, the socks | |
| Kim | Kim | winning strategy | name in drawings: «Kim −3»; used for boys and girls in German too |
| Leo | Leo | winning strategy | name in drawings: «Leo −4» |
| Grandma, Granny | Oma | clocks, parts | in drawings: «Oma», «Omas Uhr» |
| Mum, Dad | Mama, Papa | calendar, ordering's tip, "About" | nouns, with capitals: «Mama», «Papa» |
| islanders A, B, C | A, B, C | knights | letters |
| knight K, liar L | R, L | knights, `thumb.*` | R for «Ritter», L for «Lügner»; no islander is R or L |
| children A, B, C · boxes A, B | the same | enumeration · ratios | letters |
| the cook (she) · the baker (he) | eine Köchin · ein Bäcker | weighing · parts | |
| the buyers · woodcutters · a farmer | die erste, zweite, dritte Person · Holzfäller · ein Bauer | parts · gaps · percentages | |

The people of history go by their German names: Euklid, König Ptolemaios, Proklos, die „Elemente“, Archimedes, Raffael, „Die Schule von Athen“, Konfuzius, die „sechs Künste“, Aryabhata, Pune, Taqi ad-Din, Istanbul, Luca Pacioli, Jacopo de' Barbari, al-Chwarizmi, Bagdad, „Neun Kapitel der Rechenkunst“, Aristoteles, Alexander der Große, Nikolai Bogdanow-Belski and his „Kopfrechnen. In der Volksschule von S. A. Ratschinski“, St. Louis, Tokio, der Konnō-Hachimangū-Schrein, Leningrad, Moskau, Ungarn, die USA, die UdSSR, George Pólya. A photograph's credit, «Yu. D. Korolev» or «Momotarou2012», stays as its source writes it; the archive is «Hauptarchiv Moskau», as the Spanish names it in Spanish.

## Grades

`topics.grades` is **«Klassen {range}»**: «Klassen 1–2», «Klassen 3–4», «Klassen 5–6», «Klassen 1–6»; so `topic.example` reads «Beispiel 2 · Klassen 3–4», and the label on "Techniques" «Aufgabe · Klassen 2–3».

- It is the card's own phrase for a span of grades, «Aufgaben der Klassen {from}–{to}» (`progress.grades_ranks`), and its «für die Klassen 1 bis 6» (`first_run.lead`), so a chip on the site and the card's progress read alike.
- German names a span of grades in the plural, as it writes «Seiten 3–5»; the card's singular «Klasse {grade}» (`child.grade`) names one grade.
- An ordinal does not fit the slot: «{range}. Klasse» would mark only the last number, «3–4. Klasse», where German writes «3./4. Klasse».

The alternative is «Klasse {range}», the way German workbooks label a span («Mathe Klasse 3–4»), in the singular of `child.grade`. A text with two slots writes the card's form: «Klassen {first}–{last}» on "Research", and «{count} · Klassen {first}–{last}» under its sources; the home page's lead, which has the one slot, says «für die Klassen {range}», and "Why" says «Bis Klasse {last}», one grade in the singular.

## Books on "Research"

| Key | The English page | German | How sure |
|---|---|---|---|
| perelman.firsthundred | For Young Mathematicians. The First Hundred Puzzles | the English title | no German edition found |
| perelman.secondhundred | The Second Hundred Puzzles | the English title | no German edition found |
| perelman.living | Living Mathematics | «Heitere Mathematik» (Der Kinderbuchverlag, Berlin, 1962 and later editions) | fairly sure it is this book: its chapters came out first as booklets of the same publisher, translated by H. Asemissen and illustrated by W. Riegenring, who also illustrates «Heitere Mathematik»: «Ein Frühstück mit Denkaufgaben» (1950), which opens with the squirrel of «Живая математика», «Mathematik im Spiel», «Denkaufgaben mit Zahlenriesen» and «Geometrische Denkaufgaben» (1950–1951). That the 1962 volume holds the whole book is not checked |
| perelman.arithmetic | Entertaining Arithmetic | «Unterhaltsame Arithmetik» (Volk und Wissen, Berlin, 1963) | fairly sure: German Wikipedia lists it, and its title is the Russian's word for word, as «Unterhaltsame Geometrie» and «Unterhaltsame Algebra» are; its contents are not checked |
| perelman.quick | Quick Reckoning | the English title | no German edition found |
| ignatyev.wit | In the Realm of Ingenuity, or Arithmetic for Everyone | the English title | no German edition found |
| dudeney.amusements | Amusements in Mathematics | the English title | no German edition of the whole book: Anaconda's «Das große Buch der Denk- und Knobelspiele» (2016) holds 116 of its puzzles beside the Canterbury ones, and the German translation online at janko.at keeps the English title |
| dudeney.canterbury | The Canterbury Puzzles | «Die Canterbury-Rätsel» (Anaconda, Köln, 2009) | sure |
| loyd.cyclopedia | Cyclopedia of Puzzles | the English title | no German edition found; Martin Gardner's selections, «Mathematische Rätsel und Spiele» (DuMont, 1978; 2005), are other books |
| carroll.tangled | A Tangled Tale | «Geschichten mit Knoten» (Insel, Frankfurt am Main, 1978, translated by Walter E. Richartz) | sure |
| carroll.pillow | Pillow Problems | the English title | no German edition found; German writing about the book calls it «Kopfkissenprobleme», which is no edition's title |

The authors' names: Jakow Perelman and Jemeljan Ignatjew, as German writes Russian names (the GDR editions print «J. I. Perelman»); Henry E. Dudeney, Sam Loyd and Lewis Carroll as they are.

## Places for the author's reading

- The menu: «Studie» for «Forschung» and «Team» for «Über uns», in the menu, the pages' titles and the footer alike. Without them the German row is the widest of all (about 1200 px against the Russian's 1163); with them it is the narrowest of the three. The other way out is the plain words everywhere and a fold some 40 px wider for every language.
- «Dienst» for "Service": literal, as «Servicio» and «Сервис» are, but unusual as a menu entry. The idiomatic pair is «Technik» for the service's page and «Methoden» for the techniques' (measured: about 1151 and 1135 px with «Studie» and «Team», still under the Russian), at the cost of «Methode» for "technique" across that page.
- `topics.grades`, «Klassen {range}», against «Klasse {range}».
- On "Techniques", Anya and Borya become Anton and Bea, a boy and a girl swapped, because «Ritterin» is not a story's knight in German; the other way is two boys, Anton and Bruno, which keeps every type word in the masculine.
- The pigeonholes are «Käfige», as the drawings draw them; «Fächer» would echo «Schubfachprinzip» instead.
- The spaces before «€» and «%» are U+00A0, as Intl writes them; the Spanish files have a plain space there. They look the same.
- Division stays ÷ where the English writes it, because the shared card picture writes «12 ÷ 3 + 1 = 5»; German schools write «:», as the Russian writes it in its own words.
- Claude's labels come from its German help and a German guide's description of the interface, which mix «Konnektoren», «Connectoren» and «Connectors»; a look at Claude's interface in German would settle the chat menu's word, here «Konnektoren». Its warning over a prompt and its question about saving are not in its German help, so the text describes them.
- The consent screen's wording of the Drive permission comes from Google's Marketplace, not from the screen itself.
- Google's account page is «verknüpfte Apps» on the site and «Verbindungen zu Drittanbieter-Apps» in the card's `data.access_how`: both are Google's words for the same page, in two of its articles.
- The books marked unsure: «Heitere Mathematik» for Living Mathematics, «Unterhaltsame Arithmetik», and the missing editions of the rest.
- «Der Kniff» for "The main move"; «Die Schlüsselidee» is the plainer alternative.
- The make-believe app's cheer is «Du bist ein Star!!!», since «Champion» would make the child a boy.
- «Lernermodell» for "student model"; «Lernstandsmodell» names no person at all.
- Players as «wer anfängt», «das Gegenüber», «wer am Zug ist», where German olympiad texts write «der erste Spieler».
- «u. a.» in `cite.many`, as German writes it, where the Spanish kept «et al.».
- "The assumption method" as «Rechnen mit einer Annahme»: German has no fixed name for it.
- The card names the topic of enumeration «Systematisches Probieren», "trying systematically"; the technique "List systematically" is «Systematisch auflisten», close to it but not the card's word.
