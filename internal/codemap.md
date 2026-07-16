# internal/

## Responsabilidade

Esta árvore agrega as capacidades privadas da aplicação **Rateio Luz**: entrada de
dados, validação, regra de rateio proporcional, formatação para o usuário,
interface desktop e persistência do histórico. Ela concentra o núcleo reutilizado
pelos pontos de entrada CLI e GUI, mantendo a regra de negócio independente da
forma de interação e da apresentação.

## Encapsulamento

Em Go, um pacote dentro de um diretório chamado `internal` só pode ser importado
por código pertencente à árvore cujo diretório pai contém esse `internal`. Neste
projeto, os pacotes abaixo são detalhes de implementação de `rateio-luz`: podem
ser usados pelos comandos e por outros pacotes do próprio projeto, mas não por
consumidores externos como API pública.

## Mapa dos subpacotes

- `data`: fronteira de entrada da CLI; escreve os prompts e lê as três linhas
  brutas, sem interpretar os valores.
- `validation`: normaliza vírgula ou ponto decimal, rejeita entradas inválidas ou
  negativas e converte consumos para `*big.Rat` e o valor monetário para centavos
  inteiros. Também exige consumo total maior que zero.
- `calculator`: implementa o rateio proporcional exato. Calcula consumo total e
  participações com `big.Rat`, arredonda o valor do primeiro consumidor ao
  centavo por *half-up* e atribui o restante ao segundo, preservando exatamente o
  total da conta.
- `presentation`: transforma valores de domínio em textos de interface: reais,
  percentuais, decimais com vírgula e consumos em kWh.
- `gui`: monta a aplicação desktop Fyne, aplica tema e identidade visual,
  coordena validação, cálculo, apresentação de erros/resultados e operações de
  histórico.
- `history`: persiste registros de rateio em CSV; cria diretório, arquivo e
  cabeçalho quando necessário, lista registros com validação estrutural e remove
  um registro por índice mediante substituição segura por arquivo temporário.

## Fluxo compartilhado

```text
CLI: data.ReadInput ─┐
                    ├─> validation.ParseAndValidate
GUI: campos Fyne ───┘              │
                                   v
                         calculator.Calculate
                                   │
                                   v
                    presentation (BRL, %, kWh, decimal)
                                   │
                         saída CLI ou widgets GUI
```

A CLI acrescenta apenas a captura terminal por `data`; a GUI obtém os mesmos
textos diretamente dos campos. Depois disso, ambas devem seguir a mesma cadeia:
validar e converter na fronteira, calcular somente com valores de domínio e
formatar apenas na saída. Assim, nenhuma interface replica parsing, aritmética ou
regras de arredondamento.

## Histórico da GUI

Após um cálculo bem-sucedido, a GUI cria um `history.Entry` com data/hora e um
retrato dos valores já formatados exibidos ao usuário. O registro só é persistido
quando o usuário escolhe **Salvar no histórico**. O `history.Store` usa
`historico.csv` em `rateio-luz` dentro do diretório de configuração do usuário,
com fallback para `~/.config` e, por fim, o diretório temporário.

A aba **Histórico** carrega os registros ao ser aberta ou atualizada, apresenta
os mesmos consumos, percentuais e valores do retrato salvo e permite excluir por
índice. Arquivo inexistente equivale a histórico vazio; erros de leitura,
gravação ou exclusão são convertidos pela GUI em estados e mensagens para o
usuário. O histórico é uma integração exclusiva da GUI e fica fora do fluxo de
cálculo: ele armazena a apresentação de um resultado já validado e calculado.
