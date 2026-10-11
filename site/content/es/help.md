---
title: Ayuda
description: Cómo conectar MathTrail, cómo transcurre una clase, qué hacer si algo parece ir mal y cómo ponerse en contacto con nosotros.
---

# Ayuda

Qué necesita, cómo conectar MathTrail, cómo transcurre una clase, qué hacer si algo parece ir mal y cómo ponerse en contacto con nosotros.

## Qué necesita

- Una cuenta propia de Claude o de ChatGPT, y una cuenta de Google. Claude es para personas adultas, y ChatGPT requiere el consentimiento de la madre o el padre para adolescentes.
- Su hijo o hija no necesita ninguna de las dos: no tiene cuenta y no inicia sesión en ningún sitio.

MathTrail guarda el perfil de su hijo o hija en un solo archivo de su propio Google Drive, y en ningún otro sitio.

## Conectarlo en Claude

1. En Claude, abra [Personalizar → Conectores](https://claude.ai/customize/connectors).
2. Pulse **+ Añadir** y elija **Añadir conector personalizado**.
3. Escriba MathTrail como nombre y pegue la dirección: `https://mcp.mathtrail.app/mcp`
4. Pulse **Continuar**, y otra vez **Continuar**: no hace falta cambiar la configuración de autenticación que Claude detecta. Después pulse **Añadir**, luego **Conectar**, e inicie sesión con Google. MathTrail pide permiso para guardar en su Google Drive un solo archivo con el perfil de su hijo o hija. Permítalo en la pantalla de Google: en ese archivo se guarda el progreso de su hijo o hija.
5. En un chat, pulse **+** → **Conectores**, active MathTrail y diga: «Hagamos un problema de MathTrail».

El plan Free de Claude permite un solo conector personalizado. Un conector añadido en claude.ai o en Claude Desktop también funciona en las aplicaciones móviles de Claude, con la misma cuenta.

## Conectarlo en ChatGPT

MathTrail aún no está en el directorio de ChatGPT. Hasta que lo esté, el modo para desarrolladores de ChatGPT permite añadirlo por su dirección en los planes de pago, en la web: active el modo para desarrolladores y añada una app con la dirección de arriba, como [describe OpenAI](https://developers.openai.com/plugins/deploy/connect-chatgpt). MathTrail todavía se está probando en ChatGPT.

## Cómo transcurre una clase

- **La primera vez**, el chat le pide que confirme que es madre, padre o docente del niño o la niña, y después un seudónimo para su hijo o hija, nunca el nombre real, y el grado, de 1.º a 6.º. Si lo desea, añada los intereses y lo que deba quedar fuera de los problemas.
- **Los cinco primeros problemas** sirven para ver en qué punto está su hijo o hija. Después, los problemas se ajustan a las respuestas.
- **Usted escribe y su hijo o hija toca.** Pida un problema en el chat. Aparece una tarjeta y, en un minuto o dos, llega a ella el problema con cinco opciones; su hijo o hija responde tocando una. **La bombilla** abre una pista. **Las dos flechas** piden al chat el siguiente problema y, una vez dada la respuesta, **Otro problema** hace lo mismo. El problema llega más abajo, en una tarjeta nueva. Mientras su hijo o hija resuelve, el chat escribe por adelantado el siguiente problema, así que normalmente aparece enseguida en la tarjeta nueva.
- **Después de una respuesta**, la tarjeta se convierte al instante en el repaso de la respuesta: el error que hay detrás de una respuesta incorrecta, un dibujo de la solución y la solución paso a paso. Al chat no llega nada. Si su hijo o hija no sabe la respuesta, dígaselo al chat, y más abajo aparecerá una tarjeta con la solución.
- **Progreso:** pregunte «¿Cómo va mi hijo o hija?» para ver en qué punto está en cada tema, qué ha cambiado en la última semana, un análisis y los errores que se repiten. Allí, **Editar** cambia el perfil; también puede pedirlo en el chat.
- **Mantenga cerrado el registro de llamadas a herramientas del chat mientras su hijo o hija resuelve:** muestra cada problema junto con su respuesta.

## Cuando algo sale mal

- **Al volver de Google aparece una página de MathTrail, «Tick the box for Google Drive».** La casilla que permite a MathTrail guardar su archivo se quedó sin marcar. Pulse **Back to Google** y márquela: sin ese archivo, el progreso no tiene dónde guardarse.
- **MathTrail le pide que vuelva a iniciar sesión.** Un inicio de sesión dura como máximo 90 días. Además, una cuenta de Google admite a la vez alrededor de un centenar de autorizaciones para una misma aplicación, así que, después de muchas conexiones, la más antigua deja de funcionar. Volver a iniciar sesión soluciona ambas cosas.
- **Claude no permite añadir otro conector.** El plan Free de Claude tiene cabida para un solo conector personalizado: quite antes otro.
- **La tarjeta se queda esperando un minuto o dos.** El modelo del chat escribe cada problema, y MathTrail lo comprueba antes de que su hijo o hija lo vea. Eso es lo que tarda el primer problema de una clase; los siguientes se escriben mientras su hijo o hija resuelve, y llegan enseguida en cuanto están listos. Con la profundidad de razonamiento habitual del chat, la espera es corta: con la más profunda, el modelo reflexiona sobre cada problema durante un minuto o más. Si las comprobaciones rechazan un problema, se vuelve a escribir, con tres intentos en total; si fallan los tres, el chat lo dice y usted puede pedir otro.
- **La tarjeta dice «Pide un deseo para mañana».** Cada niño o niña recibe hasta 50 problemas nuevos al día, y el día también se cierra tras diez peticiones en las que no se pudo escribir ningún problema. El día termina a medianoche según el reloj del dispositivo en el que su hijo o hija responde en la tarjeta. Mientras su hijo o hija no haya respondido en ninguna tarjeta, el día termina a medianoche UTC. El progreso y el último problema siguen disponibles.
- **Claude pone «Otro problema» en el cuadro de mensaje, bajo un aviso que pide precaución antes de enviarlo.** Es la petición de la tarjeta, que espera a que usted la envíe: pulse Intro.
- **Claude pregunta si permite que MathTrail guarde el perfil.** Al guardarlo se sustituye lo que el perfil contenía antes, así que Claude pregunta primero. Permítalo si el cambio lo ha pedido usted.
- **No aparece ninguna tarjeta.** Algunos chats no muestran tarjetas. Entonces el chat da el problema en forma de texto, con sus opciones de la A a la E, y usted escribe la letra que elija su hijo o hija.
- **El perfil no se puede leer o está en la papelera de Drive.** MathTrail lo indica en el chat y no cambia nada. Restaure el archivo desde la papelera, o deje que MathTrail recupere su última versión legible o empiece un perfil nuevo, como explica [la guía sobre los datos de su hijo o hija](https://github.com/MathTrail/mathtrail-standalone/blob/main/docs/parents.md) (en inglés).
- **Un problema parece incorrecto.** Puede que lo sea: el programa comprueba lo que puede, y el sentido lo comprueba el propio modelo. Díganos qué decía el problema.

## Los datos de su hijo o hija

Todo lo que MathTrail sabe de su hijo o hija está en un solo archivo, `mathtrail-profile.json`, en una carpeta `MathTrail` de su Google Drive. Cómo descargarlo, desconectar MathTrail o eliminarlo todo se explica en [la guía sobre los datos de su hijo o hija](https://github.com/MathTrail/mathtrail-standalone/blob/main/docs/parents.md) (en inglés). Qué se guarda, dónde y durante cuánto tiempo se explica en la [política de privacidad](../privacy/).

## Contacto

- **Ayuda y preguntas sobre los datos de su hijo o hija:** escriba a [altedtech.info@gmail.com](mailto:altedtech.info@gmail.com).
- **Un error en un problema, o una idea:** escriba a la misma dirección o abra una incidencia en [GitHub](https://github.com/MathTrail/mathtrail-standalone/issues). Las incidencias son públicas, así que no incluya en ellas ningún nombre, fotografía ni otro dato de su hijo o hija.
- **Un problema de seguridad:** comuníquelo de forma privada a través de los [avisos de seguridad de GitHub](https://github.com/MathTrail/mathtrail-standalone/security/advisories/new), como pide la [política de seguridad](https://github.com/MathTrail/mathtrail-standalone/blob/main/SECURITY.md) del proyecto.

Leemos todo y respondemos lo antes posible, pero no prometemos ningún plazo de respuesta.
