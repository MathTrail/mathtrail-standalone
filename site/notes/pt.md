# Brazilian Portuguese (`pt`)

The Brazilian Portuguese half of [the brief](README.md): the rules the Portuguese texts follow on top of it, and the words they use. The card's Portuguese dictionary, [`web/locales/pt.json`](../../web/locales/pt.json), comes first wherever it has a word; this file adds only what the site needs beyond it. Portuguese is quoted here in its own marks, “…”, and English in "…"; a text whose exact characters matter is set in `code`.

## Register and variant

- **Brazilian Portuguese, the card's.** There is no European text. The card writes “Salvar”, “Excluir”, “Baixe”, “configurações”, “arquivo”, “app”, and the site goes on in the same variant: “tela”, “celular”, “ônibus”, “mouse”, “equipe”, “cachorro”, “café da manhã”, “registro”, “usuário”, “Moscou”, “Bagdá”; never “guardar” for saving a file, where Brazil says “salvar”, as the card's button does (“guardar” for keeping, as the card's “não guarda nada”, is Brazilian too), “eliminar”, “ficheiro”, “ecrã”, “aplicação”, “definições”, “telemóvel”, “autocarro”, “rato”, “equipa”, “pequeno-almoço”, “registo”, “utilizador”, “Moscovo”, “Bagdade”. Spelling as Brazil applies the 1990 agreement: “ideia”, “voo”, “autoverificação”, “pseudônimo” (card), “econômico”, “Antônio”. An action in progress is the gerund, “a criança está resolvendo”, never “está a resolver”.
- **The reader is “você”**, with the verb in the third person, as the card speaks to the parent: “Peça um problema no chat.” (`waiting.ask_in_chat`), “Exclua o arquivo e esvazie a lixeira do Drive” (`data.delete_how`), “…que você pode abrir, baixar ou excluir” (`first_run.data`). Never “o senhor” or “a senhora”, which show a gender, and never “tu”. A plural reader is “vocês”.
- **The child is “você” too**, as the card speaks to the child: “Escolha uma resposta” (`task.pick`), “Faça um pedido para amanhã” (`waiting.limit`), “Tente de novo.” (`task.answer_failed`). Brazilian Portuguese has one “você” for both; only the tone differs, short and warm for the child. That covers the quotes under “O que dizer”, the questions of a tip, Pólya's questions and the child's own lines.
- **What a parent types to the chat** is a request as people type one, with no pronoun to place, such as “Um problema novo, por favor” and `Vamos fazer um problema do MathTrail.` Every topic's `ask` takes one shape, “Quero um problema de” and the card's topic name in lower case, or “do” where the name wants its article: “Quero um problema de cavaleiros e mentirosos”, “Quero um problema de ordenação”, “Quero um problema do princípio da casa dos pombos”.
- **No gender shows, the child's or the parent's.** Portuguese verbs never show it; adjectives, participles, nouns of persons and the articles and pronouns that go with them do. So:
  - "your child" is “a criança”, the card's word everywhere (“pelo progresso da criança”, `topic_choice.coach_note`): a feminine noun for any child, so whatever agrees with it agrees with the noun, as the card writes “o que ela ainda não viu na escola” (`first_run.fill`). Keep that “ela” right after “a criança”, or repeat the noun, or drop the subject, which Portuguese allows. Never “seu filho”, “o aluno”, “o menino”: their masculine reads as a boy, and “seu filho ou filha” is clumsy where “a criança” says it all.
  - No adjective or participle agrees with the child or the parent: “para ter certeza”, not “para estar seguro”; “se a criança travar” and “Travou?”, not “se ficar travado”; “sem ajuda”, “por conta própria”, not “sozinho”; “ter curiosidade”, not “ser curioso”, while “ter medo de errar” and “com medo” agree with nothing; “Boas-vindas” (`first_run.label`), not “Bem-vindo”. A verb never agrees, the past tense included: “Errei de novo…”, “Consegui!”, “Você chegou lá”. The parent's own trade is a verb too: “Não é preciso saber ensinar”, not “ser professor”.
  - A child's question to the parent is a verb: “Você mente?”, not “Você é mentiroso?”; in the island game the parent “sempre mente”.
  - Children in general are “crianças” (card `first_run.lead`), and “crianças” where "Research" counts them: one noun for boys and girls, so no count shows a gender. A group inside a task is the task's own, in the generic masculine Brazilian textbooks write: “300 alunos”, “os lenhadores”; the rule is about the reader's child. Where the site itself speaks of pupils or teachers, “estudantes”, “quem ensina”, “o corpo docente”, not “os alunos”, “os professores”.
  - "The parent" in the third person is the card's “a mãe ou o pai” (`first_run.lead`: “mãe, pai ou docente”), “os pais” in the plural and “para os pais” on a label (`profile.label`, `progress.grades_note`); "the tutor" is “docente” (card). The one who signs in, on the technical pages, is “a pessoa adulta” (card `first_run.lead`); ownership reads best as “da família”: “o Drive da família”.
  - The people of the examples have a gender, as in the English: “Laura não é a última”. The islanders are letters, in the generic masculine: “A é mentiroso”. A question whose answer is a person asks with a verb or an invariable phrase, “Quem mede menos?”, “Quem fica em último lugar?”, not “Quem é o mais baixo?”.
  - “Juntos” would show a gender: “com a criança”, “junto com a criança” (an adverb, invariable), “em família”.
  - What the reader and the child do together is the singular imperative to the reader with “com a criança”, never the plural “vocês”: “Treine com a criança”, “pergunte junto com a criança”.
  - On a page that speaks to “você”, a third-person “seu” reads as "your": write “dela”, “do serviço” or the noun again (“com o reflexo dela”, “nenhum registro do serviço”).
- **An example's question is a task for the child**, as a task on the card is, so it speaks to the child, with the imperative of “você” or impersonally: “Calcule…”, “Quantos postes há?”. The words around it, the steps and the notes, speak to the reader.
- **What a topic teaches**, `hero.teaches`, goes on from its label “O que ensina”, so it opens “A” and an infinitive (“ensina a contar”), and a second verb after “e” takes its “a” too: “A contar dias e anos…”, “A raciocinar passo a passo e a conferir…”.
- **Sentences every topic's page shares** read the same on every page: `method.lead` ends `Os exemplos abaixo vão do mais simples a um de olimpíada.`, and `description` is `O que são os problemas de …, como se resolvem, onde as crianças costumam errar e como ajudar em casa.`, its end always `…onde as crianças costumam errar e como ajudar em casa.`
- **The card's words, exactly.** A topic's name written out in a sentence is the card's, either in “…” with its capital, as the card writes `do tema “{topic}”` (`topic_choice.ask`), or in lower case without quotes: “a enumeração de casos”, “o princípio da casa dos pombos”. Never a paraphrase, though the English sometimes paraphrases ("putting things in order" is “ordenação”).
- **Products and companies take the article** Brazilian Portuguese gives them, contracted where a preposition meets it: “o MathTrail”, “do MathTrail”, “no Claude”, “ao Claude”, “o ChatGPT”, “o Google”, “no Google Drive”, “no GitHub”; a company is feminine, “a OpenAI”, “a Anthropic”. The card mostly does (“Boas-vindas ao MathTrail”, “perfil do MathTrail”, “o MathTrail escolhe”), though not everywhere; Claude's own help writes “o Claude” (“as configurações de autenticação que o Claude detectou”), and Google's own label is “Fazer login com o Google”. A title may start with the bare name: “MathTrail: matemática de olimpíada no seu chat”.

## Typography

- **Quotes:** “…”, as the card writes them (`topic_choice.ask`, `first_run.start`); a quote inside a quote ‘…’. The English “…” stay “…”, and the straight "…" of the English documents become “…” too. Never «…», which is the European mark.
- **A sentence that ends with a quote:** a quoted word or phrase that ends the sentence takes the period after the closing mark, as the card writes `…em “Perfil e progresso”.` (`first_run.start`); a quote that is a whole sentence keeps its own end mark inside, `diga: “Vamos fazer um problema do MathTrail.”`, and takes nothing after it. A `say:` that is the quote alone is written as the English writes it: the whole sentence, its mark inside.
- **Spaces:** none before `:` `;` `?` `!`, as in the card's file, where none stands. The card puts a no-break space before a dash where its English does (`topic_choice.ask`); the site's files write plain spaces, as the English, Russian and Spanish ones do, and the only no-break spaces on the site are in the two prices on "Why", the bait app's and MathTrail's own (see Money). Questions and exclamations open with nothing: Portuguese has no ¿ ¡; the bait's “!!!” as the English. The ellipsis is one character, “…”.
- **Dashes:** the spaced em dash, the Brazilian “travessão”, as the card writes it: `Não enviado — tente de novo` (`chat.not_sent`), “texto — inciso — texto”. Where Portuguese reads better, the English dash becomes a colon, a comma or two sentences. Ranges take the en dash with no spaces: “1–6”, “70–85%”, “séculos IX–XIX”.
- **Contractions and the crase** as the norm wants them: “à criança”, “às 4:25”, “ao chat”, “do tema”; “dizer à criança”, never “dizer a criança”.
- **Ordinals:** the digit and the indicator, with no point, as the card's `{grade}º ano`: “1º”, “2ª”. The indicators are º (U+00BA) and ª (U+00AA), never the degree sign ° (U+00B0), which looks the same. In an example (its question, steps, note, answer and drawings), an ordinal the English writes in digits stays in digits, in the noun's gender: “a 23ª bandeira”, “a 5ª casa”, “o 7º lançamento”; in the flags' drawing “21ª, 22ª, 23ª”; the badge "2nd" is “2º” (o segundo jogador). Elsewhere, words up to ten (“a sétima vez”), digits beyond.
- **Dates** use cardinals, “de 3 a 10 de maio”, “o dia 25”, “os dias 2, 9, 16, 23 e 30”, but the first of a month is an ordinal, as Brazil writes it: “1º de maio”, “o dia 1º do mês que vem”. A document's date: “Última atualização: 10 de outubro de 2026.” Months and weekdays are lower case: “maio”, “segunda-feira”. In an example's text a weekday is written whole, “numa quarta-feira”; in a drawing or a badge it is short, as Intl writes it and as the site's own week drawing (pigeonhole page) shows it, with the capital: “Seg.”, “Ter.”, “Qua.”, “Qui.”, “Sex.”, “Sáb.”, “Dom.”.
- **Centuries** in Roman numerals: “século V”, “fim do século XIX”, “séculos IX–XIX”; a decade “a década de 1950”.
- **Times** as the English writes them: “4:25”, “15:10”, “0:00”, the bait's “03:59”; “at 4:25” is “às 4:25”. A length of time: “4 h 30 min” in a drawing, “4 horas e 30 minutos” in prose; “meia hora”.
- **Numbers outside the maths** as `Intl.NumberFormat("pt")` writes them, which groups from four digits on, unlike the Spanish: “12.450”, the bait's coins; “9,99”; “0,2”, so the formula reads `P = 0,2 + 0,8·σ(θ + δ − β)`; “19 milhões”. Years are never grouped, though Intl would write “2.026”: “2026”, “1894”. A version or a date of code stays as written: “MCP 2026-07-28”.
- **Numbers in the maths** — the examples, the tables of a topic's page, and the sums of its cards and tips — keep the English digits and no thousands separator, as the Russian and the Spanish do: “2026 × 2026”, “1111”, “3816 = 9 × 424”, the arithmetic page's “10, 100 ou 1000” and its table's “1900”, the percentages page's tip “de 1200”. A decimal takes the comma: “0,5 + ½ = 1”, “1,5 + ½ = 2”.
- **Percent** takes no space, everywhere: Intl writes “70%” in Portuguese and the card's own percentages come out so. “70–85%” on the home page and on "Service" (whose English writes "70–85 %"), and “20%” on the page of percentages, in its words and drawings alike.
- **Money** appears only on "Why", in the make-believe app and on MathTrail's own card (`app.cards.free.price`), in reais as `new Intl.NumberFormat("pt-BR", {style: "currency", currency: "BRL"})` writes them: “R$ 9,90” for the hat, “R$ 0” for MathTrail. Intl writes a no-break space (U+00A0) between “R$” and the number, which keeps the two on one line in the app's narrow pills: type that character there, not a plain space (in a double-quoted YAML text the escape `\u00a0` gives the same character). Intl writes “R$ 0,00” unless told to drop the cents; the English "$0" has none, so “R$ 0”.
- **Units** are symbols after a space, with no point: “12 m”, “200 kg”, “40 min”, “4 h”, “7 d” (the calendar drawings' day), and the litre “L”, as Brazilian schoolbooks write it: “5 L”, `{count} L`. In prose, “metros”, “litros”, “quilos”.
- **Signs:** × for multiplication; − (U+2212) for minus, never a hyphen; · in the formula. Division as each English page writes it: ÷ on the home page, "Why", "Techniques" and "Research"; : on the topic pages that use it (gaps, calendar, percentages, ratios, parts, arithmetic). Fractions as the English: “3/4”, “½”.
- **Capitals:** sentence case in titles and headings: “Como ajudar em casa”. Capitals keep their accents: “Álgebra”, “Às 15:10”. “olimpíada” (with the card's accent), “ensino fundamental” and the languages are lower case in prose (“em português”); a named contest is capitalised: “a Olimpíada Internacional de Matemática”, “a Olimpíada Brasileira de Matemática das Escolas Públicas”. A weekday or a month that opens an answer or a table cell takes the capital the sentence gives it, as the answer `Segunda-feira.` does.
- **The card questions on the home page and on "Why"** write no lone Latin capital inside a sentence that the English does not: the picture check reads one there as a label of the picture. A capital that opens a sentence or a quotation is taken for a word, and inside a sentence the one-letter words “a”, “à”, “e”, “é”, “o” are lower case and harmless.
- **YAML:** a text that holds `: ` or begins with `*` is quoted. A text that is a number or a time keeps the quotes the English gives it (`"87"`, `"4:30"`, `"1"`), and “12.450” and “1.900” take them too, as the Spanish quoted its “12.450”: the site reads every value as text, so the quotes only keep other readers of the file from taking them for numbers. Accented letters and the marks `“` `”` `‘` `’` need no YAML quotes, and a text may open with `“` unquoted, as the English writes `say: “…”`; quoting it is harmless. “Sim” and “Não” are no YAML booleans. Inside a `>-` block nothing needs quoting.
- **Titles of works** (a book, a painting, a treatise) go in “…” where the English writes them plain, as the Russian and the Spanish mark them: the emphasis of a text is the English one's, and the build reads `*…*` as emphasis. Where the English sets a title in italics, so does the Portuguese: *A arte de resolver problemas*.

## Platforms' labels

### Claude

Claude's interface speaks Brazilian Portuguese: its help lists “Português (Brasileiro)” among the languages of the interface ([Como usar Claude no seu idioma preferido](https://support.claude.com/pt/articles/10769299-como-usar-claude-no-seu-idioma-preferido)). So its labels are those of its Portuguese help, not the English ones the Russian keeps. The help lives under `/pt/`: the `/pt-BR/` addresses lead to the English pages. It speaks “você” with the verb in the third person, as the site does (“Clique em…”, “Digite um nome…”), and presses a button with “Clique em”, which the site follows: “Clique em **+ Adicionar** e depois em **Adicionar conector personalizado**”.

- [A] “Comece com conectores personalizados usando MCP remoto”, https://support.claude.com/pt/articles/11175166-comece-com-conectores-personalizados-usando-mcp-remoto
- [B] “Use conectores para estender as capacidades do Claude”, https://support.claude.com/pt/articles/11176164-use-conectores-para-estender-as-capacidades-do-claude

| English on the site | Brazilian Portuguese | Where the help says it |
|---|---|---|
| Customize → Connectors | Personalizar → Conectores | A: “Navegue até Personalizar > Conectores”. The site keeps its own arrow. |
| + / + Add | + Adicionar | A: “Clique em "+ Adicionar" e depois em "Adicionar conector personalizado"” |
| Add custom connector | Adicionar conector personalizado | A |
| Name it MathTrail | Digite MathTrail como nome | A: “Digite um nome para o conector”; the field's label is not quoted |
| the address | a URL do servidor MCP remoto; on the site, “o endereço” | A: “Adicione a URL do servidor MCP remoto do seu conector” |
| Continue, and Continue again | Continuar, e mais uma vez Continuar | A quotes “Continuar” twice, after the URL and after the authentication settings, as the site's two presses |
| the sign-in settings Claude detects | as configurações de autenticação que o Claude detecta | A: “Revise as configurações de autenticação que Claude detectou”; B: “…que o Claude detectou” |
| Add | Adicionar | A: “Termine de configurar seu conector clicando em "Adicionar."” |
| Connect | Conectar | A: “Clique em "Conectar" para autenticar”; B: “Clique em "Conectar" ou "Instalar"” |
| + → Connectors, in a chat | + → Conectores | A: “através do botão "+" no canto inferior esquerdo da sua interface de chat, depois em "Conectores."” |
| the free plan | o plano Free | A: “Usuários do plano Free são limitados a um conector personalizado.” (B says “Usuários gratuitos”) |
| Always allow | Permitir sempre | A, its section on security: “clique em "Permitir sempre"”. B's per-tool setting reads “Sempre permitir”, another control. Not on the site yet |
| Sign in now · Sign in when needed · No sign in | Entrar agora · Entrar quando necessário · Sem entrada | A, under “Autenticação”; not on the site yet |
| Use Claude's published identity · Register automatically · Use your own OAuth client | Usar a identidade publicada do Claude (Recomendado) · Registrar automaticamente · Usar seu próprio cliente OAuth | A, under “Cliente OAuth”; not on the site yet |
| Claude's mobile apps · Claude Desktop | os aplicativos móveis do Claude · Claude Desktop | A: “os aplicativos móveis” |
| "Use caution before running this prompt" | not found in Claude's Portuguese help | the text describes it: “sob um aviso que pede cautela antes de enviá-lo” |
| Claude asks whether to allow MathTrail to save | the question is not quoted in the help; it speaks only of “solicitações de aprovação de ferramentas” and the button “Permitir sempre” | the text describes it: “O Claude pergunta se permite que o MathTrail salve o perfil” |

### ChatGPT

| English on the site | Brazilian Portuguese | Source |
|---|---|---|
| developer mode | **o modo de desenvolvedor** | https://help.openai.com/pt-br/articles/12584461-modo-desenvolvedor-e-apps-mcp-no-chatgpt-beta, “Modo de desenvolvedor e apps MCP no ChatGPT”; the path “Configurações → Apps → Configurações avançadas → Modo de desenvolvedor”, then “Apps → Criar” |
| ChatGPT's directory | o diretório do ChatGPT | OpenAI's Portuguese help says the app directory became the “Diretório de plugins” on 9 July 2026: https://help.openai.com/pt-br/articles/11487775-apps-in-chatgpt, and in the European help https://help.openai.com/pt-pt/articles/20001256-plugins-no-chatgpt-e-no-codex (“Desde 9 de julho de 2026, a OpenAI migrou o Diretório de aplicações para o Diretório de plugins”). The generic phrase stays true. |
| app, apps | app, apps | OpenAI's Brazilian help writes “apps”, and the card “o app” (`data.how_remove`) |
| Plus, Pro, Business, Enterprise, Education | the same | names of plans |

help.openai.com refuses fetchers: these labels come from search results that quote its Portuguese pages, and their capitals are not checked. Brazilian guides elsewhere write “Modo Desenvolvedor”; OpenAI's own title has “de”, and the site says “o modo de desenvolvedor”.

### Google

| English on the site | Brazilian Portuguese | Source |
|---|---|---|
| Google Drive | Google Drive; My Drive is “Meu Drive” | https://support.google.com/drive/answer/2424384?hl=pt-BR |
| Drive's bin | a lixeira do Drive (card `data.delete_how`); in Drive, “Lixeira” on the left, the buttons “Esvaziar lixeira” and “Excluir definitivamente”; files go for good after 30 days | https://support.google.com/drive/answer/2375102?hl=pt-BR |
| version history | o histórico de versões; for a file such as the profile, “Mais” → “Gerenciar versões”, and “Manter para sempre” | https://support.google.com/drive/answer/2409045?hl=pt-BR. Its heading says “Manter indefinidamente”; the button its steps name is “Manter para sempre” |
| linked apps, third-party access | a página de apps vinculados da sua Conta do Google (myaccount.google.com/linkedapps, the address the English links); the section “Acesso à sua Conta do Google”; “Ver detalhes” → “Remover acesso” → “Confirmar” | https://support.google.com/accounts/answer/13533235?hl=pt-BR, “Gerenciar vínculos da sua Conta do Google com terceiros”. "Delete all connections" has no wording in Google's current help: it named an older page. The card's `data.access_how` says “apps de terceiros com acesso”, Google's older name |
| the box for Google Drive on the consent screen | a caixa do Google Drive; Google words the permission “Ver, editar, criar e excluir apenas os arquivos do Google Drive que você usa com este app” | as Google lists it in Portuguese in its Workspace Marketplace: https://workspace.google.com/marketplace/app/docs_sheets_editor_for_drive/703230099925?hl=pt-BR. Not seen on the live consent screen, and the English never quotes it |
| Sign in with Google | in prose, “fazer login com o Google” (the card's “faço login”, `profile.signin_country_off`); Google's own name for the button and the feature is “Fazer login com o Google” | https://developers.google.com/identity/siwg/features?hl=pt-br, https://developers.google.com/identity/branding-guidelines?hl=pt-br |
| Google account | Conta do Google, with its capital | the card (`data.access_how`) and Google |

MathTrail's own sign-in pages speak English, so “Tick the box for Google Drive” and **Back to Google** stay English.

## Glossary

"card" marks the card's own word, from `web/locales/pt.json`.

### Product and lesson

| English | Brazilian Portuguese | Note |
|---|---|---|
| task | problema | card, `task.label` |
| problem (olympiad) | problema | the same word as task |
| school sum | conta, exercício de escola | "A school sum" is “Uma conta de escola” |
| topic | tema | card |
| group of topics | grupo | card, `group.*` exactly |
| trap | armadilha | card, “A armadilha” |
| the catch (not a trap) | a pegadinha, a dificuldade | never “armadilha” outside the card's sense; “pegadinha” is the card's own (`group.numbers`) |
| technique | técnica | |
| the main move (`topic.method`) | a ideia-chave | |
| hint | dica | card |
| clue, a given of a task | pista | “Cada pista é uma seta”; “dica” stays the card's hint |
| option | alternativa | the word Brazilian tests and olympiads use for A to E (OBMEP, Canguru); the card lists them under “Respostas”. “Opção” is a choice in counting: “3 opções para o 1º algarismo” |
| answer | resposta | card |
| right, wrong answer | resposta certa, errada | card: “Certo!”, “respostas erradas”; “a alternativa correta”, as a test says it |
| solution | solução | card |
| step | passo | never “etapa”, the card's rank |
| question, the text of a task | enunciado | card: “Número do enunciado tomado como resposta” |
| question, what is asked | pergunta | card: “Resposta a outra pergunta” |
| picture of a task | desenho | “imagem” for a photograph or a painting; “figura” only for a shape, as in the card's “Figuras na malha quadriculada” |
| label of a picture | rótulo | |
| grade | ano, ano escolar | card `profile.grade`: “Ano escolar” |
| Grade 3 | 3º ano | card `child.grade` |
| grades 1 to 6 | do 1º ao 6º ano | card `first_run.lead` |
| grades 3–4, two slots | {from}º–{to}º ano; in a sentence, do {from}º ao {to}º ano | card `progress.grades`, `progress.grades_ranks` |
| level | nível | card |
| difficulty | dificuldade | card |
| rank | etapa | card, “Etapa {rank}”: a number, never a title |
| rating | pontuação | card: “Pontuação neste tema”, “Pontuação geral” |
| progress | progresso | card, “Perfil e progresso” |
| mastered | dominado | card |
| review, the card's section | análise | card, “Análise” |
| how the answer went | a análise da resposta | card `result.label` |
| trial series | série de teste | card |
| skipped | em branco | card |
| profile | perfil | card |
| pseudonym | pseudônimo | card; the Brazilian circumflex |
| interests | interesses | card |
| lesson | aula | card: “Escolhido para as aulas”; it says “lições” in “Idioma das lições”, quoted as it is |
| coach | treinador | card, `topic_choice.coach_short`; the Coach product too. The card's own pill, `app.badge`, calls MathTrail “Treino de olimpíada” |
| companion, in the coach | companheiro | a character, not the child |
| parent | a mãe ou o pai; os pais; “para os pais” on a label | card |
| tutor | docente | card `first_run.lead` |
| adult | pessoa adulta | card |
| child, your child | a criança | card; see "No gender shows" |
| children | crianças | card; on "Research" too |
| chat | o chat | card |
| card | cartão | card |
| connector, custom connector | conector, conector personalizado | Claude |
| app | app (o app), apps | card and OpenAI; “aplicativo” where Claude's help speaks of its mobile apps |
| sign in | fazer login | card (`profile.signin_country_off`) and Google |
| sign-in | login | “o login”; the Service page's tab “Login” |
| account | conta; Conta do Google | card |
| Drive file, the profile's file | arquivo do Drive, arquivo do perfil | card |
| folder | pasta | card |
| bin | lixeira | card and Google |
| delete · download · save | excluir · baixar · salvar | card and Google |
| daily limit | limite do dia | card: “esse é o limite do dia” |
| Another task | Outro problema | card's button, quoted exactly |
| Make a wish for tomorrow | Faça um pedido para amanhã | card, quoted exactly |
| Profile & progress | Perfil e progresso | card |
| Edit · Save · Cancel | Editar · Salvar · Cancelar | card |
| the bulb, the hint's button | a lâmpada | |
| the two arrows | as duas setas | |
| tap | tocar | the child's taps on the card |
| press, click | clicar: “clique em”, for a button in Claude, as its help; “pressione” for a key | |
| Enter, the key | Enter | as Brazilian keyboards name it |
| type | digitar | |
| free, of price | grátis | the chip `hero.free` |
| Free, the plan | Free | “o plano Free”: a plan's name |
| open source | código aberto | |
| in development | em desenvolvimento | |
| beta | Beta | |
| soon | em breve | |
| olympiad maths | matemática de olimpíada | card `first_run.lead`: “problemas de olimpíada de matemática” |
| olympiad-style task | problema de olimpíada, problema no estilo de olimpíada | |
| maths circle | círculo de matemática | as OBMEP names its own “Círculos de Matemática”; “clube de matemática” where clearer |
| school | escola | card |
| school mark | nota | card: “nota escolar” |
| homework | dever de casa | the most widespread Brazilian name; São Paulo says “lição de casa” |

### The site's sections and headings

| English | Brazilian Portuguese | Note |
|---|---|---|
| Topics (`nav.topics`) | Temas | 5 characters |
| Techniques (`nav.techniques`) | Técnicas | 8 |
| Why (`nav.why`) | Por quê | 7: alone, “por quê” takes its circumflex; the page's title, “Por que matemática de olimpíada”, does not |
| Coach (`nav.coach`) | Treinador | 9, beside the tag “Beta”; the card's word, and no shorter one |
| Research (`nav.research`) | Estudo | 6. “Pesquisa”, 8, is the fuller word; see below. The page's title is “Estudo” too, as a page's title is its menu entry in every language, and the footer shows it |
| Service (`nav.service`) | Serviço | 7 |
| About (`nav.about`) | Sobre | 5, and the page's title too; “Sobre nós”, 9, is the fuller form |
| Add to Claude (`nav.add`) | Adicionar ao Claude | 19: Claude's own verb, “+ Adicionar” |
| Technical pages | Páginas técnicas | said by a screen reader |
| Language | Idioma | |
| Help · Privacy policy · Terms of use | Ajuda · Política de privacidade · Termos de uso | |
| How it is solved | Como se resolve | |
| Where children go wrong | Onde as crianças erram | |
| What to say | O que dizer | to the child |
| How to help at home | Como ajudar em casa | |
| Learn first | Aprender antes | |
| Opens the way to | Abre caminho para | |
| Builds on · Opens | Depende de · Abre caminho para | the topics page |
| Ask for a task on this topic | Peça um problema deste tema | card: “do tema” |
| Show · Hide the solution | Mostrar a solução · Esconder a solução | card: “Esconder dica” |
| Example 2 | Exemplo 2 | |
| Why it matters | Por que isso importa | |
| How to solve: … | Como resolver: … | |
| The rules of … | As regras de… | |
| Check: | Conferindo: | the child's and the solver's own check is “conferir”, as the card's advice says it; “verificar” is the program's checks |
| When it helps: | Quando ajuda: | |
| In the topics: | Nos temas: | |
| You are here · On this page · All topics | Você está aqui · Nesta página · Todos os temas | |
| More → | Saiba mais → | |
| Honestly | Sinceramente | |
| See a lesson | Ver uma aula | |
| See the coach → · See the topics → | Ver o treinador → · Ver os temas → | |
| For technical readers | Para o público técnico | no gendered “leitores” |
| You, the parent's line on "Why" · C, the child's | Você · C | “C é a criança”: the English letter fits |
| The adult · The model, in the chat's illustration | Pessoa adulta · Modelo | card |
| Step 1, of the topics' map | Passo 1 | never “Etapa” |
| Foundations · Rules · Hard problems | Fundamentos · Regras · Problemas difíceis | |
| First · Topic · Then, the map's key | Antes · Tema · Depois | |
| Groups of topics | Grupos de temas | |

Not measured yet. By the advance widths of a common sans-serif, scaled by what the Spanish change measured, the Portuguese row with “Estudo” and “Sobre” comes out some 10 px narrower than the Spanish one, which holds in a row down to 1157 px in Chromium and 1143 in WebKit; with “Pesquisa” and “Sobre” it comes out level with the Russian, 1163 and 1149, and with “Pesquisa” and “Sobre nós” some 30 px past it. So the menu takes the short words, and `just site-fold` says whether “Pesquisa” fits back in, the first word to restore; the menu still folds below 1230 px.

### Maths and olympiad terms

| English | Brazilian Portuguese | Note |
|---|---|---|
| olympiad | olimpíada | the card's accent |
| reference task | problema de referência | |
| solver | resolvedor | first time, “um programa que resolve o problema” |
| brute force | força bruta | |
| worst case | o pior caso | card, `trap.best_case_not_worst` |
| good, bad luck | sorte, azar | card: “o jeito mais azarado de tudo acontecer” |
| to be sure, for certain | para ter certeza, com certeza | never “para estar seguro” |
| pigeonhole principle | princípio da casa dos pombos | card |
| pigeons, holes | pombos, casas | as OBMEP's texts say it; "Pigeons and pigeonholes" is “Pombos e casas”. On that page a cell of the grid is a “quadradinho”, so that “casa” stays the pigeonhole |
| parity · even · odd | paridade · par · ímpar | card |
| alternation, taking turns | alternância, alternar-se, revezar-se | card |
| remainder · quotient | resto · quociente | card |
| divisibility | divisibilidade | card |
| divides by 9 | é divisível por 9 | “dividir por” throughout, never the Spanish “entre” |
| divisibility rules | critérios de divisibilidade | card |
| digit · two-digit number | algarismo · número de dois algarismos | Brazilian school's word |
| tens, units digit | algarismo das dezenas, das unidades | |
| enumeration | enumeração de casos | card |
| case · list | caso · lista | |
| systematic listing | listar em ordem | card's advice: “Listar os casos numa ordem fixa” |
| tree, the diagram | árvore de possibilidades | the name Brazilian textbooks give it |
| working backwards | começar pelo fim, de trás para a frente | |
| the assumption method | o método da suposição | hens and rabbits: “suponha que todos sejam galinhas” |
| proof by contradiction | redução ao absurdo | |
| suppose · assumption · contradiction | suponha que… · suposição · contradição | |
| pattern | padrão | the BNCC also says “regularidade” |
| symmetric strategy | estratégia simétrica | |
| winning strategy | estratégia vencedora | card |
| winning, losing position | posição vencedora, perdedora | |
| move, in a game | jogada | card |
| the first, the second player | o primeiro jogador, o segundo | |
| invariant | invariante | |
| knights and liars | cavaleiros e mentirosos | card |
| knight · liar | cavaleiro · mentiroso | V and M in drawings; “cavaleiro” has no current feminine in this sense, so a named knight is a boy |
| island · islander | ilha · habitante da ilha | |
| statement | afirmação | card |
| ordering | ordenação | card |
| taller · the tallest | mede mais · quem mede mais | a verb, which agrees with no one: the drawing's word is “mede mais” |
| number line | reta numérica | card |
| grid · squared paper | malha quadriculada · papel quadriculado | card topic; a grid of numbers is a “grade” (card `picture.grid`) |
| cell | quadradinho, on squared paper and in a grid or a table; casa, in a strip, on a board or on the number line | Brazilian school counts area in “quadradinhos”; the card's `picture.grid` says “Uma grade de casas”, and a board's square is a “casa” |
| side of a cell | lado de um quadradinho | |
| area · perimeter | área · perímetro | card |
| strip | tira | “faixa” stays the range on "Research" |
| bar · box, of a ratio · counter | barra · caixinha · ficha | card: `picture.bars`, `advice.ratio_total_confusion`, `picture.piles` |
| ratio · in the ratio 3 to 4 | razão · na razão de 3 para 4 | card |
| sharing · share | divisão · parte | card: “divisão proporcional” |
| percentage · percent | porcentagem · por cento | card |
| fraction · the whole · the part | fração · o todo · a parte | card |
| a half, a third, a quarter, sevenths | a metade, um terço, um quarto, sétimos | |
| gaps and boundaries | intervalos e extremos | card |
| gap, between posts · post · fence | espaço · poste · cerca | card's advice: “os postes como pontos e os espaços entre eles”; “intervalo” is the topic's and the trap's word |
| off by one | um a mais ou a menos | card |
| clock face · hands · the whole hour | mostrador · ponteiros · a hora cheia | card: “Um mostrador de relógio” |
| floor 1, of a house · flight of stairs | o 1º andar · lance de escada | never “térreo”, which would change the count |
| calendar and age | calendário e idades | card |
| pan balance · weighing · fake coin · weight | balança de pratos · pesagem · moeda falsa · peso | card |
| jug · pour · fill · empty | jarra · despejar · encher · esvaziar | “transvasar” is the card's name for the picture (`picture.containers`) |
| state, of the jugs | estado | |
| overlapping groups | grupos que se sobrepõem | card |
| arithmetic with a trick | aritmética com truque | card |
| I thought of a number | pensei em um número | |
| carrying | o “vai um”: “sem vai um” | what parents call it; the BNCC says “reagrupamento” |
| round number | número redondo | |
| match, the stick | palito de fósforo, palito | |
| a special case | um caso particular | |
| sweets · dog · ferry · bus stop · market | balas · cachorro · balsa · ponto de ônibus · feira | the Brazilian words |

### The technical pages

| English | Brazilian Portuguese | Note |
|---|---|---|
| token · access token · refresh token | token · token de acesso · token de atualização | as Google's Portuguese docs |
| authorization code | código de autorização | Google |
| seal · sealed · sealing | selar · selado · selagem | |
| encrypted | criptografado | as Claude's help |
| state · stateless | estado · sem estado | |
| MCP server · tool · tool call | servidor MCP · ferramenta · chamada de ferramenta | Claude's help: “chamadas de ferramentas” |
| chat host | host do chat | |
| consent screen | tela de permissão | Google's own name, “tela de permissão OAuth” (https://developers.google.com/identity/protocols/oauth2/web-server?hl=pt-br); MathTrail's own consent screen is one too |
| scope | escopo | Google: “o escopo drive.file”; the privacy policy says “permissão”, as the consent screen does |
| client · client metadata document | cliente · documento de metadados do cliente | |
| client secret | chave secreta do cliente | Google |
| renewal · revocation | renovação · revogação | |
| pace · ceiling | ritmo · teto | |
| Sign-in, a scenario's tab on "Service" · target, of the profile's size | Login · meta | short: the tab and the bar's label do not wrap at 320 px |
| log · logging | registro, log | “Cloud Logging” stays |
| instance · revision · image | instância · revisão · imagem | Cloud Run |
| key rotation · ciphertext · nonce | rotação de chaves · texto cifrado · nonce | |
| fingerprint · near-duplicate | impressão digital · quase repetição | |
| readability · self-check | legibilidade · autoverificação | one word, per the 1990 spelling |
| brief, for the model | briefing | the loanword Brazilian offices use |
| package · schema · template | pacote · esquema · modelo | |
| deterministic · the rule | determinístico · a regra | |
| cookie · cache | cookie · cache | |
| graceful shutdown | encerramento controlado | |
| lifetime, of a token | validade | |
| Workload Identity Federation | Federação de identidade da carga de trabalho | as Google Cloud's Portuguese docs |
| student model | modelo do estudante | |
| chance, of a right answer | chance de acerto | |
| corridor · interval · goal | corredor · intervalo · meta | |
| the ceiling, an estimate that knows the child · oracle | o teto · oráculo | |
| bench · simulated children | bancada de testes · crianças simuladas | |
| a child who learns, who stays put | quem aprende, quem não avança | |
| lag · logit · points | atraso · logit · pontos | |
| mastery declared falsely | domínio declarado por engano | |
| live data, live numbers | dados reais, números reais | they come monthly: never “ao vivo” |
| range, of the chance promised | faixa | |
| standard error · baseline · seed | erro padrão · referência · semente | |
| Reached · On the edge · Not reached | Atingida · No limite · Não atingida | they agree with “a meta” |
| paper · public domain | artigo · domínio público | |
| commit · release | commit · versão | “commit” is the borrowed term |
| endpoint · claim, of a JWT | endpoint · claim | borrowed terms |
| request, the model's · request, HTTP · store · router · issuer | pedido (card) · requisição · armazenamento · roteador · emissor | |
| deploy · build · sandboxed | implantação · build · em sandbox, isolado | |
| in the open, unsealed · timeout | às claras · tempo limite | |
| digest · runtime · PKCE verifier · state machine | digest · ambiente de execução · verificador PKCE · máquina de estados | |
| scenario · sketch · pinned · sliding, of a lifetime | cenário · esboço · fixado · deslizante | |
| canvas · tablet | área de desenho · tablet | the coach; “tela” is a screen |
| GitHub issue · computer · mouse | issue (uma issue) · computador · mouse | GitHub's Portuguese docs call an issue “problema”, the word for a task here, so the loanword Brazilian developers use |
| the profile's notes | anotações | “nota” is a school mark |
| whisker, of an error bar · dashed | haste · tracejada | |
| item response theory · psychometrics | teoria de resposta ao item (TRI) · psicometria | the TRI Brazilian parents know from the ENEM |

## Pólya's four steps

The Brazilian edition is *A arte de resolver problemas: um novo aspecto do método matemático* (Interciência, Rio de Janeiro, 1978; rights 1977, reprinted 1986 and 1995; translated and adapted by Heitor Lisboa de Araújo). Its first part names the phases, in sections 7, 9, 11 and 13: “Compreensão do problema”, “Estabelecimento de um plano”, “Execução do plano”, “Retrospecto”; its introduction: “Primeiro, temos de compreender o problema… Segundo, … para estabelecermos um plano. Terceiro, executamos o nosso plano. Quarto, fazemos um retrospecto da resolução completa”. Checked in the scan of the 1995 reprint: https://archive.org/details/AArteDeResolverProblemasGeorgePolya. The page "Techniques" names its steps after them:

| English | Brazilian Portuguese |
|---|---|
| *How to Solve It* (1945) | *A arte de resolver problemas* (1945) |
| Understand the problem | Compreender o problema |
| Make a plan | Estabelecer um plano |
| Carry it out | Executar o plano |
| Look back | Fazer um retrospecto |

A technique's name is an infinitive, as Pólya's steps are: “Desenhar o problema”, “Listar em ordem”, “Começar pelo fim”.

The steps' questions speak to the child: “Diga com as suas palavras. O que você sabe? O que você precisa encontrar?” The coach's shorter line, "Understand → make a plan → solve → check", is “Compreender → fazer um plano → resolver → conferir”, and the picture's "understand, plan, solve, check" on "Why" is “compreender, planejar, resolver, conferir”.

## People's names

"Initial" marks a name whose first letter a drawing or a badge shows; within one drawing no two letters repeat.

| English | Brazilian Portuguese | Where | Initial, badge |
|---|---|---|---|
| Galt | Galt | home, the cards' pseudonym: “Escolhido para Galt” | kept, as the Russian and the Spanish keep it; an invented name shows no gender |
| Comet | Cometa | "Why" and "Topics", the cards' pseudonym | as the Russian “Комета” and the Spanish “Cometa”; a name, so never “o” or “a” before it |
| C, "your child" on "Why" | C | `parent.explain.child`: “C é a criança” | the speaker's letter, which “criança” keeps |
| Ann | Ana | ordering, parts, ratios, clocks, calendar | A, kept |
| Ben | Bruno | ordering, parts, ratios, parity | B, kept |
| Kate | Clara | ordering, calendar | the letter changes, K → C: `art.kate` is “C” |
| Dan | Davi | ordering's line | name in a drawing |
| Lena | Laura | ordering's queue, gaps, ratios | L, kept |
| Max | Miguel | ordering's queue, ratios | M, kept |
| Nina | Nina | ordering's queue | N, kept; Brazil uses the name as it is |
| Oleg | Otávio | ordering's queue | O, kept |
| Pavel | Paulo | ordering's queue | P, kept, and the badge “P” |
| Ada | Alice | ordering's runners | A, kept |
| Bob | Bento | ordering's runners | B, kept; not Bruno, who is on the same page |
| Cid | Caio | ordering's runners | C, kept |
| Dot | Diana | ordering's runners | D, kept |
| Eli | Enzo | ordering's runners | E, kept |
| Fay | Flora | ordering's runners | F, kept, and the badge “F” |
| Anya | André | "Techniques": the pets and the knights | A, kept. A boy, since a named knight is “cavaleiro” |
| Borya | Bia | "Techniques": the pets and the knights | B, kept. A girl, so that the pair stays a boy and a girl: “Bia é mentirosa” |
| Vera | Vera | "Techniques", calendar | name in drawings; a Portuguese name as it is |
| Tanya | Sofia | "Techniques", working backwards | |
| Petya | Pedro | "Techniques", working backwards | |
| Lea | Lívia | enumeration | L, kept |
| Tom | Tomás | enumeration, clocks, parts | T, kept |
| Masha | Maria | gaps, arithmetic | M, kept |
| Sasha | Alex | pigeonhole, the sweets | short for Alexandre and Alexandra alike; the English gives no gender, and neither does the Portuguese |
| Gleb | Gabriel | pigeonhole, the socks | |
| Kim | Kauã | winning strategy | name in drawings: “Kauã −3” |
| Leo | Léo | winning strategy | name in drawings: “Léo −4” |
| Grandma | a avó | clocks | in drawings: “avó”, “relógio da avó” |
| Granny | a vovó | parts, the eggs at the market | |
| Mum, Dad | a mãe, o pai; “mamãe”, “papai” where the family speaks | calendar, ordering's tip, "About" | in drawings, lower case: “mãe”, “pai”; on "About" the names “Mamãe”, “Papai” |
| islanders A, B, C | A, B, C | knights | letters |
| knight K, liar L | V, M | knights, `thumb.*` | V for “verdade” (he always tells the truth), M for “mentiroso”; C, the card's “Cavaleiros”, stays an islander's; so “Escreva V e M a lápis” and “V, M, V, M…” |
| children A, B, C · boxes A, B | the same | enumeration · ratios | letters |
| the cook (she) · the baker (he) | uma cozinheira · um padeiro | weighing · parts | |
| woodcutters · a farmer | lenhadores · um agricultor | gaps · percentages | |

The people of history keep their Brazilian names: Euclides, o rei Ptolomeu, Proclo, os “Elementos”, Arquimedes, Rafael, a “Escola de Atenas”, Confúcio, as “seis artes”, Aryabhata, Taqi al-Din, Istambul, Luca Pacioli, Jacopo de' Barbari, al-Khwarizmi, Bagdá, “Os nove capítulos sobre a arte matemática”, Aristóteles, “Alexandre, o Grande”, Nikolai Bogdanov-Belsky, “Cálculo mental. Na escola pública de S. A. Ratchinski”, St. Louis, Tóquio, Leningrado, Moscou, Hungria, EUA, URSS, George Pólya (the Brazilian edition prints “G. Polya”; the site keeps his accent). A photograph's credit, “Yu. D. Korolev” or “Momotarou2012”, stays as its source writes it.

## Grades

`topics.grades` is **“Anos {range}”**: “Anos 1–2”, “Anos 3–4”, “Anos 5–6”, “Anos 1–6”; so `topic.example` reads “Exemplo 2 · Anos 3–4”, and the label on "Techniques" “Problema · Anos 2–3”.

- It reads right with every range, and names the card's own noun, “ano” (`profile.grade` “Ano escolar”, `child.grade` “{grade}º ano”).
- A bare range wants no ordinal indicator: “{range}º ano” would mark only the last number, “3–4º ano”, which careful Portuguese does not write; it writes “3º e 4º anos”, as the Canguru de Matemática Brasil names its levels, or “3º–4º ano”, as the card's `progress.grades` does with its two slots.
- The noun comes first, so the chip reads as school years: an age always follows its number, as the page "Coach" says its "aged 7 to 12", “de 7 a 12 anos”.

The alternative is “{range}º ano”, closest to the card's “3º ano” pill, at the price of an indicator on the last number alone. A text with two slots writes the card's form: “{first}º–{last}º ano” on "Research" (`hero.grades`, `sources.level`), “do {from}º ao {to}º ano” inside a sentence; one grade alone, “{last}º ano”: "Up to grade {last}" is “Até o {last}º ano”. The home page's lead, which has the one slot, says “para os anos {range} do ensino fundamental”, and free text writes the years whole: “Alemanha · 3º e 4º anos”, “estudantes do 9º ano”.

## Books on "Research"

| Key | The English page | Brazilian Portuguese | How sure |
|---|---|---|---|
| perelman.firsthundred | For Young Mathematicians. The First Hundred Puzzles | the English title | no Portuguese edition found |
| perelman.secondhundred | The Second Hundred Puzzles | the English title | no Portuguese edition found |
| perelman.living | Living Mathematics | “Aprenda matemática brincando” (Hemus, São Paulo, 1970, translated by Marina de Camargo; still in print, ISBN 978-85-289-0632-5) | sure it is this book: its subtitle, “histórias e quebra-cabeças”, matches the Russian one's, “Математические рассказы и головоломки”, and its first and last chapters, “Quebra-cabeças para a hora do almoço” and “Trinta problemas diferentes”, are the Russian book's first and last; the year and the translator come from a library's record, not checked on a copy. Mir also printed it in Portuguese, as “Matemática recreativa: relatos e quebra-cabeças de matemática” (Moscou, 1989, translated by Hudson C. Lacerda) |
| perelman.arithmetic | Entertaining Arithmetic | the English title | no Portuguese edition found; “Aritmética recreativa” is the Spanish |
| perelman.quick | Quick Reckoning | the English title | no Portuguese edition found |
| ignatyev.wit | In the Realm of Ingenuity, or Arithmetic for Everyone | the English title | no Portuguese edition found; Mir printed it in Spanish only, as far as found |
| dudeney.amusements | Amusements in Mathematics | “Divertimentos matemáticos” | unsure, as the Spanish is: no Brazilian edition, and RBA's collection “Desafios Matemáticos”, in European Portuguese and sold in Brazil too (c. 2008), split the book into “Divertimentos matemáticos” I–III: “O enigma do mandarim”, “Os gatos do feiticeiro”, “O mistério do cais”. The brief takes an edition in the language, European or Brazilian |
| dudeney.canterbury | The Canterbury Puzzles | “Os enigmas de Canterbury” | fairly sure: RBA's same collection, in European Portuguese; no Brazilian edition |
| loyd.cyclopedia | Cyclopedia of Puzzles | the English title | no edition found; Martin Gardner's selections, “Os enigmas de Sam Loyd” in the RBA collection, are other books |
| carroll.tangled | A Tangled Tale | “Uma história embrulhada” (Papirus, Campinas, 1992) | sure: a Brazilian edition, ISBN 85-308-0217-9; its translator, Luiz Arthur Pagani, from a bookseller's listing |
| carroll.pillow | Pillow Problems | the English title | no edition of its own: RBA's European volume “Um conto enredado e outros problemas de almofada” holds it under the title of the Tangled Tale, and “almofada” is a cushion in Brazil |

The authors' names: Yakov Perelman, as the Revista Brasileira de Ensino de Física writes him (Hemus's books print “J. Perelman”); Iemelian Ignátiev, in the Portuguese transliteration of Russian (“Ie” for Е, as in “Ievguêni”, and the stress marked); Henry E. Dudeney, Sam Loyd and Lewis Carroll as they are.

## Places for the author's reading

- `topics.grades`, “Anos {range}”, against “{range}º ano”.
- The menu: “Estudo” for "Research" and “Sobre” for "About", in the menu, the pages' titles and the footer alike, by an estimate of the row's width, not a measurement; “Pesquisa” and “Sobre nós” are the fuller words, “Pesquisa” the first to bring back if `just site-fold` finds room under the Russian.
- The two prices on "Why", the bait app's and MathTrail's own, with Intl's no-break space, the first U+00A0s in the site's texts: the Spanish “9,99 €” and the Russian “990 ₽” have a plain space.
- `topic.child`, “Criança”: Portuguese's own noun is the same for boys and girls, so no slash form, and it fits the 64-pixel column.
- An option of a task is an “alternativa”, the Brazilian test word, though the card lists the options under “Respostas”.
- "Cell" is “quadradinho” on squared paper, in grids and tables (and so on the pigeonhole page, where “casa” is the pigeonhole), and “casa” in strips and on boards, while the card's `picture.grid` says “Uma grade de casas”.
- The article before product names, “o MathTrail”, “o Claude”: the card writes it in some texts and not in others (`first_run.lead` “MathTrail traz…”, `data.remove_how` “Desconecte MathTrail”).
- The four-digit numbers of the site's maths keep no separator, “10, 100 ou 1000”, “1900”, “de 1200”, as the Russian and the Spanish write them, though Brazilian prose would write “1.000”; the bait's coins, which are no maths, take Intl's point, “12.450”, as the Spanish does.
- On "Techniques", Anya and Borya become André and Bia, a boy and a girl swapped, because “cavaleiro” has no current feminine.
- Kate's letter, K → C.
- The bait's cheer is “Você é fera!!!”, the Brazilian praise that fits a boy and a girl alike, where “campeão” would make the child a boy; the child's own “Done!” on "Why" is “Consegui!”.
- "Look back" is “Fazer um retrospecto”, after the edition's “Retrospecto”; the coach's simpler line ends with “conferir”.
- Claude's labels come from its Portuguese help; its warning over a prompt and its question about saving are not in that help, so the text describes them.
- OpenAI's path to developer mode comes from search results quoting its Brazilian page, which refuses fetchers.
- Google's consent screen is “tela de permissão”, the name Google's Portuguese docs give it, where the Spanish took Google's “pantalla de consentimiento”; “tela de consentimento” is the common alternative. The consent screen's wording of the Drive permission comes from Google's Marketplace, not from the screen itself. The card's `data.access_how` names Google's page by its older name, “apps de terceiros com acesso”; Google's help now says “apps vinculados”, which the site follows. The fix, if wanted, belongs to the card's dictionary.
- The books: only “Living Mathematics” and “A Tangled Tale” have Brazilian editions. Dudeney's two books take the titles of RBA's European Portuguese collection, as the brief takes an edition in the language; “Pillow Problems” keeps its English title, since its one Portuguese edition is a volume named after the Tangled Tale.
- “Iemelian Ignátiev”, a transliteration no Brazilian edition fixes; the English "Yemelyan Ignatyev" is the alternative.
- The card names both a group and a topic “Partes de um todo” (`group.parts`, `topic.fractions.parts`), so the topics page will show the topic under a group of its own name; and its `group.counting` is “Contagem”, narrower than "Counting and enumeration". Fixes belong to the card's dictionary, and the site follows it.
- CLDR gives Portuguese the `one` form for 0 and 1 alike, so a count of zero, if the site ever shows one, reads “0 problema”.
