---
title: Política de privacidad
description: Qué sabe MathTrail de su hijo o hija, dónde se guarda y cómo recuperarlo o eliminarlo.
---

# Política de privacidad

Última actualización: 26 de octubre de 2026.

MathTrail es un proyecto gratuito y de código abierto, creado y gestionado por ALT EdTech. Las preguntas sobre esta política o sobre los datos de su hijo o hija se envían a [altedtech.info@gmail.com](mailto:altedtech.info@gmail.com). Cómo conectarlo, y qué hacer si algo parece ir mal, se explica en la [página de ayuda](../help/).

## En resumen

- **No guardamos nada.** El servicio no tiene ninguna base de datos ni ningún archivo. Entre una petición y la siguiente no recuerda nada.
- **El perfil de su hijo o hija está en su propio Google Drive**, en un archivo que puede abrir, copiar o eliminar en cualquier momento.
- **Un seudónimo, nunca un nombre real.** Ni fecha de nacimiento, ni escuela, ni fotografía.
- **Contamos a los niños y las niñas sin saber quiénes son.** Una línea del registro designa a un niño o una niña con un código que cambia cada mes y que no remite a nadie; lo que se conserva más de dos meses son totales.
- **No se vende nada, y nada sirve para entrenar un modelo.**
- **Eliminar el archivo elimina el perfil.** No hay una segunda copia en ninguna parte.

## Qué es MathTrail

MathTrail es una aplicación que usted añade a su propio Claude o ChatGPT. Elige un tema y una dificultad para su hijo o hija, pide al propio modelo del chat que escriba un problema, comprueba ese problema con un programa, lo muestra, registra la respuesta y explica el error.

MathTrail no ejecuta ningún modelo de lenguaje propio. El modelo que escribe los problemas es el del chat que usted ya utiliza, y su relación con el proveedor de ese chat se rige por las condiciones y la política de privacidad de dicho proveedor, no por esta.

## Quién inicia sesión

Solo una persona adulta: madre, padre o docente. El niño o la niña no tiene cuenta ni inicio de sesión propios. Las plataformas de chat fijan sus propias normas de edad: Claude es para personas adultas, y ChatGPT requiere el consentimiento de la madre o el padre para adolescentes. MathTrail no está dirigido a niños y niñas que usen un chat por su cuenta.

## Niños y niñas

Los problemas de MathTrail son para niños y niñas de 1.º a 6.º de primaria, de seis a doce años aproximadamente. MathTrail está pensado para que la persona adulta escriba en el chat mientras el niño o la niña, a su lado, responde tocando una opción en la tarjeta. El niño o la niña no tiene cuenta. El perfil contiene lo que la persona adulta indica sobre el niño o la niña, bajo un seudónimo, junto con las respuestas del niño o la niña, en un archivo de la propia persona adulta; qué conservan de esas respuestas los registros del servicio, y durante cuánto tiempo, se expone más abajo.

## Qué se guarda y dónde

### En su Google Drive

El perfil es un único archivo JSON, `mathtrail-profile.json`, en una carpeta llamada `MathTrail` de su propio Drive. Puede abrirlo, leerlo, copiarlo, moverlo o eliminarlo como cualquier otro archivo, y la papelera y el historial de versiones de Drive son suyos.

Contiene:

- un **seudónimo** para el niño o la niña, el grado, sus intereses y lo que usted haya pedido evitar;
- un identificador del niño o la niña, para que el perfil se pueda trasladar a una edición futura;
- el idioma de la interfaz, si eligió uno;
- el tema al que se ciñen las clases, si lo eligieron usted o su hijo o hija;
- el país en el que vive y, en Estados Unidos, su estado, si decidió indicarlos: solo están ahí para contar cuántas familias hay en cada país, y puede omitirlos;
- qué temas están dominados, las puntuaciones en las que se basa la dificultad y un resumen de las respuestas anteriores: por tema, cuándo se dio por última vez, cuántos intentos y aciertos hubo y qué trampas aparecieron;
- dónde estaban las puntuaciones antes de cada una de las últimas respuestas, y al comienzo de cada uno de los últimos siete días con respuestas, para que el progreso pueda mostrar qué ha cambiado desde el último problema y a lo largo de la semana;
- un contador de los problemas aceptados hoy, con el que se aplica el límite diario, y cuánto se adelanta o se atrasa respecto a UTC el reloj del dispositivo en el que responde su hijo o hija: la tarjeta lo envía con cada respuesta, para que el día termine a su medianoche;
- el problema actual, con su respuesta cifrada para que no se pueda leer en el archivo antes de que el niño o la niña responda;
- el siguiente problema, que el chat escribe por adelantado mientras el niño o la niña resuelve el actual, con su respuesta cifrada de la misma manera, hasta que se pida.

No contiene el nombre del niño o la niña, ni su fecha de nacimiento, escuela, dirección o fotografía, ni nada más que permita saber quién es. **El seudónimo es el único nombre que interviene, y nunca aparece dentro de un problema.**

### En el servidor

Nada. El servicio es un único programa que responde a una petición y la olvida. No tiene base de datos, ni disco, ni caché de sus datos.

Tampoco se guarda su inicio de sesión de Google: se cifra y se coloca dentro del token que el servicio entrega a su chat, junto con el país desde el que inició sesión, como se describe más abajo. La clave está en Google Secret Manager y nunca sale del servidor.

### En los registros

El servicio escribe recuentos, no contenido: si un problema se aceptó o se rechazó, por qué se rechazó, cuánto tardó, cuántos intentos necesitó, cuántos bytes ocupaba cada parte del problema que entregó el chat, si se alcanzó un límite y qué versión de las instrucciones se usó. Las líneas sobre las llamadas que requirió un mismo problema comparten un código aleatorio creado para ese problema, que las une y no dice nada del niño o la niña.

Para contar cuántos niños y niñas usan MathTrail cada día, cada semana y cada mes sin saber quiénes son, las líneas sobre un problema entregado, una respuesta y un tema dominado llevan, entre todas:

- un **nombre de recuento** para el niño o la niña: un código generado, con una clave propia del servicio, a partir del identificador del perfil y del mes del calendario. Es el mismo durante un mes y distinto al mes siguiente, y de él no se puede deducir nada sobre el niño o la niña: ni el identificador, ni el seudónimo, ni ninguna otra cosa;
- el chat (Claude, ChatGPT u otro), el idioma del problema, el grado, el mes en que se creó el perfil y cuántos temas están dominados;
- en la línea sobre un problema entregado, el **país** y, en Estados Unidos, el **estado**, si los indicó en el perfil, y el **país desde el que inició sesión**, cada uno como un código corto;
- para una respuesta, su tema, nivel y dificultad, si fue correcta, la trampa detrás de una incorrecta, si se usó la pista, si fue «no lo sé», con qué rapidez llegó (en menos de un minuto, en menos de tres o más tarde), la **probabilidad** de acertar que el servicio esperaba para su problema, quién eligió el problema (la regla del servicio, el chat o usted), si fue una de las primeras respuestas, las que determinan el nivel con el que empieza el niño o la niña, y, después de ellas, el tramo en el que estaba el número de respuestas del niño o la niña, por ejemplo, de 6 a 20.

Como todas las líneas sobre sus sesiones, estas líneas llevan además un código opaco que representa su cuenta: generado con la clave propia del servicio a partir de su cuenta de Google, no es ni su ID de Google ni su correo electrónico, y cambia en el primer inicio de sesión después de que se sustituya esa clave. Mientras se conservan a la vez las líneas de dos meses, ese código es lo que podría unir los dos nombres de recuento de un niño o una niña. También hace que las líneas sobre respuestas, mientras se conservan, sean un registro de las respuestas de su hijo o hija una a una: correcta o incorrecta, nunca la letra elegida.

Estas líneas se conservan como máximo 62 días. Lo que se conserva después son totales —cuántos niños y niñas, cuántos problemas, en qué países—: totales de cada día, totales de cada semana y de cada mes desglosados por un solo criterio a la vez, como el país o el grado, y, de cada mes, las trampas en las que cayeron las respuestas incorrectas, por tema y grado, y cómo salieron las respuestas a los problemas que eligió el servicio frente a la probabilidad que este esperaba, según esa probabilidad y según cuántas respuestas había dado el niño o la niña. No contienen ningún identificador de ningún tipo. Un total cuenta un grupo de niños y niñas, y un grupo puede constar de un solo niño o una sola niña; un total inferior a diez nunca se muestra públicamente.

Los registros nunca contienen el seudónimo, el texto de un problema, la letra de una respuesta, una dirección de correo electrónico, un token, el identificador del niño o la niña ni la dirección desde la que llegó una petición.

Aparte, Google Cloud Run conserva por defecto el registro estándar de cada petición a cualquier servicio que se ejecute en él: la hora, la dirección de la que procede, cómo se identifica el navegador, la página que la envió, la dirección completa solicitada, el código de estado, y el tamaño y la duración de la respuesta. Durante un inicio de sesión, la dirección solicitada lleva, tras un signo de interrogación, lo que el chat y Google ponen en ella: el código de inicio de sesión de un solo uso de Google, que solo el servicio puede canjear, el dominio de una cuenta de Google Workspace y la dirección de correo electrónico con la que el chat sugiere iniciar sesión, si sugiere alguna. MathTrail hace que Google Cloud no conserve el registro de ninguna petición cuya dirección lleve esa parte, así que ya no se conserva nada de eso en este registro. Los demás registros no contienen datos del perfil, ni texto de problemas, ni respuestas, ni direcciones de correo electrónico, y cada uno se conserva durante 30 días. Los registros de inicios de sesión hechos antes de esta actualización aún los contienen hasta que se cumplan sus 30 días.

### El país desde el que inicia sesión

Cuando usted inicia sesión, Google devuelve su navegador a MathTrail y, en ese único momento, el servicio busca a qué país pertenece la dirección de su navegador en la base de datos gratuita IP-to-Country de DB-IP, un archivo que el servicio guarda en su propio servidor, así que no se envía nada a DB-IP ni a nadie más. Solo se conserva el código de dos letras del país, sellado dentro del token que guarda su chat; la dirección en sí no se almacena ni se anota en ningún sitio. Una dirección que no pertenece a ningún país, como la de una red privada, no da ningún país. Las peticiones que hace su chat salen de los servidores del chat, no de usted, así que el país sigue siendo el de su último inicio de sesión.

Puede hacer que este país quede fuera del recuento: marque «No contar el país desde el que inicio sesión» en el formulario del perfil, o pídalo en el chat. A partir de entonces, ninguna línea que escriba el servicio lo incluirá, hasta que desmarque la casilla.

Geolocalización por IP de [DB-IP](https://db-ip.com), bajo la licencia [Creative Commons Attribution 4.0](https://creativecommons.org/licenses/by/4.0/).

## Para qué sirve cada cosa

- **El perfil**: para elegir cada problema y su dificultad, mostrar el progreso y aplicar el límite diario.
- **El inicio de sesión y el código de su cuenta**: para acceder a su único archivo y a nada más, y para aplicar los límites de cada cuenta.
- **Los países**: solo para contar cuántas familias hay en cada país. Puede omitir los dos.
- **Los registros**: para hacer funcionar el servicio, encontrar fallos y ver qué tal funcionan los problemas.
- **Los totales**: para saber cómo se usa MathTrail y en qué medida sus problemas se ajustan a los niños y las niñas. Solo se muestran como totales de diez o más niños y niñas, en la página «Estudio» de este sitio y en solicitudes de subvenciones.

## Cuánto tiempo se conserva cada cosa

- **El archivo del perfil**: hasta que usted lo elimine. Un archivo eliminado permanece 30 días en la papelera de Drive, salvo que usted la vacíe. Drive conserva durante un tiempo las versiones anteriores del archivo en su historial de versiones, y MathTrail pide a Drive que conserve para siempre la primera versión de cada día en que cambió el perfil, con un máximo de doscientas.
- **El token que guarda su chat**: 15 minutos; se renueva durante un máximo de 30 días cada vez, y como máximo hasta 90 días después del inicio de sesión.
- **Las dos cookies del inicio de sesión**: 10 minutos la que enlaza los pasos de un inicio de sesión, y 180 días la que recuerda las aplicaciones que usted aprobó.
- **Los registros del servicio y el registro de peticiones de Cloud Run**: 30 días.
- **La copia de las líneas con las que se cuenta a los niños y las niñas**: 62 días.
- **Las trazas** de cuánto tardaron las peticiones: 30 días. Abarcan aproximadamente una de cada diez peticiones, con la dirección, el navegador, la ruta y la consulta borrados.
- **Las mediciones** del funcionamiento del servicio, como cuántas llamadas tuvieron éxito: 24 meses. No contienen nada sobre ningún niño ni ninguna niña.
- **Los totales**: durante años. No contienen ningún identificador de ningún tipo.
- **Los registros de acceso de este sitio web y sus conversaciones**: el tiempo que los conserven GitHub y el proveedor de su chat, según sus propias políticas.

## Inicio de sesión con Google y Drive

Al iniciar sesión se pide a Google un solo permiso, **`drive.file`**. Es el permiso de Drive de alcance limitado: da acceso **solo a los archivos que ha creado la propia aplicación**. MathTrail no puede ver, listar ni abrir nada más de su Drive, incluidos los archivos que haya creado usted.

Para saber de quién era el inicio de sesión, el servicio pregunta después a Google qué cuenta de Google dio ese permiso. Google responde con el identificador de la cuenta, no con su correo electrónico, y el servicio solo conserva el código que genera a partir de ese identificador, descrito más arriba.

Conviene conocer de antemano dos consecuencias, porque parecen fallos y no lo son:

- **Google le pide permiso cada vez que se conecta.** El servicio no guarda ninguna credencial de Google propia de larga duración, así que cada conexión tiene que obtener una nueva, y Google muestra su pantalla de permisos cada vez.
- **Una cuenta de Google admite alrededor de un centenar de estas autorizaciones a la vez.** Después de unas cien conexiones, la más antigua deja de funcionar, sin previo aviso. Volver a iniciar sesión lo soluciona.

El token que el servicio emite para su chat es válido durante 15 minutos y se renueva automáticamente durante un máximo de 30 días cada vez, y como máximo hasta 90 días después del inicio de sesión. Después, hay que volver a iniciar sesión.

## Qué llega al modelo del chat

Todo lo que devuelve una herramienta de MathTrail lo lee el modelo del chat, y por eso llega al proveedor de su chat junto con el resto de la conversación. A lo largo de una clase, eso es casi todo lo que contiene el perfil:

- **el perfil y el progreso**: el seudónimo, el grado, los intereses, las habilidades excluidas, las anotaciones que usted escribió, el idioma de las clases, el país y el estado si los indicó, el tema al que se ciñen las clases, las puntuaciones y cómo han cambiado, los temas dominados, las últimas respuestas, los errores que se repiten y dónde está el archivo del perfil en su Drive, con su enlace;
- **el material para un problema**: el tema, la dificultad, problemas de ejemplo, los formatos que hay que seguir, el grado, los intereses y **las anotaciones libres que usted escribió sobre el niño o la niña**, que son las que permiten al modelo ajustar la redacción al nivel adecuado;
- **una respuesta**: si fue correcta, la opción correcta, la trampa detrás de una incorrecta y la solución.

Escriba las anotaciones teniéndolo en cuenta: llegan al proveedor del chat con todo lo demás. El identificador del niño o la niña no aparece en ningún resultado, y el seudónimo nunca entra en el texto de un problema.

El texto del propio problema lo escribe el modelo, y vuelve pasando por las comprobaciones de MathTrail.

## Qué puede ver usted y qué muestra el chat

Los hosts de chat muestran los argumentos sin procesar de las llamadas que hace una aplicación. Cuando el modelo envía un problema terminado, ese envío —**incluida la respuesta correcta**— es visible para cualquiera que lea el registro de llamadas a herramientas en la ventana del chat.

MathTrail no lo oculta ni puede ocultarlo. La respuesta se cifra para que no se pueda sacar del archivo del perfil ni entregar dos veces; no queda oculta para una persona adulta que lea el registro de su propio chat. Si quiere que el problema sea una sorpresa, no recorra el registro de llamadas delante del niño o la niña.

## Publicidad, elaboración de perfiles y entrenamiento

En MathTrail no hay publicidad. Nada sobre su hijo o hija se utiliza para crear un perfil publicitario, y nada se vende ni se cede a un corredor de datos. MathTrail no utiliza los datos de su hijo o hija para entrenar ningún modelo.

Lo que el proveedor del chat hace con la conversación se rige por su propia política, la misma que ya se aplica a todos los demás mensajes que usted envía en ese chat.

## Cookies

Este sitio web no instala cookies, no utiliza ninguna herramienta de analítica y no carga nada de otros dominios.

El servicio instala dos cookies, solo durante el inicio de sesión, y ambas son estrictamente necesarias: una que vincula el comienzo de un inicio de sesión con su final, y otra que recuerda qué aplicaciones ha aprobado ya, para que no se le pregunte dos veces. Ninguna se utiliza para hacer seguimiento.

## Cómo recuperar sus datos y eliminarlos

Todo está en su propio Drive, así que todo está en sus manos:

- **Exportar**: abra la carpeta `MathTrail` en Drive y descargue `mathtrail-profile.json`. Ese archivo es todo el perfil.
- **Eliminar**: elimine el archivo. Vacíe la papelera de Drive para borrarlo por completo. No queda nada del perfil en ninguna otra parte.
- **Retirar el acceso**: vaya a las [aplicaciones vinculadas](https://myaccount.google.com/linkedapps) de su cuenta de Google y quite el acceso de MathTrail a su cuenta. A partir de ahí, el servicio no puede acceder a nada.
- **Quitar la aplicación**: desconecte MathTrail en la configuración de su propio chat.

Eliminar el archivo y revocar el acceso son cosas independientes, y si hace las dos no queda nada.

## Quién más interviene

- **Google**: el inicio de sesión y Drive, donde está su archivo ([política de privacidad de Google](https://policies.google.com/privacy)).
- **Google Cloud**: donde se ejecuta el servicio (Cloud Run) y donde se conservan los totales descritos arriba (BigQuery) ([aviso de privacidad de Google Cloud](https://cloud.google.com/terms/cloud-privacy-notice)).
- **GitHub Pages**: desde donde se sirve este sitio web; GitHub guarda sus propios registros de acceso a él ([declaración de privacidad de GitHub](https://docs.github.com/en/site-policy/privacy-policies/github-general-privacy-statement)).
- **El proveedor de su chat**: Anthropic para Claude ([política de privacidad de Anthropic](https://www.anthropic.com/legal/privacy)), OpenAI para ChatGPT ([política de privacidad de OpenAI](https://openai.com/policies/privacy-policy/)).

Las personas que gestionan MathTrail pueden leer los registros del servicio y los totales en Google Cloud, que contienen lo descrito arriba y nada más. No hay nadie más.

## Cambios

Si esta política cambia, la fecha del principio cambia con ella y la redacción anterior queda en el historial público del proyecto.

## Idiomas

Esta política se publica en varios idiomas. El texto en inglés es el que prevalece: en caso de discrepancia entre una traducción y el texto en inglés, se aplica la versión en inglés.
