# Simplified Chinese (`zh-Hans`)

The Simplified Chinese half of [the brief](README.md): the rules the Simplified Chinese texts follow on top of it, and the words they use. The card's Simplified Chinese dictionary, [`web/locales/zh-Hans.json`](../../web/locales/zh-Hans.json), comes first wherever it has a word; this file adds only what the site needs beyond it. In this file «…» marks a Chinese text quoted; any “…” inside one belongs to the Chinese.

## Register and variant

- **One Simplified Chinese, the card's: mainland usage.** 账号, 软件, 视频, 网站, 信息; 档案 is the profile and 文件 a file (in Taiwan 檔案 is a file; here 档案 never is one). No Taiwan or Hong Kong forms, no 「」. Neither Claude nor ChatGPT is offered in mainland China and Google is blocked there, so many readers live in Singapore, Malaysia or further away: the text names no country's school system (年级 alone; no 小学/初中 cut, no textbook edition, no 小升初), no mainland institution, and no price but the make-believe one.
- **The reader is 您**, as the card speaks to the parent: «需要修改什么，聊天中会告诉您。» (`profile.not_saved_detail`), «您的数据» (`data.label`), «如果您愿意» (`first_run.fill`). A request to the reader opens with 请, as the card's «请在聊天中请求一道题。» (`waiting.ask_in_chat`); the numbered steps of a guide are bare imperatives, as Claude's own help writes them («输入连接器的名称。»). Chinese drops pronouns freely: 您 only where the sentence needs a subject or an owner, and never 你 for the reader.
- **The child is 你**, as the card speaks to the child: «选一个答案» (`task.pick`), «没能检查答案。再试一次吧。» (`task.answer_failed`), «为明天许个愿吧» (`waiting.limit`): an imperative with no pronoun, softened with 吧 where it suggests; 你 where a pronoun is needed. That covers the quotes under «可以这样说», the questions of a tip, Pólya's questions and the child's own lines.
- **What a parent types to the chat is plain and casual**, as people write to a chat, with no 您: «来一道新题», «来一道“排序推理”的题», «我的孩子学得怎么样？».
- **No gender shows, the child's or the parent's.** Chinese verbs and adjectives never show it; the written pronouns do (他 he, 她 she, 他们 and 她们), and so do 儿子/女儿, 男孩/女孩, 爸爸/妈妈. So:
  - "your child" is «您的孩子», then «孩子», or no subject at all, which Chinese allows; never 他 or 她, and not the internet's "TA". "Their interests" is «孩子的兴趣爱好»; the legal texts may write 其 for the child: «孩子及其兴趣爱好».
  - Children in general are «孩子» or «孩子们», never «他们»; counted on "Research", «名孩子».
  - "The parent" is the card's «家长», "a parent or a tutor" the card's «家长或老师» (`first_run.lead`), "the adult" the card's «成年人». 爸爸 and 妈妈 only where the English says Dad and Mum.
  - The people of the examples have a gender, as in the English, through their names: 小兰 is a girl, 小强 a boy. Repeat the name rather than write 他 or 她; where a pronoun is needed, it follows the English. The islanders are letters and take none: «A是骗子».
  - "Together" is «一起», «和孩子一起»: nothing to avoid.
- **An example's question is a task for the child**, as a task on the card is: a question or an imperative with no pronoun («一共有几根柱子？», «算一算……»), 你 where one is needed, never 您. The steps and the notes around it speak to the reader.
- **What a topic teaches**, `hero.teaches`, answers its label «学会什么» (`topic.teaches`): a verb phrase with no subject, no 学会 of its own and no final 。, a second verb joined with «，并» or «，也»: «把几条线索拼成一个顺序，并能正反两面理解比较», «不必一天一天地数，也能算出天数和年数».
- **Sentences every topic's page shares** read the same on every page: `method.lead` ends «下面的例题由浅入深，从最简单的一直到奥数竞赛题。», and `description` is «什么是……，怎样解，孩子常在哪里出错，以及在家可以怎样帮忙。», its first clause the English one's: «什么是排序推理题», «什么是关于钟表与时间的题».
- **The card's words, exactly.** A topic's name written out in a sentence is the card's text, never a paraphrase, though the English sometimes paraphrases ("putting things in order" is still «排序推理»). Where it is cited as a name (the English quotes it, or it stands in a list) it goes in “…”, as the card's «专题“{topic}”»; where it works as a noun of the sentence it stands plain: «排序推理和枚举法要先于真假话问题学». Names in quotes side by side take no 、 between them, as Chinese punctuation has it: «“枚举法”“奇偶性与交替”和“真假话问题”». A topic's `ask` is «来一道“{the card's name}”的题».

## Typography

- **Punctuation** is the card's full-width set: ，。、：；？！（）. No space before or after any of them, beside a Latin word too: «（MathTrail）», «Claude，». Never the half-width marks in Chinese prose.
- **Quotes:** “…”, and inside them ‘…’, as the card's «删除“{interest}”». Never 「」 or "…". A quote that is a whole sentence keeps its end mark inside: «A说：“我是骗子。”»; a quoted word or label takes the sentence's mark after the quote: «按“继续”。».
- **Spaces** as the card has them: one between Han characters and a Latin word, «MathTrail 就在», «Google 云端硬盘», «位于 {folder} 文件夹中»; none between Han characters and digits, «60分钟», «1～6年级», «{count}次». A single Latin letter that is a label (an islander, an option, a box) is not spaced either, as Chinese maths books write labels: «A说», «A和B都是骗子», «选项C». Arithmetic keeps its spaces: «1 + 3», «12 ÷ 3 = 4».
- **No full-width digits or letters.** Every digit is 0–9: the full-width digits an input method may type (U+FF10 to U+FF19) are digits of another script, which the site's checker refuses. The percent sign is %, never the full-width one (U+FF05); plus and equals are + and =; no full-width space (U+3000).
- **The ellipsis** in prose is the Chinese «……» (two of U+2026): «真、假、真、假……». In a formula, and where a text quotes the card's own status line, the single «…» as the English and the card write it: «1 + 2 + … + 10», «正在准备下一道题…».
- **Dashes:** the English spaced dash becomes a comma, a colon, parentheses or two sentences. Where a dash stays it is the Chinese «——» (two of U+2014, no spaces), as the card's «再来一题——专题“{topic}”».
- **Ranges** a text writes itself take the wave dash ～ (U+FF5E), as the card's «1～6年级»: «3～4年级», «7～12岁», «9～19世纪», «1509～1511年». A sign or unit is written at both ends where leaving one off would mislead, as GB/T 15835 asks: «70%～85%», never «70～85%». A range the build puts into a slot comes with an en dash (`{range}` is «3–4»), and a text that sets two slots apart uses the en dash too, as the card's «{from}–{to}年级»: «{first}–{last}年级», «{low}–{high}».
- **The middle dot:** « · » with a space on each side separates labels, as the English and the card's «奥数教练 · {grade}年级»; «·» with no space joins the parts of a foreign name: «乔治·波利亚».
- **Ordinals** are 第 and the number. In an example — its question, steps, note, answer and drawings — an ordinal the English writes in digits stays a digit, «第4次浇水», «第23面旗子», «第5格», and one it writes in words becomes Chinese numerals, «第七次», «第五格». Elsewhere, Chinese numerals in prose («第一步», «第三») and a digit in a label the English numbers in digits («第1步»). The badge "2nd", the second player, is «后手», the other half of the card's «先手» (`trap.first_move_assumed`).
- **Dates** in digits, as Intl writes them: «2026年10月10日», «5月3日», «25日»; "from the 3rd to the 10th" is «从3日到10日». A month is a digit too, «5月», as Intl writes a month with its year («2026年9月»), not the «五月» it gives a month alone. A weekday in prose is «星期一», in a drawing or a badge the short form Intl gives: «周一», «周二», «周三».
- **Centuries and decades** in Arabic digits, as GB/T 15835 writes them: «5世纪», «16世纪末», «20世纪初», «20世纪50年代». Never Roman numerals.
- **Times** as the English writes them: «4:25», «15:10», «0:00», the bait's «03:59». A length of time is «小时» and «分钟», in a drawing as in prose: «4小时30分钟», «40分钟», so the drawings' «{hours}小时{minutes}分钟». Never «分» alone, which is also a point or a mark, and never «时» for a length, which names a time of day.
- **Numbers outside the maths** as `Intl.NumberFormat("zh-Hans")` writes them: «12,450», the bait's coins, since Intl groups from four digits; «0.2», so the formula reads «P = 0.2 + 0.8·σ(θ + δ − β)». Millions in 万: "19 million" is «1900万». Years are never grouped: «2026年». A grade the text writes itself is a digit, as the card's «{grade}年级» and «1～6年级»: «3年级».
- **Counts on "Research".** A Chinese number takes a measure word before its noun, so the site's dictionary gives the page's nouns with theirs: `research.checks` is «项检查», `research.tasks` «道参考题», `research.children` «名孩子». A sentence writes «{count}{noun}» with nothing between and adds no measure word of its own: «通过了{count}{noun}的题». The counts the page writes by itself put the number straight before the noun, with no space in Chinese: «603道参考题».
- **Numbers in the maths** — the examples, the tables of a topic's page, and the sums of its cards and tips — keep the English digits, with no thousands separator, as the Russian and the Spanish do: «2026 × 2026», «1111», the arithmetic page's «10、100、1000» and its table's «1900», the percentages page's tip «1200»; a decimal keeps its point: «0.5 + ½ = 1». A number the English writes in words is written in Chinese numerals: «三个十是30».
- **Percent** with no space, as Intl writes it, everywhere, the page of percentages included: «70%», «20%». «百分之一» only where the English spells the number out.
- **Money** appears only in the make-believe app on "Why": «¥68» for the hat and «¥0» on the card of what MathTrail costs, the sign ¥ (U+00A5) as Intl writes it, no space and no decimals, as Chinese app stores show prices. ¥68 is the China App Store's price for an item sold at $9.99. The bike of the page of percentages costs «金币», a story's money, not 元.
- **Units** are Chinese words right after the number, in drawings as in prose: «12米», «200千克», «5升», «40分钟», «4小时», «7天», «3厘米». Not the symbols m, kg, l.
- **Signs:** × for multiplication; − (U+2212) for minus, never a hyphen; · in the formula. **Division is ÷ on every page**, those topic pages that write : (gaps, calendar, percentages, ratios, parts, arithmetic) included: Chinese school maths divides with ÷ alone and reads «3∶4» or «3:4» as a ratio, so «35 : 7 = 5» would read as a ratio on the very page of ratios. A ratio the English writes in words stays in words: «3比4», «1比2比3». Fractions as the English: «3/4», «½».
- **Short headings.** Chinese has no capitals to set and wraps at any character, so a heading or a label is as short as its meaning allows.
- **The card questions on the home page and on "Why"** write no Latin letter at all that the English does not: among Han characters every Latin capital stands apart, and the picture check reads it as a label of the picture.
- **Titles of works** — a book, a painting, a treatise — always take the title marks 《》, nested 〈〉, whether the English writes them plain or in italics: «《九章算术》», «《雅典学院》». Where the English writes the title as emphasis, the emphasis stays, inside the marks: «《*怎样解题*》».
- **Citations** (`cite.*`) name the authors as the source does, in Latin, with 和 for two and 等 for more: «Radesky 等，2022», «Agarwal 和 Gaule，2020»; in a sentence they stand in full-width parentheses: «（{source}）».
- **Emphasis beside Han characters.** The build reads `*…*` and `**…**` by CommonMark's rules, which refuse a mark standing between a Han character and a punctuation mark. Start and end the emphasis on a Han character, a letter or a digit, and put the sentence's punctuation after it: «**假设小斌是骑士**。», «**从哪里开始**？», «**矛盾**。». Where the emphasis has to start or end with a sign of its own, such as the + of «+ 添加», put a space on its outer side: «点击 **+ 添加**，». An emphasis that fails would show its asterisks on the page, and the build refuses a text whose emphasis is not the English one's.
- **YAML:** a text that holds `: ` (an ASCII colon and a space) or begins with `*`, `{` (a slot), `[`, `!`, `%`, `@` or a backquote is quoted, and so is a text that is a number or a time alone, as the English quotes them («"4:30"», «"87"»). A Chinese «：» needs no quotes. In a folded block (`>-`) every line break becomes a space, which Chinese does not want between Han characters: write such a text on one line, or break it only where a space belongs, beside a Latin word.

## Platforms' labels

### Claude

Claude's help speaks Simplified Chinese, and its interface has begun to. Claude's own list of interface languages still has eleven and no Chinese ([How to use Claude in your preferred language](https://support.claude.com/en/articles/10769299-how-to-use-claude-in-your-preferred-language), and [its Chinese copy](https://support.claude.com/zh-CN/articles/10769299)), but since early October 2026 the web and desktop apps offer Simplified and Traditional Chinese in the language menu, unannounced ([KuCoin, 6 October 2026](https://www.kucoin.com/news/flash/claude-adds-simplified-and-traditional-chinese-interfaces), citing ME AI News and BlockBeats). So the site writes Claude's labels as its Chinese help does, not in English as the Russian keeps them. The help speaks 您, as the site does; it quotes its labels in "…" and writes its paths with >, where the site uses bold or “…” and its own arrow.

- [A] «使用远程 MCP 开始自定义连接器», https://support.claude.com/zh-CN/articles/11175166-使用远程-mcp-开始自定义连接器
- [B] «使用连接器扩展 Claude 的功能», https://support.claude.com/zh-CN/articles/11176164-使用连接器扩展-claude-的功能

| English on the site | Simplified Chinese | Where the help says it |
|---|---|---|
| Customize → Connectors | 自定义 → 连接器 | A: «导航到 自定义 > 连接器». The site keeps its own arrow |
| + | + 添加 | A: «点击"+ 添加"，然后点击"添加自定义连接器"» |
| Add custom connector | 添加自定义连接器 | A |
| Name it MathTrail | 名称填写 MathTrail | A: «输入连接器的名称»; the field's label is not quoted |
| the address | 远程 MCP 服务器 URL; on the site, «地址» | A: «添加您的连接器的远程 MCP 服务器 URL» |
| Continue, and Continue again | 继续，再点一次继续 | A quotes «继续» after the URL and again after the settings |
| the sign-in settings Claude detects | Claude 检测到的身份验证设置 | A: «查看 Claude 检测到的身份验证设置» |
| Add | 添加 | A: «点击"添加"完成连接器配置» |
| Connect | 连接 | A, a team member's steps, and B: «点击"连接"或"安装"» |
| + → Connectors, in a chat | + → 连接器 | A: «通过聊天界面左下角的"+"按钮……然后选择"连接器"»; B the same |
| the free plan | 免费计划 | A: «免费用户限制为一个自定义连接器»; [«Pro 计划是什么？»](https://support.claude.com/zh-CN/articles/8325606): «比免费计划更多» |
| Always allow | 始终允许 | A, its section on security |
| Sign in now · Sign in when needed · No sign in | 立即登录 · 需要时登录 · 无需登录 | A, under «身份验证»; not on the site yet |
| Use Claude's published identity · Register automatically · Use your own OAuth client | 使用 Claude 的已发布身份（推荐） · 自动注册 · 使用您自己的 OAuth 客户端 | A, under «OAuth 客户端»; not on the site yet |
| "Use caution before running this prompt" | not found in Claude's Chinese help | the text describes it: «上方有一条提醒您谨慎发送的警告» |
| Claude asks whether to allow MathTrail to save | the question is not quoted in Claude's Chinese help; only «工具批准请求» and the button «始终允许» are | the text describes it: «Claude 会询问是否允许 MathTrail 保存档案» |

The help's Chinese reads like a machine's in places («导航到», «最大版» for Max), and the interface may word a button otherwise: a look at Claude in Simplified Chinese settles them.

### ChatGPT

| English on the site | Simplified Chinese | Source |
|---|---|---|
| developer mode | 开发者模式 | https://help.openai.com/zh-hans-cn/articles/12584461-developer-mode-and-mcp-apps-in-chatgpt-beta, «ChatGPT 中的开发者模式和 MCP 应用»; the path «设置 → 应用 → 高级设置», then «开发者模式», and «应用 → 创建» |
| ChatGPT's directory | ChatGPT 的目录 | OpenAI's Chinese help now names it «插件目录» (chatgpt.com/plugins) and the settings tab «插件»: https://help.openai.com/zh-hans-cn/articles/20001256 and https://help.openai.com/zh-hans-cn/articles/11487775-apps-in-chatgpt («很早的 iOS 或 Android 应用版本可能会显示应用，而不是插件»). The generic phrase stays true, as in the English |
| Plus, Pro, Business, Enterprise and Education plans | Plus、Pro、Business、Enterprise 和 Education 套餐 | names of plans; OpenAI's Chinese help calls a plan 套餐: «Business、Enterprise 和 Edu 套餐» |
| a free ChatGPT account | 免费的 ChatGPT 账号 | |
| app | 应用 | as OpenAI's Chinese help |

help.openai.com refuses fetchers; these labels come from search results that quote the articles and from a reader proxy's copy of them (r.jina.ai), taken on 11 October 2026. The developer-mode article still writes «设置 → 应用», while the apps article calls the tab «插件»; the site quotes no path, so nothing on it hangs on which.

### Google

| English on the site | Simplified Chinese | Source |
|---|---|---|
| Google Drive | Google 云端硬盘, short 云端硬盘; My Drive is «我的云端硬盘» | card `data.export_where`; https://support.google.com/drive/answer/2409045?hl=zh-Hans |
| Drive's bin | 云端硬盘的回收站 (card `data.delete_how`); the button «清空回收站»; files go for good after 30 days, «移至回收站的文件会在 30 天后永久删除» | https://support.google.com/drive/answer/2375102?hl=zh-Hans |
| version history | 版本记录; for a file such as the profile, the menu «管理版本» and «永久保留» | https://support.google.com/drive/answer/2409045?hl=zh-Hans |
| linked apps, third-party access | the «关联的应用» page of the Google account (myaccount.google.com/linkedapps, which the privacy policy links; the help heads it «第三方关联：由您掌控»); the filter «有权访问您的 Google 账号», then «查看详情» → «撤销访问权限» → «确认» | https://support.google.com/accounts/answer/13533235?hl=zh-Hans, «管理您的 Google 账号与其他开发者提供的应用之间的关联». "Delete all connections" is not in the current help: "Remove access" is «撤销访问权限» |
| the box for Google Drive on the consent screen | Google 云端硬盘的复选框; Google words the permission «查看、修改、创建和删除您通过此应用使用的特定 Google 云端硬盘文件» | as Google's Workspace Marketplace lists it in Chinese: https://workspace.google.com/marketplace/app/docs_sheets_editor_for_drive/703230099925?hl=zh-CN. Not seen on the live consent screen, and the English never quotes it |
| Sign in with Google | in prose, «用 Google 账号登录»; Google's name for the button is «使用 Google 账号登录» | https://developers.google.com/identity/branding-guidelines?hl=zh-cn |
| Google account | Google 账号 | Google's help writes 账号 throughout; the card's `data.access_how` writes the older 帐号 |

MathTrail's own sign-in pages speak English, so «Tick the box for Google Drive» and **Back to Google** stay English, in “…” or bold as the English has them.

## Glossary

"card" marks the card's own word, from `web/locales/zh-Hans.json`.

### Product and lesson

| English | Simplified Chinese | Note |
|---|---|---|
| task | 题、题目 | card: `task.label` «题目», «再来一题», «下一道题»; counted with 道 |
| problem (olympiad) | 题、奥数题 | the same word as task |
| school sum | 课本计算题、算式 | "A school sum" is «课本计算题» |
| topic | 专题 | card |
| group of topics | 专题分组; a group, 组 | card, `group.*` exactly |
| trap | 陷阱 | card, «陷阱» |
| the catch (not a trap) | 难点 | never 陷阱 outside the card's sense |
| technique | 技巧、解题技巧 | |
| the main move (`topic.method`) | 关键一招 | |
| hint | 提示 | card; never for "clue" |
| clue, a given of a puzzle | 线索 | «每条线索都是一个箭头» |
| condition | 条件 | card: «漏看了一个条件» |
| option | 选项 | card: «正在检查每个选项» |
| answer | 答案 | card |
| right, wrong answer | 答对、答错; 正确答案、错误答案 | card: «答对了！», «不太对哦» |
| solution | 解答; a way of solving, 解法 | card: «解答» |
| step | 步骤、一步 | |
| question, the text of a task | 题目 | card: «把题目里的数直接当成了答案» |
| question, what is asked | 问题; 题目问的 | card: «答非所问» |
| picture of a task | 图 | «照片» for a photograph |
| label of a picture | 标注 | |
| grade | 年级 | card |
| Grade 3 | 3年级 | card `child.grade` |
| grades 1 to 6 | 1～6年级 | card `first_run.lead` |
| grades 3–4, two slots | {from}–{to}年级 | card `progress.grades` |
| level | 水平 | |
| difficulty | 难度 | card: «已选好专题和难度» |
| rank | 等级 | card, «等级{rank}»: a number, never a title |
| rating | 评分 | card: «本专题评分», «总评分» |
| progress | 进度 | card, «档案和进度» |
| mastered | 已掌握 | card |
| review, the card's section | 分析 | card, `review.title` |
| how the answer went | 答案讲解 | card `result.label` |
| trial series | 定级练习 | card |
| skipped | 已跳过 | card |
| profile | 档案 | card; never 文件, which is a file |
| pseudonym | 化名 | card |
| interests | 兴趣爱好 | card |
| lesson | 课、一节课 | the card's «课程语言», quoted as it is |
| coach | 教练 | card, `topic_choice.coach_short`; the Coach product too |
| companion, in the coach | 伙伴 | a character, not the child |
| parent | 家长 | card |
| tutor | 老师 | card `first_run.lead` |
| adult | 成年人 | card |
| child, your child | 孩子; 您的孩子 | see "No gender shows" |
| children | 孩子、孩子们 | «名孩子» where "Research" counts them |
| chat | 聊天 | card: «在聊天中» |
| the chat's model | 聊天模型; 模型 | |
| card | 卡片 | card |
| connector, custom connector | 连接器、自定义连接器 | Claude |
| app | 应用 | OpenAI; card: «聊天应用» |
| sign in, sign-in | 登录 | card |
| account | 账号 | Google; see its table |
| Drive file, the profile's file | 云端硬盘里的文件、档案文件 | card: «整个档案就是……一个文件» |
| folder | 文件夹 | card |
| bin | 回收站 | card and Google |
| daily limit | 每天的上限 | card `waiting.limit_detail` |
| Another task | 再来一题 | card's button, quoted exactly |
| Make a wish for tomorrow | 为明天许个愿吧 | card, quoted exactly |
| Profile & progress | 档案和进度 | card |
| Edit | 修改 | card |
| the bulb, the hint's button | 灯泡 | |
| the two arrows | 两个箭头 | |
| tap | 点 | «点一下选项» |
| press, click | 点击、按 | Claude's help: «点击» |
| Enter, the key | 回车键 | |
| free, of price | 免费 | the chip `hero.free` |
| the free plan, Claude's | 免费计划 | Claude's help |
| plan, ChatGPT's | 套餐 | OpenAI's help |
| a plan, of either chat | no noun: «用的是您现有的额度» | so that no sentence picks one platform's word |
| open source | 开源 | |
| in development | 开发中 | |
| beta | Beta | |
| soon | 即将推出 | |
| olympiad maths | 奥数 | card: «奥数教练» |
| olympiad-style task | 奥数风格的题 | card `first_run.lead` |
| olympiad, the contest | 数学奥林匹克、数学竞赛 | the International Mathematical Olympiad is «国际数学奥林匹克» |
| maths circle | 数学兴趣小组 | |
| school | 学校 | card |
| school mark | 成绩 | card |
| homework | 家庭作业 | |
| "a new task, please", typed | 来一道新题 | the same words wherever the English has "a new task" |
| "Let's do a MathTrail task." | 我们来做一道 MathTrail 的题吧。 | |

### The site's sections and headings

| English | Simplified Chinese | Note |
|---|---|---|
| Topics (`nav.topics`) | 专题 | 2 characters; card |
| Techniques (`nav.techniques`) | 技巧 | 2; the page's title is «解题技巧» |
| Why (`nav.why`) | 为什么 | 3 |
| Coach (`nav.coach`) | 教练 | 2, beside the tag «Beta»; card |
| Research (`nav.research`) | 研究 | 2, and the page's title too |
| Service (`nav.service`) | 服务 | 2 |
| About (`nav.about`) | 关于我们 | 4, the page's title too: the footer names this page by the menu's word, as it names "Research", and «关于» alone names no page |
| Add to Claude (`nav.add`) | 添加到 Claude | Claude's «添加» |
| Technical pages | 技术页面 | said by a screen reader |
| Language | 语言 | |
| Help · Privacy policy · Terms of use | 帮助 · 隐私政策 · 使用条款 | |
| How it is solved | 怎样解题 | |
| Where children go wrong | 易错点 | the teachers' word for it |
| What to say | 可以这样说 | to the child |
| How to help at home | 在家怎样帮忙 | |
| Learn first | 前置专题 | |
| Opens the way to | 后续专题 | |
| Builds on · Opens | 前置：· 后续： | the topics page, and its map's line |
| First · Topic · Then, the map's key | 前置 · 专题 · 后续 | |
| Ask for a task on this topic | 请求一道本专题的题 | the card's verb, «请在聊天中请求一道题» |
| Show · Hide the solution | 查看解答 · 收起解答 | card: «收起提示» |
| Example 2 | 例2 | as Chinese textbooks number their worked examples |
| Problem, on "Techniques" | 例题 | |
| Answer | 答案 | |
| Why it matters | 为什么重要 | |
| How to solve: … | 解法：…… | |
| The rules of … | ……的规则 | |
| Check: | 检验： | |
| When it helps: | 适用于： | |
| In the topics: | 相关专题： | |
| You are here · On this page · All topics | 您的位置 · 本页内容 · 全部专题 | |
| More → | 了解更多 → | |
| See the topics → · See the coach → | 查看全部专题 → · 看看教练 → | |
| Honestly | 坦白说 | |
| See a lesson | 看看怎么上课 | |
| For technical readers | 写给技术读者 | |
| You, the parent's line on "Why" · C, the child's | 您 · 孩 | «“孩”代表您的孩子» |
| The adult · The model, in the chat's illustration | 成年人 · 模型 | |
| Step 1, of the topics' map | 第1步 | |
| Foundations · Rules · Hard problems | 基础 · 规律 · 难题 | |

`just site-fold` measures the Chinese menu, «关于我们» in it, in a row down to 984 px in Chromium and 971 px in WebKit, well inside the English one, so the fold at 1230 px holds.

### Maths and olympiad terms

| English | Simplified Chinese | Note |
|---|---|---|
| olympiad | 奥林匹克、奥赛 | |
| reference task | 参考题 | |
| solver | 求解程序 | first time, «一个能把题解出来的程序» |
| brute force | 穷举 | |
| worst case | 最坏的情况 | card, `trap.best_case_not_worst` |
| good, bad luck | 运气好、运气差; 最倒霉的情况 | card: «想一想最倒霉的情况» |
| to be sure, for certain | 保证、一定 | the word of Chinese pigeonhole problems: «至少要拿几颗，才能保证……» |
| pigeonhole principle | 抽屉原理 | card |
| pigeons, holes | 鸽子、鸽巢 | the textbooks' «鸽巢问题»; "Pigeons and pigeonholes" is «鸽子和鸽巢» |
| parity · even · odd | 奇偶性 · 偶数 · 奇数 | card |
| alternation, taking turns | 交替、轮流 | card |
| remainder · quotient | 余数 · 商 | card |
| divisibility | 整除 | card |
| divides by 9 | 能被9整除 | throughout |
| divisibility rules | 数的整除特征 | card |
| digit · two-digit number | 数字 · 两位数 | |
| tens, units digit | 十位、个位上的数字 | |
| enumeration | 枚举法 | card |
| case · list | 情况 · 列举 | |
| systematic listing | 有序列举 | the technique "List systematically" |
| tree, the diagram | 树状图 | |
| working backwards | 倒推法 | its name in Chinese school maths |
| the assumption method | 假设法 | the method of 鸡兔同笼, the hens and rabbits |
| proof by contradiction | 反证法 | |
| suppose · assumption · contradiction | 假设…… · 假设 · 矛盾 | |
| pattern | 规律 | «找规律» |
| symmetric strategy | 对称策略 | |
| winning strategy | 必胜策略 | card |
| winning, losing position | 必胜局面、必败局面 | |
| move, in a game | 一步 | |
| the first, the second player | 先手、后手 | card: «以为先手一定能赢» |
| the other player · opponent | 对方 · 对手 | |
| invariant | 不变量 | |
| knights and liars | 真假话问题 | card |
| knight · liar | 骑士 · 骗子 | 真 and 假 in drawings, marks and `thumb.*`; the words of Chinese competition problems, «总说真话的骑士», «总说假话的骗子» |
| island · islander | 岛 · 岛民 | |
| statement | 话、一句话 | card: «没有验证就相信了一句话» |
| ordering | 排序推理 | card |
| number line | 数轴 | card |
| grid · squared paper | 方格 · 方格纸 | card |
| cell · side of a cell | 格子 · 格子的边 | |
| area · perimeter | 面积 · 周长 | card |
| ratio · in the ratio 3 to 4 | 比 · 按3比4 | card; in words, as the English |
| sharing · share | 按比例分配 · 份 | card |
| percentage · percent | 百分数 · 百分之…… | card |
| fraction · the whole · the part | 分数 · 整体 · 部分 | card |
| a half, a third, a quarter | 一半、三分之一、四分之一 | |
| gaps and boundaries | 植树问题 | card |
| gap, between posts · post · fence | 间隔 · 柱子 · 篱笆 | card: «数间隔时差了一个», «用点表示柱子» |
| off by one | 差了一个 | card |
| clock face · hands · the whole hour | 钟面 · 指针（时针、分针） · 整点 | card: «钟面» |
| noon · midnight | 中午12点 · 午夜 | |
| floor 1, of a house | 1楼 | Chinese counts floors from the ground, as the English does: nothing changes |
| flight of stairs | 一段楼梯 | «从1楼走到9楼要爬8段楼梯» |
| calendar and age | 日历与年龄 | card |
| pan balance · weighing · fake coin | 天平 · 称一次 · 假币 | card: «天平», «称重与倒水» |
| weights | 砝码 | |
| jug · pour | 壶 · 倒 | card: «用来倒水的容器» |
| state, of the jugs | 状态 | |
| overlapping groups | 容斥原理 | card |
| arithmetic with a trick | 巧算 | card |
| I thought of a number | 我想好了一个数 | |
| carrying | 进位: «不用进位» | |
| round number | 整十数、整百数; making one, 凑整 | |
| match, the stick | 火柴 | |
| a special case | 特例 | |

### The technical pages

| English | Simplified Chinese | Note |
|---|---|---|
| token · access token · refresh token | 令牌 · 访问令牌 · 刷新令牌 | as Google's Chinese docs |
| authorization code | 授权代码 | Google's docs |
| seal · sealed · sealing | 密封 · 已密封 · 密封 | |
| encrypted | 加密 | |
| state · stateless | 状态 · 无状态 | |
| MCP server · tool · tool call | MCP 服务器 · 工具 · 工具调用 | |
| chat host | 聊天宿主 | |
| consent screen | 同意屏幕 | Google Cloud's «OAuth 同意屏幕» |
| scope | 权限范围; in prose, «drive.file 权限» | Google: «范围» |
| client · client metadata document | 客户端 · 客户端元数据文档 | |
| renewal · revocation | 续期 · 撤销 | |
| pace · ceiling | 速率 · 上限 | card: «每天的上限» |
| log · logging | 日志 · 日志记录 | «Cloud Logging» stays |
| instance · revision · image | 实例 · 修订版本 · 镜像 | as Cloud Run's Chinese docs |
| key rotation · ciphertext · nonce | 密钥轮换 · 密文 · nonce | |
| fingerprint · near-duplicate | 指纹 · 近似重复 | |
| readability · self-check | 可读性 · 自检 | |
| brief, for the model | 出题要求 | |
| package · schema · template | 出题材料包 · 模式（schema） · 模板 | |
| deterministic · the rule | 确定性的 · 规则 | |
| cookie · cache | Cookie · 缓存 | |
| graceful shutdown | 优雅关闭 | |
| lifetime, of a token | 有效期 | |
| Workload Identity Federation | 工作负载身份联合 | as Google Cloud's Chinese docs |
| student model | 学生模型 | |
| chance, of a right answer | 答对概率 | |
| corridor · interval · goal | 目标区间 · 区间 · 目标 | the corridor always with 目标; an interval with its level, «95%区间» |
| band, in a chart | 色带 | |
| the ceiling, an estimate that knows the child · oracle | 上限 · 理想估计 | |
| bench · simulated children | 测试台 · 模拟的孩子 | |
| a child who learns, who stays put | 在进步的孩子、原地不动的孩子 | |
| lag · logit · points | 滞后 · logit · 分 | |
| mastery declared falsely | 误判为已掌握 | |
| live data, live numbers | 真实数据 | they come monthly, not in real time |
| range, of the chance promised | 区段 | |
| standard error · baseline · seed | 标准误 · 基线 · 随机种子 | |
| Reached · On the edge · Not reached | 已达到 · 临界 · 未达到 | |
| paper · public domain | 论文 · 公有领域 | |
| commit · release | 提交（commit） · 版本 | |
| endpoint · claim, of a JWT | 端点 · 声明（claim） | |
| request, the model's · store · router · issuer | 请求 (card) · 存储 · 路由器 · 签发方 | |
| deploy · build · client secret · sandboxed | 部署 · 构建 · 客户端密钥 · 在沙箱中运行的 | |
| in the open, unsealed · timeout | 明文 · 超时 | |
| digest · runtime · PKCE verifier · state machine | 摘要（digest） · 运行时 · PKCE 代码验证器 · 状态机 | |
| scenario · sketch · pinned · sliding, of a lifetime | 场景 · 草图 · 固定的 · 滑动的 | |
| the way in · pure computation · around the core | 入口 · 纯计算 · 核心周边 | the map on "Service" |
| Sign-in, a scenario's tab on "Service" · target, of the profile's size | 登录 · 目标 | short: the tab and the bar's label do not wrap at 320 px |
| canvas · tablet | 画布 · 平板电脑 | the coach |
| GitHub issue · computer · mouse | issue · 电脑 · 鼠标 | «在 GitHub 上提交 issue» |
| the profile's notes | 备注 | |
| whisker, of an error bar · dashed | 误差须 · 虚线 | |
| item response theory · psychometrics | 项目反应理论（IRT） · 心理测量学 | |

## Pólya's four steps

The Chinese edition in print is «《怎样解题：数学思维的新方法》», G. 波利亚 著，涂泓、冯承天 译，上海科技教育出版社: first printed in 2002, again in 2007, 2011 and 2018 ([the 2007 printing](https://book.douban.com/subject/2124114/); [all editions](https://book.douban.com/works/1040246)). Its table names the four phases «理解题目», «拟订方案», «执行方案», «回顾», as the Chinese papers that cite it list them ([刘盼盼、陈维 2022](https://pdf.hanspub.org/ae20221200000_77825044.pdf); [Creative Education Studies, 2025](https://pdf.hanspub.org/ces_3095532.pdf)) and as a reader's note on the 2007 printing quotes its text: «我们必须理解该题目……拟定一个方案……我们执行我们的方案……我们回顾所完成的解答». The table itself was not seen; one of those papers writes the second phase «拟定方案» as well. The older translation, 阎育苏's (科学出版社, 1982), names them «弄清问题», «拟定计划», «实现计划», «回顾», which was checked no further than a search's summary. The page "Techniques" names its steps after the edition in print:

| English | Simplified Chinese |
|---|---|
| *How to Solve It* (1945) | 《*怎样解题*》（1945年） |
| Understand the problem | 理解题目 |
| Make a plan | 拟订方案 |
| Carry it out | 执行方案 |
| Look back | 回顾 |

A technique's name is a short verb phrase, as Pólya's phases are, and its "aka" the name Chinese school maths gives it: «把题画出来», «列个表格», «先试小数字» (特例), «有序列举», «从结尾往回推» (倒推法), «假设全都一样» (假设法), «找规律», «假设相反的情况» (反证法), «看奇偶», «鸽子和鸽巢» (抽屉原理), «对称地走» (对称策略).

The steps' questions speak to the child: «用自己的话说一说。你知道什么？要找的是什么？». The coach's shorter line, "Understand → make a plan → solve → check", is «理解题目 → 制订计划 → 解题 → 检查», and the picture's "understand, plan, solve, check" on "Why" is «理解、计划、解题、检查».

## People's names

The names follow the Chinese textbooks' «小» names (小明, 小红…): two characters, short enough for any drawing, their gender the English one's, and one Chinese name for each English one across the site. "Initial" marks a name whose first letter a drawing or a badge shows: in Chinese it is the name's own character after 小, so the bead of 小兰 shows «兰». No two characters repeat within a page, and no two names on the site sound alike.

| English | Simplified Chinese | Where | Initial, badge |
|---|---|---|---|
| Galt | Galt | home, the cards' pseudonym: «为 Galt 挑选» | kept, as the Spanish keeps it: an invented name shows no gender; spaced as a Latin word |
| Comet | 彗星 | "Why" and "Topics", the cards' pseudonym | a word, translated as the Russian «Комета» and the Spanish «Cometa»; a name, so no 这 or measure word before it |
| C, "your child" on "Why" | 孩 | `parent.explain.child`: «“孩”代表您的孩子» | the speaker's mark, one character as «您» is |
| Ann | 小安 | ordering, parts, ratios, clocks, calendar (where the art's key is `anya` and its value Ann) | 安 |
| Ben | 小明 | ordering, parts, ratios, parity | 明 |
| Kate | 小丽 | ordering, calendar | 丽 |
| Dan | 小刚 | ordering's line | name in a drawing |
| Lena | 小兰 | ordering's queue, gaps, ratios | 兰 |
| Max | 小军 | ordering's queue, ratios | 军 |
| Nina | 小娜 | ordering's queue | 娜 |
| Oleg | 小宇 | ordering's queue | 宇 |
| Pavel | 小强 | ordering's queue | 强, and the badge «强» |
| Ada | 小雪 | ordering's runners | 雪 |
| Bob | 小波 | ordering's runners | 波 |
| Cid | 小龙 | ordering's runners | 龙 |
| Dot | 小丹 | ordering's runners | 丹 |
| Eli | 小杰 | ordering's runners | 杰 |
| Fay | 小菲 | ordering's runners | 菲, and the badge «菲» |
| Anya | 小雅 | "Techniques": the pets and the knights | 雅, the legend's letter |
| Borya | 小斌 | "Techniques": the pets and the knights | 斌, the legend's letter |
| Vera | 小薇 | "Techniques", calendar | name in drawings |
| Tanya | 小彤 | "Techniques", working backwards | |
| Petya | 小鹏 | "Techniques", working backwards | |
| Lea | 小琳 | enumeration | 琳 |
| Tom | 小涛 | enumeration, clocks, parts | 涛 |
| Masha | 小梅 | gaps, arithmetic | 梅 |
| Sasha | 小华 | pigeonhole, the sweets | given to boys and girls alike, as Sasha is |
| Gleb | 小磊 | pigeonhole, the socks | |
| Kim | 小可 | winning strategy | name in drawings: «小可 −3» |
| Leo | 小乐 | winning strategy | name in drawings: «小乐 −4» |
| Grandma | 奶奶 | clocks | in drawings: «奶奶», «奶奶的钟» |
| Granny | 一位老奶奶 | parts, the egg seller at the market | not the family's grandmother |
| Mum, Dad | 妈妈、爸爸 | calendar, ordering's tip, "About" | |
| the older, the younger son | 大儿子、小儿子 | "About" | the family's own sons |
| islanders A, B, C | A, B, C | knights | Latin capitals, as Chinese maths labels them |
| knight K, liar L | 真, 假 | knights, `thumb.*` and the marks | the 真 of 真话 and the 假 of 假话; never an islander's letter |
| children A, B, C · boxes A, B | the same | enumeration · ratios | «A盒», «B盒» |
| the cook (she) · the baker (he) | 一位厨师 · 一位面包师 | weighing · parts | 她 and 他 only where the English has she and he |
| woodcutters · a farmer | 伐木工人 · 一位农民 | gaps · percentages | |

The people of history keep the names Chinese gives them: 欧几里得, 托勒密国王 (托勒密一世), 普罗克洛, 《几何原本》, 阿基米德, 拉斐尔, 《雅典学院》, 孔子, “六艺”, 阿耶波多, 浦那, 塔基丁, 伊斯坦布尔, 卢卡·帕乔利, 雅各布·德巴尔巴里, 花拉子米, 巴格达, 《九章算术》, 亚里士多德, 亚历山大大帝, 尼古拉·波格丹诺夫-别尔斯基, 《心算，在 S. A. 拉钦斯基的乡村学校》, S. A. 拉钦斯基, 贝茨学校, 圣路易斯, 东京, 金王八幡宫, 算额, 列宁格勒, 莫斯科, 匈牙利, 美国, 苏联, 俄国 (the chip of an 1895 painting), 乔治·波利亚. The inscription «Euclides» on Pacioli's slate stays as it reads. Euclid's saying is «几何学中没有王者之路。», the form Chinese histories of mathematics give it («几何无王者之道» is its classical variant). The word that al-Khwarizmi's title gave is the European «algebra», while Chinese 代数 is a translation of it, so the sentence says so: «书名给了世界“algebra”（代数）这个词». A photograph's credit, «Yu. D. Korolev» or «Momotarou2012», stays as its source writes it; the Main Archive of Moscow is «莫斯科总档案馆».

## Grades

`topics.grades` is **«{range}年级»**: «1–2年级», «3–4年级», «5–6年级», «1–6年级»; so `topic.example` reads «例2 · 3–4年级», and the label on "Techniques" «例题 · 3–4年级».

- It is the card's own phrase, «{from}–{to}年级» (`progress.grades`), to the character, so a chip on the site and the card's scale read alike, with the card's «3年级» (`child.grade`) beside them.
- Chinese wants no ordinal mark and no plural: «年级» after a range already reads "grades 3 to 4".
- The range comes from the build with an en dash, as the card's does; a range of grades the text writes itself takes the card's wave dash: «3～4年级».

The alternative is «适合{range}年级», "for grades 3–4", which says whom the topic suits but is longer on a chip and says what the chip's place already says. «{range} 年级», with a space, goes against the card. A text with two slots writes the card's form, «{first}–{last}年级»; the home page's lead, which has the one slot, says «面向{range}年级的孩子».

## Books on "Research"

| Key | The English page | Simplified Chinese | How sure |
|---|---|---|---|
| perelman.firsthundred | For Young Mathematicians. The First Hundred Puzzles | the English title | no Chinese edition found |
| perelman.secondhundred | The Second Hundred Puzzles | the English title | no Chinese edition found |
| perelman.living | Living Mathematics | 《生动的数学》 (于惠君 译，长江少年儿童出版社，2014) | unsure: the title is «Живая математика» word for word, and the book stands in the publisher's series of Perelman's popular science beside 《趣味物理学》 and 《趣味代数学》 ([Douban](https://book.douban.com/subject/25869611/)); its contents were not seen |
| perelman.arithmetic | Entertaining Arithmetic | the English title | no edition found: the Chinese 《趣味数学》 (余杰 编译) and 《趣味数学谜题》 (刘玉中 译，中国青年出版社，2011) are compilations, and neither shows which Russian book it is |
| perelman.quick | Quick Reckoning | the English title | no Chinese edition found |
| ignatyev.wit | In the Realm of Ingenuity, or Arithmetic for Everyone | the English title | no Chinese edition found |
| dudeney.amusements | Amusements in Mathematics | the English title | unsure: 《超多超好玩的烧脑数学谜题》 (张曙、张超斌 译，吉林出版集团，2018) has "400余道" puzzles on algebra, geometry, measuring, weighing, the chessboard and mazes, the shape of this book, but its original title was not seen ([Douban](https://book.douban.com/subject/27622476/)) |
| dudeney.canterbury | The Canterbury Puzzles | 《坎特伯雷趣题》 (陈以鸿 译，上海科技教育出版社，2007) | sure ([Douban](https://book.douban.com/subject/2288835/)) |
| loyd.cyclopedia | Cyclopedia of Puzzles | the English title | no translation found: 《劳埃德数学趣题大全》 (哈尔滨工业大学出版社，2016) reprints the English text in volumes under a Chinese title, and Gardner's selections, 《萨姆·劳埃德的数学趣题》, are other books |
| carroll.tangled | A Tangled Tale | the English title | no Chinese edition found |
| carroll.pillow | Pillow Problems | the English title | no Chinese edition found |

The authors' names: «雅科夫·别莱利曼», as every Chinese edition of his books writes him; «叶梅利扬·伊格纳季耶夫», the standard transcription of the Russian, with no edition to follow; «亨利·E. 杜德尼», as 《坎特伯雷趣题》 names him (Chinese Wikipedia writes 杜德耐); «萨姆·劳埃德»; «刘易斯·卡罗尔». A Chinese title goes in 《》, and an English one stays as the English page gives it, with no marks. `died`, «† {year}», is «卒于{year}年»: the cross reads as nothing in Chinese.

## Search words

Gathered on 2026-10-12 from Google's suggestions for Singapore and Malaysia and the titles of the first results, among them pages mainland parents read; they show that a phrase is typed, not how often, and [the brief](README.md#search-words) says how a text uses them.

**Who searches, and how.** In Singapore and Malaysia Google is over 90 % of searches. On the mainland Google is closed, Baidu is 47 % and Bing 22 %, and GitHub Pages, which serves the site, answers Baidu's crawler with an error, so there the site can be found through Bing at most. Parents type a grade in Chinese numerals, 「三年级」 or 「小学三年级」, and a range in digits, 「1-6年级」; Singapore and Malaysia also write 「小一」 to 「小六」. `topics.grades`, «{range}年级», matches a typed range.

**What parents type**, strongest first:

1. «奥数题», «小学奥数»;
2. a grade and «奥数题», «三年级奥数题»;
3. «数学思维题», «数学思维训练», the name mainland courses have sold 奥数 under since 2018;
4. «逻辑推理题»;
5. «应用题»;
6. «趣味数学题»;
7. «奥数练习题».

**The head's title.**

- A topic's page: **«{topic}（{range}年级奥数题）»**, the commonest shape of the pages that compete, with the topic, the grades and 奥数题, and with no 小学, as the glossary keeps 年级 alone. The longest, «比与按比例分配（5–6年级奥数题）», is 31 units. The other way, «{topic}：例题与解法（{range}年级奥数）», 41 at most, adds 例题 and 解法.
- Any other page: **«{page}｜MathTrail»**, full-width, with no space beside it, as the glossary sets full-width marks.

**Topics.** ⚑ marks a card's name nobody types.

| Topic | The card | What parents type |
|---|---|---|
| ordering | «排序推理» ⚑ | «逻辑推理题», «比高矮» |
| knights and liars | «真假话问题» | the same, rarely; «逻辑推理» |
| enumeration | «枚举法» | the same; «简单枚举», «搭配问题» |
| gaps and boundaries | «植树问题» | the same |
| clocks | «钟表与时间» ⚑ | «经过的时间», «时、分、秒»; not «钟表问题», the angles of a clock's hands |
| calendar and age | «日历与年龄» ⚑ | «年龄问题», «星期几问题» |
| pigeonhole principle | «抽屉原理» | the same; «鸽巢问题» |
| parity and alternation | «奇偶性与交替» | «奇偶性问题», «奇数和偶数» |
| arithmetic with a trick | «巧算» | the same; «速算与巧算» |
| weighing and pouring | «称重与倒水» ⚑ | «找次品», «倒水问题» |
| parts and shares | «分数应用题» | the same |
| percentages | «百分数» | the same; «百分数应用题» |
| ratios and sharing | «比与按比例分配» | «按比例分配», «比的应用», rarely |
| figures on a grid | «方格纸上的图形» ⚑ | «周长和面积», «数方格»; «方格纸» is paper to print |
| divisibility and remainders | «整除与余数» | «有余数的除法», «余数问题» |
| overlapping groups | «容斥原理» | the same; «重叠问题» |
| games with a winning strategy | «必胜策略» | the same; «抢30», «取火柴游戏» |

**Descriptions** can carry «奥数题», «数学思维», «逻辑推理», «解题方法», «解题思路» and «免费», a fact, and on "Why" what parents ask, «奥数有什么用».

**Not to use.**

- Contests: «数学思维大激斗», «数霸杯», SASMO, SIMOC, SMKC, NMOS, «奥数比赛».
- Exams and formats: «小升初», «题库», «大全», «100道», «pdf», «错题本»; the brand «举一反三».
- «小三» for a grade, which on the mainland also means a mistress.
- Words that search reads as something else: «天平» alone, a sign of the zodiac; «整除», programming; «按比例分配» alone, spreadsheets.

**For the author.** «小学» is the strongest framing parents type, «小学奥数», and the glossary keeps 年级 alone.

## Places for the author's reading

- Claude's labels, from its Chinese help: the Chinese interface came out unannounced in October 2026 and was not seen; a look at it settles the buttons. The other way is English labels with the Chinese help's beside them, at twice the length.
- Division as ÷ on every page, against the English pages that write : on six topic pages, since Chinese reads : as a ratio.
- `topics.grades`, «{range}年级», against «适合{range}年级».
- 骑士 and 骗子 for the knight and the liar. The other pair is 老实人 and 说谎者; «骗子», a cheat, is the harsher word, but Chinese competition problems use it.
- 账号, as Google writes it now, where the card's `data.access_how` writes 帐号: the fix belongs to the card's dictionary, and the site follows it.
- Two words for a plan: Claude's «免费计划» and ChatGPT's «套餐», each its help's; a plan of either chat is reworded.
- «ChatGPT 的目录» for the directory OpenAI's Chinese help now calls «插件目录»; its settings tab may now be «插件» rather than «应用».
- The names: «小» names, a bead showing the character after 小; Oleg as 小宇 and Lea as 小琳, chosen so that no two names sound alike (小丽 and 小莉 would).
- `topic.traps`, «易错点», a teachers' word, against the longer «孩子常在哪里出错»; `topic.method`, «关键一招».
- `topic.child`, «孩子», two characters in the 64-pixel label.
- «孩» for the child's line on "Why", a one-character mark beside «您».
- The make-believe price, «¥68», and «金币» for the coins of the bike problem.
- The books marked unsure: 《生动的数学》, and the candidate for "Amusements in Mathematics"; the Cyclopedia kept in English though an English reprint has a Chinese title.
- 塔基丁 for Taqi al-Din, for which no Chinese form was found in print; 帕乔利, the form Chinese books on accounting use, where Chinese Wikipedia writes 帕西奥利; 《心算，在 S. A. 拉钦斯基的乡村学校》 for Bogdanov-Belsky's painting, its two-part title as a Chinese art site gives it ([iqartis.cn](http://iqartis.cn/artwork/artwork_oil544.html)), where the Chinese press calls it 《拉钦斯基乡村学校的心算课》 ([知识分子](http://zhishifenzi.com/column/newsview/12194)).
- Pólya's phases from the edition in print, «理解题目 · 拟订方案 · 执行方案 · 回顾», against 阎育苏's older «弄清问题 · 拟定计划 · 实现计划 · 回顾».
- The consent screen's wording of the Drive permission comes from Google's Marketplace, not from the screen itself.
- What the code does to Chinese, which no translation can mend: the italic captions on "Techniques" and any `*…*` lean Han characters by a slant the font does not have.
