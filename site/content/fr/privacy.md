---
title: Politique de confidentialité
description: Ce que MathTrail sait de votre enfant, où c’est conservé, et comment le récupérer ou le supprimer.
---

# Politique de confidentialité

Dernière mise à jour : 26 octobre 2026.

MathTrail est un projet gratuit et open source, conçu et géré par ALT EdTech. Les questions sur cette politique ou sur les données de votre enfant sont à adresser à [altedtech.info@gmail.com](mailto:altedtech.info@gmail.com). La façon de le connecter, et ce qu’il faut faire si quelque chose semble anormal, sont expliquées sur la [page d’aide](../help/).

## En bref

- **Nous ne stockons rien.** Le service ne conserve aucune base de données ni aucun fichier. D’une requête à la suivante, il ne se souvient de rien.
- **Le profil de votre enfant se trouve dans votre propre Google Drive**, dans un fichier que vous pouvez ouvrir, copier ou supprimer à tout moment.
- **Un pseudonyme, jamais un vrai nom.** Ni date de naissance, ni école, ni photographie.
- **Nous comptons les enfants sans savoir qui ils sont.** Une ligne du journal désigne l’enfant par un code qui change chaque mois et ne permet de remonter à personne ; ce qui est conservé plus de deux mois, ce sont des totaux.
- **Rien n’est vendu, et rien ne sert à entraîner un modèle.**
- **Supprimer le fichier, c’est supprimer le profil.** Il n’en existe aucune autre copie, nulle part.

## Ce qu’est MathTrail

MathTrail est une application que vous ajoutez à votre propre Claude ou ChatGPT. Elle choisit un thème et une difficulté pour votre enfant, demande au modèle du chat lui-même de rédiger un problème, vérifie ce problème avec un programme, l’affiche, enregistre la réponse et explique l’erreur.

MathTrail ne fait tourner aucun modèle de langage qui lui soit propre. Le modèle qui rédige les problèmes est celui du chat que vous utilisez déjà, et votre relation avec le fournisseur de ce chat est régie par ses conditions et sa politique de confidentialité, et non par celle-ci.

## Qui se connecte

Un adulte uniquement : un parent ou un enseignant. L’enfant n’a ni compte ni connexion à son nom. Les plateformes de chat fixent leurs propres règles d’âge : Claude est réservé aux adultes, et ChatGPT exige le consentement d’un parent pour les adolescents. MathTrail ne s’adresse pas aux enfants qui utilisent un chat en autonomie.

## Les enfants

Les problèmes de MathTrail s’adressent aux enfants de la 1re à la 6e année, soit environ de six à douze ans. MathTrail est conçu pour être utilisé ainsi : l’adulte écrit dans le chat pendant que l’enfant, à ses côtés, répond en appuyant sur l’une des réponses proposées sur la carte. L’enfant n’a pas de compte. Le profil contient ce que l’adulte indique sur l’enfant, sous un pseudonyme, ainsi que les réponses de l’enfant, dans le propre fichier de l’adulte ; ce que les journaux du service conservent de ces réponses, et pendant combien de temps, est exposé plus bas.

## Ce qui est stocké, et où

### Dans votre Google Drive

Le profil est un unique fichier JSON, `mathtrail-profile.json`, placé dans un dossier nommé `MathTrail` de votre propre Drive. Vous pouvez l’ouvrir, le lire, le copier, le déplacer ou le supprimer comme n’importe quel autre fichier, et la corbeille et l’historique des versions de Drive vous appartiennent.

Il contient :

- un **pseudonyme** pour l’enfant, son niveau scolaire, ses centres d’intérêt et ce que vous avez demandé d’éviter ;
- un identifiant de l’enfant, pour que le profil puisse être transféré vers une future édition ;
- la langue de l’interface, si vous en avez choisi une ;
- le thème auquel les leçons s’en tiennent, si vous ou votre enfant en avez choisi un ;
- le pays où vous vivez et, aux États-Unis, votre État, si vous avez choisi de les indiquer : ils ne servent qu’à compter combien de familles compte chaque pays, et vous pouvez ne pas les renseigner ;
- les thèmes maîtrisés, les scores sur lesquels repose la difficulté, et un résumé des réponses passées : par thème, quand il a été donné pour la dernière fois, combien il y a eu de tentatives et de réussites, et quels pièges sont apparus ;
- où en étaient les scores avant chacune des dernières réponses, et au début de chacun des sept derniers jours avec des réponses, pour que l’affichage des progrès puisse montrer ce qui a bougé depuis le dernier problème et sur la semaine ;
- un compteur des problèmes acceptés aujourd’hui, qui sert à appliquer la limite de la journée, et l’avance ou le retard, par rapport à UTC, de l’horloge de l’appareil sur lequel votre enfant répond : la carte l’envoie avec chaque réponse, pour que la journée se termine à minuit chez vous ;
- le problème en cours, avec sa réponse chiffrée pour qu’on ne puisse pas la lire dans le fichier avant que l’enfant ait répondu ;
- le problème suivant, que le chat rédige à l’avance pendant que l’enfant résout le problème en cours, avec sa réponse chiffrée de la même façon, jusqu’à ce qu’on le demande.

Il ne contient ni le nom de l’enfant, ni sa date de naissance, ni son école, ni son adresse, ni sa photographie, ni quoi que ce soit d’autre qui permette de l’identifier. **Le pseudonyme est le seul nom en jeu, et il n’apparaît jamais dans un problème.**

### Sur le serveur

Rien. Le service est un programme unique qui répond à une requête puis l’oublie. Il n’a ni base de données, ni disque, ni cache de vos données.

Votre connexion Google n’est pas stockée non plus : elle est chiffrée et placée dans le jeton que le service remet à votre chat, avec le pays d’où provient votre connexion, comme décrit plus bas. La clé se trouve dans Google Secret Manager et ne quitte jamais le serveur.

### Dans les journaux

Le service enregistre des décomptes, pas du contenu : si un problème a été accepté ou refusé, pourquoi il a été refusé, combien de temps il a pris, combien d’essais ont été nécessaires, combien d’octets occupait chaque partie du problème remis par le chat, si une limite a été atteinte, et quelle version des instructions a été utilisée. Les lignes sur les appels qu’a nécessités un même problème partagent un code aléatoire créé pour ce problème, qui les relie entre elles et ne dit rien de l’enfant.

Pour compter combien d’enfants utilisent MathTrail chaque jour, chaque semaine et chaque mois sans savoir qui sont ces enfants, les lignes sur un problème donné, sur une réponse et sur un thème maîtrisé portent, à elles toutes :

- un **nom de comptage** pour l’enfant : un code calculé, avec une clé propre au service, à partir de l’identifiant du profil et du mois civil. Il reste le même pendant un mois et change le mois suivant, et rien sur l’enfant (ni l’identifiant, ni le pseudonyme, ni quoi que ce soit d’autre) ne peut en être déduit ;
- le chat (Claude, ChatGPT ou un autre), la langue du problème, le niveau scolaire, le mois où le profil a été créé, et le nombre de thèmes maîtrisés ;
- sur la ligne d’un problème donné, le **pays**, et l’**État** aux États-Unis, si vous les avez indiqués dans le profil, et le **pays de connexion**, chacun sous forme de code court ;
- pour une réponse, son thème, son niveau et sa difficulté, si elle était juste, le piège derrière une mauvaise réponse, si l’indice a été utilisé, s’il s’agissait de « je ne sais pas », le temps qu’elle a mis à venir (moins d’une minute, moins de trois, ou plus), la **probabilité de réussite** que le service prévoyait pour son problème, qui a choisi le problème (la règle du service, le chat ou vous), s’il s’agissait de l’une des premières réponses, celles qui déterminent le niveau de départ de l’enfant, et, après celles-ci, la tranche dans laquelle se trouvait le nombre de réponses de l’enfant, par exemple de 6 à 20.

Comme toutes les lignes qui concernent vos sessions, ces lignes portent aussi un code opaque qui représente votre compte : calculé avec la clé propre au service à partir de votre compte Google, ce n’est ni votre identifiant Google ni votre adresse e-mail, et il change à la première connexion qui suit le remplacement de cette clé. Tant que les lignes de deux mois sont conservées en même temps, c’est ce code qui pourrait relier entre eux les deux noms de comptage de l’enfant. Il fait aussi des lignes sur les réponses, tant qu’elles sont conservées, un relevé des réponses de votre enfant, une par une : bonne ou mauvaise, jamais la lettre choisie.

Ces lignes sont conservées 62 jours au plus. Ce qui est conservé ensuite, ce sont des totaux (combien d’enfants, combien de problèmes, dans quels pays) : des totaux pour chaque jour, des totaux pour chaque semaine et chaque mois ventilés selon un seul critère à la fois, comme le pays ou le niveau scolaire, et, pour chaque mois, les pièges dans lesquels sont tombées les mauvaises réponses, par thème et par niveau scolaire, ainsi que la façon dont les réponses aux problèmes choisis par le service se sont situées par rapport à la probabilité qu’il prévoyait, selon cette probabilité et selon le nombre de réponses que l’enfant avait données. Ils ne contiennent aucun identifiant, de quelque sorte que ce soit. Un total compte un groupe d’enfants, et un groupe peut n’avoir qu’un seul membre ; un total inférieur à dix n’est jamais montré publiquement.

Les journaux ne contiennent jamais ni le pseudonyme, ni le texte d’un problème, ni la lettre d’une réponse, ni l’adresse e-mail, ni le jeton, ni l’identifiant de l’enfant, ni l’adresse d’où provient une requête.

Par ailleurs, Google Cloud Run conserve par défaut l’enregistrement standard de chaque requête adressée à n’importe quel service qu’il héberge : l’heure, l’adresse d’où elle provient, la façon dont le navigateur s’identifie, la page qui l’a envoyée, l’adresse complète demandée, le statut, ainsi que la taille et la durée de la réponse. Pendant une connexion, l’adresse demandée porte, après un point d’interrogation, ce que le chat et Google y ont mis : le code de connexion à usage unique de Google, que seul le service peut échanger, le domaine d’un compte Google Workspace, et l’adresse e-mail avec laquelle le chat suggère de se connecter, s’il en suggère une. MathTrail fait en sorte que Google Cloud ne conserve l’enregistrement d’aucune requête dont l’adresse porte une telle partie, si bien que rien de tout cela n’est plus conservé dans cet enregistrement. Les autres enregistrements ne contiennent aucune donnée du profil, aucun texte de problème, aucune réponse et aucune adresse e-mail, et chacun est conservé 30 jours. Les enregistrements des connexions faits avant cette mise à jour les contiennent encore jusqu’à la fin de leurs 30 jours.

### Le pays depuis lequel vous vous connectez

Lorsque vous vous connectez, Google renvoie votre navigateur vers MathTrail et, à ce seul moment, le service cherche à quel pays appartient l’adresse de votre navigateur, dans la base de données gratuite IP-to-Country de DB-IP : un fichier que le service garde sur son propre serveur, si bien que rien n’est envoyé à DB-IP ni à qui que ce soit d’autre. Seul le code à deux lettres du pays est conservé, scellé dans le jeton que détient votre chat ; l’adresse elle-même n’est ni stockée ni consignée nulle part. Une adresse qui n’appartient à aucun pays, comme celle d’un réseau privé, ne donne aucun pays. Les requêtes de votre chat viennent de ses propres serveurs, pas de vous : le pays reste donc celui de votre dernière connexion.

Vous pouvez faire exclure ce pays de ce qui est compté : cochez « Ne pas compter le pays depuis lequel je me connecte » dans le formulaire du profil, ou demandez-le dans le chat. Dès lors, aucune ligne écrite par le service ne le porte, jusqu’à ce que vous décochiez la case.

Géolocalisation IP par [DB-IP](https://db-ip.com), sous licence [Creative Commons Attribution 4.0](https://creativecommons.org/licenses/by/4.0/).

## À quoi sert chaque élément

- **Le profil** : pour choisir chaque problème et sa difficulté, montrer les progrès et appliquer la limite de la journée.
- **La connexion et le code de votre compte** : pour accéder à votre unique fichier et à rien d’autre, et pour appliquer les limites de chaque compte.
- **Les pays** : uniquement pour compter combien de familles compte chaque pays. Vous pouvez les écarter tous les deux.
- **Les journaux** : pour faire fonctionner le service, trouver les anomalies et voir comment se comportent les problèmes.
- **Les totaux** : pour savoir comment MathTrail est utilisé et dans quelle mesure ses problèmes conviennent aux enfants. Ils ne sont montrés que sous forme de totaux d’au moins dix enfants, sur la page « Étude » de ce site et dans des demandes de subventions.

## Combien de temps chaque élément est conservé

- **Le fichier du profil** : jusqu’à ce que vous le supprimiez. Un fichier supprimé reste 30 jours dans la corbeille de Drive, sauf si vous la videz. Drive conserve un certain temps les versions antérieures du fichier dans son historique des versions, et MathTrail lui demande de conserver indéfiniment la première version de chaque jour où le profil a changé, deux cents au plus.
- **Le jeton que détient votre chat** : 15 minutes, renouvelé pour au plus 30 jours à chaque fois, et au plus 90 jours à compter de la connexion.
- **Les deux cookies de la connexion** : 10 minutes pour celui qui relie les étapes d’une connexion, et 180 jours pour celui qui mémorise les applications que vous avez approuvées.
- **Les journaux du service, et l’enregistrement des requêtes de Cloud Run** : 30 jours.
- **La copie des lignes à partir desquelles on compte les enfants** : 62 jours.
- **Les traces** de la durée des requêtes : 30 jours. Elles couvrent environ une requête sur dix, avec l’adresse, le navigateur, le chemin et la chaîne de requête effacés.
- **Les mesures** du fonctionnement du service, comme le nombre d’appels réussis : 24 mois. Elles ne contiennent rien sur les enfants.
- **Les totaux** : pendant des années. Ils ne contiennent aucun identifiant, de quelque sorte que ce soit.
- **Les journaux d’accès de ce site et vos conversations** : aussi longtemps que GitHub et le fournisseur de votre chat les conservent, selon leurs propres politiques.

## Connexion Google et Drive

La connexion demande à Google une seule autorisation, **`drive.file`**. C’est l’autorisation Drive restreinte : elle donne accès **uniquement aux fichiers que cette application a elle-même créés**. MathTrail ne peut ni voir, ni lister, ni ouvrir quoi que ce soit d’autre dans votre Drive, y compris les fichiers que vous avez créés vous-même.

Pour savoir à qui appartient la connexion, le service demande ensuite à Google quel compte Google a donné cette autorisation. Google lui communique l’identifiant du compte, pas votre adresse e-mail, et le service ne conserve que le code qu’il calcule à partir de cet identifiant, décrit plus haut.

Deux conséquences méritent d’être connues à l’avance, car elles ressemblent à des pannes alors qu’elles n’en sont pas :

- **Google vous demande votre autorisation à chaque connexion.** Le service ne garde aucun identifiant de connexion Google de longue durée qui lui soit propre : chaque connexion doit donc en obtenir un nouveau, et Google affiche à chaque fois son écran d’autorisation.
- **Un compte Google admet environ une centaine de ces autorisations à la fois.** Au bout d’une centaine de connexions environ, la plus ancienne cesse de fonctionner, sans avertissement. Il suffit de se reconnecter.

Le jeton que le service émet pour votre chat est valable 15 minutes et se renouvelle automatiquement pour au plus 30 jours à chaque fois, et au plus 90 jours à compter de la connexion. Ensuite, il faut se reconnecter.

## Ce qui parvient au modèle du chat

Tout ce que renvoie un outil de MathTrail est lu par le modèle du chat, et parvient donc au fournisseur de votre chat avec le reste de la conversation. Au fil d’une leçon, cela représente presque tout ce que contient le profil :

- **le profil et les progrès** : le pseudonyme, le niveau scolaire, les centres d’intérêt, les notions exclues, les remarques que vous avez écrites, la langue des leçons, le pays et l’État si vous les avez indiqués, le thème auquel les leçons s’en tiennent, les scores et leur évolution, les thèmes maîtrisés, les dernières réponses, les erreurs qui reviennent, et l’emplacement du fichier du profil dans votre Drive, avec son lien ;
- **les éléments d’un problème** : le thème, la difficulté, des exemples de problèmes, les formats à suivre, le niveau scolaire, les centres d’intérêt et **les remarques libres que vous avez écrites sur l’enfant**, qui permettent au modèle d’ajuster la formulation au bon niveau ;
- **une réponse** : si elle était juste, la bonne réponse, le piège derrière une mauvaise réponse et la solution.

Rédigez les remarques en conséquence : elles partent chez le fournisseur du chat avec tout le reste. L’identifiant de l’enfant ne figure dans aucun résultat, et le pseudonyme n’entre jamais dans le texte d’un problème.

Le texte du problème lui-même est rédigé par le modèle et revient en passant par les vérifications de MathTrail.

## Ce que vous pouvez voir, et ce que le chat affiche

Les hôtes de chat affichent les arguments bruts des appels que fait une application. Quand le modèle soumet un problème terminé, cette soumission, **réponse correcte comprise**, est visible pour quiconque lit le journal des appels d’outils dans la fenêtre du chat.

MathTrail ne le cache pas et ne peut pas le cacher. La réponse est chiffrée pour qu’on ne puisse ni l’extraire du fichier du profil ni la remettre deux fois ; elle n’est pas cachée à l’adulte qui lit le journal de son propre chat. Si vous voulez que le problème reste une surprise, ne faites pas défiler le journal des appels devant l’enfant.

## Publicité, profilage et entraînement

MathTrail ne contient aucune publicité. Rien de ce qui concerne votre enfant ne sert à construire un profil publicitaire, et rien n’est vendu ni cédé à un courtier en données. MathTrail n’utilise pas les données de votre enfant pour entraîner quelque modèle que ce soit.

Ce que le fournisseur du chat fait de la conversation relève de sa propre politique, celle-là même qui s’applique déjà à tous les autres messages que vous envoyez dans ce chat.

## Cookies

Ce site ne dépose aucun cookie, ne fait aucune mesure d’audience et ne charge rien depuis d’autres domaines.

Le service dépose deux cookies, uniquement pendant la connexion, tous deux strictement nécessaires : l’un relie le début d’une connexion à sa fin, l’autre mémorise les applications que vous avez déjà approuvées, pour qu’on ne vous pose pas deux fois la question. Aucun des deux ne sert au suivi.

## Récupérer vos données, et les supprimer

Tout se trouve dans votre propre Drive, donc tout est entre vos mains :

- **Exporter** : ouvrez le dossier `MathTrail` dans Drive et téléchargez `mathtrail-profile.json`. Ce fichier, c’est le profil tout entier.
- **Supprimer** : supprimez le fichier. Videz la corbeille de Drive pour l’effacer complètement. Rien du profil ne subsiste ailleurs.
- **Couper l’accès** : rendez-vous sur la page [Applications associées](https://myaccount.google.com/linkedapps) de votre compte Google et supprimez l’accès de MathTrail à votre compte. Le service ne peut alors plus accéder à rien.
- **Retirer l’application** : déconnectez MathTrail dans les paramètres de votre chat.

Supprimer le fichier et révoquer l’accès sont deux actions indépendantes, et si vous faites les deux, il ne reste rien.

## Qui d’autre intervient

- **Google** : la connexion et Drive, où se trouve votre fichier ([règles de confidentialité de Google](https://policies.google.com/privacy)).
- **Google Cloud** : là où le service s’exécute (Cloud Run), et où sont conservés les totaux décrits plus haut (BigQuery) ([avis de confidentialité de Google Cloud](https://cloud.google.com/terms/cloud-privacy-notice)).
- **GitHub Pages** : d’où ce site est servi ; GitHub tient pour lui ses propres journaux d’accès ([déclaration de confidentialité de GitHub](https://docs.github.com/en/site-policy/privacy-policies/github-general-privacy-statement)).
- **Le fournisseur de votre chat** : Anthropic pour Claude ([politique de confidentialité d’Anthropic](https://www.anthropic.com/legal/privacy)), OpenAI pour ChatGPT ([politique de confidentialité d’OpenAI](https://openai.com/policies/privacy-policy/)).

Les personnes qui gèrent MathTrail peuvent lire les journaux du service et les totaux dans Google Cloud, qui contiennent ce qui est décrit plus haut, et rien de plus. Il n’y a personne d’autre.

## Modifications

Si cette politique change, la date en haut de page change avec elle, et la rédaction précédente reste dans l’historique public du projet.

## Langues

Cette politique est publiée en plusieurs langues. Le texte anglais fait foi ; en cas de divergence entre une traduction et ce texte, c’est la version anglaise qui s’applique.
