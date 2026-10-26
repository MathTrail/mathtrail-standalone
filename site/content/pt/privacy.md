---
title: Política de privacidade
description: O que o MathTrail sabe sobre a criança, onde isso fica guardado e como recuperar ou excluir esses dados.
---

# Política de privacidade

Última atualização: 26 de outubro de 2026.

O MathTrail é um projeto gratuito e de código aberto, criado e mantido pela ALT EdTech. Envie perguntas sobre esta política ou sobre os dados da criança para [altedtech.info@gmail.com](mailto:altedtech.info@gmail.com). Como conectá-lo, e o que fazer quando algo parece errado, está na [página de ajuda](../help/).

## Em resumo

- **Não armazenamos nada.** O serviço não mantém nenhum banco de dados nem nenhum arquivo. De uma requisição para a seguinte, ele não se lembra de nada.
- **O perfil da criança fica no seu próprio Google Drive**, em um arquivo que você pode abrir, copiar ou excluir a qualquer momento.
- **Um pseudônimo, nunca um nome verdadeiro.** Nem data de nascimento, nem escola, nem fotografia.
- **Contamos as crianças sem saber quem elas são.** Uma linha do registro se refere a uma criança por um código que muda todo mês e não leva a ninguém; o que é guardado por mais de dois meses são totais.
- **Nada é vendido, e nada é usado para treinar um modelo.**
- **Excluir o arquivo exclui o perfil.** Não existe uma segunda cópia em lugar nenhum.

## O que é o MathTrail

O MathTrail é um app que você adiciona ao seu próprio Claude ou ChatGPT. Ele escolhe um tema e uma dificuldade para a criança, pede ao próprio modelo do chat que escreva um problema, verifica esse problema com um programa e o mostra, registra a resposta e explica o erro.

O MathTrail não roda nenhum modelo de linguagem próprio. O modelo que escreve os problemas é o do chat que você já usa, e a sua relação com o provedor desse chat é regida pelos termos e pela política de privacidade desse provedor, não por esta.

## Quem faz login

Apenas uma pessoa adulta — mãe, pai ou docente. A criança não tem conta nem login próprios. As plataformas de chat definem as próprias regras de idade: o Claude é para pessoas adultas, e o ChatGPT exige o consentimento da mãe ou do pai para adolescentes. O MathTrail não se destina a crianças que usam um chat por conta própria.

## Crianças

Os problemas do MathTrail são para crianças do 1º ao 6º ano, de seis a doze anos, aproximadamente. Ele foi pensado para ser usado com a pessoa adulta digitando no chat enquanto a criança, ao lado dela, responde tocando em uma alternativa no cartão. A criança não tem conta. O perfil contém o que a pessoa adulta informa sobre a criança, sob um pseudônimo, junto com as respostas da criança, em um arquivo da própria pessoa adulta; o que os registros do serviço guardam dessas respostas, e por quanto tempo, está descrito mais abaixo.

## O que é armazenado e onde

### No seu Google Drive

O perfil é um único arquivo JSON, `mathtrail-profile.json`, em uma pasta chamada `MathTrail` no seu próprio Drive. Você pode abri-lo, lê-lo, copiá-lo, movê-lo ou excluí-lo como qualquer outro arquivo, e a lixeira e o histórico de versões do Drive são seus.

Ele contém:

- um **pseudônimo** para a criança, o ano escolar, os interesses dela e o que você pediu para evitar;
- um identificador da criança, para que o perfil possa ser levado para uma edição futura;
- o idioma da interface, se você escolheu um;
- o tema ao qual as aulas se restringem, se você ou a criança escolheram um;
- o país onde você mora e, nos Estados Unidos, o seu estado, se você decidiu informá-los: eles estão ali só para contar quantas famílias há em cada país, e você pode deixá-los de fora;
- quais temas estão dominados, as pontuações em que a dificuldade se baseia e um resumo das respostas anteriores: por tema, quando ele foi dado pela última vez, quantas tentativas e acertos houve e quais armadilhas apareceram;
- em que ponto estavam as pontuações antes de cada uma das últimas respostas e no início de cada um dos últimos sete dias com respostas, para que o progresso possa mostrar o que mudou desde o último problema e ao longo da semana;
- um contador dos problemas aceitos hoje, que é como o limite do dia é controlado, e quanto o relógio do dispositivo em que a criança responde está adiantado ou atrasado em relação ao UTC: o cartão envia essa diferença a cada resposta, para que o dia termine à meia-noite no seu horário;
- o problema atual, com a resposta criptografada para que não possa ser lida no arquivo antes que a criança responda;
- o próximo problema, que o chat escreve com antecedência enquanto a criança resolve o atual, com a resposta criptografada da mesma forma, até que ele seja pedido.

Ele não contém o nome da criança, a data de nascimento, a escola, o endereço, uma fotografia nem qualquer outra coisa que a identifique. **O pseudônimo é o único nome envolvido, e ele nunca aparece dentro de um problema.**

### No servidor

Nada. O serviço é um único programa que responde a uma requisição e a esquece. Ele não tem banco de dados, nem disco, nem cache dos seus dados.

O seu login do Google também não é armazenado: ele é criptografado e colocado dentro do token que o serviço entrega ao seu chat, junto com o país de onde você fez login, como descrito abaixo. A chave fica no Google Secret Manager e nunca sai do servidor.

### Nos registros

O serviço registra contagens, não conteúdo: se um problema foi aceito ou rejeitado, por que foi rejeitado, quanto tempo levou, quantas tentativas foram necessárias, quantos bytes ocupava cada parte do problema que o chat entregou, se algum limite foi atingido e qual versão das instruções foi usada. As linhas sobre as chamadas que um problema exigiu compartilham um código aleatório criado para esse problema, que as liga entre si e não diz nada sobre a criança.

Para contar quantas crianças usam o MathTrail a cada dia, semana e mês sem saber quem é nenhuma delas, as linhas sobre um problema entregue, uma resposta e um tema dominado trazem, em conjunto:

- um **nome de contagem** para a criança: um código gerado, com uma chave do próprio serviço, a partir do identificador do perfil e do mês do calendário. Ele é o mesmo durante um mês e muda no mês seguinte, e nada sobre a criança — o identificador, o pseudônimo, qualquer outra coisa — pode ser deduzido dele;
- o chat (Claude, ChatGPT ou outro), o idioma do problema, o ano escolar, o mês em que o perfil foi criado e quantos temas estão dominados;
- na linha sobre um problema entregue, o **país** e, nos Estados Unidos, o **estado**, se você os informou no perfil, e o **país de onde você fez login**, cada um como um código curto;
- para uma resposta, o tema, o nível e a dificuldade, se ela estava certa, a armadilha por trás de uma resposta errada, se a dica foi usada, se foi um “não sei”, com que rapidez ela veio (em até um minuto, em até três ou depois), a **chance** de acerto que o serviço esperava para o problema, quem escolheu o problema (a regra do serviço, o chat ou você), se foi uma das primeiras respostas, as que encontram o nível em que a criança começa, e, depois delas, a faixa em que estava o número de respostas da criança, como de 6 a 20.

Como toda linha sobre as suas sessões, essas linhas também trazem um código opaco que representa a sua conta: gerado com a chave do próprio serviço a partir da sua Conta do Google, ele não é nem o seu ID do Google nem o seu e-mail, e muda no primeiro login depois que essa chave é substituída. Enquanto as linhas de dois meses estão guardadas ao mesmo tempo, é esse código que poderia ligar os dois nomes de contagem de uma criança. Ele também faz das linhas sobre respostas, enquanto estão guardadas, um registro das respostas da criança, uma a uma: certa ou errada, nunca a letra escolhida.

Essas linhas são guardadas por no máximo 62 dias. O que fica guardado depois disso são totais (quantas crianças, quantos problemas, em quais países): totais de cada dia, totais de cada semana e de cada mês separados por um critério de cada vez, como o país ou o ano escolar, e, para cada mês, as armadilhas em que caíram as respostas erradas, por tema e ano escolar, e como as respostas aos problemas escolhidos pelo serviço se saíram em comparação com a chance que ele esperava, por essa chance e pelo número de respostas que a criança já tinha dado. Eles não contêm nenhum identificador, de nenhum tipo. Um total conta um grupo de crianças, e um grupo pode ser uma única criança; um total menor que dez nunca é mostrado publicamente.

Os registros nunca contêm o pseudônimo, o texto de um problema, a letra de uma resposta, um endereço de e-mail, um token, o identificador da criança nem o endereço de onde veio uma requisição.

À parte, o Google Cloud Run mantém por padrão o registro de cada requisição a qualquer serviço que roda nele: o horário, o endereço de onde ela veio, como o navegador se identifica, a página que a enviou, o endereço completo solicitado, o código de status e o tamanho e a duração da resposta. Durante um login, o endereço solicitado traz, depois de um ponto de interrogação, o que o chat e o Google colocam nele: o código de login de uso único do Google, que só o serviço pode trocar, o domínio de uma conta do Google Workspace e o endereço de e-mail com o qual o chat sugere fazer login, se ele sugerir algum. O MathTrail faz o Google Cloud não guardar o registro de nenhuma requisição cujo endereço traga essa parte, então nada disso é mais guardado nesse registro. Os demais registros não contêm dados do perfil, nem texto de problemas, nem respostas, nem endereços de e-mail, e cada um é guardado por 30 dias. Os registros de logins feitos antes desta atualização ainda os contêm até completarem seus 30 dias.

### O país de onde você faz login

Quando você faz login, o Google manda o seu navegador de volta ao MathTrail e, nesse único momento, o serviço procura a qual país pertence o endereço do seu navegador, no banco de dados gratuito IP-to-Country da DB-IP — um arquivo que o serviço mantém no próprio servidor, então nada é enviado à DB-IP nem a mais ninguém. Só o código de duas letras do país é guardado, selado dentro do token que o seu chat mantém; o endereço em si não é armazenado nem registrado. Um endereço que não pertence a nenhum país, como o de uma rede privada, não indica nenhum país. As requisições do seu chat vêm dos servidores do chat, não de você, então o país continua sendo o do seu último login.

Você pode deixar esse país fora da contagem: marque “Não contar o país de onde faço login” no formulário do perfil, ou peça no chat. A partir daí, nenhuma linha escrita pelo serviço traz esse país, até você desmarcar a caixa.

Geolocalização por IP da [DB-IP](https://db-ip.com), sob a licença [Creative Commons Attribution 4.0](https://creativecommons.org/licenses/by/4.0/).

## Para que serve cada coisa

- **O perfil**: para escolher cada problema e a dificuldade dele, mostrar o progresso e controlar o limite do dia.
- **O login e o código da sua conta**: para acessar o seu único arquivo e nada mais, e para controlar os limites de cada conta.
- **Os países**: só para contar quantas famílias há em cada país. Você pode deixar os dois de fora.
- **Os registros**: para operar o serviço, encontrar falhas e ver como os problemas se saem.
- **Os totais**: para saber como o MathTrail é usado e quão bem os problemas dele se ajustam às crianças. Eles só são mostrados como totais de dez crianças ou mais, na página “Estudo” deste site e em pedidos de financiamento.

## Por quanto tempo cada coisa é guardada

- **O arquivo do perfil**: até você excluí-lo. Um arquivo excluído fica 30 dias na lixeira do Drive, a menos que você a esvazie. O Drive guarda por um tempo as versões anteriores do arquivo no histórico de versões, e o MathTrail pede ao Drive que mantenha para sempre a primeira versão de cada dia em que o perfil mudou, no máximo duzentas delas.
- **O token que o seu chat mantém**: 15 minutos, renovado por até 30 dias de cada vez e por no máximo 90 dias a partir do login.
- **Os dois cookies do login**: 10 minutos para o que liga as partes de um login, e 180 dias para o que lembra os apps que você aprovou.
- **Os registros do serviço e o registro de requisições do Cloud Run**: 30 dias.
- **A cópia das linhas a partir das quais as crianças são contadas**: 62 dias.
- **Rastros** do tempo que as requisições levaram: 30 dias. Eles cobrem cerca de uma requisição em cada dez, com o endereço, o navegador, o caminho e a consulta apagados.
- **Medições** do funcionamento do serviço, como quantas chamadas deram certo: 24 meses. Elas não contêm nada sobre nenhuma criança.
- **Os totais**: por anos. Eles não contêm nenhum identificador, de nenhum tipo.
- **Os registros de acesso deste site e as suas conversas**: pelo tempo que o GitHub e o provedor do seu chat os guardarem, segundo as próprias políticas.

## O login com o Google e o Drive

O login pede ao Google uma única permissão, **`drive.file`**. É a permissão do Drive de alcance limitado: ela dá acesso **somente aos arquivos que o próprio app criou**. O MathTrail não pode ver, listar nem abrir mais nada no seu Drive, nem mesmo os arquivos criados por você.

Para saber de quem era o login, o serviço então pergunta ao Google qual Conta do Google deu essa permissão. O Google informa o identificador da conta, não o seu e-mail, e o serviço guarda só o código que gera a partir desse identificador, descrito acima.

Vale a pena conhecer de antemão duas consequências, porque elas parecem falhas e não são:

- **O Google pede a sua permissão toda vez que você se conecta.** O serviço não guarda nenhuma credencial própria do Google de longa duração, então cada conexão precisa obter uma nova, e o Google mostra a tela de permissão toda vez.
- **Uma Conta do Google permite cerca de cem dessas autorizações ao mesmo tempo.** Depois de umas cem conexões, a mais antiga para de funcionar, sem aviso. Fazer login de novo resolve.

O token que o serviço emite para o seu chat vale por 15 minutos e é renovado automaticamente por até 30 dias de cada vez, e por no máximo 90 dias a partir do login. Depois disso, você faz login de novo.

## O que chega ao modelo do chat

Tudo o que uma ferramenta do MathTrail retorna é lido pelo modelo do chat e, por isso, chega ao provedor do seu chat junto com o resto da conversa. Ao longo de uma aula, isso é quase tudo o que o perfil contém:

- **o perfil e o progresso**: o pseudônimo, o ano escolar, os interesses, as habilidades deixadas de fora, as anotações que você escreveu, o idioma das aulas, o país e o estado, se você os informou, o tema ao qual as aulas se restringem, as pontuações e como elas mudaram, os temas dominados, as últimas respostas, os erros que se repetem e onde fica o arquivo do perfil no seu Drive, com o link dele;
- **o material para um problema**: o tema, a dificuldade, problemas de exemplo, os formatos a seguir, o ano escolar, os interesses e **as anotações livres que você escreveu sobre a criança**, que são o que permite ao modelo ajustar a redação ao nível certo;
- **uma resposta**: se ela estava certa, a alternativa correta, a armadilha por trás de uma resposta errada e a solução.

Escreva as anotações com isso em mente: elas vão para o provedor do chat junto com todo o resto. O identificador da criança não está em nenhum resultado, e o pseudônimo nunca entra no texto de um problema.

O texto do problema em si é escrito pelo modelo e volta passando pelas verificações do MathTrail.

## O que você pode ver e o que o chat mostra

Os hosts de chat mostram os argumentos brutos das chamadas que um app faz. Quando o modelo envia um problema pronto, esse envio — **incluindo a resposta correta** — fica visível para qualquer pessoa que leia o registro de chamadas de ferramentas na janela do chat.

O MathTrail não esconde isso, nem pode esconder. A resposta é criptografada para que não possa ser tirada do arquivo do perfil nem entregue duas vezes; ela não fica escondida de uma pessoa adulta que lê o registro do próprio chat. Se você quer que o problema seja uma surpresa, não percorra o registro de chamadas na frente da criança.

## Publicidade, criação de perfis e treinamento

Não há publicidade no MathTrail. Nada sobre a criança é usado para montar um perfil publicitário, e nada é vendido ou repassado a um corretor de dados. O MathTrail não usa os dados da criança para treinar nenhum modelo.

O que o provedor do chat faz com a conversa é regido pela política do próprio provedor — a mesma que já vale para todas as outras mensagens que você envia nesse chat.

## Cookies

Este site não grava nenhum cookie, não roda nenhuma ferramenta de análise e não carrega nada de outros domínios.

O serviço grava dois cookies, só durante o login, ambos estritamente necessários: um que liga o início de um login ao seu fim, e outro que lembra quais apps você já aprovou, para que a mesma pergunta não seja feita duas vezes. Nenhum deles é usado para rastreamento.

## Recuperar os seus dados e excluí-los

Tudo está no seu próprio Drive, então tudo está nas suas mãos:

- **Exportar**: abra a pasta `MathTrail` no Drive e baixe `mathtrail-profile.json`. Esse arquivo é o perfil inteiro.
- **Excluir**: exclua o arquivo. Esvazie a lixeira do Drive para removê-lo por completo. Nada do perfil sobrevive em outro lugar.
- **Remover o acesso**: abra a página de [apps vinculados](https://myaccount.google.com/linkedapps) da sua Conta do Google e remova o acesso do MathTrail à sua conta. A partir daí, o serviço não consegue acessar nada.
- **Remover o app**: desconecte o MathTrail nas configurações do seu próprio chat.

Excluir o arquivo e revogar o acesso são coisas independentes, e fazer as duas não deixa nada para trás.

## Quem mais participa

- **Google**: o login e o Drive, onde fica o seu arquivo ([política de privacidade do Google](https://policies.google.com/privacy)).
- **Google Cloud**: onde o serviço roda (Cloud Run) e onde ficam guardados os totais descritos acima (BigQuery) ([aviso de privacidade do Google Cloud](https://cloud.google.com/terms/cloud-privacy-notice)).
- **GitHub Pages**: de onde este site é servido; o GitHub mantém os próprios registros de acesso a ele ([declaração de privacidade do GitHub](https://docs.github.com/en/site-policy/privacy-policies/github-general-privacy-statement)).
- **O provedor do seu chat**: a Anthropic para o Claude ([política de privacidade da Anthropic](https://www.anthropic.com/legal/privacy)), a OpenAI para o ChatGPT ([política de privacidade da OpenAI](https://openai.com/policies/privacy-policy/)).

As pessoas que mantêm o MathTrail podem ler os registros do serviço e os totais no Google Cloud, que contêm o que está descrito acima e nada mais. Não há mais ninguém.

## Alterações

Se esta política mudar, a data no topo muda junto, e a redação anterior fica no histórico público do projeto.

## Idiomas

Esta política é publicada em vários idiomas. O texto em inglês é o que prevalece; quando uma tradução divergir dele, vale a versão em inglês.
