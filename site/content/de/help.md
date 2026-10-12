---
title: Hilfe
description: Wie Sie MathTrail verbinden, wie eine Lektion abläuft, was zu tun ist, wenn etwas nicht zu stimmen scheint, und wie Sie uns erreichen.
---

# Hilfe

Was Sie brauchen, wie Sie MathTrail verbinden, wie eine Lektion abläuft, was zu tun ist, wenn etwas nicht zu stimmen scheint, und wie Sie uns erreichen.

## Was Sie brauchen

- Ein eigenes Konto bei Claude oder ChatGPT und ein Google-Konto. Claude ist für Erwachsene gedacht, und ChatGPT verlangt bei Jugendlichen die Zustimmung eines Elternteils.
- Ihr Kind braucht keins von beidem: Es hat kein Konto und meldet sich nirgends an.

MathTrail bewahrt das Profil Ihres Kindes in einer einzigen Datei in Ihrem eigenen Google Drive auf und nirgendwo sonst.

## Mit Claude verbinden

1. Öffnen Sie in Claude [Anpassen → Konnektoren](https://claude.ai/customize/connectors).
2. Klicken Sie auf **+ Hinzufügen** und wählen Sie **Benutzerdefinierten Connector hinzufügen**.
3. Geben Sie als Namen MathTrail ein und fügen Sie diese Adresse ein: `https://mcp.mathtrail.app/mcp`
4. Klicken Sie auf **Weiter** und noch einmal auf **Weiter**: Die Einstellungen zur Authentifizierung, die Claude erkannt hat, müssen nicht geändert werden. Klicken Sie dann auf **Hinzufügen**, danach auf **Verbinden**, und melden Sie sich mit Google an. MathTrail bittet darum, in Ihrem Google Drive eine Datei mit dem Profil Ihres Kindes speichern zu dürfen. Erlauben Sie das auf der Seite von Google: In dieser Datei wird der Fortschritt Ihres Kindes gespeichert.
5. Klicken Sie in einem Chat auf **+** → **Konnektoren**, schalten Sie MathTrail ein und schreiben Sie: „Lass uns eine MathTrail-Aufgabe machen.“

Der Free-Plan von Claude erlaubt einen einzigen benutzerdefinierten Connector. Ein Konnektor, der auf claude.ai oder in Claude Desktop hinzugefügt wurde, funktioniert mit demselben Konto auch in den mobilen Apps von Claude.

## Mit ChatGPT verbinden

MathTrail ist noch nicht im Verzeichnis von ChatGPT. Bis dahin lässt sich MathTrail in kostenpflichtigen Tarifen über den Entwicklermodus von ChatGPT per Adresse hinzufügen, und zwar im Web: Aktivieren Sie den Entwicklermodus und fügen Sie eine App mit der Adresse oben hinzu, wie [OpenAI es beschreibt](https://developers.openai.com/plugins/deploy/connect-chatgpt). In ChatGPT wird MathTrail noch getestet.

## So läuft eine Lektion ab

- **Beim ersten Mal** bittet der Chat Sie zu bestätigen, dass Sie ein Elternteil oder die Lehrkraft des Kindes sind, und fragt dann nach einem Pseudonym für Ihr Kind, nie nach dem echten Namen, und nach der Klasse von 1 bis 6. Wenn Sie möchten, nennen Sie auch seine Interessen und alles, was in den Aufgaben nicht vorkommen soll.
- **Die ersten fünf Aufgaben** ermitteln, wo Ihr Kind steht. Danach richten sich die Aufgaben nach den Antworten.
- **Sie schreiben, und Ihr Kind tippt auf die Karte.** Bitten Sie im Chat um eine Aufgabe. Eine Karte erscheint, und nach ein bis zwei Minuten kommt darauf die Aufgabe mit fünf Antwortmöglichkeiten; Ihr Kind antwortet, indem es eine davon antippt. **Die Glühbirne** öffnet einen Tipp. **Die zwei Pfeile** bitten den Chat um die nächste Aufgabe, und sobald Ihr Kind geantwortet hat, tut **Neue Aufgabe** dasselbe. Die Aufgabe kommt darunter, in einer neuen Karte. Während Ihr Kind eine Aufgabe löst, schreibt der Chat schon die nächste, deshalb erscheint sie meist sofort auf der neuen Karte.
- **Nach einer Antwort** wird die Karte sofort zur Besprechung der Antwort: der Fehler hinter einer falschen Antwort, ein Bild der Lösung und die Lösung Schritt für Schritt. An den Chat geht dabei nichts. Wenn Ihr Kind die Antwort nicht weiß, sagen Sie das dem Chat, und darunter kommt eine Karte mit der Lösung.
- **Fortschritt:** Fragen Sie „Wie kommt mein Kind voran?“, dann sehen Sie, wo Ihr Kind in jedem Thema steht und was sich in den letzten sieben Tagen bewegt hat, dazu eine Auswertung und die wiederkehrenden Fehler. Mit **Ändern** passen Sie dort das Profil an, oder Sie bitten im Chat darum.
- **Lassen Sie das Protokoll der Tool-Aufrufe im Chat geschlossen, während Ihr Kind eine Aufgabe löst:** Es zeigt jede Aufgabe zusammen mit ihrer Antwort.

## Wenn etwas schiefgeht

- **Google schickt Sie zurück auf eine Seite von MathTrail, „Tick the box for Google Drive“.** Im Kästchen, das MathTrail erlaubt, die Datei zu speichern, fehlt das Häkchen. Klicken Sie auf **Back to Google** und setzen Sie das Häkchen: Ohne diese Datei kann der Fortschritt nirgends gespeichert werden.
- **MathTrail bittet Sie, sich erneut anzumelden.** Eine Anmeldung gilt höchstens 90 Tage. Außerdem erlaubt ein Google-Konto einer App etwa hundert Autorisierungen gleichzeitig, sodass nach vielen Verbindungen die älteste nicht mehr funktioniert. Eine erneute Anmeldung behebt beides.
- **Claude fügt keinen weiteren Konnektor hinzu.** Im Free-Plan ist Platz für einen benutzerdefinierten Connector: Entfernen Sie zuerst einen anderen.
- **Die Karte wartet ein bis zwei Minuten.** Das Modell des Chats schreibt jede Aufgabe, und MathTrail prüft sie, bevor Ihr Kind sie sieht. So lange dauert die erste Aufgabe einer Lektion; die nächsten werden geschrieben, während Ihr Kind eine Aufgabe löst, und kommen sofort, sobald sie fertig sind. Mit der üblichen Denktiefe des Chats bleibt die Wartezeit kurz: Bei der größten Denktiefe denkt das Modell über jede Aufgabe eine Minute oder länger nach. Lehnen die Prüfungen eine Aufgabe ab, wird sie neu geschrieben, mit insgesamt drei Versuchen; scheitern alle drei, sagt der Chat das, und Sie können um eine andere bitten.
- **Die Karte zeigt „Wünsch dir etwas für morgen“.** Ein Kind bekommt bis zu 50 neue Aufgaben am Tag, und außerdem ist der Tag nach zehn Anfragen vorbei, bei denen keine Aufgabe geschrieben werden konnte. Der Tag endet um Mitternacht nach der Uhr des Geräts, auf dem Ihr Kind die Karte beantwortet. Solange Ihr Kind noch auf keiner Karte geantwortet hat, endet er um Mitternacht UTC. Der Fortschritt und die letzte Aufgabe bleiben verfügbar.
- **Claude setzt „Neue Aufgabe“ ins Eingabefeld, unter einem Hinweis, der zur Vorsicht mahnt.** Das ist die Bitte der Karte, die auf Sie wartet: Drücken Sie die Eingabetaste, um sie zu senden.
- **Claude fragt, ob MathTrail das Profil speichern darf.** Beim Speichern wird ersetzt, was vorher im Profil stand, deshalb fragt Claude zuerst. Erlauben Sie es, wenn Sie um die Änderung gebeten haben.
- **Es erscheint keine Karte.** Manche Chats zeigen keine Karten. Dann gibt der Chat die Aufgabe in Worten mit den Antwortmöglichkeiten A bis E, und Sie schreiben den Buchstaben der Antwort Ihres Kindes.
- **Das Profil lässt sich nicht lesen, oder es liegt im Papierkorb in Google Drive.** MathTrail sagt das im Chat und ändert nichts. Stellen Sie die Datei aus dem Papierkorb wieder her, oder lassen Sie MathTrail die letzte lesbare Version zurückholen oder ein neues Profil anlegen, wie es die [Anleitung zu den Daten Ihres Kindes](https://github.com/MathTrail/mathtrail-standalone/blob/main/docs/parents.md) (auf Englisch) erklärt.
- **Eine Aufgabe sieht falsch aus.** Das kann sein: Das Programm prüft, was es prüfen kann, und den Sinn prüft das Modell selbst. Schreiben Sie uns, was in der Aufgabe stand.

## Die Daten Ihres Kindes

Alles, was MathTrail über Ihr Kind weiß, steht in einer einzigen Datei, `mathtrail-profile.json`, in einem Ordner `MathTrail` in Ihrem Google Drive. Wie Sie die Datei herunterladen, die Verbindung zu MathTrail trennen oder alles löschen, steht in der [Anleitung zu den Daten Ihres Kindes](https://github.com/MathTrail/mathtrail-standalone/blob/main/docs/parents.md) (auf Englisch). Was gespeichert wird, wo und wie lange, steht in der [Datenschutzerklärung](../privacy/).

## Kontakt

- **Hilfe und Fragen zu den Daten Ihres Kindes:** Schreiben Sie an [altedtech.info@gmail.com](mailto:altedtech.info@gmail.com).
- **Ein Fehler in einer Aufgabe oder eine Idee:** Schreiben Sie an dieselbe Adresse, oder erstellen Sie ein Issue auf [GitHub](https://github.com/MathTrail/mathtrail-standalone/issues). Issues sind öffentlich; nehmen Sie deshalb keinen Namen, kein Foto und keine anderen Angaben zu Ihrem Kind in ein Issue auf.
- **Ein Sicherheitsproblem:** Melden Sie es vertraulich über die [Sicherheitsempfehlungen von GitHub](https://github.com/MathTrail/mathtrail-standalone/security/advisories/new), wie die [Sicherheitsrichtlinie](https://github.com/MathTrail/mathtrail-standalone/blob/main/SECURITY.md) des Projekts darum bittet.

Wir lesen alles und antworten, so schnell wir können, aber eine bestimmte Antwortzeit versprechen wir nicht.
