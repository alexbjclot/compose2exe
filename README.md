# compose2exe

**[Español](#castellano) · [Català](#català) · [English](#english)**

---

## Español

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
| `build:` automático | ❌ | ✅ |
| Funciona en Windows | ❌ (bug) | ✅ |
| Modo embed | ✅ | ✅ |
| Parada ordenada | ❌ | ✅ |
| Preflight checks | ❌ | ✅ |

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
1. Verifica que Docker está instalado, el daemon está activo y los puertos están disponibles
2. Crea todas las redes Docker definidas en el compose
3. Descarga o carga las imágenes
4. Arranca los contenedores en orden de dependencia (`depends_on`)
5. Al pulsar `Ctrl+C`, para todos los contenedores en orden inverso

### Servicios con build:

Si tu `docker-compose.yml` usa `build:` en lugar de `image:`, compose2exe construye la imagen automáticamente y la embebe dentro del binario. No es necesario usar `--embed` ni subir la imagen a ningún registro.

```yaml
services:
  miapp:
    build: .   # ← se construye y embebe automáticamente
    ports:
      - "3000:3000"
```

> ⚠️ Si la imagen resultante supera los 2 GB, compose2exe mostrará un error indicando que debes publicar la imagen en un registro y usar `image:` en su lugar.

### Limitaciones

- El modo embed requiere que cada imagen pese menos de 2 GB (limitación de Go)
- Los volúmenes de fichero individual se detectan y se embeben automáticamente en el binario

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
| `build:` automàtic | ❌ | ✅ |
| Funciona a Windows | ❌ (bug) | ✅ |
| Mode embed | ✅ | ✅ |
| Aturada ordenada | ❌ | ✅ |
| Preflight checks | ❌ | ✅ |

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
1. Verifica que Docker està instal·lat, el daemon està actiu i els ports estan disponibles
2. Crea totes les xarxes Docker definides al compose
3. Descarrega o carrega les imatges
4. Arrenca els contenidors en ordre de dependència (`depends_on`)
5. En prémer `Ctrl+C`, atura tots els contenidors en ordre invers

### Serveis amb build:

Si el teu `docker-compose.yml` usa `build:` en lloc de `image:`, compose2exe construeix la imatge automàticament i l'embeu dins del binari. No cal usar `--embed` ni pujar la imatge a cap registre.

```yaml
services:
  mevaaplicacio:
    build: .   # ← es construeix i s'embeu automàticament
    ports:
      - "3000:3000"
```

> ⚠️ Si la imatge resultant supera els 2 GB, compose2exe mostrarà un error indicant que cal publicar la imatge a un registre i usar `image:` en el seu lloc.

### Limitacions

- El mode embed requereix que cada imatge pesi menys de 2 GB (limitació de Go)
- Els volums de fitxer individual es detecten i s'embeguen automàticament al binari

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
| Automatic `build:` | ❌ | ✅ |
| Works on Windows | ❌ (bug) | ✅ |
| Embed mode | ✅ | ✅ |
| Graceful shutdown | ❌ | ✅ |
| Preflight checks | ❌ | ✅ |

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
1. Verifies Docker is installed, the daemon is running and ports are available
2. Creates all Docker networks defined in the compose file
3. Pulls or loads images
4. Starts containers in dependency order (`depends_on`)
5. On `Ctrl+C`, stops all containers in reverse order

### Services with build:

If your `docker-compose.yml` uses `build:` instead of `image:`, compose2exe builds the image automatically and embeds it inside the binary. No need to use `--embed` or push the image to any registry.

```yaml
services:
  myapp:
    build: .   # ← built and embedded automatically
    ports:
      - "3000:3000"
```

> ⚠️ If the resulting image exceeds 2 GB, compose2exe will return a clear error message indicating that you should push the image to a registry and use `image:` instead.

### Limitations

- Embed mode requires each image to be under 2 GB (Go `//go:embed` limitation)
- File-based volume mounts are automatically detected and embedded inside the binary

---

## License

MIT
