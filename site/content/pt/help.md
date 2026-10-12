---
title: Ajuda
description: Como conectar o MathTrail, como funciona uma aula, o que fazer quando algo parece errado e como falar conosco.
---

# Ajuda

Do que você precisa, como conectar o MathTrail, como funciona uma aula, o que fazer quando algo parece errado e como falar conosco.

## Do que você precisa

- Uma conta própria do Claude ou do ChatGPT e uma Conta do Google. O Claude é para pessoas adultas, e o ChatGPT exige o consentimento da mãe ou do pai para adolescentes.
- A criança não precisa de nenhuma das duas: ela não tem conta e não faz login em lugar nenhum.

O MathTrail guarda o perfil da criança em um único arquivo no seu próprio Google Drive, e em nenhum outro lugar.

## Conectar no Claude

1. No Claude, abra [Personalizar → Conectores](https://claude.ai/customize/connectors).
2. Clique em **+ Adicionar** e escolha **Adicionar conector personalizado**.
3. Digite MathTrail como nome e cole o endereço: `https://mcp.mathtrail.app/mcp`
4. Clique em **Continuar** e mais uma vez em **Continuar**: não é preciso mudar as configurações de autenticação que o Claude detecta. Depois clique em **Adicionar**, em seguida em **Conectar**, e faça login com o Google. O MathTrail pede permissão para guardar no seu Google Drive um único arquivo com o perfil da criança. Permita isso na tela do Google: é nesse arquivo que o progresso da criança fica salvo.
5. Em um chat, clique em **+** → **Conectores**, ative o MathTrail e diga: “Vamos fazer um problema do MathTrail.”

O plano Free do Claude permite um só conector personalizado. Um conector adicionado no claude.ai ou no Claude Desktop também funciona nos aplicativos móveis do Claude, com a mesma conta.

## Conectar no ChatGPT

O MathTrail ainda não está no diretório do ChatGPT. Até lá, o modo de desenvolvedor do ChatGPT permite adicioná-lo pelo endereço, nos planos pagos, na versão web: ative o modo de desenvolvedor e adicione um app com o endereço acima, como [a OpenAI descreve](https://developers.openai.com/plugins/deploy/connect-chatgpt). O MathTrail ainda está sendo testado no ChatGPT.

## Como funciona uma aula

- **Na primeira vez**, o chat pede que você confirme que é mãe, pai ou docente da criança e depois pede um pseudônimo para ela, nunca o nome verdadeiro, e o ano escolar, de 1 a 6. Se quiser, acrescente os interesses da criança e o que deve ficar de fora dos problemas.
- **Os cinco primeiros problemas** mostram em que ponto a criança está. Depois deles, os problemas se ajustam às respostas.
- **Você digita, e a criança toca.** Peça um problema no chat. Aparece um cartão e, em um ou dois minutos, o problema chega a ele com cinco alternativas; a criança responde tocando em uma delas. **A lâmpada** abre uma dica. **As duas setas** pedem ao chat o próximo problema e, depois que a criança responde, **Outro problema** faz o mesmo. O problema vem mais abaixo, em um cartão novo. Enquanto a criança resolve, o chat escreve o próximo problema com antecedência, então ele costuma aparecer na hora no cartão novo.
- **Depois de uma resposta**, o cartão se transforma imediatamente na análise da resposta: o erro por trás de uma resposta errada, um desenho da solução e a solução passo a passo. Nada vai para o chat. Se a criança não souber, diga isso no chat, e um cartão com a solução aparece mais abaixo.
- **Progresso:** pergunte “Como a criança está se saindo?” para ver em que ponto a criança está em cada tema, o que mudou na última semana, uma análise e os erros que se repetem. Ali, **Editar** muda o perfil; você também pode pedir isso no chat.
- **Mantenha fechado o registro de chamadas de ferramentas do chat enquanto a criança resolve:** ele mostra cada problema junto com a resposta.

## Quando algo dá errado

- **O Google manda você de volta a uma página do MathTrail, “Tick the box for Google Drive”.** A caixa que permite ao MathTrail guardar o arquivo dele ficou desmarcada. Clique em **Back to Google** e marque essa caixa: sem esse arquivo, não há onde guardar o progresso.
- **O MathTrail pede que você faça login de novo.** Um login dura no máximo 90 dias. Além disso, uma Conta do Google permite cerca de cem autorizações simultâneas a um mesmo app, então, depois de muitas conexões, a mais antiga para de funcionar. Fazer login de novo resolve as duas coisas.
- **O Claude não deixa adicionar outro conector.** O plano Free do Claude tem espaço para um só conector personalizado: remova outro antes.
- **O cartão fica esperando um ou dois minutos.** O modelo do chat escreve cada problema, e o MathTrail o verifica antes que a criança o veja. O primeiro problema de uma aula leva esse tempo; os seguintes são escritos enquanto a criança resolve e chegam na hora, assim que ficam prontos. A profundidade de raciocínio habitual do chat mantém a espera curta: na mais profunda, o modelo pensa em cada problema por um minuto ou mais. Um problema recusado pelas verificações é escrito de novo, com três tentativas no total; se as três falharem, o chat avisa, e você pode pedir outro.
- **O cartão diz “Faça um pedido para amanhã”.** Cada criança recebe até 50 problemas novos por dia, e o dia também se encerra depois de dez pedidos em que nenhum problema pôde ser escrito. O dia termina à meia-noite pelo relógio do dispositivo em que a criança responde no cartão. Enquanto a criança ainda não tiver respondido em nenhum cartão, o dia termina à meia-noite UTC. O progresso e o último problema continuam disponíveis.
- **O Claude coloca “Outro problema” na caixa de mensagem, sob um aviso que pede cautela antes de enviá-lo.** É o pedido do cartão esperando por você: pressione Enter para enviar.
- **O Claude pergunta se permite que o MathTrail salve o perfil.** Salvar substitui o que o perfil tinha antes, por isso o Claude pergunta primeiro. Permita se foi você quem pediu a mudança.
- **Nenhum cartão aparece.** Alguns chats não mostram cartões. Então o chat apresenta o problema em palavras, com as alternativas de A a E, e você digita a letra da resposta da criança.
- **O perfil não pode ser lido, ou está na lixeira do Drive.** O MathTrail avisa no chat e não muda nada. Restaure o arquivo da lixeira, ou deixe o MathTrail trazer de volta a última versão legível do perfil ou começar um perfil novo, como explica [o guia sobre os dados da criança](https://github.com/MathTrail/mathtrail-standalone/blob/main/docs/parents.md) (em inglês).
- **Um problema parece errado.** Pode ser que esteja: o programa verifica o que consegue, e o sentido é verificado pelo próprio modelo. Conte para nós o que o problema dizia.

## Os dados da criança

Tudo o que o MathTrail sabe sobre a criança está em um único arquivo, `mathtrail-profile.json`, em uma pasta `MathTrail` do seu Google Drive. Como baixá-lo, desconectar o MathTrail ou excluir tudo está explicado no [guia sobre os dados da criança](https://github.com/MathTrail/mathtrail-standalone/blob/main/docs/parents.md) (em inglês). O que é guardado, onde e por quanto tempo está na [política de privacidade](../privacy/).

## Fale conosco

- **Ajuda e perguntas sobre os dados da criança:** escreva para [altedtech.info@gmail.com](mailto:altedtech.info@gmail.com).
- **Um erro em um problema, ou uma ideia:** escreva para o mesmo endereço ou abra uma issue no [GitHub](https://github.com/MathTrail/mathtrail-standalone/issues). As issues são públicas, então não coloque nelas nenhum nome, fotografia ou outro dado da criança.
- **Uma falha de segurança:** comunique-a de forma privada pelos [avisos de segurança do GitHub](https://github.com/MathTrail/mathtrail-standalone/security/advisories/new), como pede a [política de segurança](https://github.com/MathTrail/mathtrail-standalone/blob/main/SECURITY.md) do projeto.

Lemos tudo e respondemos assim que possível, mas não prometemos nenhum prazo de resposta.
