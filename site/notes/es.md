# Spanish (`es`)

The Spanish half of [the brief](README.md): the rules the Spanish texts follow on top of it, and the words they use. The card's Spanish dictionary, [`web/locales/es.json`](../../web/locales/es.json), comes first wherever it has a word; this file adds only what the site needs beyond it.

## Register and variant

- **One Spanish, the card's.** Neutral, leaning to Spain where the variants split, as the card does: «Añadir», «problema», «Rellene este campo». There is no separate Latin American text. A word every variant shares wins over a local one (escuela, teléfono, dispositivo, tareas escolares); where none exists, the Spain form, as the card's «Añadir».
- **The reader is «usted»**, as the card speaks to the parent: «Pida un problema en el chat» (`waiting.ask_in_chat`), «El chat le dirá qué cambiar» (`profile.not_saved_detail`). Never «vosotros»; a plural reader is «ustedes».
- **The child is «tú»**, as the card speaks to the child: «Elige una respuesta» (`task.pick`), «Pide un deseo para mañana» (`waiting.limit`). That covers the quotes under «Qué decirle», the questions of a tip, Pólya's questions and the child's own lines.
- **What a parent types to the chat is «tú»** too, as people write to a chat: «Dame un problema de caballeros y mentirosos», «Un problema nuevo, por favor».
- **No gender shows, the child's or the parent's.** Spanish verbs never show it; adjectives, participles and some nouns do. So:
  - "your child" is «su hijo o hija» where a noun is needed, «el niño o la niña» where the reader may be a tutor; after that the verb carries the subject, which Spanish drops, or «quien resuelve».
  - No adjective or participle agrees with the child or the parent: «para asegurarse», not «para estar seguro»; «si se atasca», not «si está atascado»; «a solas», not «solo»; «¡Ya está!», not «¡Listo!»; «Otra vez me he equivocado», not «Estoy equivocado». A participle after «haber» never agrees, so the perfect is safe.
  - A child's question to the parent is a verb: «¿Mientes?», not «¿Eres mentiroso?»; in the island game the parent «miente siempre».
  - Children in general are «niños y niñas», as the card's `first_run.lead`; counted on "Research", «estudiantes». A group inside a task is the task's own, in the generic masculine Spanish textbooks write: «24 niños», «11 alumnos»; the rule is about the reader's child.
  - "The parent" in the third person is the card's «madre o padre», «madres y padres», and «para padres» on a label. The one who signs in, on the technical pages, is «la persona adulta» (card `first_run.lead`); ownership reads best as «de la familia»: «el Drive de la familia».
  - The people of the examples have a gender, as in the English: «Lucía no es la última». The islanders are letters, in the generic masculine: «A es mentiroso».
- **An example's question is a task for the child**, as a task on the card is, so it speaks to the child with «tú»: «Calcula…», «¿Cuántos…?». The words around it, the steps and the notes, speak to the reader.
- **What a topic teaches**, `hero.teaches`, goes on from its label «Qué enseña», so it opens «A» and an infinitive, and a second verb after «y» takes its «a» too: «A contar días y años…», «A razonar paso a paso y a comprobar…».
- **Sentences every topic's page shares** read the same on every page: `method.lead` ends «Los ejemplos de abajo van del más sencillo a uno de olimpiada.», and `description` ends «…dónde suelen equivocarse los niños y las niñas y cómo ayudar en casa.».
- **The card's words, exactly.** A topic's name written out in a sentence is the card's, either in «» with its capital, as the card writes «{topic}», or in lower case without quotes: «la enumeración de casos», «el principio del palomar». Never a paraphrase, though the English sometimes paraphrases ("putting things in order").

## Typography

- **Quotes:** «…»; a quote inside a quote “…”, and inside that ‘…’. The English “…” become «…».
- **Questions and exclamations** open and close: ¿…?, ¡…!, and the bait app's «¡¡¡…!!!». No space before ? ! : ; or after ¿ ¡. The ellipsis is one character, «…».
- **Dashes:** the English spaced dash becomes a colon, a comma or two sentences. Where a dash stays it is closed, the Spanish way: «texto —inciso— texto». Ranges take the en dash with no spaces: «1–6», «70–85 %», «siglos IX–XIX».
- **«y» and «o»** become «e» before an i- or hi- sound and «u» before an o- or ho- sound: «Pulse Añadir, después Conectar, e inicie sesión»; «siete u ocho».
- **Ordinals:** 1.º, 2.ª, 1.er, 3.er, the point before the indicator, as the card's «{grade}.º de primaria». In an example — its question, steps, note and answer — an ordinal the English writes in digits stays in digits: «el 9.º caramelo», «la 5.ª casilla». Elsewhere, words up to ten («la séptima vez»), then «número»: «la bandera número 23». In a drawing, digits with the indicator in the noun's gender: «21.ª, 22.ª, 23.ª» for flags; the badge "2nd" is «2.º». An example's answer keeps in digits every number its solver proves, so an ordinal there stays a digit: «la 5.ª casilla».
- **Dates** use cardinals: «del 3 al 10 de mayo», «el día 25», «los días 2, 9, 16, 23 y 30». Months and weekdays are lower case: «lunes», «mayo». A weekday in a drawing or a badge is short, as Intl writes it, capitalised where it opens the badge: «Lun», «Mar», «Mié».
- **Centuries** in Roman numerals: «siglo V», «finales del siglo XIX»; decades in words: «los años cincuenta».
- **Times** as the English writes them: «4:25», «15:10», «0:00», the bait's «03:59». A length of time: «4 h 30 min» in a drawing, «4 horas y 30 minutos» in prose.
- **Numbers outside the examples** as `Intl.NumberFormat("es")` writes them: «12.450», but «1200», since Spanish groups only from five digits; «9,99»; «0,2», so the formula reads «P = 0,2 + 0,8·σ(θ + δ − β)»; «19 millones». Years are never grouped.
- **Numbers inside the examples** keep the English digits and no thousands separator: «2026 × 2026», «1111». A decimal takes the comma, as the Russian writes it: «0,5 + ½ = 1».
- **Percent** takes a space before the sign, as the RAE asks and as the card's own percentages come out of Intl: «70–85 %». Inside an example, and everywhere on the page of percentages, whose drawings write «20%» beside the words, it is written as the English writes it: «20%».
- **Money** appears only in the make-believe app on "Why": «9,99 €» and «0 €», the symbol after the number and a space.
- **Units** are symbols after a space, with no point: «12 m», «200 kg», «5 l», «40 min», «4 h», and «7 d», the calendar drawings' day. In prose, «metros», «litros».
- **Signs:** × for multiplication; − (U+2212) for minus, never a hyphen; · in the formula. Division as each English page writes it: ÷ on the home page, "Why", "Techniques" and "Research"; : on the topic pages that use it (gaps, calendar, percentages, ratios, parts, arithmetic). Fractions as the English: «3/4», «½».
- **Capitals:** sentence case in titles and headings: «Cómo ayudar en casa». Capitals keep their accents: «Álex», «Óscar». «primaria», «olimpiada» and the languages are lower case in prose («en español»); a named contest is capitalised: «la Olimpiada Internacional de Matemáticas».
- **The card questions on the home page and on "Why"** write no lone Latin capital inside a sentence that the English does not: the picture check reads one there as a label of the picture. A capital that opens a sentence, a quotation or follows a colon is taken for a word, and inside a sentence the one-letter words «a», «e», «o», «u», «y» are lower case and harmless.
- **A sentence that ends with a quote** takes its period after the closing guillemet, as the RAE asks, even when the quote ends in ? or !: «Pregunte: «¿Seis quiénes?».». A `say:` that is the quote alone takes a period after a statement, «Señálalo en la tira»., and nothing after a ? or !.
- **«Juntos»** would show a gender: «en familia», «con su hijo o hija», or a verb in the plural for the reader and the child together («Practiquen»).
- **YAML:** a text that holds `: ` or begins with `*` is quoted; «, ¿ and ¡ need no quotes.
- **Titles of works** — a book, a painting, a treatise — go in «» where the English writes them plain, as the Russian does: the emphasis of a text is the English one's, and the build reads `*…*` as emphasis.

## Platforms' labels

### Claude

Claude's interface speaks Spanish, so its labels are those of its Spanish help, not the English ones the Russian keeps. The help speaks «tú»; the site wraps the labels in «usted»: «Pulse **+ Añadir** y elija **Añadir conector personalizado**». The help's prose mixes «añadir» and «agregar»; its labels say «Añadir».

- [A] «Comienza con conectores personalizados usando MCP remoto», https://support.claude.com/es/articles/11175166-getting-started-with-custom-connectors-using-remote-mcp
- [B] «Usa conectores para ampliar las capacidades de Claude», https://support.claude.com/es/articles/11176164-use-connectors-to-extend-claude-s-capabilities

| English on the site | Spanish | Where the help says it |
|---|---|---|
| Customize → Connectors | Personalizar → Conectores | A: «Navega a Personalizar > Conectores». The site keeps its own arrow. |
| + | + Añadir | A: «Haz clic en "+ Añadir" y luego en "Añadir conector personalizado"» |
| Add custom connector | Añadir conector personalizado | A |
| Name it MathTrail | Escriba MathTrail como nombre | A: «Introduce un nombre para el conector»; the field's label is not quoted |
| the address | la URL del servidor MCP remoto; on the site, «la dirección» | A |
| Continue, and Continue again | Continuar, y otra vez Continuar | A quotes «Continuar» once, after the URL; the second press is the same button, as the English help has it |
| the sign-in settings Claude detects | la configuración de autenticación que Claude detecta | A: «Revisa la configuración de autenticación que Claude detectó» |
| Add | Añadir | A: «Termina de configurar tu conector haciendo clic en "Añadir"» |
| Connect | Conectar | B: «Haz clic en "Conectar" o "Instalar"» |
| + → Connectors, in a chat | + → Conectores | A and B: «el botón "+" en la parte inferior izquierda… luego "Conectores"» |
| the free plan | el plan Free | A: «Los usuarios de Free están limitados a un conector personalizado» |
| Always allow | Permitir siempre | A, its section on security |
| Sign in now · Sign in when needed · No sign in | Iniciar sesión ahora · Iniciar sesión cuando sea necesario · Sin iniciar sesión | A, under «Autenticación»; not on the site yet |
| Use Claude's published identity · Register automatically · Use your own OAuth client | Usar la identidad publicada de Claude (Recomendado) · Registrarse automáticamente · Usar tu propio cliente OAuth | A, under «Cliente OAuth»; not on the site yet |
| "Use caution before running this prompt" | not found in Claude's Spanish help | the text describes it: «bajo un aviso que pide precaución antes de enviarlo» |
| Claude asks whether to allow MathTrail to save | the question is not quoted in Claude's Spanish help; only the button «Permitir siempre» is | the text describes it: «Claude pregunta si permite que MathTrail guarde el perfil» |

### ChatGPT

| English on the site | Spanish | Source |
|---|---|---|
| developer mode | **el modo para desarrolladores** (es-ES) | https://help.openai.com/es-es/articles/12584461 — «Modo para desarrolladores y apps MCP en ChatGPT [beta]»; the path «Configuración → Apps → Configuración avanzada → Modo para desarrolladores», then «Apps → Crear» |
| (the alternative) | el modo desarrollador, el modo de desarrollador (es-419) | https://help.openai.com/es-419/articles/12584461 |
| ChatGPT's directory | el directorio de ChatGPT | OpenAI's Spanish help says the «directorio de aplicaciones» became the «Directorio de plugins» («de complementos») on 9 July 2026: https://help.openai.com/es-es/articles/11487775-apps-in-chatgpt. The generic phrase stays true. |
| Plus, Pro, Business, Enterprise, Education | the same | names of plans |

The es-ES name is the one to use: it matches the card's lean, and its article speaks «usted» as the site does («Vaya a…», «Actívelo»). help.openai.com refuses fetchers; these labels come from search results that quote those pages.

### Google

| English on the site | Spanish | Source |
|---|---|---|
| Google Drive | Google Drive; My Drive is «Mi unidad» | https://support.google.com/drive/?hl=es |
| Drive's bin | la papelera de Drive; the button «Vaciar papelera»; files go for good after 30 days | https://support.google.com/drive/answer/2375102?hl=es |
| version history | el historial de versiones; for a file such as the profile, the menu «Gestionar versiones» and «Conservar para siempre» | https://support.google.com/drive/answer/2409045?hl=es |
| linked apps, third-party access | la página de aplicaciones vinculadas de su cuenta de Google (myaccount.google.com/linkedapps); the section «Acceso a tu cuenta de Google»; «Ver detalles» → «Quitar acceso» → «Confirmar» | https://support.google.com/accounts/answer/13533235?hl=es, «Gestionar las vinculaciones entre tu cuenta de Google y aplicaciones de otros desarrolladores» |
| the box for Google Drive on the consent screen | la casilla de Google Drive; Google words the permission «Ver, editar, crear y eliminar solo los archivos de Google Drive que utilices en esta aplicación» | as Google lists it in Spanish in its Workspace Marketplace: https://workspace.google.com/marketplace/app/docs_sheets_editor_for_drive/703230099925?hl=es. Not seen on the live consent screen, and the English never quotes it. |
| Sign in with Google | in prose, «iniciar sesión con Google»; Google's own name for the button is «Acceder con Google» | https://developers.google.com/identity/siwg/features?hl=es |
| Google account | cuenta de Google | |

MathTrail's own sign-in pages speak English, so «Tick the box for Google Drive» and **Back to Google** stay English.

## Glossary

"card" marks the card's own word, from `web/locales/es.json`.

### Product and lesson

| English | Spanish | Note |
|---|---|---|
| task | problema | card, `task.label` |
| problem (olympiad) | problema | the same word as task |
| school sum | ejercicio de clase, cuenta | "A school sum" is «Un ejercicio de clase» |
| topic | tema | card |
| group of topics | grupo | card, `group.*` exactly |
| trap | trampa | card, «La trampa» |
| the catch (not a trap) | el truco, la dificultad | never «trampa» outside the card's sense |
| technique | técnica | |
| the main move (`topic.method`) | la idea clave | |
| hint | pista | card; never for "clue" |
| clue, a given of a task | dato | keeps «pista» for the card's hint: «Cada dato es una flecha» |
| option | opción | the card lists the options under «Respuestas» |
| answer | respuesta | card |
| right, wrong answer | respuesta correcta, incorrecta | card: «¡Correcto!», «No del todo» |
| solution | solución | card |
| step | paso | |
| question, the text of a task | enunciado | card: «Número del enunciado tomado como respuesta» |
| question, what is asked | pregunta | card: «Respuesta a otra pregunta» |
| picture of a task | dibujo | «imagen» for a photograph |
| label of a picture | etiqueta | |
| grade | grado | card `profile.grade`; the card's note also says «cursos» |
| Grade 3 | 3.º de primaria | card `child.grade` |
| grades 1 to 6 | de 1.º a 6.º de primaria | card `first_run.lead` |
| grades 3–4, two slots | {from}.º–{to}.º de primaria | card `progress.grades_ranks` |
| level | nivel | |
| difficulty | dificultad | card |
| rank | etapa | card, «Etapa {rank}»: a number, never a title |
| rating | puntuación | card: «Puntuación en este tema», «Puntuación general» |
| progress | progreso | card, «Perfil y progreso» |
| mastered | dominado | card |
| review, the card's section | análisis | card, «Análisis» |
| how the answer went | el repaso de la respuesta | card `result.label` |
| trial series | serie de prueba | card |
| skipped | en blanco | card |
| profile | perfil | card |
| pseudonym | seudónimo | card; not «pseudónimo» |
| interests | intereses | card |
| lesson | clase | the card also says «lecciones» in «Idioma de las lecciones», quoted as it is |
| coach | entrenador | card, `topic_choice.coach_short`; the Coach product too |
| companion, in the coach | compañero | a character, not the child |
| parent | madre o padre; madres y padres; «para padres» on a label | card |
| tutor | docente | card `first_run.lead` |
| adult | persona adulta | card |
| child, your child | el niño o la niña; su hijo o hija | see "No gender shows" |
| children | niños y niñas | card; «estudiantes» where "Research" counts them |
| chat | el chat | |
| card | tarjeta | card |
| connector, custom connector | conector, conector personalizado | Claude |
| app | aplicación, app | OpenAI writes «apps» |
| sign in | iniciar sesión | card |
| sign-in | inicio de sesión | |
| account | cuenta | |
| Drive file, the profile's file | archivo de Drive, archivo del perfil | card |
| folder | carpeta | |
| bin | papelera | card and Google |
| daily limit | límite diario | card |
| Another task | Otro problema | card's button, quoted exactly |
| Make a wish for tomorrow | Pide un deseo para mañana | card, quoted exactly |
| Profile & progress | Perfil y progreso | card |
| Edit | Editar | card |
| the bulb, the hint's button | la bombilla | |
| the two arrows | las dos flechas | |
| tap | tocar | |
| press, click | pulsar: «pulse» | |
| Enter, the key | Intro | as Google's Spanish help names it |
| free, of price | gratis | the chip `hero.free` |
| Free, the plan | Free | «el plan Free»: a plan's name |
| open source | código abierto | |
| in development | en desarrollo | |
| beta | Beta | |
| soon | próximamente | |
| olympiad maths | matemáticas de olimpiada, olimpiada matemática | card `first_run.lead` |
| olympiad-style task | problema de tipo olimpiada | |
| maths circle | círculo de matemáticas | «club de matemáticas» where clearer |
| school | escuela | card |
| school mark | nota | card |
| homework | tareas escolares | shared by every variant |

### The site's sections and headings

| English | Spanish | Note |
|---|---|---|
| Topics (`nav.topics`) | Temas | 5 characters |
| Techniques (`nav.techniques`) | Técnicas | 8 |
| Why (`nav.why`) | Por qué | 7 |
| Coach (`nav.coach`) | Entrenador | 10, beside the tag «Beta»; no shorter word the card would own |
| Research (`nav.research`) | Estudio | 7: «Investigación», 13, took the row past the Russian. The page's title is «Estudio» too, as a page's title is its menu entry in every language, and the footer shows it |
| Service (`nav.service`) | Servicio | 8 |
| About (`nav.about`) | Nosotros | 8, and the page's title too |
| Add to Claude (`nav.add`) | Añadir a Claude | 15 |
| Technical pages | Páginas técnicas | said by a screen reader |
| Language | Idioma | |
| Help · Privacy policy · Terms of use | Ayuda · Política de privacidad · Condiciones de uso | |
| How it is solved | Cómo se resuelve | |
| Where children go wrong | Dónde suelen equivocarse | |
| What to say | Qué decirle | to the child |
| How to help at home | Cómo ayudar en casa | |
| Learn first | Aprender primero | |
| Opens the way to | Abre el camino a | |
| Builds on · Opens | Se basa en · Abre paso a | the topics page |
| Ask for a task on this topic | Pida un problema de este tema | |
| Show · Hide the solution | Ver la solución · Ocultar la solución | card: «Ocultar pista» |
| Example 2 | Ejemplo 2 | |
| Why it matters | Por qué importa | |
| How to solve: … | Cómo resolverlo: … | |
| The rules of … | Las reglas de… | |
| Check: | Comprobación: | |
| When it helps: | Cuándo sirve: | |
| In the topics: | En los temas: | |
| You are here · On this page · All topics | Usted está aquí · En esta página · Todos los temas | |
| More → | Ver más → | |
| Honestly | Sinceramente | |
| See a lesson | Ver una clase | |
| For technical readers | Para público técnico | no gendered «lectores» |
| You, the parent's line on "Why" · C, the child's | Usted · N | «N es el niño o la niña» |
| The adult · The model, in the chat's illustration | Persona adulta · Modelo | |
| Step 1, of the topics' map | Paso 1 | |
| Foundations · Rules · Hard problems | Fundamentos · Reglas · Problemas difíciles | |

With «Investigación» the menu stayed in a row only down to 1197 px in Chromium and 1184 px in WebKit, wider than the Russian's 1163 and 1149, so `nav.research` is «Estudio»: with it the row holds down to 1157 and 1143 px, and the menu still folds below 1230 px. «Nosotros» → «Equipo» is the next cut, should a language ever need it.

### Maths and olympiad terms

| English | Spanish | Note |
|---|---|---|
| olympiad | olimpiada | «olímpiada» is correct too, and usual in parts of Latin America |
| reference task | problema de referencia | |
| solver | solucionador | first time, «un programa que resuelve el problema» |
| brute force | fuerza bruta | |
| worst case | peor caso | card, `trap.best_case_not_worst` |
| good, bad luck | buena, mala suerte | card: «el caso de peor suerte» |
| to be sure, for certain | para asegurarse, con seguridad | never «para estar seguro» |
| pigeonhole principle | principio del palomar | card |
| pigeons, holes | palomas, nidos | "Pigeons and pigeonholes" is «Palomas y nidos» |
| parity · even · odd | paridad · par · impar | card |
| alternation, taking turns | alternancia, alternarse | card |
| remainder · quotient | resto · cociente | card |
| divisibility | divisibilidad | card |
| divides by 9 | es divisible entre 9 | «dividir entre», throughout |
| divisibility rules | criterios de divisibilidad | card |
| digit · two-digit number | cifra · número de dos cifras | |
| tens, units digit | cifra de las decenas, de las unidades | |
| enumeration | enumeración de casos | card |
| case · list | caso · lista | |
| systematic listing | enumerar con orden | the technique "List systematically" |
| tree, the diagram | diagrama de árbol | |
| working backwards | empezar por el final, ir hacia atrás | known in Latin America as «el método del cangrejo» |
| the assumption method | el método de la falsa suposición | its usual name in Spanish school maths |
| proof by contradiction | reducción al absurdo | |
| suppose · assumption · contradiction | supongamos que… · suposición · contradicción | |
| pattern | patrón | |
| symmetric strategy | estrategia simétrica | |
| winning strategy | estrategia ganadora | card |
| winning, losing position | posición ganadora, perdedora | |
| move, in a game | jugada | card |
| the first, the second player | el primer jugador, el segundo | |
| invariant | invariante | |
| knights and liars | caballeros y mentirosos | card |
| knight · liar | caballero · mentiroso | V and M in drawings; «caballero» has no feminine for this sense, so a named knight is a boy |
| island · islander | isla · habitante de la isla | |
| statement | afirmación | card |
| ordering | ordenación | card |
| number line | recta numérica | card |
| grid · squared paper | cuadrícula · papel cuadriculado | card |
| cell · side of a cell | casilla · lado | |
| area · perimeter | área · perímetro | card |
| ratio · in the ratio 3 to 4 | razón · en la razón 3 a 4 | card |
| sharing · share | reparto · parte | card |
| percentage · percent | porcentaje · por ciento | card |
| fraction · the whole · the part | fracción · el todo · la parte | card |
| a half, a third, a quarter | la mitad, un tercio, un cuarto | |
| gaps and boundaries | intervalos y extremos | card |
| gap, between posts · post · fence | intervalo · poste · valla | the card's topic and trap say «intervalos»; «espacio» is the card's word in passing |
| off by one | uno de más o de menos | card |
| clock face · hands · the whole hour | esfera · agujas · la hora en punto | card: «Una esfera de reloj» |
| floor 1, of a house | el piso 1 | never «planta baja», which would change the count |
| calendar and age | calendario y edades | card |
| pan balance · weighing · fake coin | balanza de platillos · pesada · moneda falsa | card |
| jug · pour | jarra · trasvasar | card |
| state, of the jugs | estado | |
| overlapping groups | grupos que se superponen | card |
| arithmetic with a trick | aritmética con truco | card |
| I thought of a number | he pensado un número | |
| carrying | llevarse: «sin llevadas» | |
| round number | número redondo | |
| match, the stick | cerilla | the Spain form, as the card leans |
| a special case | un caso particular | |

### The technical pages

| English | Spanish | Note |
|---|---|---|
| token · access token · refresh token | token · token de acceso · token de actualización | as Google's Spanish docs |
| authorization code | código de autorización | |
| seal · sealed · sealing | sellar · sellado · sellado | |
| encrypted | cifrado | |
| state · stateless | estado · sin estado | |
| MCP server · tool · tool call | servidor MCP · herramienta · llamada a una herramienta | |
| chat host | host del chat | |
| consent screen | pantalla de consentimiento | Google |
| scope | permiso | Google: «el permiso drive.file» |
| client · client metadata document | cliente · documento de metadatos del cliente | |
| renewal · revocation | renovación · revocación | |
| pace · ceiling | ritmo · tope | |
| Sign-in, a scenario's tab on "Service" · target, of the profile's size | Acceso · meta | short: the tab and the bar's label do not wrap at 320 px |
| log · logging | registro | «Cloud Logging» stays |
| instance · revision · image | instancia · revisión · imagen | |
| key rotation · ciphertext · nonce | rotación de claves · texto cifrado · nonce | |
| fingerprint · near-duplicate | huella · casi duplicado | |
| readability · self-check | legibilidad · autocomprobación | |
| brief, for the model | encargo | |
| package · schema · template | paquete · esquema · plantilla | |
| deterministic · the rule | determinista · la regla | |
| cookie · cache | cookie · caché | |
| graceful shutdown | apagado ordenado | |
| lifetime, of a token | vigencia | |
| Workload Identity Federation | Federación de identidades para cargas de trabajo | as Google's Spanish docs |
| student model | modelo del estudiante | |
| chance, of a right answer | probabilidad de acertar | |
| corridor · interval · goal | corredor · intervalo · objetivo | |
| the ceiling, an estimate that knows the child · oracle | el techo · oráculo | |
| bench · simulated children | banco de pruebas · estudiantes simulados | |
| a child who learns, who stays put | quien aprende, quien no avanza | |
| lag · logit · points | retraso · logit · puntos | |
| mastery declared falsely | dominio declarado por error | |
| live data, live numbers | datos reales, cifras reales | they come monthly, not in real time |
| range, of the chance promised | tramo | |
| standard error · baseline · seed | error estándar · referencia · semilla | |
| Reached · On the edge · Not reached | Alcanzado · En el límite · No alcanzado | |
| paper · public domain | artículo · dominio público | |
| commit · release | commit · versión | «commit» is the borrowed term |
| endpoint · claim, of a JWT | endpoint · claim | borrowed terms |
| request, the model's · store · router · issuer | petición (card) · almacén · enrutador · emisor | |
| deploy · build · client secret · sandboxed | despliegue · compilación · secreto de cliente · aislado | |
| in the open, unsealed · timeout | en claro · tiempo de espera | |
| digest · runtime · PKCE verifier · state machine | resumen · entorno de ejecución · verificador · máquina de estados | |
| scenario · sketch · pinned · sliding, of a lifetime | escenario · esbozo · fijado · deslizante | |
| canvas · tablet | lienzo · tableta | the coach |
| GitHub issue · computer · mouse | incidencia · ordenador · ratón | the Spain forms, as the card leans |
| the profile's notes | anotaciones | «nota» is a school mark |
| whisker, of an error bar · dashed | bigote · discontinua | |
| item response theory · psychometrics | teoría de respuesta al ítem (TRI) · psicometría | |

## Pólya's four steps

The Spanish edition is «Cómo plantear y resolver problemas» (Trillas, México, 1965, translated by Julián Zugazagoitia). Its summary (pages 18–19) names the phases «I. Comprender el problema», «II. Concebir un plan», «III. Ejecución del plan», «IV. Examinar la solución obtenida»; the list of questions under them calls the last «Visión retrospectiva». The page "Techniques" names its steps after them:

| English | Spanish |
|---|---|
| *How to Solve It* (1945) | *Cómo plantear y resolver problemas* (1945) |
| Understand the problem | Comprender el problema |
| Make a plan | Concebir un plan |
| Carry it out | Ejecutar el plan |
| Look back | Examinar la solución |

A technique's name is an infinitive, as Pólya's steps are: «Dibujar el problema», «Enumerar con orden», «Empezar por el final».

The steps' questions speak to the child: «Dilo con tus propias palabras. ¿Qué sabes? ¿Qué tienes que encontrar?» The coach's shorter line, "Understand → make a plan → solve → check", is «Comprender → hacer un plan → resolver → comprobar», and the picture's "understand, plan, solve, check" on "Why" is «comprender, planear, resolver, comprobar».

## People's names

"Initial" marks a name whose first letter a drawing or a badge shows; within one drawing no two letters repeat.

| English | Spanish | Where | Initial, badge |
|---|---|---|---|
| Galt | Galt | home, the cards' pseudonym: «Elegido para Galt» | kept, as the Russian keeps it; an invented name shows no gender |
| Comet | Cometa | "Why" and "Topics", the cards' pseudonym | as the Russian «Комета»; a name, so never «el» or «la» |
| C, "your child" on "Why" | N | `parent.explain.child`: «N es el niño o la niña» | the speaker's letter |
| Ann | Ana | ordering, parts, ratios, clocks, calendar | A, kept |
| Ben | Bruno | ordering, parts, ratios, parity | B, kept |
| Kate | Clara | ordering, calendar | the letter changes, K → C: `art.kate` is «C» |
| Dan | Diego | ordering's line | name in a drawing |
| Lena | Lucía | ordering's queue, gaps, ratios | L, kept |
| Max | Mateo | ordering's queue, ratios | M, kept |
| Nina | Nora | ordering's queue | N, kept |
| Oleg | Omar | ordering's queue | O, kept |
| Pavel | Pablo | ordering's queue | P, kept, and the badge «P» |
| Ada | Alba | ordering's runners | A, kept |
| Bob | Blas | ordering's runners | B, kept; not Bruno, who is on the same page |
| Cid | César | ordering's runners | C, kept |
| Dot | Diana | ordering's runners | D, kept |
| Eli | Enzo | ordering's runners | E, kept |
| Fay | Fátima | ordering's runners | F, kept, and the badge «F» |
| Anya | Adrián | "Techniques": the pets and the knights | A, kept. A boy, since a named knight is «caballero» |
| Borya | Bea | "Techniques": the pets and the knights | B, kept. A girl, so that the pair stays a boy and a girl: «Bea es mentirosa» |
| Vera | Vera | "Techniques", calendar | name in drawings; Vera is common in Spain now |
| Tanya | Sofía | "Techniques", working backwards | |
| Petya | Pedro | "Techniques", working backwards | |
| Lea | Laura | enumeration | L, kept |
| Tom | Tomás | enumeration, clocks, parts | T, kept |
| Masha | María | gaps, arithmetic | M, kept |
| Sasha | Álex | pigeonhole, the sweets | used for boys and girls alike, as Sasha is; the English gives no gender, and neither does the Spanish |
| Gleb | Hugo | pigeonhole, the socks | |
| Kim | Sara | winning strategy | name in drawings: «Sara −3» |
| Leo | Leo | winning strategy | name in drawings: «Leo −4» |
| Grandma, Granny | la abuela | clocks, parts | in drawings: «la abuela», «el reloj de la abuela» |
| Mum, Dad | mamá, papá | calendar, ordering's tip, "About" | in drawings, lower case |
| islanders A, B, C | A, B, C | knights | letters |
| knight K, liar L | V, M | knights, `thumb.*` | V for «veraz», M for «mentiroso»; C stays an islander's |
| children A, B, C · boxes A, B | the same | enumeration · ratios | letters |
| the cook (she) · the baker (he) | una cocinera · un panadero | weighing · parts | |
| woodcutters · a farmer | unos leñadores · un agricultor | gaps · percentages | |

The people of history keep their Spanish names: Euclides, el rey Ptolomeo, Proclo, los «Elementos», Arquímedes, Rafael, «La escuela de Atenas», Confucio, las «seis artes», Aryabhata, Taqi al-Din, Estambul, Luca Pacioli, Jacopo de' Barbari, al-Juarismi, Bagdad, «Los nueve capítulos sobre el arte matemático», Aristóteles, Alejandro Magno, Nikolái Bogdánov-Belski, S. A. Rachinski, San Luis, Tokio, Leningrado, Moscú, EE. UU., URSS, George Pólya. A photograph's credit, «Yu. D. Korolev» or «Momotarou2012», stays as its source writes it.

## Grades

`topics.grades` is **«{range} de primaria»**: «1–2 de primaria», «3–4 de primaria», «5–6 de primaria», «1–6 de primaria»; so `topic.example` reads «Ejemplo 2 · 3–4 de primaria», and the label on "Techniques" «Problema · 2–3 de primaria».

- It keeps the card's own phrase, «… de primaria», so a chip on the site and the card's «3.º de primaria» read alike.
- «primaria» names the school stage in every variant, where the noun splits: «curso» in Spain, «grado» in most of Latin America; the card itself uses both.
- A bare range wants no ordinal indicator: «{range}.º» would mark only the last number, «3–4.º», which Spanish does not write.

The alternative is «Primaria {range}», the way textbooks label a year («Matemáticas 3 Primaria»). «Cursos {range}» and «Grados {range}» each speak one variant only. A text with two slots writes the card's form: «{from}.º–{to}.º de primaria»; the home page's lead, which has the one slot, says «para {range} de primaria».

## Books on "Research"

| Key | The English page | Spanish | How sure |
|---|---|---|---|
| perelman.firsthundred | For Young Mathematicians. The First Hundred Puzzles | the English title | no Spanish edition found |
| perelman.secondhundred | The Second Hundred Puzzles | the English title | no Spanish edition found |
| perelman.living | Living Mathematics | «Matemáticas recreativas» (Mir, Moscú; Martínez Roca, Barcelona) | sure it is this book: its first chapter, «Desayuno y rompecabezas», opens with the squirrel of «Живая математика»; the years are not checked |
| perelman.arithmetic | Entertaining Arithmetic | «Aritmética recreativa» | unsure: the Spanish text found is Patricio Barros's translation, published online (librosmaravillosos.com) and later printed on demand, not a Mir edition |
| perelman.quick | Quick Reckoning | the English title | no Spanish edition found |
| ignatyev.wit | In the Realm of Ingenuity, or Arithmetic for Everyone | «En el reino del ingenio» (Mir, Moscú, 1986; Mir and Rubiños-1860, Madrid, 1995) | sure; the subtitle «o Aritmética para todos» is not on the Spanish edition, as far as found |
| dudeney.amusements | Amusements in Mathematics | «Diversiones matemáticas» | unsure: the book came out in Spanish in parts, «El acertijo del mandarín y otras diversiones matemáticas» (translated by Carlos Gardini) and, from RBA in 2008, «Diversiones matemáticas» I–III, with «Los gatos del hechicero» and «El misterio del muelle»; no edition carries the whole book under one title |
| dudeney.canterbury | The Canterbury Puzzles | «Los acertijos de Canterbury» (RBA, 2007) | sure |
| loyd.cyclopedia | Cyclopedia of Puzzles | the English title | no Spanish edition found; Martin Gardner's selections are other books |
| carroll.tangled | A Tangled Tale | «Un cuento enmarañado» (Nivola, 2002) | sure |
| carroll.pillow | Pillow Problems | «Problemas de almohada» (Nivola, 2005) | sure |

The authors' names: Yákov Perelmán and Yemelián Ignátiev, as Spanish writes Russian names; Henry E. Dudeney, Sam Loyd and Lewis Carroll as they are.

## Places for the author's reading

- `topics.grades`, «{range} de primaria», against «Primaria {range}».
- The menu: «Estudio» for «Investigación», in the menu, the page's title and the footer alike, keeps the Spanish row narrower than the Russian, so the fold stays at 1230 px. The other way out is «Investigación» everywhere and a fold some 30 px wider for every language.
- `topic.child`, «Niño/a», a slash form the RAE accepts only where room is short, as in this 64-pixel label. The alternatives are «Intento», which names the child's try rather than the child, and «Peque», which is warm but colloquial.
- `research.children`, «estudiantes», so that no count shows a gender; «niños» reads more naturally but a single one would be a boy.
- On "Techniques", Anya and Borya become Adrián and Bea, a boy and a girl swapped, because «caballero» has no feminine in this sense.
- Kate's bead letter, K → C.
- «clue» → «dato», to keep «pista» for the card's hint.
- Percent: with a space across the site, «70–85 %», but on the page of percentages, whose drawings write «20%» beside its words, as the English writes it.
- Claude's labels come from its Spanish help, which mixes «añadir» and «agregar»; a look at Claude's interface in Spanish would settle them. Its warning over a prompt and its question about saving are not in that help, so the text describes them.
- The consent screen's wording of the Drive permission comes from Google's Marketplace, not from the screen itself.
- The books marked unsure: «Aritmética recreativa», «Amusements in Mathematics», the years of «Matemáticas recreativas», and the missing subtitle of «En el reino del ingenio».
- The make-believe app's cheer is «¡¡¡Eres una estrella!!!», since «campeón» would make the child a boy.
- The card names both a group and a topic «Partes de un todo» (`group.parts`, `topic.fractions.parts`), so the topics page will show the topic under a group of its own name. The fix belongs to the card's dictionary, and the site follows it.
