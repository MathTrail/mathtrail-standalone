---
title: Aide
description: Comment connecter MathTrail, comment se déroule une leçon, que faire si quelque chose semble anormal, et comment nous joindre.
---

# Aide

Ce qu’il vous faut, comment connecter MathTrail, comment se déroule une leçon, que faire si quelque chose semble anormal, et comment nous joindre.

## Ce qu’il vous faut

- Votre propre compte Claude ou ChatGPT, et un compte Google. Claude est réservé aux adultes, et ChatGPT exige le consentement d’un parent pour les adolescents.
- Votre enfant n’a besoin ni de l’un ni de l’autre : l’enfant n’a pas de compte et ne se connecte nulle part.

MathTrail conserve le profil de votre enfant dans un seul fichier de votre propre Google Drive, et nulle part ailleurs.

## Connecter MathTrail à Claude

1. Dans Claude, ouvrez [Personnaliser → Connecteurs](https://claude.ai/customize/connectors).
2. Cliquez sur **+ Ajouter** et choisissez **Ajouter un connecteur personnalisé**.
3. Nommez-le MathTrail et collez l’adresse : `https://mcp.mathtrail.app/mcp`
4. Cliquez sur **Continuer**, puis encore sur **Continuer** : les paramètres d’authentification que Claude détecte n’ont pas besoin d’être modifiés. Cliquez ensuite sur **Ajouter**, puis sur **Connecter**, et connectez-vous avec Google. MathTrail demande à conserver un fichier contenant le profil de votre enfant dans votre Google Drive. Autorisez-le sur l’écran de Google : c’est dans ce fichier que sont enregistrés les progrès de votre enfant.
5. Dans un chat, cliquez sur **+** → **Connecteurs**, activez MathTrail et dites : « Faisons un problème MathTrail. »

Le plan Free de Claude permet un seul connecteur personnalisé. Un connecteur ajouté sur claude.ai ou dans Claude Desktop fonctionne aussi dans les applications mobiles de Claude, avec le même compte.

## Connecter MathTrail à ChatGPT

MathTrail n’est pas encore dans l’annuaire de ChatGPT. En attendant, le mode développeur de ChatGPT permet de l’ajouter par son adresse avec les forfaits payants, sur le Web : activez le mode développeur et ajoutez une application avec l’adresse ci-dessus, comme [le décrit OpenAI](https://developers.openai.com/plugins/deploy/connect-chatgpt). MathTrail est encore en phase de test dans ChatGPT.

## Comment se déroule une leçon

- **La première fois**, le chat vous demande de confirmer que vous êtes le parent de l’enfant ou son enseignant, puis il vous demande un pseudonyme pour votre enfant, jamais son vrai nom, et son niveau, de 1 à 6. Si vous le souhaitez, ajoutez ses centres d’intérêt et ce qui ne doit pas figurer dans les problèmes.
- **Les cinq premiers problèmes** permettent de voir où en est votre enfant. Ensuite, les problèmes s’ajustent aux réponses.
- **Vous écrivez, votre enfant appuie.** Demandez un problème dans le chat. Une carte apparaît et, au bout d’une minute ou deux, le problème s’y affiche avec cinq réponses proposées ; votre enfant répond en appuyant sur l’une d’elles. **L’ampoule** ouvre un indice. **Les deux flèches** demandent au chat le problème suivant et, une fois que votre enfant a répondu, **Un autre problème** fait de même. Le problème arrive plus bas, dans une nouvelle carte. Pendant que votre enfant cherche, le chat rédige à l’avance le problème suivant, si bien qu’il apparaît en général tout de suite dans la nouvelle carte.
- **Après une réponse**, la carte se transforme aussitôt en analyse de la réponse : l’erreur à l’origine d’une mauvaise réponse, un dessin de la solution et la solution étape par étape. Rien n’est envoyé au chat. Si votre enfant ne sait pas, dites-le au chat : une carte avec la solution arrive plus bas.
- **Progrès :** demandez « Où en est mon enfant ? » pour voir où se situe votre enfant dans chaque thème, ce qui a bougé sur la dernière semaine, un bilan et les erreurs qui reviennent. Le bouton **Modifier**, au même endroit, change le profil ; vous pouvez aussi le demander dans le chat.
- **Gardez fermé le journal des appels d’outils du chat pendant que votre enfant cherche :** il montre chaque problème avec sa réponse.

## Quand quelque chose ne va pas

- **Google vous renvoie vers une page de MathTrail, « Tick the box for Google Drive ».** La case qui permet à MathTrail de conserver son fichier est restée décochée. Cliquez sur **Back to Google** et cochez-la : sans ce fichier, les progrès n’ont nulle part où être enregistrés.
- **MathTrail vous demande de vous reconnecter.** Une connexion dure 90 jours au plus. Un compte Google admet aussi une centaine d’autorisations à la fois pour une même application : après de nombreuses connexions, la plus ancienne cesse donc de fonctionner. Dans les deux cas, il suffit de se reconnecter.
- **Claude refuse d’ajouter un autre connecteur.** Son plan Free n’a de place que pour un connecteur personnalisé : supprimez-en d’abord un autre.
- **La carte attend une minute ou deux.** Le modèle du chat rédige chaque problème, et MathTrail le vérifie avant que votre enfant le voie. C’est le temps que prend le premier problème d’une leçon ; les suivants sont rédigés pendant que votre enfant cherche, et arrivent dès qu’ils sont prêts. Avec la profondeur de réflexion habituelle du chat, l’attente reste courte : au réglage le plus profond, le modèle réfléchit à chaque problème pendant une minute ou plus. Un problème refusé par les vérifications est rédigé de nouveau, avec trois essais en tout ; si les trois échouent, le chat le dit, et vous pouvez en demander un autre.
- **La carte affiche « Fais un vœu pour demain ».** Chaque enfant reçoit jusqu’à 50 nouveaux problèmes par jour, et la journée se ferme aussi après dix demandes pour lesquelles aucun problème n’a pu être rédigé. La journée se termine à minuit, selon l’horloge de l’appareil sur lequel votre enfant répond aux cartes. Tant que votre enfant n’a répondu à aucune carte, elle se termine à minuit UTC. Les progrès et le dernier problème restent disponibles.
- **Claude place « Un autre problème » dans la zone de saisie, sous un avertissement qui invite à la prudence avant de l’envoyer.** C’est la demande de la carte qui vous attend : appuyez sur Entrée pour la transmettre.
- **Claude demande s’il faut autoriser MathTrail à enregistrer le profil.** L’enregistrement remplace ce que contenait le profil auparavant, c’est pourquoi Claude demande d’abord. Autorisez-le si c’est vous qui avez demandé la modification.
- **Aucune carte n’apparaît.** Certains chats n’affichent pas de cartes. Le chat donne alors le problème sous forme de texte, avec ses réponses proposées de A à E, et vous écrivez la lettre choisie par votre enfant.
- **Le profil est illisible, ou il se trouve dans la corbeille de Drive.** MathTrail le signale dans le chat et ne modifie rien. Restaurez le fichier depuis la corbeille, ou laissez MathTrail rétablir sa dernière version lisible ou créer un nouveau profil, comme l’explique [le guide sur les données de votre enfant](https://github.com/MathTrail/mathtrail-standalone/blob/main/docs/parents.md) (en anglais).
- **Un problème semble faux.** C’est possible : le programme vérifie ce qu’il peut, et le sens est vérifié par le modèle lui-même. Dites-nous ce que disait le problème.

## Les données de votre enfant

Tout ce que MathTrail sait de votre enfant tient dans un seul fichier, `mathtrail-profile.json`, dans un dossier `MathTrail` de votre Google Drive. Pour le télécharger, déconnecter MathTrail ou tout supprimer, voyez [le guide sur les données de votre enfant](https://github.com/MathTrail/mathtrail-standalone/blob/main/docs/parents.md) (en anglais). Ce qui est conservé, où et pendant combien de temps est expliqué dans la [politique de confidentialité](../privacy/).

## Nous contacter

- **Aide et questions sur les données de votre enfant :** écrivez à [altedtech.info@gmail.com](mailto:altedtech.info@gmail.com).
- **Une erreur dans un problème, ou une idée :** écrivez à la même adresse, ou ouvrez un ticket sur [GitHub](https://github.com/MathTrail/mathtrail-standalone/issues). Les tickets sont publics : n’y mettez ni nom, ni photographie, ni aucun autre renseignement sur votre enfant.
- **Un problème de sécurité :** signalez-le en privé via les [avis de sécurité de GitHub](https://github.com/MathTrail/mathtrail-standalone/security/advisories/new), comme le demande la [politique de sécurité](https://github.com/MathTrail/mathtrail-standalone/blob/main/SECURITY.md) du projet.

Nous lisons tout et répondons dès que possible, mais nous ne nous engageons sur aucun délai de réponse.
