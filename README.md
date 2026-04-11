# compose2exe

**[Castellano](#castellano) · [Català](#català) · [English](#english)**

---

## Castellano

Convierte un `docker-compose.yml` en un único ejecutable binario que cualquiera puede ejecutar.

Inspirado en [docker2exe](https://github.com/rzane/docker2exe), pero con soporte completo de Docker Compose y binarios Windows que realmente funcionan.

### ¿Por qué compose2exe?

| Característica | docker2exe | compose2exe |
|---|---|---|
| Contenedor único | ✅ | ✅ |
| Múltiples contenedores | ❌ | ✅ |
| Redes | ❌ | ✅ |
| `depends_on` | ❌ | ✅ |
| Variables de entorno | ❌ | ✅ |
| Volúmenes | ❌ | ✅ |
| Funciona en Windows | ❌ (bug) | ✅ |
| Modo embed | ✅ | ✅ |
| Parada ordenada | ❌ | ✅ |

### Requisitos

**Dispositivo de compilación:** Go 1.21+, Docker, gzip, make

**Dispositivo de ejecución:** Docker (Engine o Desktop)

### Instalación

Descarga un binario desde la [página de releases](../../releases) o compila desde el código fuente:

```bash
git clone https://github.com/alexbjclot/compose2exe.git
cd compose2exe
make build
```

### Uso

```bash
# Modo normal (hace docker pull al ejecutarse)
compose2exe --compose docker-compose.yml --name miapp --target linux/amd64

# Modo embed (imágenes dentro del binario, sin internet en ejecución)
compose2exe --compose docker-compose.yml --name miapp --embed --target linux/amd64
```

Al ejecutar el binario generado:
1. Crea todas las redes Docker definidas en el compose
2. Descarga o carga las imágenes
3. Arranca los contenedores en orden de dependencia (`depends_on`)
4. Al pulsar `Ctrl+C`, para todos los contenedores en orden inverso

### Limitaciones

- El modo embed requiere que cada imagen pese menos de 2 GB (limitación de Go)
- Los volúmenes de fichero individual pueden requerir que el fichero exista en la máquina destino
- Los contextos de build (`build:`) no están soportados — usa `image:` en su lugar

---

## Català

Converteix un `docker-compose.yml` en un únic binari executable que qualsevol pot executar.

Inspirat en [docker2exe](https://github.com/rzane/docker2exe), però amb suport complet de Docker Compose i binaris Windows que realment funcionen.

### Per què compose2exe?

| Característica | docker2exe | compose2exe |
|---|---|---|
| Contenidor únic | ✅ | ✅ |
| Múltiples contenidors | ❌ | ✅ |
| Xarxes | ❌ | ✅ |
| `depends_on` | ❌ | ✅ |
| Variables d'entorn | ❌ | ✅ |
| Volums | ❌ | ✅ |
| Funciona a Windows | ❌ (bug) | ✅ |
| Mode embed | ✅ | ✅ |
| Aturada ordenada | ❌ | ✅ |

### Requisits

**Dispositiu de compilació:** Go 1.21+, Docker, gzip, make

**Dispositiu d'execució:** Docker (Engine o Desktop)

### Instal·lació

Descarrega un binari des de la [pàgina de releases](../../releases) o compila des del codi font:

```bash
git clone https://github.com/alexbjclot/compose2exe.git
cd compose2exe
make build
```

### Ús

```bash
# Mode normal (fa docker pull en executar-se)
compose2exe --compose docker-compose.yml --name mevaaplicacio --target linux/amd64

# Mode embed (imatges dins el binari, sense internet en execució)
compose2exe --compose docker-compose.yml --name mevaaplicacio --embed --target linux/amd64
```

En executar el binari generat:
1. Crea totes les xarxes Docker definides al compose
2. Descarrega o carrega les imatges
3. Arrenca els contenidors en ordre de dependència (`depends_on`)
4. En prémer `Ctrl+C`, atura tots els contenidors en ordre invers

### Limitacions

- El mode embed requereix que cada imatge pesi menys de 2 GB (limitació de Go)
- Els volums de fitxer individual poden requerir que el fitxer existeixi a la màquina destí
- Els contextos de build (`build:`) no estan suportats — usa `image:` en el seu lloc

---

## English

Convert a `docker-compose.yml` into a single executable binary that anyone can run.

Inspired by [docker2exe](https://github.com/rzane/docker2exe), but with full Docker Compose support and working Windows binaries.

### Why compose2exe?

| Feature | docker2exe | compose2exe |
|---|---|---|
| Single container | ✅ | ✅ |
| Multiple containers | ❌ | ✅ |
| Networks | ❌ | ✅ |
| `depends_on` | ❌ | ✅ |
| Environment variables | ❌ | ✅ |
| Volumes | ❌ | ✅ |
| Works on Windows | ❌ (bug) | ✅ |
| Embed mode | ✅ | ✅ |
| Graceful shutdown | ❌ | ✅ |

### Requirements

**Building device:** Go 1.21+, Docker, gzip, make

**Executing device:** Docker (Engine or Desktop)

### Installation

Download a binary from the [releases page](../../releases) or build from source:

```bash
git clone https://github.com/alexbjclot/compose2exe.git
cd compose2exe
make build
```

### Usage

```bash
# Pull mode (pulls images from Docker Hub at runtime)
compose2exe --compose docker-compose.yml --name myapp --target linux/amd64

# Embed mode (images bundled inside the binary, no internet needed at runtime)
compose2exe --compose docker-compose.yml --name myapp --embed --target linux/amd64
```

When running the generated binary:
1. Creates all Docker networks defined in the compose file
2. Pulls or loads images
3. Starts containers in dependency order (`depends_on`)
4. On `Ctrl+C`, stops all containers in reverse order

### Limitations

- Embed mode requires each image to be under 2 GB (Go `//go:embed` limitation)
- File-based volume mounts may require the file to exist on the target machine
- Build contexts (`build:`) are not supported — use `image:` instead

---

## License

MIT
