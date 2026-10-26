---
title: Datenschutzerklärung
description: Was MathTrail über Ihr Kind weiß, wo das gespeichert ist und wie Sie es zurückholen oder löschen.
---

# Datenschutzerklärung

Stand: 26. Oktober 2026.

MathTrail ist ein kostenloses Open-Source-Projekt, das von ALT EdTech entwickelt und betrieben wird. Fragen zu dieser Datenschutzerklärung oder zu den Daten Ihres Kindes richten Sie an [altedtech.info@gmail.com](mailto:altedtech.info@gmail.com). Wie Sie MathTrail verbinden und was zu tun ist, wenn etwas nicht zu stimmen scheint, steht auf der [Hilfeseite](../help/).

## Die Kurzfassung

- **Wir speichern nichts.** Der Dienst hat keine Datenbank und keine Dateien. Zwischen einer Anfrage und der nächsten merkt er sich nichts.
- **Das Profil Ihres Kindes liegt in Ihrem eigenen Google Drive**, in einer Datei, die Sie jederzeit öffnen, kopieren oder löschen können.
- **Ein Pseudonym, nie der echte Name.** Kein Geburtsdatum, keine Schule, kein Foto.
- **Wir zählen Kinder, ohne zu wissen, wer sie sind.** Eine Zeile des Protokolls bezeichnet ein Kind mit einem Code, der sich jeden Monat ändert und sich auf niemanden zurückführen lässt; was länger als zwei Monate aufbewahrt wird, sind Gesamtzahlen.
- **Nichts wird verkauft, und mit nichts wird ein Modell trainiert.**
- **Mit der Datei löschen Sie das Profil.** Eine zweite Kopie gibt es nirgends.

## Was MathTrail ist

MathTrail ist eine App, die Sie in Ihrem eigenen Claude- oder ChatGPT-Konto hinzufügen. MathTrail wählt für Ihr Kind ein Thema und eine Schwierigkeit, bittet das eigene Modell des Chats, eine Aufgabe zu schreiben, prüft diese Aufgabe mit einem Programm, zeigt sie an, hält die Antwort fest und erklärt den Fehler.

MathTrail betreibt kein eigenes Sprachmodell. Das Modell, das die Aufgaben schreibt, ist das des Chats, den Sie ohnehin nutzen, und Ihr Verhältnis zum Anbieter dieses Chats richtet sich nach dessen Nutzungsbedingungen und dessen Datenschutzerklärung, nicht nach dieser hier.

## Wer sich anmeldet

Nur eine erwachsene Person – ein Elternteil oder eine Lehrkraft. Das Kind hat kein eigenes Konto und keine eigene Anmeldung. Die Chat-Plattformen legen ihre eigenen Altersregeln fest: Claude ist für Erwachsene gedacht, und ChatGPT verlangt bei Jugendlichen die Zustimmung eines Elternteils. MathTrail richtet sich nicht an Kinder, die einen Chat allein nutzen.

## Kinder

Die Aufgaben von MathTrail sind für Kinder der Klassen 1 bis 6 bestimmt, also etwa im Alter von sechs bis zwölf Jahren. Gedacht ist MathTrail so: Die erwachsene Person schreibt im Chat, während das Kind daneben antwortet, indem es auf der Karte eine Antwortmöglichkeit antippt. Das Kind hat kein Konto. Das Profil enthält, was die erwachsene Person unter einem Pseudonym über das Kind angibt, dazu die Antworten des Kindes, und zwar in einer Datei, die der erwachsenen Person selbst gehört; was die Protokolle des Dienstes von diesen Antworten aufbewahren und wie lange, steht weiter unten.

## Was wo gespeichert wird

### In Ihrem Google Drive

Das Profil ist eine einzige JSON-Datei, `mathtrail-profile.json`, in einem Ordner namens `MathTrail` in Ihrem eigenen Drive. Sie können sie wie jede andere Datei öffnen, lesen, kopieren, verschieben oder löschen, und der Papierkorb und der Versionsverlauf in Google Drive gehören Ihnen.

Die Datei enthält:

- ein **Pseudonym** für das Kind, die Klasse, seine Interessen und alles, was die Aufgaben auf Ihren Wunsch meiden sollen;
- eine Kennung für das Kind, damit sich das Profil in eine künftige Ausgabe mitnehmen lässt;
- die Sprache der Benutzeroberfläche, falls Sie eine gewählt haben;
- das Thema, auf das sich die Lektionen beschränken, falls Sie oder Ihr Kind eines gewählt haben;
- das Land, in dem Sie leben, und in den USA Ihren Bundesstaat, falls Sie sie angeben wollten – sie dienen nur dazu, zu zählen, wie viele Familien es in jedem Land gibt, und Sie können sie weglassen;
- welche Themen gemeistert sind, die Wertungen, nach denen sich die Schwierigkeit richtet, und eine Zusammenfassung der bisherigen Antworten: je Thema, wann es zuletzt drankam, wie viele Versuche und Erfolge es gab und welche Fallen vorkamen;
- wo die Wertungen vor jeder der letzten Antworten standen und zu Beginn jedes der letzten sieben Tage mit Antworten, damit der Fortschritt zeigen kann, was sich seit der letzten Aufgabe und im Lauf der Woche bewegt hat;
- einen Zähler der heute angenommenen Aufgaben, über den das Tageslimit eingehalten wird, und um wie viel die Uhr des Geräts, auf dem Ihr Kind antwortet, gegenüber UTC vor- oder nachgeht: Die Karte sendet das mit jeder Antwort, damit der Tag bei Ihnen um Mitternacht endet;
- die aktuelle Aufgabe, deren Antwort verschlüsselt ist, damit sie sich nicht aus der Datei lesen lässt, bevor das Kind antwortet;
- die nächste Aufgabe, die der Chat im Voraus schreibt, während das Kind die aktuelle löst, mit einer ebenso verschlüsselten Antwort, bis sie angefordert wird.

Nicht enthalten sind der Name des Kindes, sein Geburtsdatum, seine Schule, seine Adresse, ein Foto oder irgendetwas anderes, woran man es erkennen könnte. **Das Pseudonym ist der einzige Name, der überhaupt vorkommt, und es erscheint nie in einer Aufgabe.**

### Auf dem Server

Nichts. Der Dienst ist ein einziges Programm, das eine Anfrage beantwortet und sie wieder vergisst. Er hat keine Datenbank, keine Festplatte und keinen Cache mit Ihren Daten.

Auch Ihre Google-Anmeldung wird nicht gespeichert: Der Dienst verschlüsselt sie und legt sie in das Token, das er Ihrem Chat übergibt, zusammen mit dem Land, aus dem Sie sich angemeldet haben, wie weiter unten beschrieben. Der Schlüssel liegt im Google Secret Manager und verlässt den Server nie.

### In den Protokollen

Der Dienst schreibt Zahlen, keine Inhalte: ob eine Aufgabe angenommen oder abgelehnt wurde, warum sie abgelehnt wurde, wie lange es gedauert hat, wie viele Versuche nötig waren, wie viele Bytes jeder Teil der vom Chat eingereichten Aufgabe belegte, ob ein Limit erreicht wurde und welche Version der Anweisungen verwendet wurde. Die Zeilen zu den Aufrufen, die eine Aufgabe gebraucht hat, tragen einen gemeinsamen Zufallscode, der für diese Aufgabe erzeugt wird, sie miteinander verbindet und nichts über das Kind aussagt.

Um zu zählen, wie viele Kinder MathTrail an jedem Tag, in jeder Woche und in jedem Monat nutzen, ohne zu wissen, wer sie sind, enthalten die Zeilen über eine gestellte Aufgabe, über eine Antwort und über ein gemeistertes Thema zusammengenommen Folgendes:

- einen **Zählnamen** für das Kind: einen Code, den der Dienst mit einem eigenen Schlüssel aus der Kennung des Profils und dem Kalendermonat bildet. Er bleibt einen Monat lang gleich und ist im nächsten ein anderer, und nichts über das Kind – weder die Kennung noch das Pseudonym noch sonst etwas – lässt sich daraus ableiten;
- den Chat (Claude, ChatGPT oder einen anderen), die Sprache der Aufgabe, die Klasse, den Monat, in dem das Profil angelegt wurde, und wie viele Themen gemeistert sind;
- in der Zeile über eine gestellte Aufgabe das **Land** und in den USA den **Bundesstaat**, falls Sie sie im Profil angegeben haben, sowie das **Land, aus dem Sie sich angemeldet haben**, jeweils als kurzen Code;
- zu einer Antwort ihr Thema, ihr Niveau und ihre Schwierigkeit, ob sie richtig war, die Falle hinter einer falschen, ob der Tipp genutzt wurde, ob sie „Ich weiß es nicht“ lautete, wie schnell sie kam (innerhalb einer Minute, innerhalb von drei Minuten oder später), die **Lösungswahrscheinlichkeit**, die der Dienst für ihre Aufgabe erwartet hat, wer die Aufgabe gewählt hat (die Regel des Dienstes, der Chat oder Sie), ob sie zu den ersten Antworten gehörte, mit denen das Anfangsniveau des Kindes ermittelt wird, und nach diesen den Bereich, in dem die Zahl der Antworten des Kindes lag, etwa 6 bis 20.

Wie jede Zeile über Ihre Sitzungen tragen auch diese Zeilen einen undurchsichtigen Code, der für Ihr Konto steht: Der Dienst bildet ihn mit seinem eigenen Schlüssel aus Ihrem Google-Konto, er ist weder Ihre Google-ID noch Ihre E-Mail-Adresse, und er ändert sich bei der ersten Anmeldung, nachdem dieser Schlüssel ausgetauscht wurde. Solange die Zeilen zweier Monate gleichzeitig aufbewahrt werden, ist es dieser Code, der die beiden Zählnamen eines Kindes verbinden könnte. Außerdem macht er die Zeilen über Antworten, solange sie aufbewahrt werden, zu einer Aufzeichnung der Antworten Ihres Kindes, eine nach der anderen: richtig oder falsch, nie der gewählte Buchstabe.

Diese Zeilen werden höchstens 62 Tage aufbewahrt. Was danach aufbewahrt wird, sind Gesamtzahlen – wie viele Kinder, wie viele Aufgaben, in welchen Ländern: Gesamtzahlen für jeden Tag, Gesamtzahlen für jede Woche und jeden Monat, aufgeschlüsselt nach jeweils einem Merkmal, etwa dem Land oder der Klasse, und für jeden Monat die Fallen, in die falsche Antworten getappt sind, nach Thema und Klasse, sowie wie die Antworten auf die vom Dienst gewählten Aufgaben im Vergleich zu der von ihm erwarteten Lösungswahrscheinlichkeit ausfielen, aufgeschlüsselt nach dieser Wahrscheinlichkeit und danach, wie viele Antworten das Kind schon gegeben hatte. Kennungen irgendeiner Art enthalten sie nicht. Eine Gesamtzahl zählt eine Gruppe von Kindern, und eine Gruppe kann aus einem einzigen Kind bestehen; eine Gesamtzahl unter zehn wird nie öffentlich gezeigt.

Die Protokolle enthalten nie das Pseudonym, den Text einer Aufgabe, den Buchstaben einer Antwort, eine E-Mail-Adresse, ein Token, die Kennung des Kindes oder die Adresse, von der eine Anfrage kam.

Unabhängig davon zeichnet Google Cloud Run standardmäßig jede Anfrage an jeden Dienst auf, der dort läuft: die Uhrzeit, die Adresse, von der sie kam, als was sich der Browser ausgibt, die Seite, von der sie geschickt wurde, die vollständige angefragte Adresse, den Status sowie Größe und Dauer der Antwort. Während einer Anmeldung enthält die angefragte Adresse nach einem Fragezeichen, was der Chat und Google hineingeschrieben haben: den einmal gültigen Anmeldecode von Google, den nur der Dienst einlösen kann, die Domain eines Google Workspace-Kontos und die E-Mail-Adresse, die der Chat für die Anmeldung vorschlägt, falls er eine vorschlägt. MathTrail lässt Google Cloud keine Aufzeichnung einer Anfrage aufbewahren, deren Adresse einen solchen Teil enthält, sodass nichts davon mehr in dieser Aufzeichnung aufbewahrt wird. Die übrigen Aufzeichnungen enthalten keine Profildaten, keine Aufgabentexte, keine Antworten und keine E-Mail-Adressen, und jede wird 30 Tage aufbewahrt. Aufzeichnungen von Anmeldungen aus der Zeit vor dieser Aktualisierung enthalten sie noch, bis ihre 30 Tage um sind.

### Das Land, aus dem Sie sich anmelden

Wenn Sie sich anmelden, schickt Google Ihren Browser zurück zu MathTrail, und in diesem einen Moment schlägt der Dienst nach, zu welchem Land die Adresse Ihres Browsers gehört. Er nutzt dafür die kostenlose IP-to-Country-Datenbank von DB-IP, eine Datei, die der Dienst auf seinem eigenen Server hat, sodass nichts an DB-IP oder sonst jemanden gesendet wird. Aufbewahrt wird nur der Ländercode aus zwei Buchstaben, versiegelt im Token, das Ihr Chat hält; die Adresse selbst wird weder gespeichert noch irgendwo festgehalten. Eine Adresse, die zu keinem Land gehört, etwa eine aus einem privaten Netz, ergibt kein Land. Die Anfragen Ihres Chats kommen von dessen Servern, nicht von Ihnen, deshalb bleibt das Land das Ihrer letzten Anmeldung.

Sie können dieses Land aus der Zählung herausnehmen lassen: Setzen Sie im Profilformular ein Häkchen bei „Das Land, aus dem ich mich anmelde, nicht mitzählen“, oder bitten Sie im Chat darum. Von da an enthält keine Zeile, die der Dienst schreibt, dieses Land, bis Sie das Häkchen wieder entfernen.

IP-Geolokalisierung durch [DB-IP](https://db-ip.com), unter der Lizenz [Creative Commons Attribution 4.0](https://creativecommons.org/licenses/by/4.0/).

## Wozu das jeweils dient

- **Das Profil** – um jede Aufgabe und ihre Schwierigkeit zu wählen, den Fortschritt zu zeigen und das Tageslimit einzuhalten.
- **Die Anmeldung und der Code Ihres Kontos** – um auf diese eine Datei zuzugreifen und auf nichts sonst, und um die Limits jedes Kontos einzuhalten.
- **Die Länder** – nur um zu zählen, wie viele Familien es in jedem Land gibt. Sie können beide weglassen.
- **Die Protokolle** – um den Dienst zu betreiben, Fehler zu finden und zu sehen, wie sich die Aufgaben bewähren.
- **Die Gesamtzahlen** – um zu wissen, wie MathTrail genutzt wird und wie gut die Aufgaben zu den Kindern passen. Gezeigt werden sie nur als Gesamtzahlen von zehn oder mehr Kindern, auf der Seite „Studie“ dieser Website und in Förderanträgen.

## Wie lange was aufbewahrt wird

- **Die Profildatei** – bis Sie sie löschen. Eine gelöschte Datei bleibt 30 Tage im Papierkorb in Google Drive, wenn Sie ihn nicht vorher leeren. Drive bewahrt frühere Versionen der Datei eine Zeit lang im Versionsverlauf auf, und MathTrail bittet Drive, die erste Version jedes Tages, an dem sich das Profil geändert hat, für immer zu behalten, höchstens zweihundert davon.
- **Das Token, das Ihr Chat hält** – 15 Minuten; es wird jeweils für bis zu 30 Tage erneuert und höchstens bis 90 Tage nach der Anmeldung.
- **Die zwei Cookies der Anmeldung** – 10 Minuten das Cookie, das die Schritte einer Anmeldung zusammenhält, und 180 Tage das Cookie, das sich die von Ihnen zugelassenen Apps merkt.
- **Die Protokolle des Dienstes und die Aufzeichnung der Anfragen durch Cloud Run** – 30 Tage.
- **Die Kopie der Zeilen, nach denen Kinder gezählt werden** – 62 Tage.
- **Traces** zur Dauer von Anfragen – 30 Tage. Erfasst wird etwa jede zehnte Anfrage, wobei Adresse, Browser, Pfad und Abfrageparameter unkenntlich gemacht sind.
- **Messwerte** zur Arbeit des Dienstes, etwa wie viele Aufrufe gelungen sind – 24 Monate. Über ein Kind enthalten sie nichts.
- **Die Gesamtzahlen** – jahrelang. Kennungen irgendeiner Art enthalten sie nicht.
- **Die Zugriffsprotokolle dieser Website und Ihre Unterhaltungen** – so lange, wie GitHub und der Anbieter Ihres Chats sie nach ihren eigenen Richtlinien aufbewahren.

## Anmeldung mit Google und Drive

Bei der Anmeldung wird Google um eine einzige Berechtigung gebeten, **`drive.file`**. Das ist die eng gefasste Drive-Berechtigung: Zugriff gewährt sie **nur auf Dateien, die diese App selbst erstellt hat**. MathTrail kann nichts anderes in Ihrem Drive sehen, auflisten oder öffnen, auch keine Dateien, die Sie selbst angelegt haben.

Um zu wissen, wessen Anmeldung es war, fragt der Dienst Google anschließend, welches Google-Konto diese Berechtigung erteilt hat. Google nennt ihm die Kennung des Kontos, nicht Ihre E-Mail-Adresse, und der Dienst behält nur den oben beschriebenen Code, den er aus dieser Kennung bildet.

Zwei Folgen sollten Sie vorher kennen, weil sie wie Störungen aussehen, aber keine sind:

- **Google bittet bei jeder Verbindung um Ihre Zustimmung.** Der Dienst bewahrt keine eigenen langlebigen Google-Zugangsdaten auf, deshalb muss jede Verbindung sie neu einholen, und Google zeigt jedes Mal seinen Zustimmungsbildschirm.
- **Ein Google-Konto erlaubt etwa hundert solcher Autorisierungen gleichzeitig.** Nach ungefähr hundert Verbindungen funktioniert die älteste ohne Vorwarnung nicht mehr. Eine erneute Anmeldung behebt das.

Das Token, das der Dienst Ihrem Chat ausstellt, gilt 15 Minuten und wird automatisch erneuert, jeweils für bis zu 30 Tage und höchstens bis 90 Tage nach der Anmeldung. Danach melden Sie sich erneut an.

## Was beim Modell des Chats ankommt

Alles, was ein Tool von MathTrail zurückgibt, wird vom Modell des Chats gelesen und gelangt so mit dem Rest der Unterhaltung zum Anbieter Ihres Chats. Im Lauf einer Lektion ist das fast alles, was das Profil enthält:

- **das Profil und der Fortschritt** – das Pseudonym, die Klasse, die Interessen, die ausgeschlossenen Fähigkeiten, die Notizen, die Sie geschrieben haben, die Sprache der Lektionen, Land und Bundesstaat, falls Sie sie angegeben haben, das Thema, auf das sich die Lektionen beschränken, die Wertungen und wie sie sich bewegt haben, die gemeisterten Themen, die letzten Antworten, die wiederkehrenden Fehler und wo die Profildatei in Ihrem Drive liegt, samt Link;
- **das Material für eine Aufgabe** – das Thema, die Schwierigkeit, Beispielaufgaben, die einzuhaltenden Formate, die Klasse, die Interessen und **die freien Notizen, die Sie über das Kind geschrieben haben**, mit denen das Modell die Formulierung erst auf das richtige Niveau abstimmen kann;
- **eine Antwort** – ob sie richtig war, die richtige Antwortmöglichkeit, die Falle hinter einer falschen und die Lösung.

Schreiben Sie die Notizen entsprechend, denn sie gehen mit allem anderen an den Anbieter des Chats. Die Kennung des Kindes steht in keinem Ergebnis, und das Pseudonym kommt nie in den Text einer Aufgabe.

Den Text der Aufgabe selbst schreibt das Modell, und er durchläuft auf dem Rückweg die Prüfungen von MathTrail.

## Was Sie sehen und was der Chat zeigt

Chat-Hosts zeigen die Argumente der Aufrufe, die eine App macht, unverändert an. Wenn das Modell eine fertige Aufgabe einreicht, ist diese Übermittlung – **einschließlich der richtigen Antwort** – für alle sichtbar, die im Chatfenster das Protokoll der Tool-Aufrufe lesen.

MathTrail verbirgt das nicht und kann es auch nicht. Die Antwort ist verschlüsselt, damit sie sich nicht aus der Profildatei entnehmen oder zweimal ausgeben lässt; vor einer erwachsenen Person, die das Protokoll des eigenen Chats liest, ist sie nicht verborgen. Wenn die Aufgabe eine Überraschung bleiben soll, scrollen Sie nicht vor dem Kind durch das Protokoll der Aufrufe.

## Werbung, Profilbildung und Training

In MathTrail gibt es keine Werbung. Nichts über Ihr Kind wird verwendet, um ein Werbeprofil zu erstellen, und nichts wird verkauft oder an einen Datenhändler weitergegeben. MathTrail verwendet die Daten Ihres Kindes nicht, um irgendein Modell zu trainieren.

Was der Anbieter des Chats mit der Unterhaltung macht, regelt dessen eigene Richtlinie – dieselbe, die ohnehin schon für jede andere Nachricht gilt, die Sie in diesem Chat senden.

## Cookies

Diese Website setzt keine Cookies, betreibt keine Webanalyse und lädt nichts von anderen Domains.

Der Dienst setzt nur während der Anmeldung zwei Cookies, beide unbedingt erforderlich: eines, das den Beginn einer Anmeldung mit ihrem Abschluss verbindet, und eines, das sich merkt, welche Apps Sie bereits zugelassen haben, damit Sie nicht zweimal gefragt werden. Keines davon wird zum Tracking verwendet.

## Ihre Daten zurückholen und löschen

Alles liegt in Ihrem eigenen Drive, also haben Sie alles selbst in der Hand:

- **Exportieren** – öffnen Sie in Drive den Ordner `MathTrail` und laden Sie `mathtrail-profile.json` herunter. Diese Datei ist das ganze Profil.
- **Löschen** – löschen Sie die Datei. Leeren Sie den Papierkorb in Google Drive, um sie ganz zu entfernen. Vom Profil bleibt nirgendwo sonst etwas übrig.
- **Zugriff entziehen** – öffnen Sie in Ihrem Google-Konto die Seite mit den [verknüpften Apps](https://myaccount.google.com/linkedapps) und entfernen Sie den Zugriff von MathTrail auf Ihr Konto. Danach kann der Dienst auf nichts mehr zugreifen.
- **App entfernen** – trennen Sie die Verbindung zu MathTrail in den Einstellungen Ihres Chats.

Das Löschen der Datei und der Widerruf des Zugriffs sind voneinander unabhängig; wenn Sie beides tun, bleibt nichts zurück.

## Wer sonst beteiligt ist

- **Google** – die Anmeldung und Drive, wo Ihre Datei liegt ([Datenschutzerklärung von Google](https://policies.google.com/privacy)).
- **Google Cloud** – wo der Dienst läuft (Cloud Run) und wo die oben genannten Gesamtzahlen aufbewahrt werden (BigQuery) ([Google Cloud-Datenschutzhinweise](https://cloud.google.com/terms/cloud-privacy-notice)).
- **GitHub Pages** – von dort wird diese Website ausgeliefert; GitHub führt dafür eigene Zugriffsprotokolle ([Datenschutzerklärung von GitHub](https://docs.github.com/en/site-policy/privacy-policies/github-general-privacy-statement)).
- **Der Anbieter Ihres Chats** – Anthropic für Claude ([Datenschutzerklärung von Anthropic](https://www.anthropic.com/legal/privacy)), OpenAI für ChatGPT ([Datenschutzerklärung von OpenAI](https://openai.com/policies/privacy-policy/)).

Die Menschen, die MathTrail betreiben, können die Protokolle des Dienstes und die Gesamtzahlen in Google Cloud lesen, die das oben Beschriebene enthalten und nichts weiter. Weitere Beteiligte gibt es nicht.

## Änderungen

Ändert sich diese Datenschutzerklärung, ändert sich das Datum oben mit, und die vorherige Fassung bleibt in der öffentlichen Historie des Projekts erhalten.

## Sprachen

Diese Datenschutzerklärung wird in mehreren Sprachen veröffentlicht. Maßgeblich ist der englische Text; wo eine Übersetzung von ihm abweicht, gilt die englische Fassung.
