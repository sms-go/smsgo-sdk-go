# Publicando o `github.com/sms-go/smsgo-sdk-go` (pkg.go.dev)

Guia de release do SDK Go. **Não há registry nem conta**: em Go, "publicar" é criar uma **tag semver** no repositório público. O módulo é indexado no [pkg.go.dev](https://pkg.go.dev) e servido pelo *module proxy* (`proxy.golang.org`) na primeira vez que alguém pede a versão.

## Pré-requisitos (uma vez)

1. Repositório **público** em `github.com/sms-go/smsgo-sdk-go`.
2. O `module` em [`go.mod`](go.mod) precisa ser **exatamente** `github.com/sms-go/smsgo-sdk-go` (é o caminho de import; case-sensitive).

> O handle da org `sms-go` é todo minúsculo: o import é `go get github.com/sms-go/smsgo-sdk-go`. O comando `go` cuida do escaping no proxy automaticamente.

## Passo a passo do release

Em Go a versão **não fica em arquivo** — vem da tag. Só mantenha o [`CHANGELOG.md`](CHANGELOG.md).

1. `master` verde no CI (`go vet` + `go test -race`).
2. Atualize o [`CHANGELOG.md`](CHANGELOG.md).
3. Commit + push na `master`.
4. **Tag semver** (prefixo `v` obrigatório):
   ```bash
   git tag v0.3.0 && git push origin v0.3.0
   ```

O workflow [`/.github/workflows/release.yml`](.github/workflows/release.yml) dispara na tag `v*`, cria a Release no GitHub e **aquece o proxy** (`go get …@v0.3.0`), o que faz o pkg.go.dev indexar. Se preferir manual, veja abaixo.

### Aquecer o proxy manualmente (opcional)

```bash
cd "$(mktemp -d)" && go mod init warm
GOPROXY=https://proxy.golang.org GOFLAGS=-mod=mod go get github.com/sms-go/smsgo-sdk-go@v0.3.0
```
Ou simplesmente abrir `https://pkg.go.dev/github.com/sms-go/smsgo-sdk-go@v0.3.0` no navegador (o pkg.go.dev busca sob demanda).

## Verificação pós-publicação

```bash
cd "$(mktemp -d)" && go mod init t
go get github.com/sms-go/smsgo-sdk-go@v0.3.0   # deve resolver a versão
go doc github.com/sms-go/smsgo-sdk-go
```
Página: https://pkg.go.dev/github.com/sms-go/smsgo-sdk-go

## Notas

- Tags são **imutáveis** de fato: o proxy cacheia. Nunca mova uma tag já publicada — para corrigir, suba `v0.3.1`.
- Estamos em `v0.x` (pré-1.0): a API pode mudar entre minors e **não** há sufixo `/v2` no path. Isso só passa a valer a partir de `v2.0.0`.
- Mantenha a tag alinhada às versões dos outros SDKs (release unificado). Ver o guia central [`api/docs/sdks-publicacao.md`](../api/docs/sdks-publicacao.md).
