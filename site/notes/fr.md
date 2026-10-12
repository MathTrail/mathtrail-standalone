# French (`fr`)

The French half of [the brief](README.md): the rules the French texts follow on top of it, and the words they use. The card's French dictionary, [`web/locales/fr.json`](../../web/locales/fr.json), comes first wherever it has a word; this file adds only what the site needs beyond it.

A French text this file gives stands between « and », which are this file's quotation marks and no part of the text, and is written exactly as it is to be typed, its no-break spaces included; where the text itself holds guillemets, they carry theirs: «Le mot « pas »». A source's own words stand in "…", as the source writes them.

## Register and variant

- **One French, the card's.** Standard French as written in France, as the card's is: «Téléchargez», «le chat», «Annuler», and the traditional spelling with its circumflexes, as the card's «connaît». There is no separate text for Belgium, Switzerland, Québec or Africa. Where the countries' words part, the one all of them read wins: «le chat», not «le clavardage»; «adresse e-mail», as Google's French help writes it. The school's years, which every country names its own way (CP to CM2 in France, «1re année» in Québec, «1re primaire» in Belgium, 1P to 8P in Switzerland), are the card's «niveau» (see "Grades").
- **The reader is «vous»**, as the card speaks to the parent: «Demandez un problème dans le chat» (`waiting.ask_in_chat`), «Le chat vous dira ce qu’il faut modifier» (`profile.not_saved_detail`), «Remplissez ce champ.» (`problem.required`). Claude's and OpenAI's French help speak «vous» as well, so their wording goes into the site's sentences as it stands.
- **The child is «tu»**, as the card speaks to the child: «Choisis une réponse» (`task.pick`), «Fais un vœu pour demain» (`waiting.limit`), «Impossible de charger tes progrès. Reviens en arrière et réessaie.» (`progress.failed`). That covers the quotes under «Que lui dire», the questions of a tip, Pólya's questions and the child's own lines.
- **What a parent types to the chat is «tu»** too, as people write to a chat: «Un nouveau problème, s’il te plaît». A topic's `ask` line takes the card's own way of naming a topic in the chat (`topic_choice.ask`: «Un autre problème — sur le thème « {topic} »»), so all seventeen read alike: «Donne-moi un problème sur le thème « Mise en ordre »», «Donne-moi un problème sur le thème « Principe des tiroirs »».
- **No gender shows, the child's or the parent's.** French shows it in adjectives and past participles that agree with a person (sûr, sûre; prêt, prête; seul, seule; bloqué, bloquée; curieux, curieuse; content, contente), in the perfect tense of every verb conjugated with «être» and of every reflexive verb (il est arrivé, elle est arrivée; je me suis trompé, je me suis trompée), in nouns with two forms (champion, championne; joueur, joueuse; lecteur, lectrice), in the pronouns il, elle, celui, celle, and in un, une, le, la before a consonant. It does not show in «enfant» after «votre», «l’», «les» or «chaque», in a possessive (son, sa, ses agree with the thing owned: «son profil», «sa réponse»), in «lui» and «leur» («dites-lui»), in the present tense, in the perfect with «avoir» («a trouvé», «a choisi»), in the impersonal «on», or in «ensemble». So:
  - "your child" is «votre enfant», "the child" «l’enfant»; after that the text names «l’enfant» again, or turns to «on», a verbal noun or the child's answer, rather than writing «il» or «elle». "A child" in general is «chaque enfant» or «l’enfant».
  - No adjective or participle agrees with the child or the parent: «à coup sûr», «pour s’en assurer», never «pour être sûr»; «si votre enfant bloque», «en cas de blocage», not «s’il est bloqué»; «sans aide», not «seul»; «se trompe» or «a fait une erreur», not «s’est trompé»; «C’est bon !», not «Je suis prêt !». The reader is never described either: «Vous êtes ici» is safe, «Vous êtes sûr» is not.
  - The child's own lines: «J’ai encore fait une erreur…», «C’est bon ! La réponse, c’est cinq.» The parent's answer to the second: «Le plus beau, ce n’est pas la réponse, c’est le chemin pour la trouver.», never «comment tu y es arrivé».
  - A child's question to the parent is a verb: «Tu mens ?», not «Tu es menteur ?»; in the island game the parent «ment à chaque réponse».
  - Children in general are «les enfants», and an adjective after them takes the masculine plural a mixed group takes («les enfants sont curieux»), which speaks of no one child. Counted on "Research", «enfants» too. A group inside a task is the task's own: «24 enfants», «11 élèves».
  - "The parent" is «le parent», one noun for a mother and a father, as the card's «un parent»; «les parents», and «pour les parents» on a label (card). "A parent or a tutor" is the card's «un parent ou un enseignant». The one who signs in, on the technical pages, is «l’adulte» (card); ownership reads «le Drive du parent» or «le Drive de la famille».
  - "Together" is «ensemble», which shows nothing; never «tous les deux» for the reader and the child.
  - The bait app's cheer is «Tu es une star !!!», since «champion» would make the child a boy.
  - The people of the examples have a gender, as in the English: «Léna n’est pas la dernière». The islanders are letters, in the generic masculine: «A est un menteur», «A et B sont tous les deux menteurs».
- **An example's question is a task for the child**, as a task on the card is, so it speaks to the child with «tu»: «Calcule…», «Trouve-le sans multiplier les grands nombres.», «Combien y a-t-il de poteaux ?». The steps and the notes speak to the reader, with «on» and the reasoning's own «Supposons que…»: «On dessine les 6 arrêts comme des points.» A tip under «Comment aider à la maison» speaks to the reader with «vous»: «Écrivez V et M au crayon».
- **A trap told from the child's side** names nobody. The card's lines on the home page and on "Why" (`task.traps.*`, `task.trap`, `pillars.tiles.diagnosis.trap`) take «on» and the perfect with «avoir»: «On a compté les intervalles au lieu des poteaux.» A trap's `shows` speaks of «l’enfant» in the present: «L’enfant compte les arrêts au lieu des intervalles…».
- **What a topic teaches**, `hero.teaches`, goes on from its label «Ce qu’on y apprend», so it opens «À» and an infinitive, and a second verb after «et» takes its «à» too: «À compter les jours et les années sans les compter un par un», «À raisonner pas à pas et à vérifier son raisonnement». The main move, `hero.method`, is an infinitive: «Tracer une ligne et y placer tout le monde».
- **Sentences every topic's page shares** read the same on every page: `method.lead` ends «Les exemples ci-dessous vont du plus simple à un problème d’olympiade.», and `description` reads «Ce que sont les problèmes de…, comment on les résout, où les enfants se trompent et comment les aider à la maison.», its middle following the English: «Ce que sont les problèmes de mise en ordre, comment on les résout, où les enfants se trompent et comment les aider à la maison.»
- **The card's words, exactly.** A topic's name is the card's: the page's `title` as the card writes it («Mise en ordre»); in a sentence, either in « … » with its capital, as the card writes « {topic} », or in lower case without guillemets, with the article French needs: «la mise en ordre», «l’énumération des cas», «le principe des tiroirs», «les chevaliers et menteurs». Never a paraphrase, though the English sometimes paraphrases ("putting things in order", "counting all the cases").

## Typography

- **Quotes:** « … », with a no-break space inside each guillemet, as the card writes « {topic} »; a quote inside a quote “…”, and inside that ‘…’, with no spaces inside these two. The English “…” become « … ».
- **The no-break space, U+00A0.** The card puts U+00A0 NO-BREAK SPACE, never U+202F and never a plain space, before : ; ? ! and », and after «: «Exact !», «Quel thème ensuite ?», «Thème :», «Le mot « pas »». The site does the same everywhere: in prose, headings, drawings and quotes («Comment le résoudre : tracer la ligne», «Que faire ?»). It also stands before € and % where those take a space, and before an em dash (below). In the pages' YAML and the documents' Markdown it is the character itself, since YAML reads the escape `\u00a0` only inside double quotes and Markdown shows it as text; in the site's dictionary, `web/src/site/locales/fr.json`, it is the JSON escape `\u00a0`, as the card's `fr.json` writes it. To type it, write a plain space, then convert, then check; both commands take the files' paths, from the worktree's root:

  ```sh
  # A plain space before : ; ? ! » and after « becomes U+00A0. The YAML files' English comments are left alone.
  perl -CSD -i -pe 'unless ($ARGV =~ /\.yaml$/ && /^\s*#/) { s/ (?=[:;?!\x{BB}])/\x{A0}/g; s/\x{AB} /\x{AB}\x{A0}/g }' site/content/fr/index.yaml

  # Prints every line still missing a no-break space, or holding an ASCII apostrophe or a written-out \u00a0. Nothing printed means all is well.
  node -e '
  const fs = require("fs");
  for (const f of process.argv.slice(1)) fs.readFileSync(f, "utf8").split("\n").forEach((line, i) => {
    if (f.endsWith(".yaml") && /^\s*#/.test(line)) return;
    const text = line.replace(/^\s*(?:- )?"?[\w-]+"?:(?= |$)/, "");
    if (/ [:;?!»]|« |[^\s!?;][;?!»]|«[^\u00a0]|[^\s]:(?= )|\\u00a0|\x27/.test(text)) console.log(`${f}:${i + 1}: ${line}`);
  });' site/content/fr/index.yaml
  ```

  A division colon gets one too, which keeps the sign with its number: «45 : 9 = 5». Never break a line of a folded YAML block (`>-`) between a word and the : ; ? ! » after it, right after «, or inside «9,99 €» or «70–85 %»: the line break becomes a plain space.
- **Apostrophe:** ’ (U+2019) alone, as the card: «l’enfant», «aujourd’hui», «Jacopo de’ Barbari»; never the ASCII '.
- **Ellipsis:** one character, «…», as the card's «Vérification…».
- **Dashes:** the English spaced dash becomes a colon, a comma, parentheses or two sentences, which French reads better. Where a single dash stays, it is the em dash with U+00A0 before it and a plain space after it, as the card's «Les problèmes du jour sont finis — c’est la limite de la journée.»; an aside between two dashes takes commas or parentheses instead. Ranges take the en dash with no spaces: «1–6», «1509–1511», «IXe–XIXe siècles», «70–85 %».
- **Ordinals:** «1er», «1re», «2e», «3e», «23e», «XIXe»: the letters straight after the number, with no superscript and never «ème», as the card's «de la 1re à la 6e année». In an example — its question, steps, note and answer — an ordinal the English writes in digits stays in digits: «la 4e fois», «le 23e drapeau», «la 5e case»; one it writes in words stays in words: «le septième lancer». In a drawing, «21e, 22e, 23e»; the badge "2nd" is «2e». An example's answer keeps in digits every number its solver proves.
- **An example's title** takes the article the English one implies: a bare title takes the indefinite one, «Des bonbons dans un bol», «Des champignons»; a title with "The" the definite one; a title that opens with a number none.
- **Dates** use cardinals, but for the first of a month: «du 3 au 10 mai», «le 25», «les 2, 9, 16, 23 et 30», «le 1er du mois prochain», «10 octobre 2026». Months and weekdays are lower case: «lundi», «mai». A weekday in a drawing or a badge is short, as Intl writes it, with its point, and opens with a capital as a label does: «Lun.», «Mar.», «Mer.». The calendar drawings' "7 d" is «7 j».
- **Centuries** in Roman numerals: «au Ve siècle», «fin du XIXe siècle», «IXe–XIXe siècles»; decades in digits: «les années 1950».
- **Times** as the English writes them: «4:25», «15:10», «0:00», «23:59», the bait's «03:59»; never «4 h 25». A length of time is «4 h 30 min» in a drawing (the clocks page's `art.*` keep "{hours} h {minutes} min") and «4 heures 30 minutes» in prose.
- **Numbers on the topic pages and on "Techniques"** — the examples, and every rule, table and tip around them — keep the English digits with no separator: «2026 × 2026», «1111», «1000», «1200», as the Russian writes them. A decimal takes the comma, as the Russian writes it: «0,5 + ½ = 1».
- **Numbers elsewhere** are written as `Intl.NumberFormat("fr")` writes them: the comma for decimals, «0,2», so the formula reads «P = 0,2 + 0,8·σ(θ + δ − β)»; U+202F NARROW NO-BREAK SPACE between thousands, which outside the topic pages only the bait's coins need: «12 450». Years are never grouped: «1894», «2009». Millions in words: «19 millions».
- **Percent**, outside the examples, as Intl writes it, with U+00A0 before the sign: «70–85 %», as the build writes the home page's {chance}, «78 %». Inside an example, and everywhere on the page of percentages, whose drawings write «20%» beside its words, as the English writes it: «20%».
- **Money** appears only in the make-believe app on "Why": «9,99 €» for the hat and «0 €» for MathTrail, as `new Intl.NumberFormat("fr", {style: "currency", currency: "EUR"})` writes them, U+00A0 between the number and the sign (Intl writes «0,00 €»; a price of nothing needs no cents). The coins are «12 450», with U+202F.
- **Units** are symbols after a plain space, as the English writes them, with no point: «12 m», «200 kg», «5 L», «40 min», «4 h», «7 j». The litre is «L», as French schools write it. In prose, «mètres», «litres». In the texts of the cards on the home page and on "Why", the space before a unit's capital stays a plain one: the picture check takes a capital after a no-break space for a label.
- **Signs:** × for multiplication; − (U+2212) for minus, never a hyphen; · in the formula. Division as each English page writes it: ÷ on the home page, "Why", "Techniques" and "Research"; : on the topic pages that use it (gaps, calendar, percentages, ratios, parts, arithmetic). Fractions as the English: «3/4», «½».
- **Capitals:** sentence case in titles and headings: «Comment aider à la maison». Capitals keep their accents: «À propos», «Étude», «Énumération des cas», as the card's «Étape». «olympiade», «primaire» and the languages are lower case in prose («en français»); a named contest takes its capital: «les Olympiades internationales de mathématiques».
- **The card questions on the home page and on "Why"** write no lone Latin capital A–Z inside a sentence that the English does not: the picture check reads one there as a label of the picture. A capital that opens a sentence, a quotation or follows a colon is taken for a word; the one-letter words «à», «a» and «y» are lower case and harmless, and «À» is no A–Z.
- **A sentence that ends with a quote:** a quote that is a whole sentence, brought in by a colon, ends with its own final mark inside the guillemets, and nothing follows: «Dites : « Faisons un problème MathTrail. »». A quote that is part of the sentence takes the sentence's point after the guillemet: «vous répondez toujours « Non ».». A `say:` that is the quote alone has its final mark inside: « Montre-le sur la bande. », « Alors, qui est le plus grand ? Dis-le dans l’autre sens. ».
- **YAML:** a text that holds `: ` is quoted, and in French that is every text with a colon, since the no-break space goes before it and a plain space after it; a text that begins with `*` is quoted too. «, — and À need no quotes. Quote with "…": the French texts hold no ASCII double quote.
- **Titles of works** — a book, a painting, a treatise — go in « … » where the English writes them plain, as the Russian does; where the English writes `*…*`, the emphasis stays and takes no guillemets: «*Comment poser et résoudre un problème*». A French title takes a capital on its first word, and on the first noun after a definite article: «L’École d’Athènes», «Les Neuf Chapitres sur l’art mathématique», «les Éléments».

## Platforms' labels

### Claude

Claude's interface speaks French, and its French help names every label the site quotes, so the labels are those of the French help, not the English ones the Russian keeps. The help speaks «vous», as the site does, so its wording goes into the steps as it stands: «Cliquez sur **+ Ajouter** et choisissez **Ajouter un connecteur personnalisé**.» The help says "Cliquez sur", and so do the site's steps for a button of Claude or ChatGPT; a press on the card is «appuyer».

- [A] "Commencer avec les connecteurs personnalisés utilisant MCP distant", https://support.claude.com/fr/articles/11175166-getting-started-with-custom-connectors-using-remote-mcp
- [B] "Utiliser les connecteurs pour étendre les capacités de Claude", https://support.claude.com/fr/articles/11176164-use-connectors-to-extend-claude-s-capabilities

| English on the site | French | Where the help says it |
|---|---|---|
| Customize → Connectors | Personnaliser → Connecteurs | A and B: "Accédez à Personnaliser > Connecteurs." The site keeps its own arrow. |
| + | + Ajouter | A: "Cliquez sur « + Ajouter » puis sur « Ajouter un connecteur personnalisé »." |
| Add custom connector | Ajouter un connecteur personnalisé | A |
| Name it MathTrail | Nommez-le MathTrail | A: "Entrez un nom pour le connecteur."; the field's label is not quoted |
| the address | l’URL du serveur MCP distant; on the site, «l’adresse» | A: "Ajoutez l'URL du serveur MCP distant de votre connecteur." |
| Continue, and Continue again | Continuer, puis encore Continuer | A says "Cliquez sur « Continuer »." twice: after the URL, and after the sign-in settings |
| the sign-in settings Claude detects | les paramètres d’authentification que Claude détecte | A: "Vérifiez les paramètres d'authentification détectés par Claude" |
| Add | Ajouter | A: "Terminez la configuration de votre connecteur en cliquant sur « Ajouter »." |
| Connect | Connecter | A: "Cliquez sur « Connecter » pour vous authentifier"; B: "Cliquez sur « Connecter » ou « Installer »" |
| + → Connectors, in a chat; switch it on | + → Connecteurs; activer | A: "via le bouton « + » en bas à gauche de votre interface de chat, puis « Connecteurs »", with "des bascules vous permettant de les activer/désactiver" |
| the free plan | le plan Free | A: "pour les utilisateurs des plans Free, Pro, Max, Team et Enterprise. Les utilisateurs Free sont limités à un connecteur personnalisé." |
| Always allow | Toujours autoriser | A, its section "Considérations de sécurité et de confidentialité"; B lists "Toujours autoriser, Approbation requise ou Bloqué" |
| Sign in now · Sign in when needed · No sign in | Se connecter maintenant · Se connecter si nécessaire · Pas de connexion | A, under "Authentification"; not on the site yet |
| Use Claude's published identity · Register automatically · Use your own OAuth client | Utiliser l’identité publiée de Claude (Recommandé) · S’enregistrer automatiquement · Utiliser votre propre client OAuth | A, under "Client OAuth"; not on the site yet |
| Claude Desktop · Claude's mobile apps | Claude Desktop · les applications mobiles de Claude | A: "Claude, Cowork et Claude Desktop" |
| "Use caution before running this prompt" | not found in Claude's French help, nor quoted anywhere found | the text describes it: «sous un avertissement qui invite à la prudence avant de l’envoyer» |
| Claude asks whether to allow MathTrail to save | the question is not quoted in Claude's French help; only the setting "Toujours autoriser" is | the text describes it: «Claude demande s’il faut autoriser MathTrail à enregistrer le profil» |

The help writes its labels in "« … »" with plain spaces and the ASCII apostrophe; the site writes them in bold, as the English does, with ’.

### ChatGPT

| English on the site | French | Source |
|---|---|---|
| developer mode | **le mode développeur**; it is turned on under Paramètres → Applications → Paramètres avancés → Mode développeur, and an app is added under Applications → Créer | https://help.openai.com/fr-fr/articles/12584461-developer-mode-and-mcp-apps-in-chatgpt-beta, "Mode développeur et applications MCP dans ChatGPT": "Accédez à Paramètres → Applications → Paramètres avancés pour l’activer ou le désactiver", and an app is made under "Applications → Créer" |
| ChatGPT's directory | l’annuaire de ChatGPT | OpenAI's French help names it "l’annuaire des plugins", reached from "Plugins" in ChatGPT's side bar: https://help.openai.com/fr-fr/articles/20001256, "Les plugins dans ChatGPT". The generic phrase stays true. |
| an app, apps | une application, des applications | https://help.openai.com/fr-fr/articles/11487775-apps-in-chatgpt, "Applications connectées dans ChatGPT" |
| Plus, Pro, Business, Enterprise, Education plans | les forfaits Plus, Pro, Business, Enterprise et Education | names of plans; «forfait» is OpenAI's French word for a ChatGPT plan, while Claude's help says «plan» |
| on the web | sur le Web | 12584461: "sur la version web de ChatGPT" |

The fr-FR help is the one to use; its articles say they are machine-translated ("Cette page a été traduite automatiquement"), and the fr-CA ones call the plugins "Plugiciels". help.openai.com refuses fetchers: these labels come from the fr-FR pages read through a reader proxy on 11 October 2026.

### Google

| English on the site | French | Source |
|---|---|---|
| Google Drive · My Drive | Google Drive · Mon Drive | https://support.google.com/drive/answer/2409045?hl=fr: "Sur la gauche, cliquez sur Mon Drive." |
| Drive's bin | la corbeille de Drive; the button «Vider la corbeille»; files go for good after 30 days | https://support.google.com/drive/answer/2375102?hl=fr, "Supprimer des fichiers dans Google Drive"; the card says «videz la corbeille de Drive» (`data.delete_how`) |
| version history | l’historique des versions; for a file such as the profile, «Gérer les versions» and «Conserver indéfiniment» | https://support.google.com/drive/answer/2409045?hl=fr, "Suivre l'activité et les versions d'un fichier" |
| linked apps, third-party access | la page Applications associées de votre compte Google (myaccount.google.com/linkedapps, which the English links); the section «Accès à votre compte Google»; «Afficher les détails» → «Supprimer l’accès» → «Confirmer» | https://support.google.com/accounts/answer/13533235?hl=fr, "Gérer les associations entre votre compte Google et les applications d'autres développeurs", whose heading "Connexions tierces : vous gardez le contrôle" names the older page, myaccount.google.com/connections, which https://support.google.com/accounts/answer/3466521?hl=fr still calls "Connexions" |
| Delete all connections | not in Google's French help as read; others quote Android's "Supprimer toutes les connexions associées" | the English never quotes it |
| the box for Google Drive on the consent screen | la case de Google Drive; Google words the permission «Consulter, modifier, créer et supprimer uniquement les fichiers Google Drive spécifiques que vous utilisez avec cette application» | as Google lists it in French in its Workspace Marketplace: https://workspace.google.com/marketplace/app/docs_sheets_editor_for_drive/703230099925?hl=fr. Not seen on the live consent screen, and the English never quotes it. |
| Sign in with Google | «Se connecter avec Google», the button and the service alike; in prose, «se connecter avec Google» | https://developers.google.com/identity/siwg/features?hl=fr; 13533235 above |
| Google account | compte Google | |
| email address | adresse e-mail | as Google's French help writes it |

MathTrail's own sign-in pages speak English, so «Tick the box for Google Drive» and **Back to Google** stay English. The card's own words for Google's pages are quoted as the card writes them: «parmi les applications tierces ayant accès, supprimez MathTrail» (`data.access_how`).

## Glossary

"card" marks the card's own word, from `web/locales/fr.json`.

### Product and lesson

| English | French | Note |
|---|---|---|
| task | problème | card, `task.label` |
| problem (olympiad) | problème | the same word as task |
| school sum | exercice scolaire; calcul | "A school sum" is «Un exercice scolaire»; "1 sum out of 1" is «1 calcul sur 1» |
| topic | thème | card |
| group of topics | groupe | card, `group.*` exactly |
| trap | piège | card, «Le piège» |
| the catch (not a trap) | la difficulté, l’astuce | never «piège» outside the card's sense |
| technique | technique | |
| skill | compétence | «les compétences de Galt»; never «notion» |
| part, of a whole · share, a strip's equal piece | partie · part | «Une partie de quoi ?», «la partie que Tom a lue»; «3 parts», «une part vaut 5»: «part» is the strip's equal piece alone, and a share of an amount in prose is «une fraction» |
| the main move (`topic.method`) | l’idée clé | |
| hint | indice | card; never for "clue" |
| clue, a given of a task | donnée | keeps «indice» for the card's hint: «Chaque donnée est une flèche» |
| option | réponse proposée; réponse | the card says «Choisis une réponse» and lists them under «Réponses» |
| answer | réponse | card |
| right, wrong answer | bonne réponse, mauvaise réponse | card: «Exact !», «Pas tout à fait», «Juste», «Faux» |
| solution | solution | card |
| step, of a solution | étape | never numbered where a rank could be read: the card's rank is «Étape {rank}» |
| question, the text of a task | énoncé | card: «Nombre de l’énoncé pris comme réponse» |
| question, what is asked | question | card: «Réponse à une autre question» |
| picture of a task | dessin | «image», «photo» for a photograph |
| label of a picture | étiquette | |
| grade, with a number | niveau: «Niveau 3», «niveaux 3–4» | card `child.grade`, `progress.grades_ranks` |
| the grade, with no number | le niveau scolaire | card `profile.grade`; «niveau» alone is also "level" |
| grades 1 to 6, in words | de la 1re à la 6e année | card `first_run.lead` |
| level | niveau | card: «à son niveau»; «difficulté» where a grade stands near |
| difficulty | difficulté | card |
| rank | étape | card, «Étape {rank}»: a number, never a title |
| rating | score | card: «Score dans ce thème», «Score global» |
| progress | progrès | card, «Profil et progrès» |
| mastered | maîtrisé | card |
| review, the card's section | bilan | card, «Bilan» |
| how the answer went | l’analyse de la réponse | card `result.label` |
| trial series | série d’essai | card |
| skipped | sans réponse | card |
| profile | profil | card |
| pseudonym | pseudonyme | card |
| interests | centres d’intérêt | card |
| lesson | leçon | card, «Langue des leçons» |
| coach | coach | card, `topic_choice.coach_short`; the Coach product too |
| companion, in the coach | compagnon | a character, not the child |
| parent | parent; les parents; «pour les parents» on a label | card; «un parent» is a mother or a father |
| tutor | enseignant | card `first_run.lead`: «un parent ou un enseignant» |
| adult | adulte | card |
| child, your child | l’enfant; votre enfant | see "No gender shows" |
| children | les enfants | card; on "Research" too |
| chat | le chat | card; on "Techniques" «le chat» is also the cat, and the sentence tells them apart |
| card | carte | card |
| connector, custom connector | connecteur, connecteur personnalisé | Claude |
| app | application | Claude and OpenAI |
| sign in | se connecter | card, Claude, Google |
| sign-in | connexion | |
| account | compte | |
| Drive file, the profile's file | fichier Drive, fichier du profil | card |
| folder | dossier | card |
| bin | corbeille | card and Google |
| daily limit | limite de la journée | card |
| request, what a parent asks | demande | card: «Dès que la demande arrive dans le chat» |
| Another task | Un autre problème | card's button, quoted exactly |
| Make a wish for tomorrow | Fais un vœu pour demain | card, quoted exactly |
| Profile & progress | Profil et progrès | card |
| Edit | Modifier | card |
| Don't count the country I sign in from | Ne pas compter le pays depuis lequel je me connecte | card, `profile.signin_country_off`, quoted exactly |
| the bulb, the hint's button | l’ampoule | |
| the two arrows | les deux flèches | |
| tap | appuyer: «d’un simple appui» | the card's buttons |
| press, click | cliquer: «cliquez sur» | Claude's and ChatGPT's buttons, as their help writes it |
| Enter, the key | Entrée | |
| free, of price | gratuit | the chip `hero.free`: «Gratuit» |
| Free, the plan | Free | «le plan Free»: a plan's name |
| open source | open source | the chip `hero.open`: «Open source» |
| in development | en développement | |
| beta | Beta | |
| soon | bientôt | |
| olympiad maths | maths d’olympiade | card: «Coach d’olympiades», «problèmes d’olympiade» |
| olympiad-style task | problème dans l’esprit des olympiades | card `first_run.lead` |
| maths circle | club de maths | what French schools call one; «cercle mathématique», the Russian tradition's name, is unknown to most parents |
| school | école | card |
| school mark | note | card: «une note scolaire» |
| homework | devoirs | |
| I don't know | je ne sais pas | the answer "?" |

### The site's sections and headings

| English | French | Note |
|---|---|---|
| Topics (`nav.topics`) | Thèmes | 6 characters; card `progress.topics` |
| Techniques (`nav.techniques`) | Techniques | 10 |
| Why (`nav.why`) | Pourquoi | 8 |
| Coach (`nav.coach`) | Coach | 5, beside the tag «Beta»; the card's word |
| Research (`nav.research`) | Étude | 5. «Recherche», 9, took the row past the Russian, and a French menu reads it as "Search". The page's title is «Étude» too, since the English page's title is the menu's word, and the footer shows it |
| Service (`nav.service`) | Service | 7 |
| About (`nav.about`) | À propos | 8, and the page's title too |
| Add to Claude (`nav.add`) | Ajouter à Claude | 16; Claude's own «Ajouter» |
| Technical pages | Pages techniques | said by a screen reader |
| Language | Langue | |
| Help · Privacy policy · Terms of use | Aide · Politique de confidentialité · Conditions d’utilisation | |
| How it is solved | Comment on le résout | |
| Where children go wrong | Où les enfants se trompent | |
| What to say | Que lui dire | to the child; «lui» shows no gender |
| How to help at home | Comment aider à la maison | |
| Learn first | À apprendre d’abord | |
| Opens the way to | Ouvre la voie à | |
| Builds on · Opens | S’appuie sur : · Mène à : | the topics page |
| Ask for a task on this topic | Demandez un problème sur ce thème | |
| Show · Hide the solution | Voir la solution · Masquer la solution | card: «Masquer l’indice» |
| Example 2 | Exemple 2 | |
| What it teaches (`topic.teaches`) | Ce qu’on y apprend | «Ce qu’il apprend» would read as the child |
| Why it matters | Pourquoi c’est important | |
| How to solve: … | Comment le résoudre : … | |
| The rules of … | Les règles de… | «Les règles du calendrier», «Les règles de l’île» |
| Check: | Vérification : | card: «Vérification…» |
| When it helps: | Quand l’utiliser : | |
| In the topics: | Dans les thèmes : | |
| You are here · On this page · All topics | Vous êtes ici · Sur cette page · Tous les thèmes | |
| More → | En savoir plus → | |
| Honestly | Honnêtement | |
| See a lesson | Voir une leçon | |
| For technical readers | Pour un public technique | no gendered «lecteurs» |
| You, the parent's line on "Why" · C, the child's | Vous · E | «E désigne votre enfant.» |
| The adult · The model, in the chat's illustration | L’adulte · Le modèle | |
| Step 1, of the topics' map | Palier 1 | «Étape 1» would read as a rank |
| Foundations · Rules · Hard problems | Les bases · Les règles · Les problèmes difficiles | |
| Pólya's four steps, on "Coach" | Les quatre phases de Pólya | the French edition's «quatre phases» |

Measured in the site's font and sizes, the header's row of entries and its button take, in Chromium and WebKit: Russian 850.4 and 849.2 px, Spanish 846.4 and 845.6 px, and the French above 842.4 and 839.2 px. With «Recherche» the French takes 877.4 and 873.4 px, past the Russian by nearly as much as the Spanish «Investigación» was (886.4 and 885.9 px), so `nav.research` is «Étude». Should the menu ever need more room, the next cuts are «À propos» → «Équipe» (16 px less), then «Techniques» → «Méthodes» (9 px less); «Pourquoi» has no shorter word, and «Ajouter à Claude» keeps Claude's verb. `just site-fold` measures the real fold.

### Maths and olympiad terms

| English | French | Note |
|---|---|---|
| olympiad | olympiade | «les olympiades» for the competitions in general, as the card's «dans l’esprit des olympiades» |
| International Mathematical Olympiad · IMO | les Olympiades internationales de mathématiques · OIM | the French name and its initials, also on "Why"'s big line |
| reference task | problème de référence | |
| solver | solveur | the first time, «un programme qui résout le problème» |
| brute force | force brute | |
| worst case | le pire des cas | card, `trap.best_case_not_worst` |
| good, bad luck | la chance, la malchance | card: «le cas le plus malchanceux» |
| to be sure, for certain | à coup sûr, avec certitude | never «pour être sûr», which agrees with its subject; "To be sure", a card's title, is «À coup sûr» |
| pigeonhole principle | principe des tiroirs | card |
| pigeons, holes | pigeons, cages | the drawings show pigeons in cages: "Pigeons and pigeonholes" is «Les pigeons et les cages»; never «cases», which are a grid's cells |
| parity · even · odd | parité · pair · impair | card |
| alternation, taking turns | alternance, à tour de rôle | card |
| remainder · quotient | reste · quotient | card |
| divisibility | divisibilité | card |
| divides by 9 | est divisible par 9 | «diviser par», throughout |
| divisibility rules | critères de divisibilité | card |
| digit · two-digit number | chiffre · nombre à deux chiffres | |
| tens, units digit | chiffre des dizaines, des unités | |
| enumeration | énumération des cas | card |
| case · list | cas · liste | |
| systematic listing | énumérer avec méthode | the technique "List systematically" |
| tree, the diagram | arbre | «arbre des possibles», as French school maths calls it |
| working backwards | raisonner à rebours; partir de la fin | |
| the assumption method | la méthode de fausse position | its name in French arithmetic since the old «règle de fausse position» |
| proof by contradiction | le raisonnement par l’absurde | |
| suppose · assumption · contradiction | supposons que… · hypothèse · contradiction | |
| pattern | régularité; motif | «motif» for the piece that repeats |
| symmetric strategy | stratégie symétrique | |
| winning strategy | stratégie gagnante | card |
| winning, losing position | position gagnante, perdante | |
| move, in a game | coup | card |
| the first, the second player | le premier joueur, le second | |
| invariant | invariant | |
| knights and liars | chevaliers et menteurs | card |
| knight · liar | chevalier · menteur | V and M in drawings. «Chevalier» has no feminine French puzzles use («chevalière» is a signet ring), so a named knight is a boy |
| island · islander | île · habitant de l’île | |
| statement | affirmation | card |
| ordering | mise en ordre | card |
| taller, on the ordering page's arrows (`art.taller`) | dépasse | a verb, the same for a girl and a boy, since the arrows join both; the line's two ends are «qui mesure le plus» and «qui mesure le moins» |
| number line | droite graduée | card |
| grid · squared paper | quadrillage · papier quadrillé | card: «Figures sur quadrillage», «Une grille de cases» |
| cell · side of a cell | case · côté d’une case | card |
| area · perimeter | aire · périmètre | card |
| corner, of a cube · of a board | sommet · coin | |
| ratio · in the ratio 3 to 4 | rapport · dans le rapport 3 à 4 | card |
| sharing · share | partage · part | card |
| percentage · percent | pourcentage · pour cent | card |
| fraction · the whole · the part | fraction · le tout · la partie | card |
| a half, a third, a quarter | la moitié, un tiers, un quart | |
| strip, for parts | bande | |
| gaps and boundaries | intervalles et bornes | card |
| gap, between posts · post · fence | intervalle · poteau · clôture | card: «les poteaux en points et les intervalles entre eux»; a fence's "span" is «intervalle» too |
| off by one | un de plus ou de moins | card |
| clock face · hands · the whole hour | cadran · aiguilles · l’heure pile | card: «Un cadran d’horloge» |
| floor 1, of a house · flight of stairs | l’étage 1 · volée d’escalier | never «le 1er étage» or «le rez-de-chaussée»: French counts floors from a ground floor, which would change the count |
| calendar and age | calendrier et âges | card |
| pan balance · weighing · fake coin | balance à plateaux · pesée · fausse pièce | card |
| jug · pour | bidon · verser, transvaser | card: «Des récipients pour transvaser»; «bidon», as in the classic «problème des bidons» |
| fill · empty · top up | remplir · vider · compléter | |
| state, of the jugs | état | |
| overlapping groups | groupes qui se chevauchent | card |
| arithmetic with a trick | calculs astucieux | card |
| I thought of a number | j’ai pensé à un nombre | |
| carrying | retenue: «sans retenue» | |
| round number | nombre rond | |
| match, the stick | allumette | |
| a special case | un cas particulier | |

### The technical pages

| English | French | Note |
|---|---|---|
| token · access token · refresh token | jeton · jeton d’accès · jeton d’actualisation | as Google's French docs |
| authorization code | code d’autorisation | |
| seal · sealed · sealing | sceller · scellé · scellement | |
| encrypted | chiffré | never «crypté» |
| state · stateless | état · sans état | |
| MCP server · tool · tool call | serveur MCP · outil · appel d’outil | |
| chat host | hôte du chat | |
| consent screen | écran de consentement | Google |
| scope | champ d’application | Google's French term: «le champ d’application drive.file» |
| client · client metadata document | client · document de métadonnées du client | |
| renewal · revocation | renouvellement · révocation | |
| pace · ceiling | rythme · plafond | |
| Sign-in, a scenario's tab on "Service" · target, of the profile's size | Connexion · cible | short: the tab and the bar's label do not wrap at 320 px |
| log · logging | journal · journalisation | «Cloud Logging» stays |
| instance · revision · image | instance · révision · image | |
| key rotation · ciphertext · nonce | rotation des clés · texte chiffré · nonce | |
| fingerprint · near-duplicate | empreinte · quasi-doublon | |
| readability · self-check | lisibilité · autovérification | |
| brief, for the model | consigne | |
| package · schema · template | paquet · schéma · modèle | |
| deterministic · the rule | déterministe · la règle | |
| cookie · cache | cookie · cache | |
| graceful shutdown | arrêt propre | |
| lifetime, of a token | durée de vie | |
| Workload Identity Federation | fédération d’identité de charge de travail | as Google's French docs |
| student model | modèle de l’élève | |
| chance, of a right answer | probabilité de réussite | |
| corridor · interval · goal | couloir · intervalle · objectif | |
| the ceiling, an estimate that knows the child · oracle | le plafond · oracle | |
| bench · simulated children | banc d’essai · enfants simulés | |
| a child who learns, who stays put | un enfant qui progresse, qui stagne | |
| lag · logit · points | retard · logit · points | |
| mastery declared falsely | maîtrise déclarée à tort | |
| live data, live numbers | données réelles, chiffres réels | they come monthly, not in real time |
| range, of the chance promised | tranche | |
| standard error · baseline · seed | erreur type · référence · graine | |
| Reached · On the edge · Not reached | Atteint · À la limite · Non atteint | |
| paper · public domain | article · domaine public | |
| commit · release | commit · version | «commit» is the borrowed term |
| endpoint · claim, of a JWT | point de terminaison · claim | |
| request, over HTTP · store · router · issuer | requête · stockage · routeur · émetteur | the card's «demande» for what a parent asks |
| deploy · build · client secret · sandboxed | déploiement · compilation · secret client · isolé | «bac à sable» for the sandbox itself |
| in the open, unsealed · timeout | en clair · délai d’expiration | |
| digest · runtime · PKCE verifier · state machine | condensé · environnement d’exécution · vérificateur PKCE · machine à états | |
| scenario · sketch · pinned · sliding, of a lifetime | scénario · esquisse · épinglé · glissante | |
| canvas · tablet | zone de dessin · tablette | the coach |
| GitHub issue · computer · mouse | ticket GitHub · ordinateur · souris | |
| the profile's notes | remarques | «note» is a school mark |
| whisker, of an error bar · dashed | moustache · en pointillés | |
| item response theory · psychometrics | théorie de la réponse à l’item (TRI) · psychométrie | |

## Pólya's four steps

The French edition is «Comment poser et résoudre un problème» (Dunod, Paris, 1957, translated by C. Mesnage, with a preface by Georges Darmois; a second, enlarged edition in 1965, reprinted by Éditions Jacques Gabay in 2011). Its contents name "Les quatre phases": "Comprendre le problème", "Concevoir un plan", "Exécution du plan", "Revenir en arrière"; the dialogue of its second part heads the last two "Mettre le plan à exécution" and "Revenir sur la solution" (the reprint's contents: https://www.gabay-editeur.com/POLYA-Comment-poser-et-resoudre-un-probleme-2e-ed-1965). The page "Techniques" names its steps after them, each an infinitive, in the edition's own words:

| English | French |
|---|---|
| *How to Solve It* (1945) | *Comment poser et résoudre un problème* (1945) |
| Understand the problem | Comprendre le problème |
| Make a plan | Concevoir un plan |
| Carry it out | Mettre le plan à exécution |
| Look back | Revenir sur la solution |

The fourth is the dialogue's «Revenir sur la solution» rather than the contents' «Revenir en arrière», which beside the technique «Partir de la fin» would read as undoing. A technique's name is an infinitive too: «Dessiner le problème», «Énumérer avec méthode», «Partir de la fin».

The steps' questions speak to the child: «Dis-le avec tes propres mots. Qu’est-ce que tu sais ? Qu’est-ce que tu dois trouver ?» The coach's shorter line, "Understand → make a plan → solve → check", is «Comprendre → faire un plan → résoudre → vérifier», and the picture's "understand, plan, solve, check" on "Why" is «comprendre, planifier, résoudre, vérifier».

## People's names

"Initial" marks a name whose first letter a drawing or a badge shows; within one drawing no two letters repeat. The pets table on "Techniques" draws the first letter of each child's name in a disc, so its three names start with three different letters.

| English | French | Where | Initial, badge |
|---|---|---|---|
| Galt | Galt | home, the cards' pseudonym: «Choisi pour Galt» | kept, as the Russian and the Spanish keep it; an invented name shows no gender |
| Comet | Comète | "Why" and "Topics", the cards' pseudonym | as the Russian «Комета» and the Spanish «Cometa»; a name, so never «la» or «une», and no word that agrees with it |
| C, "your child" on "Why" | E | `parent.explain.child`: «E désigne votre enfant.» | the speaker's letter, from «enfant» |
| Ann | Anna | ordering, parts, ratios, clocks, calendar | A, kept |
| Ben | Basile | ordering, parts, ratios, parity | B, kept |
| Kate | Clara | ordering, calendar | the letter changes, K → C: `art.kate` is «C» |
| Dan | David | ordering's line | name in a drawing |
| Lena | Léna | ordering's queue, gaps, ratios | L, kept |
| Max | Max | ordering's queue, ratios | M, kept; as common in French as in English |
| Nina | Nina | ordering's queue | N, kept |
| Oleg | Oscar | ordering's queue | O, kept |
| Pavel | Paul | ordering's queue | P, kept, and the badge «P» |
| Ada | Alice | ordering's runners | A, kept |
| Bob | Bruno | ordering's runners | B, kept; not Basile, who is on the same page |
| Cid | César | ordering's runners | C, kept |
| Dot | Diane | ordering's runners | D, kept |
| Eli | Enzo | ordering's runners | E, kept |
| Fay | Fanny | ordering's runners | F, kept, and the badge «F» |
| Anya | Adam | "Techniques": the pets and the knights | A, kept. A boy, since a named knight is «un chevalier» |
| Borya | Billie | "Techniques": the pets and the knights | B, kept. A girl, so that the pair stays a boy and a girl: «Billie est une menteuse». Short, since the pets table is narrow |
| Vera | Jade | "Techniques", calendar | J in the pets table, apart from A and B |
| Tanya | Sophie | "Techniques", working backwards | |
| Petya | Pierre | "Techniques", working backwards | |
| Lea | Léa | enumeration | L, kept |
| Tom | Tom | enumeration, clocks, parts | T, kept |
| Masha | Manon | gaps, arithmetic | M, kept |
| Sasha | Sacha | pigeonhole, the sweets | the same name, given to boys and girls alike; the English gives no gender, and neither does the French: no pronoun, no word that agrees («Combien de bonbons faut-il que Sacha prenne…») |
| Gleb | Gabriel | pigeonhole, the socks | |
| Kim | Zoé | winning strategy | name in drawings: «Zoé −3» |
| Leo | Léo | winning strategy | name in drawings: «Léo −4» |
| Grandma, Granny | Mamie | clocks, parts | with its capital, as a name: «chez Mamie», «l’horloge de Mamie» |
| Mum, Dad | Maman, Papa | calendar, ordering's tip, "About" | with their capitals, as names, in drawings too |
| islanders A, B, C | A, B, C | knights | letters |
| knight K, liar L | V, M | knights, `thumb.*` | V for «il dit toujours vrai», M for «menteur»; C stays an islander's |
| children A, B, C · boxes A, B | the same | enumeration · ratios | letters |
| the cook (she) · the baker (he) | une cuisinière · un boulanger | weighing · parts | |
| woodcutters · a farmer | des bûcherons · un fermier | gaps · percentages | |

The people of history keep their French names: Euclide, le roi Ptolémée, Proclus, les « Éléments », Archimède, Raphaël, « L’École d’Athènes », Confucius, les « six arts », Aryabhata, Taqi al-Din, Istanbul, Luca Pacioli, Jacopo de’ Barbari, al-Khwârizmî, Bagdad, « Les Neuf Chapitres sur l’art mathématique », Aristote, Alexandre le Grand, Nikolaï Bogdanov-Belski, « Calcul mental. Dans l’école populaire de S. A. Ratchinski » (the painting's French title), le sanctuaire Konnō Hachiman-gū, Tokyo, l’école Bates, Saint-Louis, Léningrad, Moscou, les Archives principales de Moscou, George Pólya; the places Grèce, Chine, Inde, le monde islamique, Italie, Japon, Russie, États-Unis, URSS, Hongrie, Allemagne, Brésil, Royaume-Uni. A photograph's credit, «Yu. D. Korolev» or «Momotarou2012», stays as its source writes it.

## Grades

`topics.grades` is **«Niveaux {range}»**: «Niveaux 1–2», «Niveaux 3–4», «Niveaux 5–6», «Niveaux 1–6»; so `topic.example` reads «Exemple 2 · Niveaux 3–4», and the label on "Techniques" «Problème · Niveaux 2–3».

- It is the card's own word for a grade: «Niveau {grade}» (`child.grade`), «problèmes des niveaux {from}–{to}» (`progress.grades_ranks`), «Niv. {from}–{to}» (`progress.grades`), so a chip on the site and the card's pill read alike.
- No school year has one name across the French-speaking countries (CP to CM2 in France, «1re année» in Québec, «1re primaire» in Belgium, 1P to 8P in Switzerland); «niveau» and a number read the same in all of them, as on the card.
- A bare range takes no ordinal letter: «{range}e année» would mark only the last number, «3–4e», which French does not write.

The alternative is «Niv. {range}», the card's own abbreviation, should a chip run out of room. A text with two slots writes the card's form: «Niveaux {first}–{last}» (`hero.grades` on "Research"), «{count} · niveaux {first}–{last}» (`sources.level`). The home page's lead, which has the one slot, says «pour les niveaux {range}»; "up to grade {last}" on "Why" is «jusqu’au niveau {last}»; "grades 1 to 6" written out in words is the card's «de la 1re à la 6e année»; "the grade" with no number is «le niveau scolaire», since «niveau» alone is also "level".

## Books on "Research"

| Key | The English page | French | How sure |
|---|---|---|---|
| perelman.firsthundred | For Young Mathematicians. The First Hundred Puzzles | the English title | no French edition found |
| perelman.secondhundred | The Second Hundred Puzzles | the English title | no French edition found |
| perelman.living | Living Mathematics | «Oh, les maths !» (Dunod and Mir, 1992, translated by Catherine Eméri; reissued by Dunod in 2001 as «Oh, les maths ! 200 contes, énigmes et casse-tête pour mourir de rire… et tuer le temps») | fairly sure it is this book: its chapters, from "Des casse-tête en apéritif" through "Savez-vous compter ?", "Géométrie par temps de pluie" and "Le Déluge passé au crible des maths" to "Trente problèmes divers", are those of «Живая математика»; the volume adds his algebra puzzles, so it holds more than this one book |
| perelman.arithmetic | Entertaining Arithmetic | the English title | no French edition found; the prefaces of Mir's French editions call it "L'arithmétique récréative", a name for the Russian book rather than an edition |
| perelman.quick | Quick Reckoning | the English title | no French edition found |
| ignatyev.wit | In the Realm of Ingenuity, or Arithmetic for Everyone | the English title | no French edition found; Mir published it in Spanish, «En el reino del ingenio», but no French one was found |
| dudeney.amusements | Amusements in Mathematics | the English title | only a selection came out in French, «120 casse-tête d’autrefois inédits» (Fantaisium, 2019, translated by François Montmirel), which carries neither the book's title nor the whole book |
| dudeney.canterbury | The Canterbury Puzzles | «Les Énigmes de Canterbury» (Fantaisium, 2018, two volumes, translated by François Montmirel) | sure |
| loyd.cyclopedia | Cyclopedia of Puzzles | the English title | no French edition found; «Les Casse-tête mathématiques de Sam Loyd» (Dunod, from 1970) is Martin Gardner's selection, another book |
| carroll.tangled | A Tangled Tale | «Un Conte embrouillé» (in Lewis Carroll, «Œuvres», Bibliothèque de la Pléiade, Gallimard, 1990, edited by Jean Gattégno) | sure; the capital is the Pléiade's |
| carroll.pillow | Pillow Problems | «Énigmes mathématiques de Lewis Carroll» (Pôle, 1999, adapted by Élisabeth Busser) | fairly sure: its 72 problems, "pour vos nuits blanches", are the book's 72, but grouped by theme and reworded, an adaptation rather than a translation |

The authors' names: Yakov Perelman, as Dunod writes him (Mir's French editions write "Ya. Pérelman"); Emelian Ignatiev, as French writes Russian names; Henry E. Dudeney, as Fantaisium writes him; Sam Loyd and Lewis Carroll as they are.

## Search words

Gathered on 2026-10-12 from Google's suggestions for France, Belgium, Switzerland and Canada, and the titles of the first results; they show that a phrase is typed, not how often, and [the brief](README.md#search-words) says how a text uses them. The country changes the lists little.

**Who searches, and how.** In France Google is 88 % of searches and Bing 6 %; in Belgium, Canada and Switzerland Bing is 9–12 %. French parents type the class's name, «cm2», «cm1», «ce2», «cp», «6ème» or «cycle 3»; Belgian ones «3e primaire», those of Québec «3e année», those of Romandie «5P». `topics.grades`, «Niveaux {range}», matches none of them, and in France «niveau 3» reads as the collège's «3e».

**What parents type**, strongest first:

1. «énigmes mathématiques» or «énigmes maths» with a class;
2. «problèmes de logique» with a class or «avec correction»;
3. «défi maths», «défis mathématiques»;
4. «casse-tête mathématiques»;
5. «jeux mathématiques», «jeux de logique»;
6. «devinettes mathématiques»;
7. «résolution de problèmes».

«Maths d’olympiade» and «problèmes d’olympiades» are never suggested: in French search «olympiades» are contests.

**The head's title.**

- A topic's page: **«{topic} : énigmes de maths, niveaux {range}»**, with a no-break space before the colon. «Énigmes … maths» is the strongest query, and «niveaux {range}» is `topics.grades`. The longest, «Groupes qui se chevauchent : énigmes de maths, niveaux 5–6», is 58 units; «énigmes mathématiques» would push three topics past 60. The other way, «{topic} : défis maths, niveaux {range}», 53 at most, takes the parents' «défis».
- Any other page: **«{page} — MathTrail»**, with a no-break space before the dash, as the glossary writes a dash alone; the longest, «Techniques de résolution de problèmes de maths — MathTrail», is 58.

**Topics.** ⚑ marks a card's name nobody types.

| Topic | The card | What parents type |
|---|---|---|
| ordering | «Mise en ordre» ⚑ | only «problèmes de logique»; the card's name is a crossword's clue |
| knights and liars | «Chevaliers et menteurs» ⚑ | «énigme du menteur», «l’île aux menteurs» |
| enumeration | «Énumération des cas» ⚑ | «problèmes de dénombrement» with a class, «arbre des possibles» |
| gaps and boundaries | «Intervalles et bornes» | «problème des intervalles», «piquets et intervalles»; «bornes» is analysis |
| clocks | «Horloges» ⚑ | «calcul de durées», «problèmes de durées», «lire l’heure»; «horloges» are wall clocks |
| calendar and age | «Calendrier et âges» ⚑ | «problèmes d’âge», which lean to the collège's equations |
| pigeonhole principle | «Principe des tiroirs» | the same |
| parity and alternation | «Parité et alternance» ⚑ | «pair et impair», rarely |
| arithmetic with a trick | «Calculs astucieux» | the same; «astuces de calcul mental» |
| weighing and pouring | «Pesées et transvasements» ⚑ | «énigme des bidons», «transvasement», «énigme de la balance» |
| parts and shares | «Parties d’un tout» ⚑ | «problèmes de fractions»; the card's name is a crossword's clue |
| percentages | «Pourcentages» | the same |
| ratios and sharing | «Rapports et partages» ⚑ | «proportionnalité», «partage proportionnel», «problèmes de partage» |
| figures on a grid | «Figures sur quadrillage» | «aire et périmètre»; the card's name leads to copying figures |
| divisibility and remainders | «Divisibilité et restes» | «critères de divisibilité», «division euclidienne» |
| overlapping groups | «Groupes qui se chevauchent» ⚑ | «diagramme de Venn» |
| games with a winning strategy | «Jeux à stratégie gagnante» ⚑ | «jeu de Nim», «jeu des allumettes» |

**Descriptions** can carry «énigmes mathématiques», «problèmes de logique», «primaire», which reads alike in all four countries, on "Techniques" «résolution de problèmes», and «raisonnement» as a noun. «Avec correction» fits a topic's page and "Techniques" only.

**Not to use.**

- Contests: «Olympiades de mathématiques», «maths olympiades», «Kangourou des mathématiques», «rallye maths».
- What the site does not give: «pdf», «à imprimer».
- «développer le raisonnement», a promise.

**For the author.** The site's own words for its kind, «maths d’olympiade» and «problèmes d’olympiade», written 34 times, among them the titles of the home page and "Why", are never suggested. France searches by «CP» to «CM2» and «6e», which the glossary leaves out to stay alike in every country; a description can say «primaire», but «du CP à la 6e» writes grades a description cannot take.

## Places for the author's reading

- The menu: «Étude» for «Recherche», in the menu, the page's title and the footer alike, keeps the French row under the Russian, and keeps a French reader from taking the entry for a search; the other way out is «Recherche» everywhere and a fold some 30 px wider for every language.
- `topics.grades`, «Niveaux {range}», against «Niv. {range}».
- `topic.teaches`, «Ce qu’on y apprend», chosen because «Ce qu’il apprend» would read as the child; `topic.child`, «Enfant», which shows no gender and takes 41 of its 64 px; `topic.right`, the card's «Juste».
- The ask lines, «Donne-moi un problème sur le thème « … »», the card's own way of naming a topic in the chat, against the lighter «Donne-moi un problème de mise en ordre», which several names would not take without a change of their words.
- On "Techniques", Anya and Borya become Adam and Billie, a boy and a girl swapped, because «chevalier» has no feminine in this sense; Vera becomes Jade. The pets table's «perroquet» is about 56 px at its 12 px size, and its column is 40 to 58 px wide: at 320 px it may stick out of its cell, which a look at the page settles.
- Kate's bead letter, K → C.
- «clue» → «donnée», to keep «indice» for the card's hint.
- «Palier 1» for the map's layers, since «Étape» is the card's rank.
- Floors written «l’étage 9», so that no French reader counts from a ground floor.
- On the ordering page, the arrows' «dépasse» and the line's «qui mesure le plus» and «qui mesure le moins», chosen because «plus grand» would agree with one child and not the other.
- Percent: «70–85 %» with its no-break space across the site, as Intl writes it, but «20%» on the page of percentages, whose drawings write it so, and in examples.
- The litre as «L», as French schools write it, against the English «l».
- Claude's labels come from its French help; its warning over a prompt and its question about saving are not in that help, so the text describes them.
- ChatGPT's labels come from OpenAI's machine-translated fr-FR help, read through a proxy.
- The consent screen's wording of the Drive permission comes from Google's Marketplace, not from the screen itself.
- The books marked fairly sure: «Oh, les maths !», which holds Living Mathematics and more, and «Énigmes mathématiques de Lewis Carroll», an adaptation; and «Amusements in Mathematics», kept in English though a selection exists in French.
- «OIM» on "Why"'s big line, the French initials, for the English «IMO».
- The make-believe app's cheer is «Tu es une star !!!», since «champion» would make the child a boy.
- Mamie, Maman and Papa written with capitals, as names, in drawings too.
- The card names both a group and a topic «Parties d’un tout» (`group.parts`, `topic.fractions.parts`), so the topics page will show the topic under a group of its own name, as the Spanish does. The fix belongs to the card's dictionary, and the site follows it.
