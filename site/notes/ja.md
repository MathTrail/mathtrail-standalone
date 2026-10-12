# Japanese (`ja`)

The Japanese half of [the brief](README.md): the rules the Japanese texts follow on top of it, and the words they use. The card's Japanese dictionary, [`web/locales/ja.json`](../../web/locales/ja.json), comes first wherever it has a word; this file adds only what the site needs beyond it. The site's Japanese stands in 「」 in this file, as the Spanish notes put theirs in «»; what is inside the outer brackets is the site's text, its own brackets included.

## Register and variant

- **One Japanese, the card's.** Standard modern Japanese (共通語), as the card writes it; there is no regional variant.
- **The reader is spoken to in です・ます**, with the card's light honorifics: 「チャットで問題を頼んでください。」 (`waiting.ask_in_chat`), 「何を直せばよいかは、チャットでお知らせします。」 (`profile.not_saved_detail`), 「お子さまの本名は入れないでください。」 (`profile.pseudonym_note`). So 「お使いの Google ドライブ」, 「ご確認ください」, 「お子さま」. Never あなた: Japanese leaves the reader out of the sentence, as the card does. The privacy policy, the terms and the help are in です・ます too. "We", the family behind the project, is 「私たち」, as the page "About" calls itself.
- **The child is spoken to in the plain form**, as the card speaks to the child: 「答えを1つ選びましょう」 (`task.pick`), 「明日へのお願いごとをしよう」 (`waiting.limit`), 「とき方を見てみましょう」 (`result.idk`). What a parent says to the child (the quotes under 「声かけ」, the questions of a tip, Pólya's questions) is the parent's own voice: the plain form, 「〜しよう」, 「〜してみよう」, a question ending in 「？」: 「まだ使っていない手がかりはどれ？」.
- **What a parent types to the chat** is short and polite, as people write to a chat in Japanese: 「新しい問題をお願いします」, 「正直者とうそつきの問題を出してください」, 「MathTrail の問題をやりましょう」. In the parent's own message the child is 「うちの子」: 「うちの子の進み具合はどうですか？」.
- **Kanji for the parent, kana where the card writes kana for the child.** The card writes its words for the child in kana (まちがい, せいかい, とじる, こたえ, とき方) and its words for the parent in kanji (`first_run.lead` 「間違いを一つひとつ解説します」). The site is read by parents, so its prose uses ordinary kanji: 間違い, 正解, 閉じる, 答え, 解き方. A card's label quoted on the site is copied exactly, kana and all (「ヒントをとじる」, 「こたえのふりかえり」). What speaks to the child (an example's question, a hint, a 声かけ, Pólya's questions) uses no kanji outside the 1,026 taught in grades 1 to 6 (教育漢字), and writes かく (draw) and 子ども in kana, as the card does. 子ども is always so, never 子供.
- **No gender shows, the child's or the parent's.** Japanese verbs and adjectives never show it, but these words do:
  - Third-person pronouns 彼 and 彼女: never for the child, and not for anyone else either. Japanese drops the subject or repeats the name: 「マキは数を1つ思いうかべました。その数を2倍して…」.
  - First-person pronouns: ぼく and おれ are a boy's, あたし a girl's, and わたし in a child's mouth sounds like a girl. The child's own lines on "Why" use none: 「この問題、なんでやるの？」, 「またまちがえちゃった…」, 「できた！答えは5だよ。」. A speaker in an example (an islander, Borya) says わたし, as Japanese logic puzzles write a statement.
  - Sentence-final particles: わ, わよ, のよ, かしら are feminine, ぜ, ぞ masculine. None anywhere; よ, ね, の？ are neutral.
  - Name suffixes: くん is a boy's, ちゃん leans to girls. The reader's child gets none; the people of the examples are katakana names with no suffix (see "People's names").
  - Kinship nouns: 息子 and 娘 say son or daughter. "Your child" is the card's 「お子さま」 (`first_run.pseudonym`, `profile.pseudonym_note`), and 「子ども」 where the sentence is general, as the card's 「子どもの進み具合に合わせて」 (`topic_choice.coach_note`). 男の子 and 女の子 appear only where the English says boy and girl (the parity ring, the choir).
  - "The parent" is the card's 「保護者」 (`progress.for_parent` 「保護者の方へ」), "the tutor" the card's 「先生」 and "an adult" 「大人」 (`first_run.lead` 「保護者や先生などの大人」). お父さん and お母さん appear only where the English says Dad and Mum. 「親」 is used only as the mark of the parent in the dialogue on "Why" and in 親子.
  - "Children" is 「子どもたち」 or 「子ども」; counted on "Research", the counter 「人」. "Together" is 「一緒に」, "the family" 「ご家庭」 (「ご家庭のチャット」).
  - The people of the examples have the gender their names suggest; since no pronoun follows them, the Japanese text shows it no further.
- **An example's question is a task for the child**, as a task on the card is: です・ます, a question ending in 「〜か。」 as Japanese maths problems end, never in 「？」: 「柱は何本ありますか。」. The words around it, the steps and the notes, speak to the reader in です・ます: 「6つの停留所を点で表します。」. Headings, labels, the items of a list and the words of a drawing are nouns or the plain dictionary form: 「線を引いて、全員をその上に並べる」. An example's answer is a short polite sentence that states the numbers of the English one: 「30です。」, 「6通りです。」, 「AもBもうそつきです。」.
- **What a topic teaches**, `hero.teaches`, goes on from its label 「身につく力」 (`topic.teaches`), so it is a noun phrase that ends in 「力」, two verbs joined by the 連用形: 「順を追って考え、自分で確かめる力」, 「手がかりを1本の線にまとめ、比べる文を両方から読む力」.
- **Sentences every topic's page shares** read the same on every page:
  - `method.lead` ends 「下の例題は、いちばんやさしい問題から算数オリンピックの問題へと進みます。」
  - `description` is 「〇〇の問題とはどんなものか、どう解くのか、子どもがつまずきやすいところ、家庭でできることを紹介します。」, its first part the topic's own (「曜日・日付・年れいの問題とはどんなものか、…」).
  - The ask is 「〇〇の問題を出してください」, with the card's name: 「順番の推理の問題を出してください」, never a paraphrase (the English "a task on putting things in order").
- **The card's words, exactly.** A topic's name written out in a sentence is the card's, character for character. Where the sentence names it as a topic, it stands in 「」, as the card writes 「{topic}」 (`topic_choice.ask`): 「テーマ「場合の数」の問題」. Where the sentence speaks of the matter itself, it stands bare: 「場合の数は、ハトの巣原理にも正直者とうそつきにも必要です」. A topic page's `title` is the card's name alone.

## Typography

Checked character by character against the card's dictionary, which the site follows; where the card is not consistent, the site follows its majority, as noted.

- **Quotes:** 「…」, and 『…』 inside them. The English “…” become 「…」. A list of quoted names runs on with no 、 between the brackets: 「場合の数」「偶奇と交互」「正直者とうそつき」.
- **A quotation that ends a sentence** keeps no 。 inside its closing bracket; the sentence goes on and takes its own: 「Aは「わたしはうそつきです」と言いました。」. ？ and ！ stay inside. A `say:` that is the quote alone is the bracket and nothing after it: 「まだ使っていない手がかりはどれ？」.
- **Full-width punctuation:** 、 (U+3001) and 。 (U+3002), never ，．; ！ and ？ full width (the card's 「せいかい！」, 「次はどのテーマ？」), with no space after them inside a paragraph, as the card's 「せいかい！答えは{correct}です。」. The card writes one half-width ? (`advice.percent_wrong_base`); the site does not follow it.
- **Colons:** the full-width ： after a Japanese word, with no space on either side, as the card writes it eleven times (「完了：」, 「ランク{first}〜{last}：{from}〜{to}年生の問題。」). The card writes a half-width : twice (`topic_choice.label` 「テーマ:」, `progress.link_refused`); the site does not follow those. The half-width : stays only in a time (4:25), a ratio (3:4), code and addresses.
- **Brackets:** full-width （） around Japanese text and around a citation: 「（1〜6年生）」 as the card, 「（Radesky et al., 2022）」. Half-width () stay inside a formula: 「(19 + 81) + (23 + 77)」, 「σ(θ + δ − β)」.
- **Spaces:** a half-width space between a word in Latin letters and Japanese on either side, as the card writes it in eight keys: 「MathTrail は、」, 「MathTrail へようこそ」, 「お使いの Google ドライブ」, 「Claude に追加」, 「MCP サーバー」, 「チャットの AI」. No space next to full-width punctuation (「（Claude）」, 「MathTrail、」), none around digits (「1つ」, 「3年生」, 「2026年」), a label of one letter (「Aは」) or a unit's symbol (「12m」, 「20%」). Three keys of the card leave the space out (`review.step_practice`, `problem.not_in_catalog`, `first_run.start`); the site follows the other eight. No space inside a katakana compound: アクセストークン, though Google's own docs write アクセス トークン. No ideographic space (U+3000) anywhere.
- **Dots:** ・ (U+30FB) joins nouns of one list without spaces, as the card's 「倍数・約数とあまり」 and 「多い・少ない、前・後」. The middle dot · (U+00B7) with a space on each side separates the parts of a label, as the card's 「算数オリンピックコーチ · {grade}年生」: 「例題2 · 3–4年生」, 「論文全文 · PDF、14ページ」.
- **Ellipsis:** one … (U+2026), as the card's 「チェック中…」, not the book form ……. The long vowel mark ー (U+30FC) in katakana, never a dash or a minus.
- **Dashes:** the English spaced dash becomes 「。」, 「、」, 「：」 or （）. A title's dash becomes 「：」: 「MathTrail：いつものチャットで算数オリンピック」. The card's own 「次の問題 —「{topic}」のテーマで」 is quoted as it is where the site quotes it.
- **Ranges:** 〜 (U+301C WAVE DASH) with no spaces, as the card writes 「1〜6年生」 and 「2〜3問」: 「70〜85%」, 「9〜19世紀」, 「5分〜24時間」. Not ～ (U+FF5E), which `Intl.NumberFormat("ja").formatRange` writes. A range the build puts into a slot (`{range}`) comes with an en dash, 「3–4年生」, and stays so (see "Grades").
- **Ordinals** take a counter after digits, inside the examples and outside: 「9個目のあめ」, 「4回目の水やり」, 「23番目の旗」, 「5番目のマス」, 「3人目のお客さん」, 「9階」, 「25日」. In a drawing that labels things side by side the counter may be shared: "21st, 22nd, 23rd" is 「21、22、23番目」. "The first Saturday" of a month is 「第1土曜日」. The first and the second player are the card's 「先手」 and 「後手」; the badge "2nd" is 「後手」. Kanji numerals appear only inside set words, as the card's 「一日」, 「一つひとつ」, 「一列」, 「一度」; every count is in digits.
- **Dates:** 「2026年10月10日」, 「5月3日から5月10日まで」, 「その月の25日」, 「2日、9日、16日、23日、30日」. A weekday is 「月曜日」 in prose, and 「月曜」, 「水曜」 in a drawing, a chip or a badge, never the lone 「月」 or 「水」, which read as "month" and "water" (the badge of "Watering the flowers" is 「水曜」). A month is 「5月」.
- **Centuries** in digits, as Japanese writes them: 「5世紀」, 「6世紀まで」, 「9〜19世紀」, 「16世紀末」, 「19世紀末〜20世紀」; decades 「1950年代」; years 「1894年から」.
- **Times** as the English writes them, with digits and a half-width colon: 「4:25」, 「15:10」, 「0:00」, the bait's 「03:59」. A time on the clock is 時刻 and a length of time 時間, the two words Japanese teaches in grade 2. A length of time is 「4時間30分」 in prose and in a drawing alike (`{hours} h {minutes} min` is 「{hours}時間{minutes}分」, `{count} min` 「{count}分」, `{count} h` 「{count}時間」). "Noon" is 「正午」, "midnight" 「午前0時」; "half an hour" and "a quarter of an hour" are 「30分」 and 「15分」.
- **Numbers outside the maths** as `Intl.NumberFormat("ja")` writes them: 「12,450」, the bait's coins, since Japanese groups from four digits; 「0.2」, so the formula reads 「P = 0.2 + 0.8·σ(θ + δ − β)」. Years are never grouped. Millions are counted in 万, as Intl's compact form writes them: 「1900万人」.
- **Numbers in the maths** — the examples, the tables of a topic's page, and the sums of its cards and tips — keep the English digits and no separator, as the Russian and the Spanish do: 「2026 × 2026」, 「1111」, the arithmetic page's 「10、100、1000」 and its table's 「1900」, the percentages page's tip 「1200」; a decimal keeps the point: 「0.5 + ½ = 1」. Never a full-width digit (U+FF10 to U+FF19), which a Japanese keyboard may type and the site's checker refuses.
- **Percent** takes the half-width % with no space, everywhere, as Intl writes it: 「20%」, 「70〜85%」, and so inside the examples and on the page of percentages, whose drawings write 「20%」 beside the words. Never the full-width ％.
- **Money** appears only in the make-believe app on "Why": 「980円」 for the hat and 「0円」 for MathTrail. 980円 is the just-under price Japanese shops and in-app stores write, as $9.99 is in English, and 「0円」 is how Japanese advertising says "free"; `Intl.NumberFormat("ja", {style: "currency", currency: "JPY", currencyDisplay: "name"})` writes both exactly so. Not 「￥980」, the full-width sign Intl's symbol gives, nor 「¥980」.
- **Units:** the symbol directly after the digits, with no space: 「12m」, 「200kg」, 「5L」, 「100KB」. The litre is 「L」, as Japanese school books write it since 2011, never ℓ. In prose, where the English writes the word, the word: 「12メートル」, 「3リットル」. On the card questions of the home page and "Why" the fence's length is in words, 「12メートル」, as the English writes "12 metres".
- **Signs:** × for multiplication, with spaces as the English: 「5 × 50 = 250」; − (U+2212) for minus, never a hyphen or ー; · in the formula. **Division is always ÷**, also on the topic pages where the English writes ":" (gaps, calendar, percentages, ratios, parts, arithmetic): in Japanese school maths ":" is the sign of a ratio, and 「35 : 7 = 5」 would read as a ratio. A ratio is written 「3:4」, half-width, no spaces, and read 「3:4の比」; "1 to 50" on a map is 「1:50」. Fractions as the English: 「3/4」, 「½」; in words, 「3分の2」.
- **Capitals:** Japanese has none; a Latin name keeps its own case (MathTrail, ChatGPT).
- **Emphasis** stays where the English has it, and Markdown (`marked`) sees it only on a boundary it can tell. Brackets and the sentence's 。 and 、 go outside the emphasis, never at its edges inside: 「**次の問題**」, 『*いかにして問題をとくか*』, 「**ケンタが正直者だと仮定します**。すると…」. Where the text inside begins or ends with a sign (+, →), the character outside it on that side is a space or a punctuation mark: 「、**+ 追加**を押し」 and 「**+** → **コネクタ**を押し」 work; 「で**+ 追加**」, 「**矛盾。**次に」 and 「**「次の問題」**を」 do not: the page would show their asterisks, and the build refuses a text whose emphasis is not the English one's. No space is needed between a bold label and the particle after it, but for the space a Latin label takes anyway (「**Back to Google** を押して」).
- **YAML:** a text that holds an ASCII `: ` or begins with `*` is quoted. Japanese text with the full-width ： or opening with 「 needs no quotes. In a folded block (`>-`) every line break becomes a space, which Japanese does not want between its characters: write such a text on one line, however long, or break it only where a space belongs, beside a Latin word.
- **Titles of works** (a book, a painting, a treatise) go in 『』, with the English emphasis inside the brackets where the English has one: 「著書『*いかにして問題をとくか*』（1945年）」.
- **The card questions on the home page and on "Why"** write no lone Latin capital inside a sentence that the English does not: the picture check reads one there as a label of the picture. So the fence there is 「12メートル」, never 「12M」, and no option letter or name in Latin letters is added.

## Platforms' labels

### Claude

Claude's interface speaks Japanese: 日本語 is one of the languages its help lists, chosen from the profile icon's 「言語」 section (https://support.claude.com/ja/articles/10769299-how-to-use-claude-in-your-preferred-language, 「好みの言語でClaudeを使用する方法」), and its help is published in Japanese, so its labels are those of its Japanese help, not the English ones the Russian keeps. The help speaks です・ます, as the site does; the site writes Claude's buttons in bold and the rest in its own words: 「**+ 追加**を押し、**カスタムコネクタを追加**を選びます。」 The help writes Latin words with no space (「Claudeが」); the labels below hold none, and the site's own sentences keep the card's space.

- [A] 「リモートMCPを使用したカスタムコネクタを始める」, https://support.claude.com/ja/articles/11175166-getting-started-with-custom-connectors-using-remote-mcp
- [B] 「コネクタを使用してClaudeの機能を拡張する」, https://support.claude.com/ja/articles/11176164-use-connectors-to-extend-claude-s-capabilities

| English on the site | Japanese | Where the help says it |
|---|---|---|
| Customize → Connectors | カスタマイズ → コネクタ | A: 「カスタマイズ > コネクタに移動します」. The site keeps its own arrow. |
| + | + 追加 | A, the Pro and Max steps: 「「+ 追加」をクリックしてから「カスタムコネクタを追加」をクリックします」 |
| Add custom connector | カスタムコネクタを追加 | A |
| Name it MathTrail | 名前に MathTrail と入力し | A: 「コネクタの名前を入力します」; the field's label is not quoted |
| the address | アドレス | A says 「コネクタのリモートMCPサーバーURLを追加します」; the site says 「アドレス」, the card's word (`progress.link_refused`) |
| Continue, and Continue again | 続行、もう一度 続行 | A: 「「続行」をクリックします」, once after the URL and once after the authentication settings |
| the sign-in settings Claude detects | Claude が検出した認証設定 | A: 「Claudeが検出した認証設定を確認し、サーバーが異なるものが必要な場合は変更します」 |
| Add | 追加 | A: 「「追加」をクリックしてコネクタの設定を完了します」 |
| Connect | 接続 | A, the members' steps: 「「接続」をクリックして認証し」. B leaves it in English, 「「Connect」または「Install」をクリックして」 |
| + → Connectors, in a chat | + → コネクタ | A: 「チャットインターフェースの左下の「+」ボタンをクリックしてから「コネクタ」をクリック」; B: 「「コネクタ」にマウスを合わせて」 |
| the free plan | Free プラン | A: 「Freeユーザーは1つのカスタムコネクタに制限されています」. B writes 「無料ユーザー」; the site keeps the plan's name, as the brief keeps every plan's |
| Always allow | 常に許可 | A, its section on security: 「「常に許可」をクリックします」; B: 「「常に許可」「承認が必要」「ブロック」」 |
| Sign in now · Sign in when needed · No sign in | 今すぐサインイン · 必要に応じてサインイン · サインインなし | A, under 「認証」; not on the site yet |
| Use Claude's published identity · Register automatically · Use your own OAuth client | Claudeの公開されたアイデンティティを使用（推奨） · 自動的に登録 · 独自のOAuthクライアントを使用 | A, under 「OAuthクライアント」; B words the first two 「Claudeの公開ID（推奨）を使用」 and 「自動登録」; not on the site yet |
| "Use caution before running this prompt" | not found in Claude's Japanese help | the text describes it: 「注意をうながす警告の下に」 |
| Claude asks whether to allow MathTrail to save | the question is not quoted in Claude's Japanese help; only the button 「常に許可」 is | the text describes it: 「MathTrail にプロフィールの保存を許可するか、Claude が確認します」 |

Claude's labels say サインイン; the site's own word for signing in is the card's and Google's 「ログイン」, and サインイン appears only inside a quoted Claude label.

### ChatGPT

| English on the site | Japanese | Source |
|---|---|---|
| developer mode | **開発者モード** | https://help.openai.com/ja-jp/articles/12584461 — 「ChatGPT の開発者モードと MCP アプリ」. Its page title now reads 「デベロッパーモード」 and its text 「開発者モード」 throughout; the site uses the text's word. The path is 「設定 → アプリ → 詳細設定」, then 「開発者モード」; an app is made from 「アプリ → 作成」 |
| ChatGPT's directory | ChatGPT のディレクトリ | The directory is now 「プラグインディレクトリ」 (https://help.openai.com/ja-jp/articles/20001256, 「ChatGPT のプラグイン」), reached from 「設定 > プラグイン」 (https://help.openai.com/ja-jp/articles/11487775, 「ChatGPT の接続済みアプリ」). The generic phrase stays true |
| Plus, Pro, Business, Enterprise, Education | the same | names of plans; the Japanese help writes 「Business、Enterprise、Edu プラン」 |

help.openai.com refuses fetchers; these labels come from a reader proxy's copy of those three pages (r.jina.ai) and from search results that quote them. OpenAI's Japanese writes a space between Latin and Japanese, as the card does.

### Google

| English on the site | Japanese | Source |
|---|---|---|
| Google Drive | Google ドライブ, and 「ドライブ」 alone; My Drive is 「マイドライブ」 | https://support.google.com/drive/answer/2409045?hl=ja, and the card's `data.export_where` |
| Drive's bin | ドライブのゴミ箱; the button 「ゴミ箱を空にする」; files go for good after 30 days (「ゴミ箱に移動したファイルは、30 日後に完全に削除されます」) | https://support.google.com/drive/answer/2375102?hl=ja, and the card's `data.delete_how` |
| version history | 版の履歴; for a file such as the profile, the menu 「版を管理」 and 「この履歴を削除しない」 for "keep forever" | https://support.google.com/drive/answer/2409045?hl=ja |
| linked apps, third-party access | the page 「リンク済みアプリ」 of a Google account (myaccount.google.com/linkedapps); 「Google アカウントへのアクセス権」 → MathTrail → 「詳細を表示」 → 「アクセス権を削除」 → 「確認」 | https://support.google.com/accounts/answer/13533235?hl=ja, 「Google アカウントと他のデベロッパーのアプリとのリンクを管理する」 |
| the box for Google Drive on the consent screen | Google ドライブのチェックボックス; Google words the permission 「このアプリで使用する Google ドライブ上の特定のファイルのみの参照、編集、作成、削除」 | as Google lists it in Japanese in its Workspace Marketplace: https://workspace.google.com/marketplace/app/docs_sheets_editor_for_drive/703230099925?hl=ja. Not seen on the live consent screen, and the English never quotes it |
| Sign in with Google | the button is 「Google でログイン」; in prose 「Google でログインする」 | https://developers.google.com/identity/siwg/features?hl=ja |
| Google account | Google アカウント | card, Google |

The card's `data.access_how` says 「サードパーティ製アプリのアクセス権」, Google's earlier wording; the site uses the page's current name, 「リンク済みアプリ」, which the English links. Google's Japanese writes 「30 日」 with a space; the site writes digits with no space, as the card does.

MathTrail's own sign-in pages speak English, so 「Tick the box for Google Drive」 and **Back to Google** stay English.

## Glossary

"card" marks the card's own word, from `web/locales/ja.json`.

### Product and lesson

| English | Japanese | Note |
|---|---|---|
| task | 問題 | card, `task.label` |
| problem (olympiad) | 問題 | the same word as task |
| school sum | 学校の計算問題 | "A school sum" is 「学校の計算問題」 |
| topic | テーマ | card |
| group of topics | グループ | card, `group.*` exactly; "Groups of topics" is 「テーマのグループ」 |
| trap | ひっかけ | the card's label "The trap" is 「ひっかけポイント」 (`result.trap`), quoted so where the site names that label |
| the catch (not a trap) | 気をつけたいところ | "the catch is that…" is 「気をつけたいのは、…」; never ひっかけ outside the card's sense |
| technique | 解法 | |
| the main move (`topic.method`) | 解き方のカギ | |
| hint | ヒント | card |
| clue, a given of a task | 手がかり | ordering and "Techniques" |
| condition | 条件 | card, `trap.ignored_condition` |
| option | 選択肢 | the card lists the options under 「答え」 (`result.answers`) |
| answer | 答え | card |
| right, wrong answer | 正解, 間違った答え | the card writes 「せいかい！」 and 「まちがい」 for the child |
| solution | 解き方 | the card's label is 「とき方」 (`result.solution`) |
| step, of a solution | 手順 | card, `advice.stopped_early` |
| step, of Pólya or of the map | ステップ | 「問題解決への4つのステップ」, as Pólya's publisher writes it |
| question, the text of a task | 問題文 | card: 「問題文の数をそのまま答えにする」 |
| question, what is asked | 聞かれていること | card: 「聞かれていないことに答える」 |
| picture of a task | 図 | 「写真」 for a photograph, 「絵」 for a painting |
| label of a picture | 記号 | |
| grade | 学年 | card, `profile.grade` |
| Grade 3 | 3年生 | card, `child.grade` |
| grades 1 to 6 | 小学1〜6年生 | card, `first_run.lead` |
| grades 3–4, two slots | {from}〜{to}年生 | card, `progress.grades` |
| third- and fourth-graders · ninth-graders · six-year-olds | 3・4年生 · 9年生 · 6歳児 | the findings on "Why" |
| pupil, in a task about a school · pupils of research | 児童 · 生徒 | 「ある学校の児童は300人です」; 生徒 for older pupils, as Brazil's |
| level | レベル | card |
| difficulty | 難易度 | the card's waiting step says 「テーマとレベルを決めました」 (`waiting.step.topic`), quoted as it is |
| rank | ランク | card, 「ランク{rank}」: a number, never a title |
| 11 ranks | 11段階のランク | |
| rating | レーティング | card: 「このテーマのレーティング」, 「全体のレーティング」 |
| progress, the card's screen | 記録 | card, 「プロフィールと記録」 |
| progress, how a child advances | 進み具合 | card, `topic_choice.coach_note` |
| mastered | マスター | card |
| review, the card's section | ふりかえり | card, `review.title` |
| how the answer went | こたえのふりかえり | card, `result.label` |
| trial series | 実力チェック | card |
| skipped | スキップ | card |
| profile | プロフィール | card |
| pseudonym | ニックネーム | card |
| interests | 好きなこと | card |
| lesson | レッスン | card: 「レッスンの言語」 |
| coach | コーチ | card, `topic_choice.coach_short`; the Coach product too |
| companion, in the coach | 相棒 | a character, not the child |
| parent | 保護者 | card |
| tutor | 先生 | card, `first_run.lead` |
| adult | 大人 | card |
| child, your child | 子ども; お子さま | card; see "No gender shows" |
| children | 子どもたち | 「人」 where "Research" counts them |
| chat | チャット | card |
| the chat's model | チャットの AI | on the pages for parents; 「モデル」 on "Service" and "Research" |
| card | カード | card |
| connector, custom connector | コネクタ, カスタムコネクタ | Claude |
| app | アプリ | card, OpenAI, Google |
| sign in | ログインする | card, `profile.signin_country_off`; Google |
| sign-in | ログイン | |
| account | アカウント | |
| your Google Drive, your Google account, your chat | お使いの Google ドライブ, お使いの Google アカウント, お使いのチャット | card, `data.export_where`, `data.access_how` |
| Drive file, the profile's file | ドライブのファイル, プロフィールのファイル | card |
| folder | フォルダ | card, Google |
| bin | ゴミ箱 | card and Google |
| daily limit | 一日の上限 | card |
| Another task | 次の問題 | card's button, quoted exactly; also the words it puts in the chat |
| Make a wish for tomorrow | 明日へのお願いごとをしよう | card, quoted exactly |
| Profile & progress | プロフィールと記録 | card |
| Edit | 編集 | card |
| Preparing the next task… | 次の問題を用意しています… | card, `waiting.title` |
| the bulb | 電球のボタン | |
| the two arrows | 2つの矢印のボタン | |
| tap | タップする | |
| press, click | 押す | Claude's help says クリック; 押す holds for a phone too |
| Enter, the key | Enter キー | |
| free, of price | 無料 | the chip `hero.free` |
| Free, the plan | Free プラン | a plan's name |
| open source | オープンソース | |
| in development | 開発中 | |
| beta | ベータ版 | the menu's tag stays Beta |
| soon | 準備中 | |
| olympiad maths | 算数オリンピック | card, `app.badge` 「算数オリンピックコーチ」 |
| olympiad-style task | 算数オリンピック風の問題 | card, `first_run.lead` |
| olympiads of history and research | 数学オリンピック | Hungary, Leningrad, Germany, Brazil; the International Mathematical Olympiad is 「国際数学オリンピック」 |
| maths, the subject | 算数; 数学 | 算数 for grades 1 to 6; 数学 for older pupils, mathematicians and research |
| maths circle | 算数クラブ; 数学サークル | 算数クラブ today, 数学サークル for the circles of Hungary and Leningrad |
| school | 学校 | card |
| school mark | 成績 | card, `progress.grades_note` |
| homework | 宿題 | |

### The site's sections and headings

| English | Japanese | Note |
|---|---|---|
| Topics (`nav.topics`) | テーマ | 3 characters; card, `progress.topics` |
| Techniques (`nav.techniques`) | 解法 | 2 |
| Why (`nav.why`) | 学ぶ意味 | 4; the next cut is 「なぜ」, 2 |
| Coach (`nav.coach`) | コーチ | 3, beside the tag Beta; card |
| Research (`nav.research`) | 研究 | 2, and the page's title too |
| Service (`nav.service`) | 仕組み | 3; the page's title is 「サービスの仕組み」 |
| About (`nav.about`) | 私たち | 3, and the page's title too, as «Nosotros» is in Spanish; the fuller 「私たちについて」 is 7 |
| Add to Claude (`nav.add`) | Claude に追加 | Claude's own 「追加」 |
| Technical pages | 技術者向けのページ | said by a screen reader |
| Language | 言語 | |
| Help · Privacy policy · Terms of use | ヘルプ · プライバシーポリシー · 利用規約 | |
| How it is solved | 解き方 | |
| Where children go wrong | つまずきやすいところ | |
| What to say | 声かけ | to the child; the word Japanese parenting uses |
| How to help at home | 家庭でできること | |
| Learn first | 先に学ぶテーマ | |
| Opens the way to | 次に進めるテーマ | |
| Builds on · Opens | 先に学ぶ： · 次に進む： | the topics page |
| first · opens, under the map · First · Topic · Then, its legend | 先に · 次に · 先に · テーマ · 次に | |
| Ask for a task on this topic | このテーマの問題を頼んでください | the card's 「頼んでください」 |
| Show · Hide the solution | 解き方を見る · 解き方を閉じる | the card's own button says 「ヒントをとじる」 |
| Example 2 | 例題2 | |
| Problem · {grades}, on "Techniques" | 問題 · {grades} | |
| Answer | 答え | |
| Trap, the badge | ひっかけ | |
| Why it matters | なぜ大切か | |
| How to solve: … | 解き方：… | |
| The rules of … | …のルール | 「島のルール」 |
| Check: | 確かめ： | |
| When it helps: | 使える場面： | |
| In the topics: | 関連するテーマ： | |
| You are here · On this page · All topics | 現在地 · 目次 · テーマ一覧 | |
| More → | 詳しく → | |
| Honestly | 正直に言うと | |
| See a lesson | レッスンを見る | |
| For technical readers | 技術者向け | |
| You, the parent's mark on "Why" · C, the child's | 親 · 子 | each in a circle of 30 px; 「「子」はお子さまのことです。」 |
| The adult · The model, in the chat's illustration | 大人 · AI | |
| Step 1, of the topics' map | ステップ1 | |
| Foundations · Rules · Hard problems | 基礎 · きまり · 難問 | |

The menu's seven entries come to 20 Japanese characters, each 1 em wide, and with the button 「Claude に追加」 to some 26 em, against some 36 em for the Spanish and the Russian rows counted the same way (about 0.55 em a Latin letter, 0.6 a Cyrillic one), so the Japanese row is the narrowest yet, and 「私たちについて」 would still leave it so; `just site-fold` will say by how much.

A drawing's narrow column of row names (`.s-sk-name`, 64 px at 13 px, or 52 px at 12 px) holds four Japanese characters on a line, and the wide one eight; a fifth wraps. So "blocks" is 「組」, "4 blocks" 「4組」, "6 runners" 「6人」, "order" 「順番」, "from the end" 「後ろから」, "put on · take off" 「はく · ぬぐ」, "before · after" 「前 · 後」, "first Saturday" 「第1土曜日」, "row · col. · diag." 「横 · たて · ななめ」, and Grandma there is 「祖母」 (see "People's names").

### Maths and olympiad terms

| English | Japanese | Note |
|---|---|---|
| olympiad | オリンピック | 算数オリンピック or 数学オリンピック, as in "Product and lesson" |
| reference task | お手本問題 | not 基準問題 or 標準問題, which read as a level of difficulty |
| solver | ソルバー | first time, 「答えを総当たりで確かめるプログラム（ソルバー）」 |
| brute force | 総当たり | |
| worst case | 最悪の場合 | card, `trap.best_case_not_worst` |
| good, bad luck | 運がいい, 運が悪い | card: 「いちばん運が悪い場合」 |
| to be sure, for certain | 確実に, 必ず | |
| pigeonhole principle | ハトの巣原理 | card; Japanese olympiad books also call it 部屋割り論法 |
| pigeons, holes | ハト, 巣箱 | "Pigeons and pigeonholes" is 「ハトと巣箱」 |
| parity · even · odd | 偶奇 · 偶数 · 奇数 | card |
| alternation, taking turns | 交互 | card |
| remainder · quotient | あまり · 商 | card |
| divides by 9 | 9でわり切れる | わり切れる, in kana, as school writes it |
| divisibility rules | 倍数の見分け方 | card |
| digit · two-digit number | 数字 · 2けたの数 | |
| tens, units digit | 十の位, 一の位 | |
| enumeration | 場合の数 | card |
| case · list | 場合 · 書き出し | card: 「書き出しで場合が1つもれる」 |
| systematic listing | 順序よく書き出す | the technique "List systematically"; "none missed, none twice" is 「もれなく、重なりなく」 |
| tree, the diagram | 樹形図 | |
| working backwards | 後ろから考える; 逆算 | the technique "Work from the end", aka 逆算 |
| the assumption method | つるかめ算 | the classic name for the heads-and-legs problem; the hens and rabbits stay |
| proof by contradiction | 背理法 | |
| suppose · assumption · contradiction | 〜と仮定する · 仮定 · 矛盾 | "Suppose this one is a knight…" is 「この人が正直者だと仮定すると…」 |
| pattern | きまり | 「きまりを見つける」 |
| strategy | 作戦 | card, `group.games` |
| symmetric strategy | 対称な作戦 | |
| winning strategy | 必勝法 | card, `topic.games.strategy` |
| winning, losing position | 勝ちの局面, 負けの局面 | card: 局面 (`advice.first_move_assumed`) |
| the player to move | 手番の人 | "loses for the player to move" is 「手番の人が負ける」 |
| move, in a game | 手 | 「一手」 |
| the first, the second player | 先手, 後手 | card, `trap.first_move_assumed` |
| invariant | 変わらないもの | 不変量 on "Techniques" only if needed |
| knights and liars | 正直者とうそつき | card |
| knight · liar | 正直者 · うそつき | 正 and う in drawings; see below |
| island · islander | 島 · 島の住人 | |
| statement | 発言 | card, `advice.trusted_statement` |
| ordering | 順番の推理 | card |
| number line | 数直線 | card |
| grid · squared paper | マス目 · 方眼紙 | card |
| cell · side of a cell | マス · マスの辺 | |
| area · perimeter | 面積 · まわりの長さ | card |
| L shape | L字形 | |
| ratio · in the ratio 3 to 4 | 比 · 3:4の比で | |
| sharing · share | 分ける · 分け前 | card topic 「比と比例配分」 |
| percentage · percent | 百分率 · パーセント | card |
| percent of what? · the base | 何のパーセント？ · もとにする量 | card, `advice.percent_wrong_base` and `trap.percent_wrong_base` |
| 20% off | 20%引き | |
| fraction · the whole · the part | 分数 · 全体 · 部分 | card |
| a half, a third, a quarter | 半分, 3分の1, 4分の1 | |
| strip, bar | 帯 | card, `picture.bars` and `advice.part_whole_swap`; not テープ図 |
| box, of a ratio | 箱 | card, `advice.ratio_total_confusion` |
| gaps and boundaries | 植木算 | card |
| gap · post · fence | 間 · 柱 · さく | card: 「柱を点で、その間のすき間も」 |
| off by one | 1つずれる | card |
| clock face · hands · the whole hour | 文字盤 · 長針, 短針 · ちょうどの時刻 | card: 「時計の文字盤」; "step to the whole hour" is 「ちょうどの時刻まで進める」 |
| time on the clock · length of time | 時刻 · 時間 | |
| floor 1, of a house | 1階 | Japanese counts floors as the English does |
| calendar and age | カレンダーと年れい | card |
| pan balance · a weighing · fake coin | てんびん · はかる回数 · にせのコイン | card |
| jug · pour | 入れ物 · 移しかえる | card |
| state, of the jugs | 状態 | |
| overlapping groups | 集合算 | card |
| two circles that overlap | ベン図 | |
| arithmetic with a trick | 計算のくふう | card |
| I thought of a number | ある数を思いうかべました | |
| carrying | くり上がり | 「くり上がりなしで」 |
| round number | きりのいい数 | |
| match, the stick | マッチ棒 | |
| sweets | あめ | 「レモン味のあめ」, 「オレンジ味のあめ」 |
| a special case | 特別な場合 | |
| row · column · diagonal, of a grid | 横 · たて · ななめ | |
| times as many | 〜倍 | 「4倍」 |

Japanese counts things with counters, which every translator would choose differently: people 人; sweets, eggs, apples, rolls, marbles, beads, mushrooms 個; posts, logs and their pieces, trees, bushes, matches, darts, legs, cell sides 本; coins, cards, petals, shirts, scarves, socks one at a time 枚; hens 羽; rabbits 匹; cars and buses 台; seats 席; tasks 問; times, cuts, throws, jumps, weighings, handshakes, breaks 回; ways, flags of different designs, queues, codes 通り; sets and pairs 組; pages ページ; stops 「6つの停留所」; flights of stairs 「8階分」; days 日, weeks 週間, months か月, years 年, ages 歳; points 点. "How many different flags?" is 「何通り」, and its answer 「6通り」.

### The technical pages

| English | Japanese | Note |
|---|---|---|
| token · access token · refresh token | トークン · アクセストークン · 更新トークン | 更新トークン as Google's Japanese docs |
| authorization code | 認可コード | Google |
| seal · sealed · sealing | 封印する · 封印された · 封印 | 「答えは子どもが答えるまで封印されています」 |
| encrypted | 暗号化された | |
| state · stateless | 状態 · ステートレス | |
| MCP server · tool · tool call | MCP サーバー · ツール · ツール呼び出し | |
| chat host | チャットホスト | |
| consent screen | 同意画面 | Google |
| scope | スコープ | Google; 「権限」 for parents: 「drive.file の権限」 |
| client · client metadata document | クライアント · クライアントメタデータ文書 | |
| renewal · revocation | 更新 · 取り消し | |
| pace · ceiling | ペース · 上限 | |
| Sign-in, a scenario's tab on "Service" · target, of the profile's size | ログイン · 目標 | short: the tab and the bar's label do not wrap at 320 px |
| log · logging | ログ · ログ記録 | 「Cloud Logging」 stays |
| instance · revision · image | インスタンス · リビジョン · イメージ | as Cloud Run's Japanese docs |
| key rotation · ciphertext · nonce | 鍵のローテーション · 暗号文 · ノンス | |
| fingerprint · near-duplicate | フィンガープリント · ほぼ重複 | for parents, "not a repeat" is 「くり返しでない」 |
| readability · self-check | 読みやすさ · セルフチェック | |
| brief, for the model | 指示書 | 指示書 where the brief is named as a thing the service hands over; 指示 where the text speaks of the model's instructions in general or of their version: 「モデル向けの指示」, 「指示のバージョン」 |
| package · schema · template | パッケージ · スキーマ · テンプレート | |
| deterministic · the rule | 決定的 · ルール | |
| cookie · cache | Cookie · キャッシュ | |
| graceful shutdown | グレースフルシャットダウン | |
| lifetime, of a token | 有効期間 | |
| Workload Identity Federation | Workload Identity 連携 | as Google Cloud's Japanese docs |
| student model | 学習者モデル | |
| skill model | スキルのモデル | the home page |
| chance, of a right answer | 正解する確率 | |
| corridor · interval · goal | 目標帯 · 区間 · 目標 | |
| range, of the chance or of the answers | 範囲 | |
| the ceiling, an estimate that knows the child · oracle | 天井 · オラクル | |
| bench · simulated children | ベンチ · 仮想の子ども | |
| a child who learns, who stays put | 学んで伸びる子, 力が変わらない子 | |
| lag · logit · points | 遅れ · ロジット · ポイント | |
| mastery declared falsely | 誤ったマスター判定 | |
| live data, live numbers | 実データ, 実際の数値 | they come monthly, not in real time |
| standard error · baseline · seed | 標準誤差 · ベースライン · シード | |
| Reached · On the edge · Not reached | 達成 · 境界 · 未達 | |
| paper · public domain | 論文 · パブリックドメイン | |
| † {year}, a book's author | {year}年没 | |
| commit · release | コミット · リリース | |
| endpoint · claim, of a JWT | エンドポイント · クレーム | |
| request · store · router · issuer | リクエスト · ストア · ルーター · 発行者 | the card says 「おねがい」 for the child's ask (`task.another_coming`) |
| deploy · build · client secret · sandboxed | デプロイ · ビルド · クライアントシークレット · サンドボックス化された | |
| in the open, unsealed · timeout | 平文で · タイムアウト | |
| digest · runtime · PKCE verifier · state machine | ダイジェスト · ランタイム · PKCE のコード検証子 · ステートマシン | |
| scenario · sketch · pinned · sliding, of a lifetime | シナリオ · スケッチ · 固定 · スライド式 | |
| canvas · tablet | キャンバス · タブレット | the coach |
| GitHub issue · computer · mouse | Issue · パソコン · マウス | "open an issue on GitHub" is 「GitHub で Issue を作成する」 |
| the profile's notes | メモ | |
| whisker, of an error bar · dashed | ひげ · 破線 | |
| item response theory · psychometrics | 項目反応理論（IRT） · 心理測定学 | |
| Elo on top of IRT | IRT（項目反応理論）の上に組んだ Elo | the home page and "Service"; Elo stays |

## Pólya's four steps

The Japanese edition is 『いかにして問題をとくか』 (G・ポリア 著, 柿内賢信 訳, 丸善, 1954, from the fifth English edition, 第5版; now 丸善出版, 第11版, its type renewed in 2022), as its publisher's page gives it: https://www.maruzen-publishing.co.jp/contents/howtosolveit/index.html. Its summary, 「問題解決への4つのステップ」, names the phases 「問題を理解すること」, 「計画を立てること」, 「計画を実行すること」, 「ふり返ってみること」, under the imperatives 「第1に 問題を理解しなければならない」 … 「第4に 得られた答えを検討せよ」. The page "Techniques" names its steps after them, in the dictionary form without こと:

| English | Japanese |
|---|---|
| *How to Solve It* (1945) | 『*いかにして問題をとくか*』（1945年） |
| Understand the problem | 問題を理解する |
| Make a plan | 計画を立てる |
| Carry it out | 計画を実行する |
| Look back | ふり返ってみる |

A technique's name is a verb phrase in the dictionary form, as Pólya's steps are, and 「図をかく」 writes かく in kana as Pólya's 「図をかけ」 and the card do: 「図をかく」, 「表をつくる」, 「小さい数で試す」 (aka 「特別な場合を考える」), 「順序よく書き出す」, 「後ろから考える」 (aka 「逆算」), 「全部同じだと考える」 (aka 「つるかめ算」), 「きまりを見つける」, 「反対を仮定する」 (aka 「背理法」), 「偶数・奇数に目をつける」, 「ハトと巣箱」 (aka 「ハトの巣原理」), 「相手のまねをする」 (aka 「対称な作戦」). The groups are 「問題を見えるようにする」, 「答えを見つける」, 「証明する」.

The steps' questions speak to the child, as Pólya's 「未知のものは何か」 does in its own way: 「自分の言葉で言ってみよう。わかっていることは何？求めるものは何？」. The coach's shorter line, "Understand → make a plan → solve → check", is 「理解する → 計画を立てる → 解く → 確かめる」, and the picture's "understand, plan, solve, check" on "Why" is 「理解する、計画を立てる、解く、確かめる」. George Pólya is 「ジョージ・ポリア」.

## People's names

Every person of an example has a Japanese given name in katakana with no suffix: a name in katakana stands out from the kana around it, so it needs no さん to be read as a name, and prose and drawings write it alike; くん and ちゃん would show a gender. "Initial" marks a name whose first character a drawing or a badge shows: the first katakana of the Japanese name takes the place of the Latin initial. Within one drawing no two initials repeat, and none is a katakana that reads as a numeral or as its neighbour in a small circle (ニ as 二, ハ as 八, ミ as 三, チ as 千, ナ as 十, ロ as a box, and シ beside ツ or ソ beside ン). Names in drawings are two or three characters; the six runners, who stand side by side, are two each. The same English name is the same Japanese name on every page, and no name serves two people.

| English | Japanese | Where | Initial, badge |
|---|---|---|---|
| Galt | ガルト | home, the cards' pseudonym: 「ガルト向けに選んだ問題」 | as the Russian «Галт»; an invented name shows no gender |
| Comet | コメット | "Why" and "Topics", the cards' pseudonym | the word as a nickname, as the Russian «Комета» and the Spanish «Cometa» |
| C, "your child" on "Why" | 子 | `parent.explain.child`, and in the lead 「「子」はお子さまのことです。」 | the speaker's mark; the parent's, "You", is 「親」 |
| Ann | ユイ | ordering, parts, ratios, clocks, calendar | ユ, for A |
| Ben | リク | ordering, parts, ratios, parity | リ, for B |
| Kate | エマ | ordering, calendar | エ: `art.kate` is 「エ」 |
| Dan | ゴウ | ordering's line | name in a drawing |
| Lena | レナ | ordering's queue, gaps, ratios | レ, for L |
| Max | ソウ | ordering's queue, ratios | ソ, for M |
| Nina | マオ | ordering's queue | マ, for N |
| Oleg | ジン | ordering's queue | ジ, for O |
| Pavel | タク | ordering's queue | タ, for P, and the badge 「タ」 |
| Ada | アヤ | ordering's runners | ア, for A |
| Bob | カイ | ordering's runners | カ, for B |
| Cid | ヒロ | ordering's runners | ヒ, for C |
| Dot | メイ | ordering's runners | メ, for D |
| Eli | セナ | ordering's runners | セ, for E |
| Fay | フウ | ordering's runners | フ, for F, and the badge 「フ」 |
| Anya | アオイ | "Techniques": the pets and the knights | ア, the legend's letter for A |
| Borya | ケンタ | "Techniques": the pets and the knights | ケ, the legend's letter for B; says わたし: 「アオイもわたしも正直者です」 |
| Vera | リン | "Techniques", calendar | name in drawings |
| Tanya | マイ | "Techniques", working backwards | |
| Petya | ユウ | "Techniques", working backwards | |
| Lea | サナ | enumeration | サ, for L |
| Tom | トモ | enumeration, clocks, parts | ト, for T |
| Masha | マキ | gaps, arithmetic | マ, for M; differs from Lena's レ |
| Sasha | ヒカル | pigeonhole, the sweets | used for boys and girls alike, as Sasha is |
| Gleb | ゲン | pigeonhole, the socks | |
| Kim | ユキ | winning strategy | name in drawings: 「ユキ −3」 |
| Leo | レオ | winning strategy | name in drawings: 「レオ −4」 |
| Grandma, Granny | おばあちゃん | clocks, parts | 「おばあちゃんの時計」; in the clocks drawing's row of names, four characters wide, 「祖母」 |
| Mum, Dad | お母さん, お父さん | calendar, ordering's tip | |
| Mum, Dad, Older son, Younger son | ママ, パパ, 長男, 次男 | "About" | the family's own names for itself |
| islanders A, B, C | A, B, C | knights | Latin capitals, as Japanese logic puzzles label them: 「Aは「わたしはうそつきです」と言いました。」 |
| knight K, liar L | 正, う | knights, `thumb.*`, `art.knight`, `art.liar`, `basis.cards[].mark` | the first character of the card's 「正直者」 and 「うそつき」 |
| children A, B, C · boxes A, B | the same | enumeration · ratios | letters; 「箱A」, 「箱B」 |
| the brother and the sister | 兄, 妹 | "Techniques", drawing the problem | |
| the cook (she) · the baker (he) | コックさん · パン屋さん | weighing · parts | no pronoun follows them |
| woodcutters · a farmer | きこりたち · 農家の人 | gaps · percentages | |

The knight is 「正」 and the liar 「う」: each is the first character of the card's word, as K and L are of the English; 正 is taught in grade 1 and う is a kana, so a child can write both over a name, as the tip "Write K and L in pencil" asks, where 嘘 is no school kanji and blurs in a circle of 32 px. In prose the two marks always stand in 「」, since a bare 「正とう」 reads as the word 正当: 「「正」と「う」をえんぴつで書く」, and a sequence in one pair, 「正、う、正、う…」. The islanders stay A, B, C and take no suffix.

The people of history keep their usual Japanese names: ユークリッド, プトレマイオス1世 (the English "King Ptolemy", 「プトレマイオス王」), プロクロス, 『原論』, 「幾何学に王道なし」, アルキメデス, ラファエロ, 『アテナイの学堂』, 孔子, 六芸, アーリヤバタ, プネー, タキー・アッディーン, イスタンブール, オスマン帝国, ルカ・パチョーリ, ヤコポ・デ・バルバリ, フワーリズミー, バグダード, 『九章算術』, 算額, 東京の金王八幡宮, アリストテレス, アレクサンドロス大王, ニコライ・ボグダノフ＝ベルスキー, 『暗算』, S・A・ラチンスキー, セントルイス, レニングラード, モスクワ, アメリカ, ソ連, ジョージ・ポリア. The word that came from al-Khwarizmi's title is the European "algebra", not the Japanese 代数, so the sentence quotes it: 「「algebra（代数学）」という言葉は、その本の題名から生まれました」. Sangaku hung in shrines and temples, 「寺社」, as the caption's own example, a shrine, shows. A photograph's credit, 「Yu. D. Korolev」 or 「Momotarou2012」, stays as its source writes it, and the licence with it.

## Grades

`topics.grades` is **「{range}年生」**: 「1–2年生」, 「3–4年生」, 「5–6年生」, 「1–6年生」; so `topic.example` reads 「例題2 · 3–4年生」, and the label on "Techniques" 「問題 · 2–3年生」.

- It is the card's own word, 年生, after the range, as the card writes 「{from}〜{to}年生」 (`progress.grades`) and 「{grade}年生」 (`child.grade`), so a chip on the site and the card's pill read alike.
- It is the shortest phrase a chip can carry: three characters after the numbers.
- The range comes from the build with an en dash, 「3–4」, where the card writes 〜. Japanese reads 「3–4年生」 without trouble, but the two forms will stand side by side: "Research", whose texts write the range themselves with two slots, says 「{first}〜{last}年生」, as the card. Only the build can give the slot a 〜; until then the text keeps the en dash.

The alternative is 「小学{range}年生」, 「小学3–4年生」, as the card's 「小学1〜6年生」 (`first_run.lead`): clearer that the grades are those of elementary school, two characters longer on every chip. The home page's lead, which has the one slot, says it in full: 「小学{range}年生向けの」.

## Books on "Research"

| Key | The English page | Japanese | How sure |
|---|---|---|---|
| perelman.firsthundred | For Young Mathematicians. The First Hundred Puzzles | the English title | no Japanese edition found |
| perelman.secondhundred | The Second Hundred Puzzles | the English title | no Japanese edition found |
| perelman.living | Living Mathematics | 『おもしろい数学』 (三橋重男 訳, 東京図書, 1977) | sure it is this book: Japanese Wikipedia (ヤコブ・ペレルマン) and the puzzle-book list of Kagoshima University (https://www.sci.kagoshima-u.ac.jp/q_chan/PUZZLE/books.html) give Живая математика as its original; the year is Wikipedia's, the list gives 1979, with a second volume, 『続おもしろい数学』. The same translation first came out as 『生きた数学』 (東京図書〈数学新書〉, 1967); other translations are 『イワンの数学パズル』 (金沢養 訳, 白揚社, 1959) and 『パズルで鍛える数学力』 (白揚社, 2000) |
| perelman.arithmetic | Entertaining Arithmetic | 『おもしろい計算術』 (金光不二夫 訳, 東京図書, 1980) | fairly sure: Japanese Wikipedia lists it under Занимательная арифметика; an earlier translation is 『数の魅力』 (もののべながおき 訳, ダイヤモンド社, 1965) |
| perelman.quick | Quick Reckoning | the English title | no Japanese edition found |
| ignatyev.wit | In the Realm of Ingenuity, or Arithmetic for Everyone | 『数学おもちゃ箱』 (松野武 訳, 東京図書, 1980–81) | sure it is this book: the Kagoshima list gives В царстве смекалки as its original; it came in two volumes, the second 『続数学おもちゃ箱』, and the years are not checked; the subtitle is not on the Japanese edition |
| dudeney.amusements | Amusements in Mathematics | 『パズルの王様』 (藤村幸三郎・林一 訳, ダイヤモンド社, 1965) | fairly sure: a Japanese page on Dudeney says the book came out from ダイヤモンド社 under this title (https://www.torito.jp/puzzles/121.shtml); it came in several volumes, and whether they hold the whole book is not checked. 『パズルの王様傑作集』 (高木茂男 編訳, ダイヤモンド社, 1986) is a selection |
| dudeney.canterbury | The Canterbury Puzzles | 『カンタベリー・パズル』 (伴田良輔 訳, ちくま学芸文庫, 2009) | sure |
| loyd.cyclopedia | Cyclopedia of Puzzles | the English title | no Japanese edition of the whole book: 東京図書's five volumes (松野武 訳, 1981–82, 『サム・ロイドのおもしろパズル』 and others) and 『サム・ロイドの「考える」パズル』 (伴田良輔 編訳, 青山出版社, 2008) are selections, and Martin Gardner's 『サムロイドのパズル百科』 is another book |
| carroll.tangled | A Tangled Tale | 『もつれっ話』 (柳瀬尚紀 訳, ちくま文庫, 1989) | sure; first published by れんが書房新社 |
| carroll.pillow | Pillow Problems | 『枕頭問題集』 (柳瀬尚紀 訳, 朝日出版社〈エピステーメー叢書〉, 1978) | sure |

The authors' names: ヤコフ・ペレリマン, as his Japanese publishers write the surname (Japanese Wikipedia writes ヤコブ and also ペレルマン); エメリヤン・イグナーチェフ, the surname as his Japanese edition writes it and the given name as Japanese writes Емельян; ヘンリー・E・デュードニー, サム・ロイド and ルイス・キャロル as Japanese writes them.

## Search words

Gathered on 2026-10-12 from Google's suggestions for Japan and the titles of the first results; they show that a phrase is typed, not how often, and [the brief](README.md#search-words) says how a text uses them.

**Who searches, and how.** Google is 59 % of searches in Japan, Bing 31 % and Yahoo! 8 %, and Yahoo! JAPAN answers with Google's engine. Parents type a grade as 「小3」, 「3年生」 or 「小学3年生」, a band as 「低学年」 or 「高学年」, and frame the whole with 「小学生」. `topics.grades`, 「{range}年生」, matches them; the chips leave out the 「小学」 the card's 「小学1〜6年生」 has.

**What parents type**, strongest first:

1. 「論理パズル」 and 「算数パズル」, with 「小学生」;
2. 「算数 難問」;
3. 「思考力」 and 「思考力問題」;
4. 「算数 おもしろ問題」;
5. 「頭の体操」, also the title of a series of books;
6. 「応用問題」 and 「発展問題」, the school's words for problems beyond the textbook;
7. 「文章題」, the school's word for a problem in words, best with a topic;
8. 「特殊算」, the cram schools' name for the named kinds of problem, seven of which are the site's topics.

**The head's title.**

- A topic's page: **「{topic}の問題（小学{range}年生）」**. 「〇〇の問題」 is how parents type a topic and how every topic's description names it, and 「小学」 sets the pages apart from the same names at high school, 「場合の数と確率」 or 「鳩の巣原理 大学入試」. The longest, 「てんびんと水の移しかえの問題（小学3–6年生）」, is 43 units. The other way, 「{topic}の問題と解き方（{range}年生）」, 47 at most, takes 解き方 and loses 小学.
- Any other page: **「{page}｜MathTrail」**, a full-width bar with no space beside it, as the glossary sets full-width marks; Japanese sites put no dash before the site's name.

**Topics.** ⚑ marks a card's name nobody types. Where the card writes a word in kana and parents in kanji, search finds both, and the kanji may stand once beside the card's word.

| Topic | The card | What parents type |
|---|---|---|
| ordering | 「順番の推理」 ⚑ | 「推理算」, 「論理パズル」 |
| knights and liars | 「正直者とうそつき」 | 「正直者と嘘つき」, 「嘘つき問題」 |
| enumeration | 「場合の数」 | the same; 「樹形図」 |
| gaps and boundaries | 「植木算」 | the same |
| clocks | 「時計と時間」 | 「時刻と時間」; not 「時計算」, the angles of a clock's hands |
| calendar and age | 「カレンダーと年れい」 ⚑ | 「日暦算」, 「年齢算」 |
| pigeonhole principle | 「ハトの巣原理」 | 「鳩の巣原理」 |
| parity and alternation | 「偶奇と交互」 ⚑ | 「偶数と奇数の問題」 |
| arithmetic with a trick | 「計算のくふう」 | 「計算の工夫」, 「工夫して計算」 |
| weighing and pouring | 「てんびんと水の移しかえ」 ⚑ | 「天秤の問題」, 「油分け算」 |
| parts and shares | 「部分と全体」 ⚑ | 「相当算」, 「分数の文章題」; the card's name is a book's title |
| percentages | 「百分率」 | the same; 「割合」 |
| ratios and sharing | 「比と比例配分」 ⚑ | 「比の問題」; 「比例配分」 alone is shares in a company |
| figures on a grid | 「方眼紙の図形」 ⚑ | 「面積と周りの長さ」 |
| divisibility and remainders | 「倍数・約数とあまり」 | 「倍数と約数」, 「あまりのあるわり算」 |
| overlapping groups | 「集合算」 | the same; 「ベン図」 |
| games with a winning strategy | 「必勝法のあるゲーム」 | 「必勝法」, 「石取りゲーム」 |

The cram schools' names, 「推理算」, 「日暦算」, 「年齢算」, 「相当算」, 「油分け算」, name a kind of problem only, never with 「受験」 or 「入試」.

**Descriptions** can carry 「小学生」, 「思考力」, 「算数パズル」, 「論理パズル」, 「解き方」, 「考え方」 and 「無料」, a fact. Today only the description of "Service" runs past 160 units, to 178, and that of "Topics" carries none of these words.

**Not to use.**

- Contests and exams: 「キッズBEE」, 「ジュニア算数オリンピック」, 「思考力検定」, 「算数検定」, 「中学受験」, 「中学入試」.
- Brands: 「思考力プラス」, 「なぞぺー」, 「算数脳」.
- What the site does not give: 「プリント」, 「問題集」, 「ドリル」.
- 「てんびん算」, a cram schools' method for concentrations.

**For the author.** 「算数オリンピック」 is the name of a Japanese contest, and parents who type it look for that contest. The card and this glossary use it for the kind of problem, and the home page's title and description, "Techniques" and "Why" lead with it. The search words would keep it in the prose and lead titles and descriptions with 「思考力」, 「算数パズル」 or 「論理パズル」.

## Places for the author's reading

- The people of the examples: Japanese names in katakana with no suffix, and the first katakana in place of an initial, so ordering's badges become 「タ」 and 「フ」 and Kate's bead 「エ」. The alternative is hiragana names with さん, the textbook style, which would have to drop the さん in every drawing.
- The knight and the liar as 「正」 and 「う」; 「嘘」 was the other candidate.
- `topics.grades`, 「{range}年生」, which shows the build's en dash, 「3–4年生」, beside the card's and "Research"'s 〜; against 「小学{range}年生」.
- The menu: 「学ぶ意味」 for "Why" (the shorter 「なぜ」 reads oddly among nouns), 「仕組み」 for "Service", 「解法」 for "Techniques", and 「私たち」 for "About", in the menu and as the page's title, as the Spanish «Nosotros»; the fuller 「私たちについて」 fits the row too.
- 「Google ドライブ」, Google's own Japanese name and the card's, though the brief lists Google Drive among the names never translated; the Russian writes «Google Диск» the same way.
- 「お子さま」 for "your child", the card's word; お子さん, a shade less formal, is not used.
- Division always ÷, on every page, where the English topic pages write ":": in Japanese ":" is the sign of a ratio.
- The make-believe app's 「980円」 and 「0円」.
- Claude's labels come from its Japanese help, where "Connect" is 「接続」 in [A] and left in English in [B], and the free plan is 「Free」 in [A] and 「無料」 in [B]; a look at Claude's interface in Japanese would settle both. Its warning over a prompt and its question about saving are not in that help, so the text describes them.
- 「開発者モード」, the word of OpenAI's Japanese article, whose title now says 「デベロッパーモード」.
- 「リンク済みアプリ」, Google's current name for the page, against the card's 「サードパーティ製アプリのアクセス権」 in `data.access_how`, Google's earlier wording; the card may want the new one.
- 「算数オリンピック」 is also the name of a Japanese contest, held since 1992 by its own committee. The site uses the words only as the card does, for olympiad-style maths, and calls the contests of history and research 数学オリンピック; no text may name that contest or promise preparation for it.
- The card's own inconsistencies, which the site does not copy: the half-width ":" of `topic_choice.label` and `progress.link_refused`, the half-width "?" of `advice.percent_wrong_base`, and the missing space after MathTrail in `review.step_practice`, `problem.not_in_catalog` and `first_run.start`.
- Times as 「4:25」, as decided, where Japanese school maths writes 「4時25分」.
- Grandma is 「祖母」 in the one row of names in the clocks drawing that holds four characters, and 「おばあちゃん」 everywhere else.
- "Reference task" as 「お手本問題」, "the chat's model" as 「チャットの AI」 on the pages for parents, "What to say" as 「声かけ」, "Where children go wrong" as 「つまずきやすいところ」.
- The books marked less than sure: 『おもしろい計算術』, the volumes and years of 『数学おもちゃ箱』, and whether 『パズルの王様』 holds the whole of Amusements in Mathematics.
- The counters of "Research": the site's dictionary gives each noun with its counter (「問のお手本問題」, 「人」), and the page writes the number straight before it, 「120問のお手本問題」, 「12人」, with no space in Japanese.
