# Hindi (`hi`)

The Hindi half of [the brief](README.md): the rules the Hindi texts follow on top of it, and the words they use. The card's Hindi dictionary, [`web/locales/hi.json`](../../web/locales/hi.json), comes first wherever it has a word, and the site's own, [`web/src/site/locales/hi.json`](../../web/src/site/locales/hi.json), already holds the menu, the groups of topics, the letters of the knights' page and the nouns of "Research"; this file adds only what the pages need beyond them. Hindi verbs show the gender of their subject in most tenses, as Arabic verbs do in every person, so much of this file says how a sentence is worded so that neither the reader's gender nor the child's shows.

## Register and variant

- **One Hindi, the card's: standard modern Hindi.** The everyday word wherever the card uses one — «सवाल», «जवाब», «हल», «संकेत», «कोशिश», «ज़्यादा», «मदद», «स्कूल» — and the card's loans as they are: «सेव», «प्रोफ़ाइल», «चैट», «कार्ड», «लिंक», «फ़ाइल», «फ़ोल्डर», «रेटिंग», «कोच», «ऐप», «ऐक्सेस». A concept of maths takes the term of India's Hindi-medium school books (NCERT): «भिन्न», «अंश», «हर», «शेषफल», «भागफल», «भाजक», «विभाज्यता», «क्षेत्रफल», «परिमाप», «अनुपात», «समानुपात», «प्रतिशत», «सम», «विषम», «इकाई», «दहाई», «सैकड़ा», and «अंक» for a digit against «संख्या» for a number. No heavier Sanskrit word where an everyday one is usual: «मदद», not «सहायता»; «सवाल», not «प्रश्न»; «हफ़्ता» and «साल» in prose, not «सप्ताह» and «वर्ष». No English loan where a common Hindi word exists, but the card's loans stay.
- **Spelling, as the card writes it:**
  - The nukta on ज़ and फ़, and on ड़ and ढ़, always: «ज़्यादा», «ज़रूरत», «फ़ाइल», «हफ़्ता», «मुफ़्त», «टॉफ़ी», «घड़ी», «पढ़ना». It is the separate sign U+093C after the letter, as all 120 of the card's are, never a precomposed letter of U+0958–U+095F, which some keyboards type and which would make a topic's name differ from the card's by its bytes. A search for `[\u0958-\u095F]` finds them.
  - क, ख and ग without the nukta, as most of the card's words write them: «किस्मत», «बाकी», «खत्म», «गलत», «गलती», «कागज़». The card's «आख़िरी» (`review.moving`) and «बदक़िस्मत» (`advice.best_case_not_worst`) keep theirs only where the site quotes them.
  - The chandrabindu where nothing stands above the headline, the bindu where a vowel sign does: «जाँच», «पहुँच», «माँगें», «यहाँ», «बँटवारा», «पाँच», «दिखाएँ», but «में», «हैं», «नहीं», «करें», «पाँचवीं».
  - The anusvara before a consonant, never a half nasal, as in every word of the card: «अंक», «अंतराल», «संख्या», «पंक्ति», «खंभा», «हिंदी». The language picker's «हिन्दी» (`language.name`) is the name Intl gives the language, and stays there alone.
  - The polite verb ends in «-एँ» after a vowel and «-ें» after a consonant: «हटाएँ», «दिखाएँ», «जाएँ»; «करें», «बदलें», «लिखें».
- **The reader is «आप»**, in the polite imperative and subjunctive, as the card speaks to the parent: «चैट में सवाल माँगें।» (`waiting.ask_in_chat`), «चाहें तो बच्चे की रुचियाँ…» (`first_run.fill`), «असली नाम, जन्मतिथि या स्कूल कभी न लिखें।» (`first_run.pseudonym`). Hindi verbs show their subject's gender in the habitual, the progressive, the future, the perfective of an intransitive verb and with «सकना», so «आप» is never the subject of such a form: not «आप कर सकते हैं», «आप चाहेंगे», «आप देख रहे हैं», «आप पहुँचे». Instead:
  - the polite imperative or the subjunctive: «लिखें», «देखें», «चुनें», «चाहें तो», «अगर आप चाहें कि सवाल बच्चे के लिए नया रहे…»;
  - the passive, as the card's «जिसे खोला, डाउनलोड किया या मिटाया जा सकता है» (`first_run.data`): "you can open it, copy it or delete it" is «इसे खोला, कॉपी किया या मिटाया जा सकता है»; "you may leave them out" «इन्हें छोड़ा भी जा सकता है»; "ask in the chat" in a statement «चैट में सवाल माँगा जा सकता है»;
  - the dative: «आपको Claude या ChatGPT खाता चाहिए», «आपको हर विषय का पड़ाव दिखेगा», «आपको फिर से साइन इन करना होगा»;
  - the ergative of a transitive verb, which agrees with its object: «आपने विषय चुना», «अगर बदलाव आपने माँगा हो»;
  - the genitive, which agrees with the thing possessed: «आपका चैट», «आपकी Google Drive», «आपके बच्चे की प्रगति»;
  - an impersonal turn: "you do not have to know olympiad topics" is «ओलंपियाड के विषय जानना ज़रूरी नहीं»; "you can prepare a child yourself" «बच्चे को खुद तैयार करना संभव है»; «…करना काफ़ी है»;
  - the copula, which shows no gender: «आप यहाँ हैं» (`topic.path`), «…कि आप बच्चे के अभिभावक या ट्यूटर हैं».

  No variable adjective or participle about the reader: never «पक्का», «थका हुआ», «झूठा»; «तैयार», «खुद» and «सहमत» do not change and may stand.
- **The child** is the card's generic «बच्चा»: «बच्चे की प्रगति के हिसाब से» (`topic_choice.coach_note`), «बच्चे के स्तर के हिसाब से» (`review.step_practice`), «बच्चे का असली नाम कभी नहीं।» (`profile.pseudonym_note`); «आपका बच्चा», «आपके बच्चे» may stand where the English says "your child". But the child, the reader's or any one child, is never the subject of a verb form that shows gender, and takes no variable adjective or participle: not «बच्चा जवाब देता है», «बच्चा उलझ जाता है», «बच्चा कर सकता है», «सीखने वाला बच्चा». Instead:
  - a genitive and a verbal noun: "once your child answers" is «बच्चे के जवाब देते ही»; "before your child sees it" «बच्चे के देखने से पहले»; "how the child reasoned" «बच्चे के सोचने का तरीका»;
  - the dative: "the child notices" «बच्चे को समझ आ जाता है»; "if the child needs a hint" «बच्चे को संकेत की ज़रूरत हो»; "what your child already knows" «बच्चे को पहले से क्या आता है»; "where your child went wrong" «बच्चे से कहाँ गलती हुई»; "your child needs neither" «बच्चे को इनमें से कोई नहीं चाहिए»;
  - the ergative of a transitive verb: «बच्चे ने विकल्प चुना», «बच्चे ने कैसे सोचा»;
  - the subjunctive, which shows no gender: "until your child commits to one" «जब तक बच्चा कोई एक जवाब न चुन ले»; "if the child insists" «बच्चा ज़िद करे, तो»; "while your child solves" «जब तक बच्चा हल करे»; "so that the child thinks it through" «ताकि बच्चा खुद सोचे»;
  - the passive, or the thing as the subject, as every trap of the card is named: «सूची बनाते समय कोई स्थिति छूट गई» (`trap.missed_case`), «एक ही चीज़ दो बार गिनी गई» (`trap.double_count`), «“नहीं” शब्द पर ध्यान नहीं गया» (`trap.ignored_not`), «हल आधे रास्ते में रुक गया» (`trap.stopped_early`). So "the child uses one clue and passes over another" is «एक सुराग पर ध्यान जाता है और दूसरा छूट जाता है»; "the child works out who the first islander is and stops there" «पहले द्वीपवासी का पता चलते ही हल वहीं रुक जाता है»; "a tap records the answer" «जवाब एक टैप से दर्ज होता है»;
  - an invariant participle: «जवाब देते ही», «हल करते समय», «लिखते हुए».

  The pronouns «वह», «उसे», «उसका», «उसने» show no gender and stand for the child freely. Children in general, «बच्चे», may be the subject of a verb in the generic masculine plural: «बच्चे कहाँ गलती करते हैं» (`topic.traps`), «बच्चे अक्सर किस्मत पर भरोसा करते हैं»; "every child went to school" goes into that plural, «सभी बच्चे स्कूल जाने लगे». A group of children inside a task is the task's own, in the generic masculine plural: «24 बच्चे», «150 छात्र», «कुछ बच्चे शतरंज क्लब जाते हैं».
- **Speaking to the child** — the quotes under «क्या कहें», the questions of a tip, Pólya's questions, the parent's lines on "Why", a card's hint — is the card's informal «तुम», in the imperative, as «एक जवाब चुनो» (`task.pick`), «कल के लिए एक मन्नत माँगो» (`waiting.limit`), «देखो, इसे ऐसे हल करते हैं।» (`result.idk`): «सोचो», «गिनो», «बताओ», «देखो», «बनाओ». A question asks with no "you": «“कितने खंभे हैं?”», «“कौन-सा सुराग अभी इस्तेमाल नहीं हुआ?”». "Let's" is «चलो» and the subjunctive: «“चलो, इन्हें नियम से दोबारा लिखें।”». Never «तुम कर सकते हो», «तुम पक्के हो», «तुम जीतोगे»; the ergative («तुमने सही चुना») and the dative («तुम्हें पक्का पता है?») show no gender and may stand. "And if you are as unlucky as can be?" is «“और अगर किस्मत बिलकुल साथ न दे तो?”».
- **The child's own words**, in the first person, take only forms that show no gender: the ergative, «मैंने फिर गलती की», «मैंने एक संख्या सोची»; the dative, «मुझे लगता है», «मुझे पक्का पता है», «मुझे ये सवाल क्यों चाहिए? किसी ने तो दिए भी नहीं।»; «मुझसे फिर गलती हो गई…»; «हो गया! जवाब पाँच है।». Never «मैं गया», «मैं गई», «मैं कर रहा हूँ», «मैं पक्का हूँ», «मैं हार गया». "I might take two red ones straight away" is «शायद मुझे पहली बार में ही दो लाल गेंदें मिल जाएँ।». A child's question to the parent asks about the words, not the person: «“क्या आपकी बात झूठी है?”», not «“क्या आप झूठे हैं?”», which shows the parent's gender.
- **What a parent types in the chat** speaks to the chat in the polite imperative, which shows no gender: «एक नया सवाल दें», «कक्षा 3 के लिए ओलंपियाड का एक सवाल दें», «“सच्चे और झूठे” विषय पर एक सवाल दें», "Let's do a MathTrail task" «चलिए, MathTrail का एक सवाल हल करें।». In the parent's own message the child is «मेरा बच्चा», never a subject: "How is my child doing?" is «मेरे बच्चे की प्रगति कैसी है?», not «मेरा बच्चा कैसा कर रहा है?». A question about a step asks about the step: «एक खंभा ज़्यादा क्यों?», «6 क्यों नहीं?».
- **An example's question is a task for the child**, as a task on the card is: it asks with no "you", or in the «तुम» imperative: «कितने खंभे हैं?», «कौन-सी नाव जीती?», «बड़ी संख्याओं को गुणा किए बिना निकालो।». The steps of a solution speak in the card's impersonal present, as its «देखो, इसे ऐसे हल करते हैं।»: «6 स्टॉपों को बिंदुओं से बनाते हैं।», «पहले सुराग को उलटते हैं…», «मान लेते हैं कि A सच्चा है।»; or in plain statements: «4:00 के बाद 25 मिनट बचते हैं।». An answer is a short sentence with the numbers of the English one: «15 मिनट।», «हरी नाव जीती।», «A और B दोनों झूठे हैं।». The notes speak to the reader, in statements or in the polite imperative.
- **The people of the examples** have the gender the English gives them, and their verbs show it: «कविता हर 3 दिन पर फूलों को पानी देती है», «मधु ने एक संख्या सोची।», «रवि कहता है कि उसके 15 अंक हुए।». Where the English gives none, "People's names" chooses the name and the gender. Mum, Dad and Grandma take the honorific plural Hindi gives elders: «माँ 30 साल की हैं», «पापा वर्षा से 28 साल बड़े हैं», «दादी बाज़ार में अंडे बेचती हैं». The islanders are letters, in the generic masculine, and speak in it: «A झूठा है», «मैं सच्चा हूँ», «मेरा दाईं ओर वाला पड़ोसी झूठा है।»; a knight with a name may be a girl: «अदिति सच्ची है».
- **"The parent"** is the card's «अभिभावक» (`progress.for_parent`), "a parent or a tutor" the card's «अभिभावक या ट्यूटर» (`first_run.lead`), "an adult" the card's «बड़ा व्यक्ति», as in «इसे कोई बड़ा व्यक्ति सेट करता है». These generic role nouns may take a verb in the generic masculine on the technical pages and in the documents, as the card's «सेट करता है» does: «बड़ा व्यक्ति एक बार साइन इन करता है». A turn without one reads better where it fits: «साइन इन सिर्फ़ एक बार होता है».
- **"We"**, the family behind the project and ALT EdTech in the documents, takes the masculine plural Hindi gives a mixed group: «हम सब कुछ पढ़ते हैं», «हम कुछ भी स्टोर नहीं करते». On "About" the four of us are named people, and the verbs show who each is: «माँ … सँभालती हैं», «पापा … लिखते हैं», «बड़ा बेटा … हल करता है». The rule about the child is about the reader's child.
- **"Together"** is «साथ मिलकर», which does not change: «सूची माँगें और साथ मिलकर जाँचें».
- **Names and nouns that take a verb.** MathTrail, Claude, ChatGPT and the chat are masculine, as the card writes «MathTrail बच्चे को … देता है» (`first_run.lead`), «MathTrail उन्हें … चुनता है» (`review.step_practice`), «चैट क्या पूछेगा» (`first_run.fill_label`) and «चैट आपको दे सकता है» (`data.export_where`): «आपका चैट», «मॉडल लिखता है, MathTrail जाँचता है». Drive is feminine, as the card's «आपकी Google Drive» (`data.export_file`) and Google's «अपनी Drive». Google, OpenAI, Anthropic and ALT EdTech are masculine, or worded so that they take no verb: «जैसा OpenAI बताता है». The nouns the pages lean on:
  - masculine: «सवाल», «जवाब», «हल», «संकेत», «जाल», «विषय», «पड़ाव», «तरीका», «सुराग», «कार्ड», «मॉडल», «ऐप», «खाता», «टोकन», «सर्वर», «पेज», «अनुरोध»;
  - feminine: «सेवा», so "the service keeps nothing" is «सेवा कुछ नहीं रखती» while «MathTrail कुछ नहीं रखता»; «प्रोफ़ाइल», as the card's «प्रोफ़ाइल नहीं बनी» (`first_run.not_created`); «फ़ाइल», «रेटिंग», «जाँच», «गलती», «कठिनाई», «कक्षा», «स्थिति», «पट्टी», «कॉल», «कुकी», «कुंजी», «सील», and «चाल», as the card's «पहली चाल» (`trap.first_move_assumed`).
- **What a topic teaches**, `hero.teaches`, goes on from its label «क्या सिखाता है» (`topic.teaches`, whose subject is the topic) with an infinitive phrase, a second infinitive joined by «और»: «सुरागों को जोड़कर एक तस्वीर बनाना, और तुलना को दोनों ओर से पढ़ना»; «घंटे और आधी रात के पार समय गिनना, और घड़ी के समय को बीते समय से अलग पहचानना». The main move, `hero.method`, is an infinitive phrase too, as the card's advice is («…एक छोटा-सा चित्र बनाना: … फिर दोनों को गिनना।», `advice.off_by_one`): «एक रेखा खींचना और सबको उस पर रखना», «छोटी स्थिति का चित्र बनाकर उस पर गिनना». Where the English quotes, it quotes: «“मान लेते हैं कि यह सच्चा है…”». The three steps of a method (`method.steps`) are infinitives as well: «हर सुराग को एक ही दिशा में मोड़ना». None of these ends with a danda.
- **Headings.** A heading that tells the reader to do something is a polite imperative, as the dictionary's «इस विषय पर सवाल माँगें» (`topic.ask`) and «पहले सीखें» (`topic.before`): «सवाल माँगें», «परिवार को कतार में खड़ा करें». A heading about what the child does is a noun phrase: "Answer with a button" is «बटन से जवाब»; the moments on "Why", "Asks “what for?”", "Gets it wrong", "Gets it right", are «सवाल: “किसलिए?”», «गलत जवाब के बाद», «सही जवाब के बाद». "How to solve: …" (`method.title`) is the dictionary's «हल कैसे करें» (`topic.solving`) and then a noun phrase or an infinitive: «हल कैसे करें: रेखा खींचना», «हल कैसे करें: नियम से सूची», «हल कैसे करें: “मान लेते हैं कि…”».
- **The patterns, once for all:**
  - "Ask for a task": a heading «सवाल माँगें»; a sentence to the reader «चैट में सवाल माँगें», or impersonally «चैट में सवाल माँगा जा सकता है».
  - "Tap the bulb": to the reader «बल्ब वाला बटन दबाएँ»; of the child «बल्ब वाले बटन से संकेत खुलता है».
  - "You sit alongside; your child taps": «आप पास बैठें, और टैप बच्चा करे।».
  - "Write in the chat", "Press **Add**": «चैट में लिखें», «**Add** दबाएँ».
  - "Say to the child: …": «बच्चे से कहें: “…”», and the quote speaks with «तुम» or with no "you": «“तीन स्टॉप बनाओ। उनके बीच कितने अंतराल हैं?”».
  - A question to the child: «“क्या रेखा पर आगे कोई और है?”», «“यह घड़ी का समय है या बीता हुआ समय?”».
  - A tip for home: the title a polite imperative or a noun phrase, «परिवार को कतार में खड़ा करें», «जवाब से पहले रेखा»; the text in the polite imperative, what the child does in the subjunctive or the dative: «घर के सभी लोगों को पहले लंबाई के हिसाब से, फिर उम्र के हिसाब से खड़ा करें, और हर सुराग ज़ोर से बोलें: “पापा माँ से लंबे हैं, इसलिए पापा पहले खड़े होंगे।”»; «…और बच्चे को अंदाज़ा लगाने दें»; «जब बच्चा स्थितियाँ लिखे, तो पूछें: “इन्हें किस क्रम में लिखा है?”».
  - A trap's `shows` puts the thing that went wrong in the subject, as the card's trap names do; its `say:` speaks to the child.
- **Sentences every topic's page shares** read the same on every page:
  - `method.lead` ends «नीचे दिए उदाहरण सबसे आसान से ओलंपियाड के स्तर तक जाते हैं।». «नीचे» is the card's word (`task.another_coming`).
  - `description` begins with the topic's name in quotes, «विषय» and the kind of problem, on every page alike, as the rule below for a name before a postposition says: «“घड़ी” विषय के सवाल क्या होते हैं», «“सच्चे और झूठे” विषय के सवाल…». Where the English adds the words parents search by, in brackets or beside the topic's own, the Hindi adds the Hindi parents' words in the same place, from "Search words": «“अंतराल और किनारे” विषय के सवाल (…) क्या होते हैं». It ends «…, इन्हें कैसे हल करते हैं, बच्चे कहाँ गलती करते हैं और घर पर कैसे मदद करें।», the last two clauses the dictionary's headings (`topic.traps`, `topic.home`). The ending is 53 units of the checker's 160, so the beginning has about a hundred.
  - `basis.title` is «… के नियम»: «रेखा के नियम», «द्वीप के नियम».
  - `idea.note.title` is «यह क्यों ज़रूरी है»; a note headed "Why not 18?" is «18 क्यों नहीं?»; "Check:" in a step is «जाँच:».
  - `ask` is «“{topic}” विषय पर एक सवाल दें», the card's name in quotes, as its own «“{topic}” विषय पर» (`topic_choice.ask`): «“घड़ी” विषय पर एक सवाल दें», «“क्रम में लगाना” विषय पर एक सवाल दें». The English sometimes paraphrases the name ("a task on putting things in order"); the Hindi never does.
- **The card's words, exactly.** A topic's name written out in a sentence is the card's (`topic.*`), and it stands in “…”, as the card writes «“{topic}” विषय पर»: three of the names are verb phrases that a sentence would read as verbs («क्रम में लगाना», «सभी विकल्प गिनना», «तौलना और उड़ेलना»), and a plural one would change its ending before a postposition («सच्चे और झूठे» → «सच्चों और झूठों»). So the name keeps its letters inside the quotes, and «विषय» follows it where a postposition comes: «“सच्चे और झूठे” विषय के सवाल», «“सभी विकल्प गिनना” विषय “कबूतरखाना सिद्धांत” और “सच्चे और झूठे”, दोनों के लिए ज़रूरी है». A trap's name in a sentence is the card's too (`trap.*`), in quotes: «“अंतराल गिनने में एक की चूक”». Never a paraphrase, though the English sometimes paraphrases. A topic page's `title` is the card's name alone, with no quotes.

## Typography

Measured with Node 24's Intl for `hi` (ICU 78, CLDR 48) and checked character by character against the card's dictionary.

- **Sentence marks:** a sentence ends with the danda «।» (U+0964), as the card's sentences do; «?» and «!» are the ASCII marks, as are the comma, the colon and the semicolon. No space before a mark and one after it; never the double danda «॥». A heading, a label, a chip, an item of a list and the words of a drawing take no danda where the English has no full stop. A decimal keeps its point: «0.2».
- **Quotes:** “…”, as the card (`topic_choice.ask`, `profile.interest_remove`); a quote inside a quote ‘…’. No space inside the marks.
- **A sentence that ends with a quote:** a quoted sentence keeps its own mark inside the closing quote, and the sentence takes none after it, as Hindi prints dialogue: «A कहता है: “B झूठा है।”», «पूछें: “इन्हें किस क्रम में लिखा है?”». A quoted word or phrase is followed by the sentence's own mark: «इसे “जाल” कहते हैं।». A `say:` that is the quote alone is the quote and nothing after it: «“कौन-सा सुराग अभी इस्तेमाल नहीं हुआ?”», «“तीन स्टॉप बनाओ। उनके बीच कितने अंतराल हैं?”».
- **Dashes:** the card's em dash « — », with an ordinary space on each side, as the English pages write it: «चैट का मॉडल एक सेकंड में सवाल लिख देता है — और उतनी ही जल्दी चूक भी जाता है». The card puts a no-break space before some of its dashes; the site's texts type none. Ranges take the en dash with no spaces, as the card's «कक्षा {from}–{to}» (`progress.grades`): «1–6», «70–85%», «1509–1511», «9वीं–19वीं सदी». The ellipsis is one character, «…», as the card's «जाँच रहे हैं…». A hyphen joins the pairs Hindi joins: «सम-विषम», «बारी-बारी», «कौन-सा», «छोटा-सा», «एक-तिहाई», «कदम-दर-कदम», and the noun «साइन-इन».
- **Digits:** 0–9 alone, everywhere. Intl writes `hi` in Latin digits, and the site's checker refuses any other: never the Devanagari digits, U+0966 to U+096F, which a Hindi keyboard may type.
- **Numbers in the maths** — the examples, the tables of a topic's page, the sums of its cards and tips — keep the English digits and no separator, as the Russian and the Spanish do: «2026 × 2026», «1111», the arithmetic table's «1900», the percentages tip's «1200», «0.5 + ½ = 1».
- **Numbers outside the maths** as `new Intl.NumberFormat("hi")` writes them, in the Indian grouping, from four digits on: «12,450» (the bait's coins), «1,23,456»; «0.2», so the formula reads «P = 0.2 + 0.8·σ(θ + δ − β)», as the English. Years are never grouped: «1894», «2009». A count in millions goes into lakhs and crores, as Intl's compact form writes it (`notation: "compact"`, `compactDisplay: "long"`): "19 million" is «1.9 करोड़», as the Japanese «1900万人» follows its own counting.
- **A noun counted by digits** is plural from 2 on: «1 खंभा», «5 खंभे»; «1 टॉफ़ी», «9 टॉफ़ियाँ»; «सवाल», «जवाब», «दिन», «छात्र» do not change in the direct case. Before a postposition the plural is oblique: «5 खंभों के बीच», «20 सवालों के बाद».
- **Ordinals:** words outside the examples, agreeing with the noun: «पहला, पहली, पहले», «दूसरा», «तीसरा», «चौथा», «पाँचवाँ, पाँचवीं, पाँचवें», «छठा», «सातवीं बार», «दसवाँ». In an example — its question, steps, note and answer — an ordinal the English writes in digits stays in digits, its suffix joined with no space or hyphen and agreeing with the noun, in the forms CLDR's ordinal categories for `hi` give:

  | Number | Masculine | Feminine | Oblique, plural |
  |---|---|---|---|
  | 1 | 1ला | 1ली | 1ले |
  | 2, 3 | 2रा, 3रा | 2री, 3री | 2रे, 3रे |
  | 4 | 4था | 4थी | 4थे |
  | 6 | 6ठा | 6ठी | 6ठे |
  | any other: 5, 7–20, 21, 23, 100 | 5वाँ, 23वाँ | 5वीं, 9वीं | 5वें, 10वें |

  So «4थी बार पानी देना», «6ठे स्टॉप के बाद», «9वीं टॉफ़ी», «10वाँ मोज़ा», «23वाँ झंडा», «5वाँ खाना», «25वाँ दौर», «4था स्थान», «10वीं चाल». In a drawing the number stands alone: «21, 22, 23». The badge "2nd" is the word «दूसरा»: its answer, "The second player.", has no number, and the build lets a badge without one through.
- **Floors** are numbered as the English numbers them, the number after the noun: «मंज़िल 1», «मंज़िल 3», «मंज़िल 9», in the question, the drawing and the tip alike. Never «भूतल», the ground floor, and never «पहली मंज़िल», which in India is the floor above the ground: the task counts the flights from floor 1, where both girls start.
- **Dates** as Intl writes them: «3 मई», «10 अक्टूबर 2026»; "Last updated: 10 October 2026." is «अंतिम अपडेट: 10 अक्टूबर 2026।». A day of the month is a cardinal with «तारीख»: "the 3rd" «3 तारीख», "from the 3rd to the 10th" «3 से 10 तारीख तक», "the 2nd, 9th, 16th, 23rd and 30th" «2, 9, 16, 23 और 30 तारीख», "the 25th of that month" «उस महीने की 25 तारीख». The months: «जनवरी, फ़रवरी, मार्च, अप्रैल, मई, जून, जुलाई, अगस्त, सितंबर, अक्टूबर, नवंबर, दिसंबर». The weekdays: «सोमवार, मंगलवार, बुधवार, गुरुवार, शुक्रवार, शनिवार, रविवार», and in a badge or a drawing Intl's short ones, «सोम, मंगल, बुध, गुरु, शुक्र, शनि, रवि»: "Wed" is «बुध», "Mon" «सोम», "Tue" «मंगल».
- **Centuries** in digits with the ordinal suffix, never in Roman numerals: «5वीं सदी», «6ठी सदी तक», «9वीं–19वीं सदी», «16वीं सदी के अंत में», «19वीं सदी का एक चित्र». Decades: «1950 का दशक». Years in digits, never grouped: «1894 से».
- **Times** as the English writes them: «4:25», «15:10», «0:00», the bait's «03:59». Prose may add «बजे» where Hindi wants it: «4:25 बजे», «यह कितने बजे खत्म होगी?»; never the «am» or «pm» that Intl's 12-hour form adds. "Noon" is «दोपहर 12 बजे», "midnight" «आधी रात». A length of time in a drawing is in Intl's short units, «40 मि॰», «4 घं॰», and `{hours} h {minutes} min` is «{hours} घं॰ {minutes} मि॰», with no comma between; in prose the words, «4 घंटे 30 मिनट», «आधा घंटा», «पंद्रह मिनट». «॰» is U+0970, the abbreviation sign Intl writes; never a full stop in its place.
- **Units** in a drawing as Intl's short ones: «12 मी», «200 कि॰ग्रा॰», «5 ली॰» (`{count} l` is «{count} ली॰»), and «7 दिन» for the calendar's "d". In prose the word, even where the English writes the symbol: «12 मीटर», «10 मीटर की सड़क», «200 किलोग्राम», «5 लीटर». On "Service", «KB» and «px» stay as the English writes them, though Intl writes «kB».
- **Percent** as Intl writes it, the digits and then «%», with no space: «20%», «70–85%». Intl's range form, «70% – 85%», is not used.
- **Money** appears only in the make-believe app on "Why" (`app.bait.hat`, `app.cards.free.price`): Indian rupees, as a price tag writes them, «₹799» for the hat and «₹0» on the free app's card. That is Intl's symbol form with `maximumFractionDigits: 0`, without the paise of «₹799.00». The Russian writes «990 ₽» and the Japanese «980円»: a local price, not a converted one.
- **Signs:** × for multiplication; − (U+2212) for minus, never a hyphen; · in the formula. **Division is always ÷**, also on the topic pages where the English writes ":" (gaps, calendar, percentages, ratios, parts, arithmetic): in India's school books ":" is the sign of a ratio, and «35 : 7 = 5» would read as one. So «8 ÷ 2 = 4», «20 ÷ 2 = 10», «30 ÷ 3 × 2 = 20», and the note "Why not 45 : 5?" is «45 ÷ 5 क्यों नहीं?». The site's dictionary already gives the drawings «÷» (`sign.divide`). A ratio keeps the colon, with a space on each side, as NCERT writes it: «3 : 4»; "in the ratio 3 to 4" is «3 : 4 के अनुपात में», "1 to 2 to 3" «1 : 2 : 3», "1 to 50" on a map «1 : 50». Fractions as the English: «3/4», «2/7», «½»; in words «आधा», «एक-तिहाई», «दो-तिहाई», «एक-चौथाई». The unknown is the Latin «x», as Hindi-medium books write it: «यश x उठाए, तो किरण 5 − x उठाती है». «>» as the English: «अनु > रवि > कविता».
- **Arrows:** «→» as the English: «और जानें →» (`topics.more`), «12 → 6 → 6 − 3 − 1 = 2», «विफल → बच्चे तक कभी नहीं पहुँचता». So are «↓», «↑» and «↗»: «पूरी स्क्रीन पर खोलें ↗». The ferry's badge "→" stays: it means the right bank.
- **Capitals:** Devanagari has none; a Latin name keeps its own case.
- **No italic.** Devanagari has none, and the site draws an emphasis upright in Hindi, as it does in Japanese, Chinese and Arabic. So a title of a work — a book, a painting, a treatise — goes in “…” wherever the English names one, plain or emphasized. Where the English emphasizes a title, the asterisks stay, outside the quotes, `*“…”*`, so that the count of emphases is the English one, which the build checks. A space or a mark follows the closing asterisk, never a letter: `*“…”* (1945)`, `*“…”* नाम की किताब में`. An emphasis that closes after a quote mark and runs straight into a letter is not read as one, and the page would show its asterisks.
- **Bold, at its edges.** A bold phrase that ends with a danda closes before the space after it: `**मान लेते हैं कि B सच्चा है।** तब…` works, while `**विरोधाभास।**अब` shows its asterisks. A bold label that begins with a sign has a space before it: «**+ Add** दबाएँ», «**+** → **Connectors** दबाएँ».
- **Letters as labels.** The islanders, the children of enumeration, the boxes of ratios and the card's options keep the Latin capitals A, B, C…, as Hindi reasoning books and NCERT write labels and persons, in the words and the drawings alike: «A कहता है: “मैं झूठा हूँ, और B सच्चा है।”», «डिब्बा A», «पाँच विकल्प, A से E तक». Enumeration's lost case reads «BC?», and two letters of one case stand together as the English writes them: «AB और BA दो अलग कतारें हैं». The knight and the liar are «स» and «झ», the first letters of the card's «सच्चे और झूठे», as `thumb.knight` and `thumb.liar` draw them: in the drawings (`art.knight`, `art.liar`), on the rules' cards (`basis.cards[].mark`), in "K, L, K, L…", «स, झ, स, झ…», and in the tip "Write K and L in pencil", «पेंसिल से स और झ लिखें». A knight is «सच्चा», a liar «झूठा», a girl «सच्ची», «झूठी». A person's initial in a drawing or a badge is the first letter of the Hindi name (see "People's names"). The letter for "your child" on "Why" is «ब», for «बच्चा», and the parent's is «आप».
- **The card questions on the home page and on "Why"** write no Latin letter, as the English writes none there: the picture check reads a lone Latin capital in them as a label of the picture.
- **The counted nouns of "Research"**, `research.*` in the site's dictionary, are written in the direct case. A slot that carries one — `{count} {noun}`, or a single slot such as `{trial}`, `{children}`, `{answers}`, `{after}` — stands where Hindi wants the direct case: as a subject, as an object without «को», alone, or before a colon; never before a postposition, where «बच्चे», «जगहें», «जाँचें», «सवाल», «जवाब» would need the oblique «बच्चों», «जगहों», «जाँचों», «सवालों», «जवाबों». So:
  - `hero.lead`: «… {count} {noun} पास करने के बाद ही …», not «{count} {noun} के बाद»;
  - `model.rows.error_static.name`: «{after} दर्ज होने पर अनुमान की चूक», not «{after} के बाद»; «दर्ज» and «होने» do not change;
  - `theses.reach.text`, `live.ready.weighed`: «शुरुआती दौर, यानी पहले {trial}, …», not «{trial} का शुरुआती दौर»;
  - `live.rule`: «जब उसमें कम से कम {children} और कम से कम {answers} हों»;
  - `live.ready.month`: «{month}: नियम के चुने सवालों पर {children} और {answers}।»;
  - `sources.total`: «कुल {count} {noun}»; `hero.paper`: «… · PDF, {count} {noun}»; `model.run.size`: «{children} × {answers}».

  A verb may agree with such a slot only where the count is above one on every page, as the thresholds of `live.rule` are. CLDR's Hindi "one" covers both 0 and 1, so a count of 0 takes the "one" form, «0 बच्चा»; no sentence of the page shows a count of 0, since a month with too few children shows its own frame. `research.places` stands in `theses.twice.text` as a distance, «{count} {noun} आगे», and Hindi says «2 जगह आगे», so the dictionary gives «जगह» in both forms.
- **YAML:** a text that holds `: ` or begins with `*`, `{`, `[`, `]`, `&`, `!`, `|`, `>`, `%`, `@` or a backtick is quoted. A text with a ratio, «3 : 4», holds `: ` and is quoted too. The danda, Devanagari and “” need no quotes, and a text may begin with “.
- **What a text never types:** a no-break space, a zero-width joiner or non-joiner (Hindi needs neither), a mark of direction, a soft hyphen, a Devanagari digit or a precomposed nukta letter.

## Platforms' labels

A quoted label is written exactly as its platform writes it, in its spelling, though the site's own sentences follow the card's: Google writes «हटाएं» with an anusvara where the card writes «हटाएँ», and a label keeps Google's. A label is set apart as the English sets it apart — in bold where the English bolds it, in quotes or as a link's text where the English does that — since the build holds every text to the English one's count of emphases; the sentence around it speaks to the reader in the polite imperative: «**+ जोड़ें** दबाएँ और **कस्टम कनेक्टर जोड़ें** चुनें।». No sentence of a platform's help is copied whole: their Hindi is machine-made, ends with «.», and some of it shows the reader's gender («क्या मैं … जाँच कर सकता हूँ?»).

### Claude

Claude's interface speaks Hindi, and a browser set to Hindi gets it (claude.ai/login opens in Hindi for `hi-IN`); its help does not. The help lists Hindi among the interface's languages [C], but its own language picker offers twelve languages and no Hindi, every `/hi/` address of an article redirects to the English page, and support.claude.com/hi is an empty help center. So Claude's labels are taken from Claude's Hindi interface itself, the strings claude.ai loads for a Hindi browser [D], as the content files ask ("Claude's own names for the page, as Claude writes them in the page's language", the comment on `connect.connectors`). Each of [D]'s strings is keyed by a hash of its English source, so a label below marked "hash" is the very string Claude shows for that English; "meaning" marks a string matched by what it says. The strings change with Claude's releases; they were read on 2026-10-12.

- [A] "Get started with custom connectors using remote MCP", https://support.claude.com/en/articles/11175166-get-started-with-custom-connectors-using-remote-mcp (`/hi/` serves the English page)
- [B] "Use connectors to extend Claude's capabilities", https://support.claude.com/en/articles/11176164-use-connectors-to-extend-claude-s-capabilities
- [C] "How to use Claude in your preferred language", https://support.claude.com/en/articles/10769299-how-to-use-claude-in-your-preferred-language
- [D] Claude's Hindi interface strings, `https://claude.ai/i18n/hi-IN.json` and `hi-IN.overrides.json`, as https://claude.ai/login loads them for a browser set to Hindi; a direct request gets a challenge page

| English on the site | Hindi | Where it comes from |
|---|---|---|
| Customize → Connectors | **कस्टमाइज़ → कनेक्टर**: «Claude में **कस्टमाइज़ → कनेक्टर** खोलें।» In help the label is the link's text: «Claude में [कस्टमाइज़ → कनेक्टर](https://claude.ai/customize/connectors) खोलें।» | D: the pattern "Customize → {section}" is «कस्टमाइज़ → {section}» (hash), and "Connectors" «कनेक्टर» (hash); a message of Claude's writes the path whole, «इसे कस्टमाइज़ → कनेक्टर में कनेक्ट करें» (meaning). "Customize" alone is «अनुकूलित करें» (hash), which the path does not use. A: "Navigate to Customize > Connectors" |
| + / + Add | **+ जोड़ें**: «**+ जोड़ें** दबाएँ» | D, "+ Add" (hash). A: "Click "+ Add" then "Add custom connector"" |
| Add custom connector | **कस्टम कनेक्टर जोड़ें** | D: "Add custom connector…" is «कस्टम कनेक्टर जोड़ें…» (hash), and the plain words are a string of their own (meaning); the site writes no ellipsis, as the English writes none. A |
| Name it MathTrail | «इसका नाम MathTrail रखें» | D: the field "Name" is «नाम» (hash); the field is not quoted. A: "Enter a name for the connector" |
| the address | «पता», the card's word (`progress.link_refused`); once, «रिमोट MCP सर्वर का URL» | D: "MCP server URL" is «MCP सर्वर URL» (hash). A: "your connector's remote MCP server URL" |
| Continue, and Continue again | **जारी रखें**: «**जारी रखें** दबाएँ, और फिर से **जारी रखें**» | D, "Continue" (hash). A quotes it after the address and again after the sign-in settings |
| the sign-in settings Claude detects | described: «Claude साइन-इन की जो सेटिंग खुद पहचानता है, उन्हें बदलने की ज़रूरत नहीं है।» | D names the panel «प्रमाणीकरण» ("Authentication", hash); the site describes it, as the English does. A: "Review the authentication settings Claude detected" |
| Add | **जोड़ें**: «फिर **जोड़ें** दबाएँ» | D, "Add" (hash). A: "Finish configuring your connector by clicking "Add"" |
| Connect | **कनेक्ट करें**: «उसके बाद **कनेक्ट करें** दबाएँ और Google से साइन इन करें» | D, "Connect" (hash). B: "Click "Connect" or "Install"" |
| + → Connectors, in a chat | «किसी चैट में **+** → **कनेक्टर** दबाएँ, MathTrail चालू करें और लिखें: “…”» | D: "Connectors" is «कनेक्टर» (hash). A and B: "the "+" button on the lower left of your chat interface, then "Connectors"" |
| the free plan | «Free प्लैन»: «Free प्लैन में एक ही कस्टम कनेक्टर जोड़ा जा सकता है।» The plan's name stays Free | D: "Free plan" is «Free प्लैन» (hash, in the overrides). A: "Free users are limited to one custom connector" |
| Always allow | **हमेशा अनुमति दें** | D, "Always allow" (hash); B writes "Always allow", A's section on security "Allow always". Not quoted on the site yet |
| "Use caution before running this prompt" | «इस प्रॉम्प्ट को चलाने से पहले सावधानी बरतें», in quotes: «“इस प्रॉम्प्ट को चलाने से पहले सावधानी बरतें” वाली चेतावनी» | D (meaning): the warning's first sentence. Not in Claude's help in any language |
| Claude asks whether to allow MathTrail to save | described: «प्रोफ़ाइल सेव करने से पहले Claude पूछता है कि MathTrail को इसकी अनुमति दी जाए या नहीं।» | the question is not quoted in Claude's help |

Claude takes a masculine verb, as MathTrail does: «Claude … पहचानता है», «Claude पूछता है». Claude's English help disagrees with itself since 1 October 2026, "+ Add → Add custom connector" in A against "Add → Custom → Web" in B's own section, and "Allow always" against "Always allow"; the site follows A's steps, as every language does.

### ChatGPT

OpenAI's help speaks Hindi, and each of its Hindi pages says it was machine-translated («यह पेज मशीन द्वारा अनुवादित है.»). help.openai.com refuses fetchers; the pages below were read in a browser on 12 October 2026.

| English on the site | Hindi | Source |
|---|---|---|
| developer mode | **डेवलपर मोड** | https://help.openai.com/hi-in/articles/12584461-developer-mode-and-mcp-apps-in-chatgpt — «ChatGPT में डेवलपर मोड और MCP ऐप्स». The path to turn it on: «सेटिंग → ऐप्स → उन्नत सेटिंग»; to add an app, «ऐप्स → बनाएँ», then «बनाएँ पर क्लिक करें» |
| ChatGPT's directory | «ChatGPT की डायरेक्टरी»; its name now «प्लगइन डायरेक्टरी» | https://help.openai.com/hi-in/articles/20001256-plugins-in-chatgpt — «ChatGPT में प्लगइन»: «प्लगइन डायरेक्टरी सभी ChatGPT प्लान पर उपलब्ध है.» The app directory became the plugin directory on 9 July 2026 (see [the Spanish notes](es.md)); the generic «ChatGPT की डायरेक्टरी» stays true |
| apps · connected apps | «ऐप», «ऐप्स» · «कनेक्टेड ऐप» | https://help.openai.com/hi-in/articles/11487775-connected-apps-in-chatgpt — «ChatGPT में कनेक्टेड ऐप» |
| a plan · Free, Plus, Pro, Business, Enterprise, Education | «प्लान» · the same | the plans' names stay; the help writes them in Latin letters in most places |

A sentence that quotes the path: «**सेटिंग → ऐप्स → उन्नत सेटिंग** में **डेवलपर मोड** चालू करें, फिर **ऐप्स → बनाएँ** से ऐप जोड़ें।»

### Google

Read with `?hl=hi` on 12 October 2026; each page says its Hindi may be machine-made.

| English on the site | Hindi | Source |
|---|---|---|
| Google Drive | **Google Drive**, in Latin letters, feminine, as the card writes «आपकी Google Drive» (`data.export_where`) and Google «अपनी Drive». My Drive is «मेरी ड्राइव». Google's older Hindi name, «Google डिस्क», is not used | https://support.google.com/drive/answer/2375102?hl=hi («Google Drive की फ़ाइलें मिटाना»), https://support.google.com/drive/answer/2409045?hl=hi («बाईं ओर, मेरी ड्राइव पर क्लिक करें.») |
| Drive's bin | «Drive का ट्रैश», the card's phrase (`data.delete_how`); the button **ट्रैश खाली करें**; "Restore" **वापस लाएं**; files go for good after 30 days, «ट्रैश में भेजी गई फ़ाइलें, 30 दिनों के बाद हमेशा के लिए मिट जाती हैं.» | 2375102; https://support.google.com/drive/answer/1716222?hl=hi |
| version history | «फ़ाइल के पिछले वर्शन»; the menu **वर्शन मैनेज करें** and **हमेशा के लिए सेव रखें** ("Keep forever"). «वर्शन इतिहास» is the history of Docs, Sheets and Slides, which the profile's file does not have | 2409045 («फ़ाइल के वर्शन और गतिविधि देखना»); https://support.google.com/docs/answer/190843?hl=hi |
| linked apps, third-party access | the page Google names «'कनेक्ट किए गए ऐप्लिकेशन' पेज» (myaccount.google.com/linkedapps), which the site sets apart as the English does, as a link: «अपने Google खाते के [कनेक्ट किए गए ऐप्लिकेशन](…) पेज पर जाएँ»; the section «आपके Google खाते का ऐक्सेस», then **पूरी जानकारी देखें**, **ऐक्सेस हटाएं** and **पुष्टि करें**, which a sentence joins with «→». The card's older words, «तीसरे पक्ष के ऐप वाले पेज» (`data.access_how`), are not used on the site. "Delete all connections" is on no page, and the site needs only «ऐक्सेस हटाएं» | https://support.google.com/accounts/answer/13533235?hl=hi («अपने Google खाते और अन्य डेवलपर के ऐप्लिकेशन के बीच के कनेक्शन मैनेज करना») |
| the box for Google Drive on the consent screen | the box of Google Drive. Google's Hindi wording of the permission, «…आप Google Drive की जिन फ़ाइलों का इस्तेमाल करते हैं…», shows the reader's gender and is not copied; the English never quotes it | Google's Workspace Marketplace, `?hl=hi`; not seen on the live consent screen |
| Sign in with Google | **Google से साइन इन करें**, the button's words, which a sentence takes as they are | https://developers.google.com/identity/branding-guidelines?hl=hi |
| Google account | «Google खाता», as the card writes «अपने Google खाते में» (`data.access_how`) | 13533235 |
| Enter, the key | «Enter दबाएँ» | https://support.google.com/mail/answer/6594?hl=hi («Enter दबाएं») |

MathTrail's own sign-in pages speak English, so "Tick the box for Google Drive" and **Back to Google** stay English.

## Glossary

"card" marks the card's own word, from `web/locales/hi.json`; "site" a word of the site's dictionary, `web/src/site/locales/hi.json`.

### Product and lesson

| English | Hindi | Note |
|---|---|---|
| task | सवाल | card, `task.label`; plural «सवाल», oblique «सवालों» |
| problem (olympiad) | सवाल | the same word as task; never «समस्या», which reads as a trouble |
| puzzle, riddle | पहेली | plural «पहेलियाँ»; "logic puzzles" «तर्क की पहेलियाँ» |
| school sum | स्कूल का सवाल | "A school sum" is «स्कूल का सवाल»; "a usual lesson" «आम पाठ» |
| topic | विषय | card; plural «विषय» (`progress.topics`) |
| group of topics | समूह | `group.*` exactly; "Groups of topics" «विषयों के समूह» |
| trap | जाल | card, `result.trap`; plural «जाल» |
| the catch (not a trap) | पेच | never «जाल» outside the card's sense: "the catch is that an hour has 60 minutes" is «पेच यह है कि घंटे में 60 मिनट होते हैं, 100 नहीं» |
| technique | तरीका | plural «तरीके», as `nav.techniques` and `techniques.count` |
| the main move (`topic.method`) | मुख्य विचार | site |
| hint | संकेत | card, `task.hint`; never for "clue" |
| clue, a given of a puzzle | सुराग | plural «सुराग»; keeps «संकेत» for the card's hint: «हर सुराग एक तीर है» |
| condition | शर्त | card, `trap.ignored_condition` |
| statement, what an islander says | कथन; बात | card: «कथन» (`advice.trusted_statement`), «बात» (`trap.trusted_statement`) |
| option | विकल्प | card, `advice.double_count`; the card lists them under «जवाब» (`result.answers`); "Five options, A to E" is «पाँच विकल्प, A से E तक» |
| answer | जवाब | card |
| right, wrong answer | सही जवाब, गलत जवाब | card: «सही!», «पूरी तरह सही नहीं» |
| solution | हल | card, `result.solution` |
| explanation | समझाना; व्याख्या | a verbal noun, as the card's «हर गलती को समझाता है» (`first_run.lead`): "the explanation starts from…" is «समझाना वहीं से शुरू होता है…»; «व्याख्या» where the technical pages count the model's explanations |
| diagnosis, of a mistake | निदान | "A mistake is a diagnosis" is «गलती एक निदान है» |
| step, of a solution or a method | कदम | card, `advice.stopped_early`; "step by step" «कदम-दर-कदम»; a layer of the topics' map is «चरण» |
| question, the text of a task | सवाल | card: «सवाल की कोई संख्या ही जवाब मान ली गई» (`trap.number_from_text`) |
| question, what is asked | सवाल क्या पूछ रहा है | card, `advice.answered_other_question`: "the question asks for minutes" is «सवाल मिनट पूछ रहा है» |
| picture of a task, a drawing | चित्र | card, `advice.off_by_one`; «फ़ोटो» for a photograph, «पेंटिंग» for a painting |
| label of a picture | लेबल | card, `profile.grade_note` |
| grade | कक्षा | card, `profile.grade` |
| Grade 3 | कक्षा 3 | card, `child.grade` |
| grades 1 to 6 | कक्षा 1 से 6 | card, `first_run.lead` |
| grades 3–4 | कक्षा 3–4 | card, `progress.grades`; site, `topics.grades` |
| level | स्तर | card, `review.step_practice` |
| difficulty | कठिनाई | card, `waiting.step.topic` |
| rank | पड़ाव | card, «पड़ाव {rank}»: a number, never a title; "11 ranks" «11 पड़ाव» |
| rating | रेटिंग | card, feminine: «इस विषय में रेटिंग», «कुल रेटिंग» |
| progress | प्रगति | card, «प्रोफ़ाइल और प्रगति» |
| mastered · mastery | सीख लिया · महारत | card, `progress.mastered`: "a topic counts as mastered" is «विषय सीख लिया गया माना जाता है»; the noun of "Research" is «महारत» |
| review, the card's section | समीक्षा | card, `review.title`; "what goes well and what to develop" is the card's «मज़बूत पक्ष», «जिन पर काम करना है» |
| how the answer went | जवाब का विश्लेषण | card, `result.label` |
| trial series | शुरुआती दौर | card, `result.trial` |
| skipped, a skip | छोड़ा गया | card, `progress.skipped` |
| profile | प्रोफ़ाइल | card, feminine |
| pseudonym | छद्म नाम | card |
| interests | रुचियाँ | card, `profile.interests` |
| skill | कौशल | card, `problem.not_in_catalog` |
| notes, the parent's about the child | नोट्स | «अंक» is a school mark |
| lesson | पाठ | card, «पाठों की भाषा» (`profile.language`) |
| coach | कोच | card, `topic_choice.coach_short`, «ओलंपियाड कोच» (`app.badge`); the Coach product too |
| companion, in the coach | साथी | a character, not the child; "the fox companion" «लोमड़ी साथी» |
| parent | अभिभावक | card |
| tutor | ट्यूटर | card |
| adult | बड़ा व्यक्ति | card, `first_run.lead` |
| child, your child | बच्चा; आपका बच्चा | card; see "The child" |
| children | बच्चे | card; counted on "Research", «बच्चा», «बच्चे» (site) |
| pupils, in a task | छात्र | the plural the same, oblique «छात्रों»: «300 छात्र» |
| class, of pupils | कक्षा | the same word as grade, as Hindi has it: "the whole class" «पूरी कक्षा» |
| chat | चैट | card, masculine: «आपका चैट» |
| card | कार्ड | card |
| connector, custom connector | कनेक्टर, कस्टम कनेक्टर | Claude's own words (see Platforms' labels); masculine |
| developer mode · ChatGPT's directory · the plans | see Platforms' labels | the plans' names stay: Free, Plus, Pro, Business, Enterprise, Education |
| Claude's message box · "Use caution before running this prompt" · Always allow | see Platforms' labels | |
| Google Drive · My Drive · version history · linked apps · Sign in with Google · Enter, the key | see Platforms' labels | the card says «आपकी Google Drive» and «Google खाता» |
| app | ऐप | card, masculine |
| sign in · sign-in | साइन इन करना · साइन-इन | card, «साइन-इन वाले देश» (`profile.signin_country_off`): the noun hyphenated, the verb not |
| account | खाता | card, «अपने Google खाते में» (`data.access_how`) |
| Drive file, the profile's file | प्रोफ़ाइल की फ़ाइल | card, `data.export_file` |
| folder | फ़ोल्डर | card |
| bin | ट्रैश | card, «Drive का ट्रैश» (`data.delete_how`); see Platforms' labels |
| delete · download · copy · save | मिटाना · डाउनलोड करना · कॉपी करना · सेव करना | card |
| access · settings · disconnect | ऐक्सेस · सेटिंग · डिसकनेक्ट करना | card, `data.how_cut_off`, `data.remove_how` |
| daily limit | रोज़ की सीमा | card, `waiting.limit_detail` |
| country · state, in the United States | देश · राज्य | card |
| Another task | अगला सवाल | the card's button, quoted exactly (`task.another`) |
| Make a wish for tomorrow | कल के लिए एक मन्नत माँगो | card, quoted exactly |
| Preparing the next task… | अगला सवाल तैयार हो रहा है… | card, `waiting.title` |
| Profile & progress | प्रोफ़ाइल और प्रगति | card |
| Edit | बदलें | card, `profile.edit` |
| Don't count the country I sign in from | साइन-इन वाले देश को गिनती में न लें | card, quoted exactly in the privacy policy |
| I don't know, as an answer | मुझे नहीं पता | the dative shows no gender |
| the bulb, the hint's button | बल्ब वाला बटन | |
| the two arrows | दो तीरों वाला बटन | |
| tap | टैप करना; एक टैप | "A tap records the answer" is «जवाब एक टैप से दर्ज होता है» |
| press, click | दबाना: «दबाएँ» | |
| free, of price | मुफ़्त | the chip `hero.free` |
| Free, the plan | Free | the plan's name |
| open source | ओपन सोर्स | the chip `hero.open` |
| code, of the project | कोड | "View the code on GitHub" is «GitHub पर कोड देखें» |
| in development | तैयारी में | the badge on "Coach" |
| beta | Beta | `nav.beta` |
| soon | जल्द ही | site, `topics.soon` |
| olympiad maths | ओलंपियाड गणित | site, `title.topic`; card, «ओलंपियाड कोच» |
| olympiad-style task | ओलंपियाड जैसा गणित का सवाल | card, «ओलंपियाड जैसे गणित के सवाल» (`first_run.lead`) |
| maths circle | गणित क्लब | the word Indian schools use |
| school | स्कूल | card; not «विद्यालय» |
| school mark | अंक; नंबर | card, «स्कूल का अंक» (`progress.grades_note`); «अच्छे नंबर» in a phrase, and «नंबर» wherever «अंक» could be read as a digit |
| homework | होमवर्क | |
| subscription · ads | सब्सक्रिप्शन · विज्ञापन | |
| home page · site · page | होम पेज · साइट · पेज | site, `share.picture`, `nav.pages`; a book's page is «पन्ना» |

### The site's sections and headings

| English | Hindi | Note |
|---|---|---|
| Topics (`nav.topics`) | विषय | site; card, `progress.topics` |
| Techniques (`nav.techniques`) | तरीके | site |
| Why (`nav.why`) | क्यों | site |
| Coach (`nav.coach`) | कोच | site, beside the tag Beta; card |
| Research (`nav.research`) | शोध | site; the page's title and the footer's label too, as a page's title is its menu entry in every language; masculine, «शोध बताता है». A study the page cites is «अध्ययन» |
| Service (`nav.service`) | सेवा | site; feminine |
| About (`nav.about`) | हमारे बारे में | site; the page's title and the footer's label too |
| Add to Claude (`nav.add`) | Claude में जोड़ें | site |
| Technical pages (`nav.technical`) | तकनीकी पेज | site; said by a screen reader |
| Pages · Menu · Language | साइट के पेज · मेनू · भाषा | site, `nav.pages`, `nav.menu`, `nav.language` |
| Help · Privacy policy · Terms of use | मदद · निजता नीति · उपयोग की शर्तें | the documents' titles |
| Links | लिंक | site, `footer.links` |
| How it is solved | हल कैसे करें | site, `topic.solving` |
| Where children go wrong | बच्चे कहाँ गलती करते हैं | site, `topic.traps` |
| What to say | क्या कहें | site, `topic.say`; the quotes under it speak to the child |
| How to help at home | घर पर कैसे मदद करें | site, `topic.home` |
| Learn first | पहले सीखें | site, `topic.before` |
| Opens the way to | आगे का रास्ता | site, `topic.after` |
| Builds on · Opens | आधार: · आगे का रास्ता: | site, `topics.builds_on`, `topics.opens` |
| Ask for a task on this topic | इस विषय पर सवाल माँगें | site, `topic.ask` |
| Show · Hide the solution | हल दिखाएँ · हल छिपाएँ | the card's «संकेत छिपाओ» says it to the child |
| Example 2 | उदाहरण 2 | site, `topic.example` |
| Problem · {grades}, on "Techniques" | सवाल · {grades} | |
| Answer · Trap · Child · Right | जवाब · जाल · बच्चा · सही | site, `topic.answer`, `topic.trap`, `topic.child`, `topic.right` |
| Why it matters | यह क्यों ज़रूरी है | |
| How to solve: … | हल कैसे करें: … | then a noun phrase or an infinitive |
| The rules of … | … के नियम | «रेखा के नियम», «द्वीप के नियम» |
| Check: | जाँच: | |
| When it helps: | कब काम आता है: | the technique, «तरीका», is masculine |
| In the topics: | विषयों में: | |
| If the problem has… · …try | अगर सवाल में… · …तो आज़माएँ | the table of cues on "Techniques" |
| You are here · On this page · All topics | आप यहाँ हैं · इस पेज पर · सभी विषय | site |
| More → | और जानें → | site, `topics.more` |
| See the coach → · See the topics → | कोच देखें → · विषय देखें → | |
| Honestly | सच कहें तो | the subjunctive shows no gender |
| See a lesson | एक पाठ देखें | |
| For technical readers | तकनीकी पाठकों के लिए | the generic plural |
| You · C, the dialogue on "Why" | आप · ब | «ब यानी आपका बच्चा।» |
| The adult · The model, in the chat's illustration | बड़ा व्यक्ति · मॉडल | card, «बड़ा व्यक्ति» |
| Your chat · An illustration of the chat | आपका चैट · चैट का एक उदाहरण | the chat is masculine |
| How the next task was picked · Picked for Galt · Galt's skills · Why this task · Auto-picked | अगला सवाल कैसे चुना गया · गाल्ट के लिए चुना गया · गाल्ट के कौशल · यही सवाल क्यों · अपने-आप चुना गया | |
| Step 1, of the topics' map | चरण 1 | a step of a solution is «कदम» |
| Foundations · Rules · Hard problems | आधार · नियम · कठिन सवाल | «आधार», as `topics.builds_on`; "nothing, it is a foundation" «कुछ नहीं, यह खुद आधार है» |
| First · Topic · Then, the map's key | पहले · विषय · फिर | |
| Free · Open source, the chips | मुफ़्त · ओपन सोर्स | |
| Copy · Copied | कॉपी करें · कॉपी हो गया | site |
| Task · Answer · Sign-in, the scenarios on "Service" | सवाल · जवाब · साइन-इन | short: the tabs do not wrap at 320 px |

The menu fits with room to spare: measured with `just site-fold`, the row of the seven entries stays whole down to 910 px, the narrowest of all the site's languages, «हमारे बारे में» its longest entry.

### Maths and olympiad terms

| English | Hindi | Note |
|---|---|---|
| olympiad | ओलंपियाड | card; masculine; "the International Mathematical Olympiad" «अंतरराष्ट्रीय गणित ओलंपियाड» |
| reference task | संदर्भ सवाल | site, `research.tasks` |
| solver | सॉल्वर | the first time on a page «सवाल हल करने वाला प्रोग्राम (सॉल्वर)» |
| brute force | हर विकल्प आज़माना | "reproduces that answer by brute force" is «हर विकल्प आज़माकर वही जवाब निकालता है» |
| case | स्थिति | card, `trap.missed_case`; also the state of the jugs and a position of a game, as the card's «छोटी स्थितियों से खेल» (`advice.first_move_assumed`) |
| list · in order | सूची · तय क्रम में | card, `advice.missed_case` |
| worst case | सबसे बुरी स्थिति | card, `trap.best_case_not_worst` |
| good, bad luck · unlucky | अच्छी किस्मत, बुरी किस्मत · बदकिस्मती | card, «किस्मत पर भरोसा» |
| to be sure, for certain | पक्का | agreeing with the thing made sure, «ताकि एक संतरे वाली टॉफ़ी पक्की हो जाए»; never with a person |
| pigeonhole principle | कबूतरखाना सिद्धांत | card |
| pigeons, holes | कबूतर, खाने | the holes are the «खाने» of the card's name; "Pigeons and pigeonholes" «कबूतर और खाने» |
| parity · even · odd | सम-विषम · सम · विषम | card |
| alternation, taking turns | बारी-बारी | card |
| remainder · quotient · divisor | शेषफल · भागफल · भाजक | card; NCERT |
| divisibility · divisibility rules | विभाज्यता · विभाज्यता के नियम | card, `skill.divisibility_rules` |
| divides by 9 | 9 से विभाज्य है | NCERT; "with nothing left over" «बिना शेष के» |
| multiple · factor · common divisor | गुणज · गुणनखंड · सार्व गुणनखंड | NCERT: "3 and 5 have no common divisor but 1" is «1 के सिवा 3 और 5 का कोई सार्व गुणनखंड नहीं» |
| digit · number · two-digit number | अंक · संख्या · दो अंकों की संख्या | NCERT; «अंक» is a digit and «संख्या» a number, never swapped |
| tens, units digit | दहाई का अंक, इकाई का अंक | NCERT |
| sum · difference · product | योग · अंतर · गुणनफल | NCERT: "the sum of its digits" «अंकों का योग» |
| enumeration | सभी विकल्प गिनना | card; "counting problems" «गिनती के सवाल» |
| systematic listing | तय क्रम में लिखना | card, `advice.missed_case` |
| tree, the diagram | वृक्ष आरेख | the first time on a page «पेड़ जैसा चित्र (वृक्ष आरेख)» |
| working backwards | अंत से शुरू करना | |
| suppose · the assumption · contradiction | मान लेते हैं कि… · मानी गई बात · विरोधाभास | "the assumption you started from is wrong" is «शुरू में मानी गई बात गलत है» |
| proof · prove · proof by contradiction | प्रमाण · सिद्ध करना · विरोधाभास से प्रमाण | NCERT's «सिद्ध कीजिए» |
| pattern | पैटर्न | NCERT's own loan in its books for the first grades |
| strategy · winning strategy · symmetric strategy | रणनीति · जीतने की रणनीति · सममित रणनीति | card, `group.games`, `topic.games.strategy` |
| winning, losing position | जीतने वाली स्थिति, हारने वाली स्थिति | |
| move, in a game | चाल | card, feminine: «पहली चाल» (`trap.first_move_assumed`) |
| the first, the second player | पहला खिलाड़ी, दूसरा खिलाड़ी | |
| the other player, the opponent | सामने वाला खिलाड़ी | never «दूसरा खिलाड़ी», which is the second player |
| the player to move | जिसकी चाल हो | "loses for the player to move" is «जिसकी चाल हो, उसकी हार» |
| invariant, what does not change | जो नहीं बदलता | |
| knights and liars | सच्चे और झूठे | card |
| knight · liar | सच्चा · झूठा | «स» and «झ»; a girl «सच्ची», «झूठी» |
| island · islander | द्वीप · द्वीपवासी | the plural the same |
| ordering | क्रम में लगाना | card |
| taller, shorter · the tallest | लंबा, छोटा · सबसे लंबा | "Ann is taller than Ben" is «अनु रवि से लंबी है»; the arrow's `art.taller` «लंबा» |
| number line | संख्या रेखा | card |
| grid · squared paper | ग्रिड · खानेदार कागज़ | card, «वर्गाकार ग्रिड», «खानों का जाल» (`picture.grid`) |
| cell · side of a cell | खाना · खाने की भुजा | card, `picture.grid`, `advice.ratio_total_confusion`; "perimeter, in cell sides" «परिमाप, खानों की भुजाओं में» |
| row · column · diagonal | पंक्ति · स्तंभ · विकर्ण | NCERT; card, «पंक्ति» |
| area · perimeter | क्षेत्रफल · परिमाप | card |
| rectangle · square · L shape | आयत · वर्ग · L आकार | NCERT |
| 2 by 4, of a grid | 2 × 4 | as NCERT writes it: «2 × 4 खानों की पट्टी», «5 × 5 का बोर्ड» |
| ratio · in the ratio 3 to 4 | अनुपात · 3 : 4 के अनुपात में | card; see Typography |
| proportion | समानुपात | card, `skill.ratios` |
| sharing · share | बँटवारा · हिस्सा | card |
| percentage · percent of what? | प्रतिशत · किसका प्रतिशत? | card, `advice.percent_wrong_base` |
| 20% off | 20% छूट | |
| fraction · numerator · denominator | भिन्न · अंश · हर | card; NCERT: "the fraction's bottom number" is «हर» |
| the whole · the part | पूरा · हिस्सा | card, «हिस्से और पूरे में उलझन» (`trap.part_whole_swap`) |
| a half, a third, a quarter, two thirds, a seventh | आधा, एक-तिहाई, एक-चौथाई, दो-तिहाई, सातवाँ हिस्सा | NCERT |
| strip, bar | पट्टी | card, `picture.bars`, `advice.part_whole_swap` |
| box, of a ratio | खाना | card, «बराबर खानों में» (`advice.ratio_total_confusion`); a real box is «डिब्बा» |
| gaps and boundaries | अंतराल और किनारे | card |
| gap, span · post · fence | अंतराल · खंभा · बाड़ | card, `advice.off_by_one`; «बाड़» is feminine |
| off by one | एक की चूक | card, `trap.off_by_one` |
| floor · flight of stairs · step | मंज़िल · ज़ीना · सीढ़ी | see Typography for floors; "8 flights" «8 ज़ीने» |
| clock face · hands · the whole hour | घड़ी का डायल · सुइयाँ · पूरा घंटा | card, `picture.clock`; "step to the whole hour" «पहले पूरे घंटे तक जाना» |
| time on the clock · length of time | घड़ी का समय · बीता हुआ समय | "Is this a time on the clock or a length of time?" is «यह घड़ी का समय है या बीता हुआ समय?» |
| calendar and age | कैलेंडर और उम्र | card |
| day of the week · date | हफ़्ते का दिन · तारीख | |
| week · month · year | हफ़्ता · महीना · साल | card, «पिछला हफ़्ता» (`progress.period_week`) |
| pan balance · pan · weights | दो पलड़ों वाला तराज़ू · पलड़ा · बाट | card, `picture.balance` |
| a weighing | एक बार तौलना | counted with «बार»: «2 बार तौलना», «कम से कम कितनी बार तौलना पड़ेगा?» |
| fake coin | नकली सिक्का | |
| jug · pour · fill · empty · top up | जग · उड़ेलना · भरना · खाली करना · ऊपर तक भरना | card, «तौलना और उड़ेलना», «उड़ेलने के बर्तन» (`picture.containers`) |
| overlapping groups · Venn diagram | आपस में मिलते समूह · वेन आरेख | card; NCERT |
| arithmetic with a trick | पेचीदा अंकगणित | card |
| I thought of a number | मैंने एक संख्या सोची | the ergative shows no gender |
| carrying | हासिल | NCERT: "with no carrying at all" is «बिना किसी हासिल के» |
| round number | गोल संख्या | 10, 100, 1000 |
| doubled, halved · times as many | दुगुना, आधा · गुना | «4 गुना» |
| match, the stick | तीली | «माचिस की तीली» |
| a special case | विशेष स्थिति | |
| coin, as the money of a task | सिक्का | "400 coins" «400 सिक्के» |
| sweets | टॉफ़ियाँ | "lemon, orange sweets" «नींबू वाली, संतरे वाली टॉफ़ियाँ» |
| counter, on a board | गोटी | card, `picture.piles` |
| die · throw | पासा · फेंकना | "7 throws" «7 बार फेंकना» |
| x, the unknown | x | |

### The technical pages

| English | Hindi | Note |
|---|---|---|
| token · access token · refresh token | टोकन · ऐक्सेस टोकन · रीफ़्रेश टोकन | «ऐक्सेस» as the card writes it |
| authorization code | ऑथराइज़ेशन कोड | |
| seal · sealed · sealing | सील करना · सीलबंद · सील करना | «सीलबंद» does not change: «जवाब सीलबंद रहता है» |
| encrypted | एन्क्रिप्ट किया हुआ | |
| state · stateless | स्टेट · स्टेटलेस | a case of a task is «स्थिति» |
| MCP server · tool · tool call | MCP सर्वर · टूल · टूल कॉल | «कॉल» is feminine |
| chat host | चैट होस्ट | |
| consent screen | सहमति स्क्रीन | |
| scope · permission | स्कोप · अनुमति | «drive.file स्कोप»; for parents «अनुमति» |
| client · client metadata document | क्लाइंट · क्लाइंट का मेटाडेटा दस्तावेज़ | |
| renewal · revocation | रिन्यू करना · रद्द करना | |
| pace · ceiling | रफ़्तार · ऊपरी सीमा | |
| target, of the profile's size | लक्ष्य | |
| log · logging | लॉग · लॉगिंग | «Cloud Logging» stays |
| instance · revision · image | इंस्टेंस · रिविज़न · इमेज | |
| key · key rotation · ciphertext · nonce | कुंजी · कुंजी बदलना · सिफ़रटेक्स्ट · nonce | «कुंजी» is feminine |
| fingerprint · near-duplicate | फ़िंगरप्रिंट · लगभग दोहराव | card, «दोहराव» (`advice.double_count`); "not a repeat" «दोहराव नहीं» |
| readability · self-check | पढ़ने में आसानी · खुद की जाँच | "readable for the grade" is «कक्षा के हिसाब से पढ़ने लायक» |
| brief, for the model · instructions | ब्रीफ़ · निर्देश | "the instructions version" «निर्देशों का संस्करण» |
| package · schema · template · catalog · format | पैकेज · स्कीमा · टेंपलेट · कैटलॉग · फ़ॉर्मैट | «टेंपलेट» with the anusvara, as Spelling says |
| service account | सेवा खाता | Google's Hindi; on "Service" it stands beside «सेवा», the service itself |
| deterministic · the rule | तय नियम से चलने वाला · नियम | |
| version · release · commit | संस्करण · रिलीज़ · commit | card, «संस्करण» (`app.version`) |
| cookie · cache | कुकी · कैश | «कुकी» is feminine, plural «कुकियाँ»: «दो कुकियाँ» |
| identifier · Google ID · provider · credential · trace · opaque | आईडी · Google आईडी · प्रदाता · क्रेडेंशियल · ट्रेस · अपारदर्शी | «आईडी» is feminine; as the privacy policy writes them |
| graceful shutdown | व्यवस्थित ढंग से बंद होना | |
| lifetime, of a token · sliding | वैधता · आगे खिसकती | "30 days, sliding, 90 at most" «30 दिन, आगे खिसकती, अधिक से अधिक 90» |
| Workload Identity Federation | Workload Identity Federation | the name stays |
| student model · skill model | छात्र मॉडल · कौशल मॉडल | |
| chance, of a right answer | सही जवाब की संभावना | "{chance} chance — in the zone" is «संभावना {chance} — सही दायरे में» |
| corridor · goal | दायरा · लक्ष्य | "the corridor aimed at" «लक्ष्य दायरा» |
| interval, statistical · range, of the chance · band, in a chart | अंतराल · श्रेणी · पट्टी | «अंतराल» is a gap on the topics' pages |
| the ceiling, an estimate that knows where the child stands | ऊपरी सीमा | |
| bench · simulated children | बेंच · काल्पनिक बच्चे | |
| a child who learns · who stays put | सीखते हुए · बिना बढ़त के | no participle agrees with a child: "a child who learns · a share of tasks" is «सीखते हुए · सवालों का हिस्सा» |
| lag · logit · points | पिछड़ना · लॉजिट · अंक | |
| mastery declared falsely | महारत की गलत घोषणा | |
| live data, live numbers | असली आँकड़े | they come monthly, not in real time |
| standard error · baseline · seed | मानक त्रुटि · आधार रेखा · सीड | |
| Reached · On the edge · Not reached | पूरा · सीमा पर · पूरा नहीं | agreeing with «लक्ष्य» |
| paper · public domain | शोध-पत्र · सार्वजनिक डोमेन | |
| study · finding | अध्ययन · नतीजा | |
| psychometrics · item response theory | मनोमिति · आइटम रिस्पॉन्स थ्योरी (IRT) | |
| Elo on top of IRT | IRT के ऊपर Elo | Elo stays |
| endpoint · claim | एंडपॉइंट · क्लेम | |
| request · store · router · issuer | अनुरोध · स्टोर · राउटर · जारीकर्ता | card, «अनुरोध» (`task.another_coming`) |
| deploy · build · client secret · sandboxed | डिप्लॉय · बिल्ड · क्लाइंट सीक्रेट · सैंडबॉक्स में | |
| in the open, unsealed · timeout | खुले रूप में · टाइमआउट | |
| digest · runtime · PKCE verifier · state machine | डाइजेस्ट · रनटाइम · PKCE verifier · स्टेट मशीन | |
| scenario · sketch · pinned | परिदृश्य · स्केच · पिन किया हुआ | |
| binary · database · queue | बाइनरी · डेटाबेस · कतार | |
| the way in · pure computation · around the core | प्रवेश · सिर्फ़ गणना · कोर के आसपास | the groups of the map on "Service" |
| canvas · tablet · phone · computer · mouse | कैनवास · टैबलेट · फ़ोन · कंप्यूटर · माउस | |
| browser · tab · web | ब्राउज़र · टैब · वेब | |
| GitHub issue | GitHub issue | "open an issue on GitHub" is «GitHub पर issue खोलें» |
| whisker, of an error bar · dashed | त्रुटि रेखा · टूटी रेखा | |
| counting name, of the logs · totals | गिनती का नाम · कुल आँकड़े | the privacy policy |
| email address | ईमेल पता | |
| † {year}, a book's author | निधन {year} | a noun, which shows no gender |
| KB · px | KB · px | as the English |

## Pólya's four steps

No Hindi edition of *How to Solve It* was found: Open Library's 38 editions hold none in an Indian language, archive.org has none, and Hindi Wikipedia has no article on the book or on Pólya. Azim Premji University's Hindi magazine names it by its English title, «'हाऊ टू सॉल्व इट'», and Pólya «जॉर्ज पोल्या» (https://archive.org/details/apu.anuvadasampada.hin.3043.1), as Hindi Wikipedia's link to him does. So the page names the book as the English page does, in its English title, set in quotes since Hindi has no italic: *“How to Solve It”* (1945). Pólya is «जॉर्ज पोल्या».

The steps take plain Hindi infinitives, built on the card's «सवाल» and «जाँच»; no edition gives them names, and an exam-preparation site's «समस्या को समझें…» reads as a form:

| English | Hindi |
|---|---|
| *How to Solve It* (1945) | *“How to Solve It”* (1945) |
| Understand the problem | सवाल समझना |
| Make a plan | योजना बनाना |
| Carry it out | योजना पर अमल करना |
| Look back | पीछे मुड़कर जाँचना |

"Look back" keeps Pólya's picture of looking back and the checking his text describes, and stays clear of «समीक्षा», the card's name for its review (`review.title`). The steps' questions speak to the child with «तुम» or with no "you", and in the card's words: «अपने शब्दों में सवाल दोबारा बताओ। क्या दिया गया है? क्या पूछा गया है?». The coach's shorter line, "Understand → make a plan → solve → check", is «समझना → योजना बनाना → हल करना → जाँचना», and the picture's "understand, plan, solve, check" on "Why" «समझना, योजना बनाना, हल करना, जाँचना».

A technique's name is an infinitive phrase or a noun phrase, as these steps are: «चित्र बनाना» (Draw the problem), «सारणी बनाना» (Make a table), «छोटी संख्याओं से शुरू करना» (Start with small numbers), «तय क्रम में सूची बनाना» (List systematically), «अंत से शुरू करना» (Work from the end), «मान लेना कि सब एक जैसे हैं» (Suppose they are all the same), «पैटर्न ढूँढना» (Look for a pattern), «उलटा मान लेना» (Suppose the opposite), «सम-विषम पर नज़र रखना» (Watch the parity), «कबूतर और खाने» (Pigeons and pigeonholes), «चालों को आईने की तरह दोहराना» (Mirror the moves). The card's «सारणी» (`picture.table`) is a table.

## People's names

Every person of an example has a common North Indian name of the gender the English gives, and the same English name is the same Hindi name on every page; no name serves two people. "Initial" marks a name whose first letter a drawing or a badge shows. Each such name begins with a bare consonant or a vowel letter, so the initial is a single letter with no vowel sign. Within one drawing no two initials repeat, and none differ by a single stroke (ब and व, म and भ, म and प, घ and ध, य and थ).

| English | Hindi | Where | Initial, badge |
|---|---|---|---|
| Galt | गाल्ट | home, the cards' pseudonym: «गाल्ट के लिए चुना गया», «गाल्ट के कौशल» | transliterated, as the Russian «Галт»; an invented name shows no gender. Never the subject of a verb that shows one |
| Comet | धूमकेतु | "Why" and "Topics", the cards' pseudonym | the Hindi word for a comet, as the Russian «Комета»; a pseudonym, so never the subject of a verb that shows gender |
| C, "your child" on "Why" | ब | `parent.explain.child`: «ब यानी आपका बच्चा।» | the first letter of «बच्चा»; the parent's mark is «आप» |
| Ann | अनु | ordering, parts, ratios, clocks, calendar | «अ» |
| Ben | रवि | ordering, parts, ratios, parity | «र» |
| Kate | कविता | ordering, calendar | «क»: `art.kate` is «क» |
| Dan | दीपक | ordering's line | name in a drawing |
| Lena | लता | ordering's queue, gaps, ratios | «ल» |
| Max | हर्ष | ordering's queue, ratios | «ह» |
| Nina | नमिता | ordering's queue | «न» |
| Oleg | ओम | ordering's queue | «ओ» |
| Pavel | पवन | ordering's queue | «प», and the badge «प» |
| Ada | आशा | ordering's runners | «आ» |
| Bob | भरत | ordering's runners | «भ» |
| Cid | जय | ordering's runners | «ज» |
| Dot | सपना | ordering's runners | «स» |
| Eli | ईशान | ordering's runners | «ई». A boy: the English gives no gender, and a Hindi verb must: «ईशान आखिरी नहीं था» |
| Fay | फ़रहा | ordering's runners | «फ़», and the badge «फ़» |
| Anya | अदिति | "Techniques": the pets and the knights | «अ». A girl and a knight: «अदिति सच्ची है» |
| Borya | मोहन | "Techniques": the pets and the knights | «म»: «मोहन झूठा है», «अदिति और मैं, दोनों सच्चे हैं।» A liar whose initial is not «स», the knight's mark on the knights' page |
| Vera | वर्षा | "Techniques", calendar | «व» on the pets' faces; a name in the calendar's drawings |
| Tanya | तनु | "Techniques", working backwards | |
| Petya | राहुल | "Techniques", working backwards | |
| Lea | रचना | enumeration | «र» |
| Tom | तरुण | enumeration, clocks, parts | «त»; the handshake drawing shows «र» and «त» alone |
| Masha | मधु | gaps, arithmetic | «म» |
| Sasha | सुमन | pigeonhole, the sweets | a name boys and girls share, as Sasha is. The sentences take the dative, «सुमन को कितनी टॉफ़ियाँ निकालनी होंगी?», «सुमन को संतरे वाली टॉफ़ी चाहिए», which shows no gender; where a verb must show one, a girl |
| Gleb | गौरव | pigeonhole, the socks | |
| Kim | किरण | winning strategy | a name boys and girls share, here a girl: «पहले किरण उठाती है». Name in drawings: «किरण −3» |
| Leo | यश | winning strategy | name in drawings: «यश −4» |
| Grandma, Granny | दादी | clocks, parts | the honorific plural, «दादी बेचती हैं»; in drawings «दादी», «दादी की घड़ी» |
| Mum, Dad | माँ, पापा | calendar, ordering's tip | the honorific plural: «पापा माँ से लंबे हैं» |
| Mum, Dad, Older son, Younger son | माँ, पापा, बड़ा बेटा, छोटा बेटा | "About" | the family's own names for itself |
| the brother and the sister | भाई, बहन | "Techniques", drawing the problem | «भाई अपनी बहन से 4 साल बड़ा है» |
| islanders A, B, C | A, B, C | knights | Latin capitals, in the generic masculine |
| knight K, liar L | स, झ | knights: `art.knight`, `art.liar`, `basis.cards[].mark`, the tip "Write K and L"; `thumb.*` | the first letters of the card's «सच्चे» and «झूठे» |
| children A, B, C · boxes A, B | the same | enumeration · ratios | «डिब्बा A», «डिब्बा B» |
| the cook (she) · the baker (he) | रसोइन · बेकरी वाला | weighing · parts | «एक रसोइन के पास…», «बेकरी वाला दोपहर तक आधे बन बेच देता है» |
| woodcutters · a farmer · the buyers | लकड़हारे · किसान · ग्राहक | gaps · percentages · parts | the generic masculine |

The people of history keep their Hindi names, as Hindi Wikipedia writes them where it has them: «यूक्लिड», «राजा टॉलेमी» (Wikipedia names him «टॉलेमी प्रथम सोटर»), «प्रोक्लस», “एलिमेंट्स”, «आर्किमिडीज़», «राफेल», “एथेंस का विद्यालय” (a fresco, «फ्रेस्को»), «कन्फ़्यूशियस», «छह कलाएँ», «आर्यभट», «पुणे», «तक़ी अल-दीन» and his observatory in «इस्तांबुल», an Ottoman miniature, «उस्मानी लघुचित्र», «लूका पचोली», «याकोपो दे' बारबारी», «अल-ख़्वारिज़्मी», «बग़दाद», “गणितीय कला पर नौ अध्याय”, «अरस्तु», «सिकंदर महान», «निकोलाई बोगदानोव-बेल्स्की» and his painting “मौखिक गणना। एस. ए. राचिंस्की के जन-विद्यालय में”, «सेंट लुईस» and its «बेट्स स्कूल», «टोक्यो», the shrine «कोन्नो हाचिमान-गू मंदिर», «सांगाकु», «लेनिनग्राद», «मास्को», «हंगरी», «संयुक्त राज्य अमेरिका» («अमेरिका» on a chip), «सोवियत संघ», «जॉर्ज पोल्या». The chips of the eras are «यूनान», «चीन», «भारत», «इटली», «जापान», «रूस», «इस्लामी दुनिया», and the Greeks «यूनानी». Wikipedia has no article on Taqi al-Din, de' Barbari, Bogdanov-Belsky, Rachinsky, the Bates School, the shrine, sangaku or the six arts: those names are transliterated or translated here. The book "whose title gave us the word "algebra"" gave Hindi no word, since «बीजगणित» is Sanskrit's: the Hindi says that the world took the word from it, «वह किताब, जिसके नाम से दुनिया को “algebra” (बीजगणित) शब्द मिला», as Hindi Wikipedia tells it («'अल-जब्र' … यह बिगड़कर 'ऐल्जेब्रा' (algebra) बन गया»). "Euclides", written on the slate, is an inscription and stays. A photograph's credit, «Yu. D. Korolev» or «Momotarou2012», stays as its source writes it.

## Grades

`topics.grades` is **«कक्षा {range}»**: «कक्षा 1–2», «कक्षा 3–4», «कक्षा 5–6», «कक्षा 1–6». So `topic.example` reads «उदाहरण 2 · कक्षा 3–4», the label on "Techniques" «सवाल · कक्षा 2–3», and "Research" writes «कक्षा {first}–{last}» (`hero.grades`, `sources.level`).

- It is the card's own chip, «कक्षा {from}–{to}» (`progress.grades`), and its «कक्षा {grade}» (`child.grade`), so a chip on the site and the card's pill read alike.
- «कक्षा» before a range does not change before a postposition: «कक्षा 3–4 के बच्चे», the home page's «कक्षा {range} के लिए», "Why"'s «कक्षा {last} तक».
- "grades 1 to 6" in words is the card's «कक्षा 1 से 6» (`first_run.lead`), and "the grade, from 1 to 6" in help the card's «कक्षा, 1 से 6 तक» (`first_run.fill`).
- "third- and fourth-graders" on "Why" are «कक्षा 3 और 4 के बच्चे», "ninth-graders" «कक्षा 9 के छात्र», "six-year-olds" «6 साल के बच्चे».
- The head's title of a topic's page, `title.topic`, is «{topic}: ओलंपियाड गणित के सवाल, कक्षा {range}», already in the site's dictionary. The longest, «जीतने की रणनीति वाले खेल: ओलंपियाड गणित के सवाल, कक्षा 5–6», is 51 of the checker's 60 units: it counts a Devanagari syllable as one, and one more for a vowel sign that stands beside its letter, as ा, ि, ी and ो do, while a sign above or below it, a nasal and the virama of a conjunct count none.

The alternative is the plural «कक्षाएँ {range}», which the card never writes for a range.

## Books on "Research"

| Key | The English page | Hindi | How sure |
|---|---|---|---|
| perelman.firsthundred | For Young Mathematicians. The First Hundred Puzzles | the English title | no Hindi edition found |
| perelman.secondhundred | The Second Hundred Puzzles | the English title | no Hindi edition found |
| perelman.living | Living Mathematics | «सरस गणित» (Mir, Moscow, 1982 and 1990; translated by «देवेन्द्र प्र॰ वर्मा») | sure it is this book: the scan reproduces the Russian title page, «Живая математика», and its first chapter, «नाश्ते पर पहेलियां», opens with «मैदान में गिलहरी», the squirrel in the meadow (https://archive.org/details/ya.-perelman-maths-can-be-fun-in-hindi-mir-1990). Hindi Wikipedia's «जीवंत गणित» is its own gloss, not an edition's title |
| perelman.arithmetic | Entertaining Arithmetic | the English title | no Hindi edition found; Mir's Hindi Perelman of algebra, «मनोरंजक बीजगणित», is another book |
| perelman.quick | Quick Reckoning | the English title | no Hindi edition found |
| ignatyev.wit | In the Realm of Ingenuity, or Arithmetic for Everyone | the English title | no Hindi edition found |
| dudeney.amusements | Amusements in Mathematics | the English title | no Hindi edition found |
| dudeney.canterbury | The Canterbury Puzzles | the English title | no Hindi edition found |
| loyd.cyclopedia | Cyclopedia of Puzzles | the English title | no Hindi edition found |
| carroll.tangled | A Tangled Tale | the English title | no Hindi edition found |
| carroll.pillow | Pillow Problems | the English title | no Hindi edition found |

A Hindi title goes in “…”, «“सरस गणित”»; an English one stays plain, as the Spanish and Arabic pages write them. mirtitles.org, which lists Mir's books in India's languages, refused a reader that was not a person, so a Hindi edition of another of these books may still exist.

The authors' names, as Hindi Wikipedia writes them where it has an article: «याकोव पेरेल्मान» (the Hindi edition prints «पेरेलमान»), «लुइस कैरल»; and transliterated where it has none: «येमेल्यान इग्नात्येव», «हेनरी ई. ड्यूडनी», «सैम लॉयड». `sources.died`, "† {year}", is «निधन {year}»: Hindi has no dagger for a death, and the noun shows no gender.

## Search words

Gathered on 2026-10-12 from Google's suggestions for India in Hindi (`hl=hi`, `gl=IN`), some 130 queries typed in Devanagari and in Latin letters, the way Hindi is often typed online («ganit paheli», «dimagi paheli»). The titles of the first results could be read for one query only, «बच्चों के लिए गणित की पहेलियाँ»: the search engines refuse a script, and those titles say what the suggestions say. The lists show that a phrase is typed, not how often, and [the brief](README.md#search-words) says how a text uses them.

**Who searches, and how.** Google is India's search engine; Bing barely counts there. Parents of India search maths for children in English at least as much as in Hindi («maths olympiad questions for class 3», «olympiad maths for class 1»), so the Hindi pages are found by the Hindi queries alone. A class is typed with its digits: «कक्षा 3», «class 3», «क्लास 3»; in words, «तीसरी कक्षा», rarely. `topics.grades`, «कक्षा {range}», is the word of Hindi-medium schools and of the suggestions, and a title keeps its digits through the slot; a description names the stage in words or leaves the classes out.

**What parents type**, strongest first:

1. «गणित की पहेलियाँ», «गणित पहेली», with «उत्तर सहित», «बच्चों के लिए» or a class;
2. «गणित के सवाल» with a class, «गणित के सवाल कक्षा 3», which leans to homework;
3. «दिमागी पहेली» and «तार्किक पहेली», riddles and logic puzzles;
4. «गणित ओलंपियाड प्रश्न», «ओलंपियाड गणित», whose suggestions are nearly all exams and their papers;
5. the parents' own questions, «बच्चों को गणित कैसे पढ़ाएं» and «गणित में रुचि कैसे पैदा करें»;
6. «गणित ट्रिक्स», maths tricks;
7. «रीजनिंग», the reasoning section of India's competitive exams: clocks, calendars, ranking and Venn diagrams are suggested with it, and with exam papers.

**The head's title.**

- A topic's page: **«{topic}: ओलंपियाड गणित के सवाल, कक्षा {range}»**. «ओलंपियाड गणित» is typed as it stands, and «गणित के सवाल» is Hindi's commonest phrase for maths problems; together they name the kind of problem and no contest. The longest, «जीतने की रणनीति वाले खेल: ओलंपियाड गणित के सवाल, कक्षा 5–6», is 51 units, each syllable of Devanagari counted once and each vowel sign beside its letter once more. «गणित की पहेलियाँ», typed more, would call the site's tasks riddles; it goes into the descriptions.
- Any other page: **«{page} — MathTrail»**, with the card's dash.

**Topics.** ⚑ marks a card's name nobody types.

| Topic | The card | What parents type |
|---|---|---|
| ordering | «क्रम में लगाना» ⚑ | «क्रम व्यवस्था», «रैंकिंग», «बैठक व्यवस्था»: the reasoning exams' words for order and seating |
| knights and liars | «सच्चे और झूठे» ⚑ | nothing in Hindi, and «सच्चे और झूठे» is suggested as botany's true and false fruits, «सच्चे फल और झूठे फल»; «knights and knaves puzzles» in English |
| enumeration | «सभी विकल्प गिनना» ⚑ | «गिनती के सवाल»; «counting figures» counts the triangles of a figure |
| gaps and boundaries | «अंतराल और किनारे» ⚑ | nothing found |
| clocks | «घड़ी» | «घड़ी के प्रश्न»; «घड़ी रीजनिंग», the angle between the hands, is the exams' |
| calendar and age | «कैलेंडर और उम्र» | «उम्र के सवाल», «आयु संबंधी प्रश्न», «कैलेंडर के प्रश्न» |
| pigeonhole principle | «कबूतरखाना सिद्धांत» ⚑ | nothing in Hindi; «pigeonhole principle in hindi», in English |
| parity and alternation | «सम-विषम और बारी-बारी» ⚑ | «सम विषम संख्या», «सम और विषम संख्याएँ»: definitions for small children |
| arithmetic with a trick | «पेचीदा अंकगणित» ⚑ | «गणित ट्रिक्स», «गणित के जादुई ट्रिक» |
| weighing and pouring | «तौलना और उड़ेलना» ⚑ | nothing that is maths: «नकली सिक्का» is counterfeit money |
| parts and shares | «पूरे के हिस्से» ⚑ | «भिन्न के सवाल», with a class |
| percentages | «प्रतिशत» | «प्रतिशत के सवाल» |
| ratios and sharing | «अनुपात और बँटवारा» ⚑ | «अनुपात और समानुपात», «अनुपात के सवाल» |
| figures on a grid | «वर्गाकार ग्रिड पर आकृतियाँ» ⚑ | «क्षेत्रफल और परिमाप» |
| divisibility and remainders | «विभाज्यता और शेषफल» | «विभाज्यता के नियम», «शेषफल» |
| overlapping groups | «आपस में मिलते समूह» ⚑ | «वेन आरेख», «वेन डायग्राम» |
| games with a winning strategy | «जीतने की रणनीति वाले खेल» ⚑ | «गणितीय खेल»; «जीतने की रणनीति» alone is the elections' |

**Descriptions** can carry «गणित की पहेलियाँ», «गणित के सवाल», «बच्चों के लिए», the stage in words, «प्राथमिक कक्षाएँ», on "Techniques" «सवाल हल करने के तरीके», on "Why" «सोचने वाले सवाल» and how to teach a child maths, which the page's own sources back. «उत्तर सहित» and «हल सहित» fit a topic's page and "Techniques" only; the home page's honest words are the card's, «हर गलती को समझाता है».

**Not to use.**

- Contests and exams: «IMO», «SOF», «हिंदुस्तान ओलंपियाड», «मैट्रिक्स ओलंपियाड», «नेशनल ओलंपियाड», «गणित ओलंपियाड.इन», «ओलंपियाड परीक्षा», «एसएससी», «प्रश्न पत्र», «सैंपल पेपर», «सिलेबस».
- School books and what the site does not give: «एनसीईआरटी», «गणित मेला» and «गणित का जादू», titles of NCERT's books for the first classes, «किताब», «pdf», «वर्कशीट», «डाउनलोड».
- «रीजनिंग» alone, which brings readers who prepare for exams; «बुद्धि परीक्षा», a test; «प्रतिभाशाली», which sorts children; and promises such as «दिमाग तेज़ करें».
- Words that search reads as something else: «सच्चे और झूठे» without «पहेली» or «विषय», the fruits; «नकली सिक्का», counterfeit money; «जीतने की रणनीति» without «खेल», elections; «वर्ग पहेली», a crossword; «पहेली बुझाओ», a riddle to guess; «गणित हल करो» and «सॉल्वर», which solve homework.

**For the author.** India's parents type maths searches in English more than in Hindi; the Hindi pages compete for the Hindi queries alone. Ten of the card's seventeen names are typed by nobody, and «सच्चे और झूठे» is read as botany. The suggestions for «ओलंपियाड» are about exams; the title keeps «ओलंपियाड गणित» because the site's tasks are of that style, and names no contest.

**Decided by the author (2026-10-26, T114.1).** The Hindi pages keep their words as they are, and the title keeps «ओलंपियाड गणित», the style of the site's tasks, naming no contest.

## Places for the author's reading

- Claude's labels in Hindi, «**+ जोड़ें**», «**कस्टम कनेक्टर जोड़ें**», «**जारी रखें**», «**कनेक्ट करें**», «**कस्टमाइज़ → कनेक्टर**», as Claude's own Hindi interface writes them, read from the strings claude.ai loads for a browser set to Hindi: Claude's help has no Hindi. A parent whose Claude speaks English sees the English labels; the alternative is to keep the English, as the Russian and Arabic pages do, or to write both, «**+ जोड़ें** (+ Add)».
- Google's labels quoted in Google's spelling, «ऐक्सेस हटाएं», «ट्रैश खाली करें», beside the card's «हटाएँ» in the site's own sentences; and «'कनेक्ट किए गए ऐप्लिकेशन' पेज», Google's Hindi name for the linked apps page, rather than the card's older «तीसरे पक्ष के ऐप वाले पेज».
- Pólya's steps, «सवाल समझना», «योजना बनाना», «योजना पर अमल करना», «पीछे मुड़कर जाँचना», and the techniques' names: no Hindi edition of the book names them.
- «सरस गणित», the one Hindi edition found of the books on "Research"; the others keep their English titles. «पेरेल्मान» as Hindi Wikipedia writes it, against the edition's «पेरेलमान».
- The "algebra" sentence on "Why", which says that the world took the word from the book, since Hindi's «बीजगणित» did not come from it.
- The gender rules and their cost. Hindi shows the subject's gender in most verb forms, so the texts lean on the passive, the dative («आपको…», «बच्चे को…»), the subjunctive and the card's passive trap names far more than plain Hindi would; the reader is never «आप कर सकते हैं» and the child never «बच्चा करता है». A few sentences will read more formal than a Hindi parent's own speech.
- «बच्चा» as the generic noun for the child, masculine in form, as the card writes it, against «बच्चा या बच्ची», which no text of the card uses.
- The generic role nouns «अभिभावक», «बड़ा व्यक्ति» and «ट्यूटर» with a masculine verb on the technical pages and in the documents, as the card's «सेट करता है».
- ÷ for every division, though the English topic pages write ":", and the ratio's colon «3 : 4» where the English writes "3 to 4" in words.
- The money: «₹799» and «₹0», a local price as the Russian and the Japanese write theirs.
- The Latin letters A, B, C for the islanders, the children of enumeration and the boxes, and «स» and «झ» for the knight and the liar. The speaker's letter on "Why" stands for "your child", and every language writes it with the first letter of its own word, so the Hindi is «ब», for «बच्चा», with «आप» for the parent.
- The names, all chosen here: «अनु», «रवि», «कविता», «दीपक», «लता», «हर्ष», «नमिता», «ओम», «पवन», «आशा», «भरत», «जय», «सपना», «ईशान», «फ़रहा», «अदिति», «मोहन», «वर्षा», «तनु», «राहुल», «रचना», «तरुण», «मधु», «सुमन», «गौरव», «किरण», «यश»; the genders chosen where the English gives none, Sasha («सुमन», a name for either, a girl if a verb must say) and Kim («किरण», a girl), and Eli («ईशान», a boy); «फ़रहा», a Muslim name of North India among the rest; the pseudonyms «गाल्ट» and «धूमकेतु».
- The head's title of a topic's page, «{topic}: ओलंपियाड गणित के सवाल, कक्षा {range}», with «ओलंपियाड गणित» though its suggestions are about exams (see "Search words").
- «शोध» for "Research" and «हमारे बारे में» for "About", in the menu, the pages' titles and the footer.
- «तरीके» for the techniques, the everyday word, against «युक्तियाँ» or «रणनीतियाँ».
- `research.places` is «जगह» in both forms: its one slot, in `theses.twice.text`, is a distance, «{count} {noun} आगे», and Hindi says «2 जगह आगे», not «2 जगहें». The other `research.*` nouns read right wherever their slots stand in the direct case.
- The ordinal suffixes in digits for 1, 2, 3, 4 and 6, «1ला», «2रा», «3री», «4थी», «6ठी सदी», as CLDR's ordinal categories for `hi` give them, beside «-वाँ», «-वीं», «-वें» for every other number. «1ली» is rarer in print than «पहली»; the digits keep the example's numbers as the English writes them.
- «1.9 करोड़» for "19 million", Intl's compact form for `hi`.
- Floors as «मंज़िल 1», «मंज़िल 9», so that neither «भूतल» nor India's «पहली मंज़िल» changes the count of flights.
- The spelling: क, ख and ग without the nukta, «किस्मत», «आखिरी», as most of the card writes them, though the card has «आख़िरी» and «बदक़िस्मत»; the card itself might be made consistent.
- The nukta typed as U+093C, as the card's file has it, never the precomposed letters.
- «खानेदार कागज़» for squared paper, the everyday word that goes with the card's «खाना», against NCERT's «वर्गांकित कागज़».
- «विरोधाभास» for a contradiction and «विरोधाभास से प्रमाण» for the proof: NCERT's own Hindi term for the method is not checked.
- «ज़ीना» for a flight of stairs, which not every reader uses; «सीढ़ी» is a step.
- «रसोइन» for the cook, who is a woman in the English; «रसोइया» would make her a man.
- «सुराग» for "clue", to keep «संकेत» for the card's hint.
- «सामने वाला खिलाड़ी» for "the other player" and "the opponent", to keep «दूसरा खिलाड़ी» for "the second player".
- Every topic's name in a sentence in “…”, also where no postposition would change it: three of the card's names are verb phrases that a sentence would read as verbs.
- The ask, «“{topic}” विषय पर एक सवाल दें», the same on all seventeen pages, against shorter forms such as «घड़ी पर एक सवाल दें».
- «सेवा», feminine, for "the service" on the technical pages, beside the masculine MathTrail and chat: a sentence's verb follows whichever is its subject.
- «सच कहें तो» for "Honestly", «तैयारी में» for the badge "In development", «तकनीकी पाठकों के लिए» for "For technical readers".
- «हिन्दी» in the language picker, as Intl names the language, and «हिंदी» in prose, as the card spells every such word.
- The card names both a group and a topic «पूरे के हिस्से» (`group.parts`, `topic.fractions.parts`), so the topics page will show the topic under a group of its own name, as in Spanish. The fix belongs to the card's dictionary, and the site follows it.
- The card's «अंक» is a school mark (`progress.grades_note`), a digit in the maths and a point in darts; the site keeps «अंक» for all three and writes «नंबर» for a mark wherever it could be read as a digit.
