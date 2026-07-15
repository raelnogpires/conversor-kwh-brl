# Rateio Luz

Rateio Luz divide uma conta de energia entre dois moradores de forma
proporcional ao consumo individual em kWh. O projeto oferece uma interface
gráfica para Linux e Windows, feita com [Fyne](https://fyne.io/), e mantém a
interface de linha de comando (CLI) original. As duas interfaces reutilizam a
mesma validação e a mesma regra de cálculo.

Não há cadastro, banco de dados, acesso à rede ou armazenamento de histórico.

## Como funciona

O usuário informa:

1. o consumo do morador 1 em kWh;
2. o consumo do morador 2 em kWh;
3. o valor total da conta em reais.

O Rateio Luz soma os dois consumos, calcula a participação proporcional de
cada morador e aplica essa proporção ao valor da conta. A interface mostra o
consumo total, os dois percentuais, os valores individuais em `R$` e uma
conferência de que os pagamentos recompõem exatamente o total informado.

Em forma simplificada:

```text
consumo total = consumo 1 + consumo 2
proporção 1 = consumo 1 / consumo total
proporção 2 = consumo 2 / consumo total
valor 1 = conta em centavos × proporção 1
valor 2 = conta em centavos - valor 1 arredondado
```

### Exemplo com dois moradores

Para os consumos de `105,5 kWh` e `67,2 kWh` e uma conta de `R$ 184,72`, o
resultado é:

```text
Consumo total: 172,7 kWh
Morador 1: 61,09% — R$ 112,84
Morador 2: 38,91% — R$ 71,88
Conferência: R$ 112,84 + R$ 71,88 = R$ 184,72
```

## Interfaces disponíveis

### Interface gráfica

A janela do Rateio Luz usa azul-marinho e branco-gelo, contém exemplos nos
campos, mensagens de validação em português e uma área organizada para o
resultado. O botão **Calcular rateio** é a ação principal e o botão **Limpar**
reinicia o formulário. É possível corrigir os valores e calcular novamente sem
reiniciar o programa.

A tecla `Enter` avança pelos campos e, no campo do valor da conta, executa o
cálculo. A navegação por teclado e o foco do campo inválido são preservados.

### Interface de linha de comando

A CLI continua disponível para uso em terminal e mantém o fluxo original de
três perguntas. Ela usa os mesmos pacotes de validação, cálculo e apresentação
da interface gráfica.

## Precisão monetária e arredondamento

Dinheiro não é armazenado em ponto flutuante. O valor da conta e os dois
pagamentos usam `int64` em centavos, o que evita artefatos binários como
`184,719999...` e torna exata a reconciliação final.

Os consumos e as proporções usam `math/big.Rat`. Assim, os decimais digitados
são representados como números racionais exatos e não é necessário duplicar ou
aproximar a fórmula em cada interface.

O valor exato do morador 1 é arredondado para o centavo mais próximo. Quando o
resultado fica exatamente no meio do centavo, o arredondamento é para cima
(*half-up*). O morador 2 recebe o restante da conta:

```text
valor do morador 2 = total em centavos - valor do morador 1
```

Essa política é determinística e garante, inclusive em contas com centavos e
consumos decimais, que os dois pagamentos sempre somem exatamente o valor
original. Como consequência, qualquer centavo residual do arredondamento é
atribuído ao morador 2.

## Validação de entrada

- campos vazios e textos não numéricos são rejeitados com uma mensagem clara;
- consumos negativos e contas negativas são inválidos;
- zero é aceito para apenas um dos moradores e também para o valor da conta;
- os dois consumos iguais a zero são inválidos, pois não existe proporção a
  calcular;
- vírgula ou ponto pode ser usado como único separador decimal;
- a conta aceita no máximo duas casas decimais;
- consumos podem ter mais casas decimais;
- espaços antes e depois do valor são ignorados;
- separadores de milhares, formatos mistos, notação científica, `NaN` e
  infinito não são aceitos;
- valores monetários fora do intervalo de centavos de `int64` são rejeitados.

Para evitar ambiguidade, digite `1234,56`, e não `1.234,56`. Formas incompletas
como `.5` e `5,` também são inválidas.

## Requisitos de desenvolvimento

- Go 1.24 ou uma versão posterior compatível com a diretiva do `go.mod`;
- um compilador C e os cabeçalhos gráficos exigidos pelo Fyne para executar ou
  compilar a GUI;
- Docker em execução apenas para o fluxo opcional de cross-build com
  `fyne-cross`.

Baixe as dependências Go na raiz do projeto:

```bash
go mod download
```

### Dependências do Fyne no Linux

Depois de instalar Go 1.24 ou posterior, em Debian, Ubuntu, Linux Mint e
Raspberry Pi OS instale as bibliotecas nativas de desenvolvimento com:

```bash
sudo apt-get install gcc libgl1-mesa-dev xorg-dev libwayland-dev libxkbcommon-dev
```

Consulte o [guia oficial de início do Fyne](https://docs.fyne.io/started/quick/)
para os pacotes equivalentes no Fedora, Arch Linux, openSUSE e outras
distribuições. Esses pacotes são necessários para desenvolver e compilar; o
usuário final não precisa instalar o ambiente Go.

## Executar localmente

Todos os comandos desta seção partem da raiz do repositório.

### Executar a CLI

```bash
go run ./cmd
```

### Executar a GUI

```bash
go run .
```

A primeira compilação do Fyne pode demorar mais porque inclui código C do
driver gráfico.

## Testes

Para executar toda a suíte sem depender de uma sessão gráfica ou de um servidor
de exibição, use a tag `ci` do Fyne:

```bash
go test -tags ci ./...
```

Os testes cobrem a validação, a CLI, a apresentação em português, a GUI e a
regra de domínio, incluindo consumos iguais e diferentes, kWh decimais, valores
BRL decimais, negativos, consumo total zero, arredondamento e reconciliação do
total.

## Compilar no Linux

Para gerar o executável nativo da interface gráfica:

```bash
mkdir -p dist
go build -trimpath -ldflags="-s -w" -o dist/rateio-luz .
./dist/rateio-luz
```

Para gerar também um executável da CLI:

```bash
go build -trimpath -o dist/rateio-luz-cli ./cmd
```

O comando compila para a arquitetura da máquina Linux atual e requer as
dependências de desenvolvimento do Fyne listadas acima.

## Compilar nativamente no Windows

O Fyne requer Go, um compilador C e os recursos gráficos do sistema. O fluxo
recomendado pelo projeto Fyne é instalar o
[MSYS2](https://www.msys2.org/), abrir o terminal **MSYS2 MinGW 64-bit** e
instalar as ferramentas:

```bash
pacman -Syu
pacman -S git mingw-w64-x86_64-toolchain mingw-w64-x86_64-go
```

Depois, no diretório raiz do projeto no Windows:

```bash
go mod download
mkdir -p dist
go build -trimpath -ldflags="-H=windowsgui -s -w" -o dist/rateio-luz.exe .
```

O parâmetro `-H=windowsgui` evita abrir um console junto com a janela. Para
compilar a CLI no Windows, preserve o console:

```bash
go build -trimpath -o dist/rateio-luz-cli.exe ./cmd
```

## Cross-build do Windows no Linux

Uma aplicação desktop Fyne usa CGO. Portanto, apenas executar
`GOOS=windows go build` no Linux não basta: também seria necessária uma cadeia
C completa para o alvo Windows. O fluxo documentado usa
[`fyne-cross`](https://github.com/fyne-io/fyne-cross), que fornece essa cadeia
em contêineres Docker.

Com o Docker instalado e o serviço em execução, na raiz do projeto:

```bash
go install github.com/fyne-io/fyne-cross@latest
fyne-cross windows \
  -engine docker \
  -arch=amd64 \
  -name "Rateio Luz" \
  -app-id com.raelpires.rateioluz \
  -app-version 1.0.0 \
  -icon assets/rateio-luz.png \
  .
```

Os artefatos ficam sob `fyne-cross/bin/windows-amd64` e
`fyne-cross/dist/windows-amd64`; o ZIP contém `Rateio Luz.exe` como aplicativo
gráfico, sem janela de console. Não acrescente `-release` nesse cross-build: no
Windows, esse modo é reservado à geração e assinatura de AppX em um host
Windows.

O mesmo ambiente em contêiner pode gerar o pacote Linux sem instalar os
cabeçalhos gráficos na máquina hospedeira:

```bash
fyne-cross linux \
  -engine docker \
  -arch=amd64 \
  -release \
  -name "Rateio Luz" \
  -app-id com.raelpires.rateioluz \
  -app-version 1.0.0 \
  -icon assets/rateio-luz.png \
  .
```

Os artefatos ficam em `fyne-cross/bin/linux-amd64` e
`fyne-cross/dist/linux-amd64`. Para uma distribuição pública assinada no
Windows, faça a etapa de release e assinatura em um ambiente Windows
configurado para esse fim.

## Ícone, metadados e empacotamento

A identidade da aplicação está centralizada em `FyneApp.toml`, na raiz do
projeto: nome **Rateio Luz**, ID
`com.raelpires.rateioluz`, versão e metadados de integração com o desktop. O
ícone minimalista de eletricidade e rateio está disponível em:

- `assets/rateio-luz.svg`: fonte vetorial editável;
- `assets/rateio-luz.png`: recurso RGBA de 512 × 512 pixels usado pela janela e
  pelas ferramentas de empacotamento;
- `assets/resources.go`: incorpora o PNG ao executável, sem caminhos externos
  em tempo de execução.

O Fyne converte o PNG para os recursos de ícone exigidos pelo destino. Para
instalar a ferramenta de empacotamento:

```bash
go install fyne.io/tools/cmd/fyne@latest
```

No Linux, gere um pacote com o ícone e os metadados executando na raiz:

```bash
fyne package -os linux -release
```

No Windows nativo, use o mesmo manifesto:

```bash
fyne package -os windows -release
```

O comando deve ser executado na raiz, onde o Fyne encontra o `FyneApp.toml`.
O pacote Linux inclui a integração de desktop; o pacote Windows incorpora o
ícone e os metadados no executável. Consulte também a
[documentação de empacotamento do Fyne](https://docs.fyne.io/started/packaging/).

## Estrutura do projeto

```text
main.go                      entrada da GUI (`go run .`)
FyneApp.toml                 nome, ID, versão, ícone e metadados desktop
assets/
  rateio-luz.svg             fonte vetorial do ícone
  rateio-luz.png             ícone 512 × 512 para janela e pacotes
  resources.go               recurso incorporado ao executável
cmd/
  main.go                    entrada da CLI
internal/
  calculator/                regra proporcional e reconciliação monetária
  data/                      leitura das três entradas da CLI
  gui/                       janela, tema, interação e mensagens da GUI
  presentation/              formatação compartilhada de BRL, kWh e percentual
  validation/                conversão e validação compartilhadas
```

Os pacotes `validation`, `calculator` e `presentation` não dependem da GUI. A
camada `gui` apenas coleta os dados, chama esses pacotes e exibe o resultado;
ela não contém uma cópia das fórmulas de rateio.

## Limitações atuais

- o rateio é fixo para exatamente dois moradores;
- não há histórico, exportação, impressão nem persistência dos cálculos;
- não são aceitos separadores de milhares nem notação científica;
- o morador 2 recebe a eventual diferença de um centavo do arredondamento;
- os pacotes gerados localmente não são assinados digitalmente e não constituem
  um instalador certificado para distribuição pública.

## Licença

Distribuído sob a licença MIT. Consulte [LICENSE.md](LICENSE.md).
Ela permite usar, modificar, redistribuir, sublicenciar e explorar
comercialmente o projeto, desde que o aviso de copyright e o texto da licença
sejam preservados nas cópias ou partes substanciais do software.
