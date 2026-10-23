# cf-genimage

MCP server (transport stdio) untuk generate image via Cloudflare Workers AI.
Ditulis dengan Go memakai arsitektur clean architecture.

**Role yang di-hardcode** (`internal/domain/image.go`):

| Parameter | Nilai |
| --------- | ----- |
| Model | `@cf/black-forest-labs/flux-1-schnell` |
| Steps | `4` |
| Format output | WebP lossy, quality 90 |

Ketiganya tidak diekspos sebagai parameter tool — klien hanya mengirim prompt.

## Struktur

```text
cmd/cf-genimage/main.go            composition root (wiring + run stdio)
internal/
  domain/image.go                  entity, port (interface), konstanta model/steps
  usecase/generate_image.go        business rule: validasi → generate → simpan
  config/config.go                 kredensial dari env / .env
  adapter/
    cloudflare/client.go           ImageGenerator via Workers AI REST API
    webpconv/generator.go          decorator: transcode hasil API ke WebP
    filestore/store.go             ImageStore ke filesystem lokal
    mcp/server.go                  tool generate_image + server stdio
```

## Tool

`generate_image`

| Param | Wajib | Deskripsi |
| ----- | ----- | --------- |
| `prompt` | ya | Deskripsi gambar |
| `output_path` | tidak | Path lokal untuk menyimpan WebP (cth. `/tmp/cat.webp`) |

Respons: teks ringkasan + gambar (`image` content, WebP base64) + structured output
(`mime_type`, `bytes`, `saved_path`, `model`, `steps`).

## Build & run

Butuh `libwebp-dev` (encoder memakai cgo):

```bash
cd cf-genimage
go build -o cf-genimage ./cmd/cf-genimage
```

Kredensial via environment (atau file `.env` di working directory):

```bash
export CF_ACCOUNT_ID=...
export CF_API_TOKEN=...
./cf-genimage   # serve MCP over stdio
```

## Konfigurasi klien (contoh)

```json
{
  "mcpServers": {
    "cf-genimage": {
      "command": "/home/doko/tmp/cfimagegen/cf-genimage/cf-genimage",
      "env": {
        "CF_ACCOUNT_ID": "...",
        "CF_API_TOKEN": "..."
      }
    }
  }
}
```

## Test & lint

```bash
go test -race ./...
go vet ./...
gofmt -l .
golangci-lint run ./...
```
