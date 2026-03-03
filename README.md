# RabbitMQ Messaging System - POC

Sistema de mensajería distribuido usando RabbitMQ con servidor y cliente en Go, diseñado para comunicación a través de internet con soporte TLS/SSL.

## 📋 Tabla de Contenidos

- [Arquitectura](#arquitectura)
- [Estructura del Proyecto](#estructura-del-proyecto)
- [Requisitos](#requisitos)
- [Quick Start](#quick-start)
- [Desarrollo Local](#desarrollo-local)
- [Docker Bake](#docker-bake)
- [Configuración TLS](#configuración-tls)
- [Deployment en VPS](#deployment-en-vps)
- [Taskfile Commands](#taskfile-commands)
- [API Reference](#api-reference)
- [Troubleshooting](#troubleshooting)

## 🏗️ Arquitectura

### Componentes

1. **Server (Publicador)**
   - API REST con Gin framework
   - Publica mensajes a RabbitMQ
   - Dockerizado con multi-stage builds
   - Hot-reload en desarrollo con Air

2. **Client (Consumidor)**
   - Goroutine controlado con loop infinito
   - Consume mensajes de RabbitMQ
   - Graceful shutdown implementado
   - Reconexión automática
   - Dockerizado con multi-stage builds

3. **RabbitMQ**
   - Cola de mensajería persistente
   - Soporte TLS/AMQPS (puerto 5671)
   - Management UI
   - Configuración para producción

### Flujo de Datos

```
Cliente HTTP → [Server API] → RabbitMQ Queue → [Client Consumer] → Procesamiento
   (POST)         (VPS)         (VPS)           (Localhost/VPS)
```

## 📁 Estructura del Proyecto

```
POC-AMQPS/
├── cmd/
│   ├── server/main.go              # Entry point del servidor
│   └── client/main.go              # Entry point del cliente
├── internal/
│   ├── config/config.go            # Configuración centralizada
│   ├── models/message.go           # Modelos de datos
│   ├── rabbitmq/
│   │   ├── publisher.go            # Lógica de publicación
│   │   └── consumer.go             # Lógica de consumo
│   ├── server/server.go            # HTTP server y handlers
│   └── client/client.go            # Consumer logic
├── certs/                          # Certificados TLS (gitignored)
│   ├── server-cert.pem
│   ├── server-key.pem
│   └── ca-cert.pem
├── docs/
│   └── TLS-SETUP.md                # Guía detallada de TLS
├── scripts/
│   └── test-api.sh                 # Script de pruebas
├── Dockerfile.server               # Multi-stage build para server
├── Dockerfile.client               # Multi-stage build para client
├── docker-compose.dev.yml          # Ambiente de desarrollo
├── docker-compose.prod.yml         # Ambiente de producción
├── docker-bake.hcl                 # Docker Bake configuration
├── Taskfile.yml                    # Task runner (reemplaza Makefile)
├── rabbitmq.conf                   # Configuración RabbitMQ con TLS
├── .air.server.toml                # Hot-reload config para server
├── .air.client.toml                # Hot-reload config para client
└── README.md                       # Este archivo
```

## 🔧 Requisitos

- **Go** 1.21 o superior
- **Docker** 20.10+ y **Docker Compose** v2+
- **Docker Buildx** (para Docker Bake)
- **Task** (Taskfile runner) - [Instalación](https://taskfile.dev/installation/)
- **jq** (opcional, para scripts de testing)

### Instalar Task

```bash
# macOS
brew install go-task/tap/go-task

# Linux
sh -c "$(curl --location https://taskfile.dev/install.sh)" -- -d -b /usr/local/bin

# Windows (PowerShell)
choco install go-task

# O usando Go
go install github.com/go-task/task/v3/cmd/task@latest
```

## 🚀 Quick Start

### Opción 1: Desarrollo Local con Docker (Recomendado)

```bash
# 1. Clonar y navegar al proyecto
cd POC-AMQPS

# 2. Descargar dependencias
task deps

# 3. Levantar ambiente completo (RabbitMQ + Server + Client)
task dev:up

# 4. Ver logs
task dev:logs

# 5. Probar API
task test:api

# 6. Detener todo
task dev:down
```

### Opción 2: Ejecutar sin Docker

```bash
# Terminal 1: Levantar RabbitMQ
docker compose up -d rabbitmq

# Terminal 2: Ejecutar servidor
task run:server

# Terminal 3: Ejecutar cliente
task run:client

# Terminal 4: Probar API
task test:api
```

## 💻 Desarrollo Local

### Ambiente de Desarrollo con Hot-Reload

El ambiente de desarrollo usa **Air** para hot-reload automático:

```bash
# Levantar todo el stack de desarrollo
task dev:up

# Los cambios en el código se recargan automáticamente
# Server: http://localhost:8086
# RabbitMQ Management: http://localhost:15672 (guest/guest)
```

### Estructura de Docker Compose Dev

```yaml
services:
  rabbitmq:   # Puerto 5672 (AMQP), 15672 (Management)
  server:     # Puerto 8086, hot-reload habilitado
  client:     # Consumidor con hot-reload
```

### Ver logs específicos

```bash
# Todos los servicios
task dev:logs

# Solo server
task dev:logs -- server

# Solo client
task dev:logs -- client

# Solo rabbitmq
task dev:logs -- rabbitmq
```

## 🐳 Docker Bake

Este proyecto usa **Docker Bake** con HCL para builds optimizados y multi-plataforma.

### ¿Qué es Docker Bake?

Docker Bake permite definir builds complejos en un archivo HCL, con soporte para:
- Multi-stage builds
- Multi-plataforma (amd64, arm64)
- Build cache optimization
- Targets y grupos
- Variables y configuración centralizada

### Configuración (`docker-bake.hcl`)

```hcl
group "dev" {
  targets = ["server-dev", "client-dev"]
}

group "prod" {
  targets = ["server-prod", "client-prod"]
}
```

### Comandos de Build

```bash
# Build de imágenes de desarrollo
task docker:build:dev

# Build de imágenes de producción
task docker:build:prod

# Build de todas las imágenes
task docker:build:all

# Build con plataformas específicas
docker buildx bake -f docker-bake.hcl --set *.platform=linux/amd64

# Build y push a registry
docker buildx bake -f docker-bake.hcl prod --push

# Build con variables custom
REGISTRY=myregistry.com VERSION=v1.0.0 task docker:build:prod
```

### Targets Disponibles

| Target | Descripción | Plataformas | Tags |
|--------|-------------|-------------|------|
| `server-dev` | Server con hot-reload | linux/amd64 | messaging-server:dev |
| `server-prod` | Server optimizado | linux/amd64, linux/arm64 | messaging-server:prod, :latest |
| `client-dev` | Client con hot-reload | linux/amd64 | messaging-client:dev |
| `client-prod` | Client optimizado | linux/amd64, linux/arm64 | messaging-client:prod, :latest |

### Variables de Docker Bake

```bash
# Cambiar registry
REGISTRY=ghcr.io/myuser task docker:build:prod

# Cambiar prefijo de imagen
IMAGE_PREFIX=myapp task docker:build:prod

# Cambiar versión
VERSION=1.2.3 task docker:build:prod
```

### Build Cache

Docker Bake está configurado para usar registry cache:

```bash
# El cache se guarda automáticamente en:
# - messaging-server:buildcache-dev
# - messaging-server:buildcache-prod
# - messaging-client:buildcache-dev
# - messaging-client:buildcache-prod

# Esto acelera builds subsecuentes significativamente
```

## 🔐 Configuración TLS

### Quick Setup con Certificados Self-Signed (Testing)

```bash
# Generar certificados de prueba
task tls:generate

# Levantar ambiente de producción con TLS
task prod:up
```

### Setup con Certificados DigiCert (Producción)

Para configuración completa con certificados DigiCert wildcard y dominio de GoDaddy, ver:

📖 **[docs/TLS-SETUP.md](docs/TLS-SETUP.md)** - Guía detallada de configuración TLS

**Resumen rápido:**

1. **Preparar certificados DigiCert:**
   ```bash
   mkdir -p certs
   cp /path/to/your_domain.crt certs/server-cert.pem
   cp /path/to/your_domain.key certs/server-key.pem
   cp /path/to/DigiCertCA.crt certs/ca-cert.pem
   chmod 600 certs/server-key.pem
   ```

2. **Configurar DNS en GoDaddy:**
   - Tipo A: `@` → IP de tu VPS
   - Tipo A: `rabbitmq` → IP de tu VPS
   - Tipo A: `api` → IP de tu VPS

3. **Configurar variables de entorno:**
   ```bash
   cp .env.prod .env
   # Editar .env con tus credenciales
   ```

4. **Levantar con TLS:**
   ```bash
   task docker:build:prod
   task prod:up
   ```

### Ubicación de Certificados

```
certs/
├── server-cert.pem    # Certificado del servidor (+ intermedios)
├── server-key.pem     # Clave privada
└── ca-cert.pem        # Certificado CA raíz
```

**⚠️ IMPORTANTE:** Los certificados están en `.gitignore` y NUNCA deben commitearse.

## 🌐 Deployment en VPS

### Opción 1: Docker Compose (Recomendado)

```bash
# 1. En tu máquina local, build las imágenes
task docker:build:prod

# 2. Guardar imágenes
docker save messaging-server:prod -o server.tar
docker save messaging-client:prod -o client.tar

# 3. Subir al VPS
scp server.tar client.tar root@your-vps:/tmp/
scp -r POC-AMQPS root@your-vps:/opt/

# 4. En el VPS, cargar imágenes
ssh root@your-vps
cd /opt/POC-AMQPS
docker load -i /tmp/server.tar
docker load -i /tmp/client.tar

# 5. Configurar certificados (ver docs/TLS-SETUP.md)
# Copiar tus certificados DigiCert a ./certs/

# 6. Configurar variables de entorno
cp .env.prod .env
nano .env

# Actualizar con tus ajustes de producción:
# RABBITMQ_URL=amqps://admin:password@pocmq.yourdomain.com:5671/
# API_DOMAIN=pocapi.yourdomain.com

# 7. Levantar servicios
docker compose -f docker-compose.prod.yml up -d

# 8. Verificar
docker compose -f docker-compose.prod.yml ps
docker compose -f docker-compose.prod.yml logs -f
```

### Opción 2: Binarios Nativos

```bash
# En el VPS
cd /opt/POC-AMQPS

# Build
task build:all

# Crear servicio systemd (ver documentación anterior)
# O ejecutar directamente
./bin/server &
./bin/client &
```

### Configurar Firewall

```bash
# Permitir puertos necesarios
sudo ufw allow 8086/tcp   # API Server
sudo ufw allow 5671/tcp   # RabbitMQ AMQPS
sudo ufw allow 15672/tcp  # RabbitMQ Management (opcional)
sudo ufw enable
```

### Nginx Reverse Proxy (Opcional)

```nginx
server {
    listen 80;
    server_name api.yourdomain.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name api.yourdomain.com;

    ssl_certificate /opt/POC-AMQPS/certs/server-cert.pem;
    ssl_certificate_key /opt/POC-AMQPS/certs/server-key.pem;

    location / {
        proxy_pass http://localhost:8086;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

## 📝 Taskfile Commands

### Gestión de Dependencias

```bash
task deps              # Descargar dependencias Go
```

### Build

```bash
task build:server      # Build solo server
task build:client      # Build solo client
task build:all         # Build ambos binarios
```

### Ejecución Local

```bash
task run:server        # Ejecutar server (sin Docker)
task run:client        # Ejecutar client (sin Docker)
```

### Docker - Desarrollo

```bash
task dev:up            # Levantar stack de desarrollo
task dev:down          # Detener stack de desarrollo
task dev:logs          # Ver logs (agregar -- service para filtrar)
```

### Docker - Producción

```bash
task prod:up           # Levantar stack de producción
task prod:down         # Detener stack de producción
task prod:logs         # Ver logs de producción
```

### Docker Bake

```bash
task docker:build              # Build con argumentos custom
task docker:build:dev          # Build imágenes dev
task docker:build:prod         # Build imágenes prod
task docker:build:all          # Build todas (multi-plataforma)
```

### Testing

```bash
task test              # Ejecutar tests Go
task test:api          # Probar API con script
```

### TLS

```bash
task tls:generate      # Generar certificados self-signed
```

### Limpieza

```bash
task clean             # Limpiar binarios y recursos Docker
```

### Flags y Argumentos

```bash
# Pasar argumentos a comandos
task dev:logs -- server              # Ver solo logs del server
task docker:build -- --no-cache      # Build sin cache
```

## 📚 API Reference

### POST /api/messages

Publica un mensaje en la cola de RabbitMQ.

**Request:**
```json
{
  "content": "string (required)",
  "metadata": {
    "source": "string (optional)",
    "type": "string (optional)"
  }
}
```

**Response (200 OK):**
```json
{
  "success": true,
  "message": "Message published successfully",
  "id": "uuid-del-mensaje"
}
```

**Ejemplo:**
```bash
curl -X POST http://localhost:8086/api/messages \
  -H "Content-Type: application/json" \
  -d '{
    "content": "Hola desde el API!",
    "metadata": {
      "source": "curl",
      "type": "test"
    }
  }'
```

### GET /health

Health check del servidor.

**Response (200 OK):**
```json
{
  "status": "healthy",
  "time": "2024-01-01T12:00:00Z"
}
```

## 🔍 Troubleshooting

### Task no encontrado

```bash
# Verificar instalación
task --version

# Reinstalar
go install github.com/go-task/task/v3/cmd/task@latest
```

### Docker Buildx no disponible

```bash
# Verificar
docker buildx version

# Instalar/habilitar
docker buildx install
docker buildx create --use
```

### Hot-reload no funciona en desarrollo

```bash
# Verificar que Air está instalado en el container
docker compose -f docker-compose.dev.yml exec server which air

# Rebuild containers
task dev:down
task docker:build:dev
task dev:up
```

### Certificados TLS no funcionan

```bash
# Verificar permisos
ls -la certs/

# Verificar certificado
openssl x509 -in certs/server-cert.pem -text -noout

# Verificar que RabbitMQ carga los certificados
task prod:logs -- rabbitmq
```

### Cliente no puede conectarse a RabbitMQ

```bash
# Verificar que RabbitMQ está corriendo
docker ps | grep rabbitmq

# Verificar logs
task dev:logs -- rabbitmq

# Test de conexión
telnet localhost 5672
```

### Build falla con Docker Bake

```bash
# Limpiar cache
docker buildx prune -f

# Build sin cache
docker buildx bake -f docker-bake.hcl --no-cache

# Verificar sintaxis HCL
docker buildx bake -f docker-bake.hcl --print
```

## 🎯 Mejores Prácticas

### Desarrollo

1. **Usa el ambiente Docker para desarrollo:**
   ```bash
   task dev:up
   ```
   - Hot-reload automático
   - Ambiente aislado
   - Fácil de limpiar

2. **Commits frecuentes:**
   - Los certificados están gitignored
   - `.env.prod` está gitignored
   - Usa `.env.example` como template

3. **Testing:**
   ```bash
   task test:api
   ```

### Producción

1. **Usa TLS/AMQPS siempre:**
   - Puerto 5671 para AMQPS
   - Certificados válidos de DigiCert
   - Ver `docs/TLS-SETUP.md`

2. **Variables de entorno seguras:**
   - No hardcodear passwords
   - Usar secrets management
   - Rotar credenciales regularmente

3. **Monitoreo:**
   - RabbitMQ Management UI
   - Logs centralizados
   - Health checks

4. **Backups:**
   - Datos de RabbitMQ en volumes
   - Backup regular de configuración

## 📖 Documentación Adicional

- **[docs/TLS-SETUP.md](docs/TLS-SETUP.md)** - Configuración detallada de TLS con DigiCert y GoDaddy
- **[Taskfile.yml](Taskfile.yml)** - Todos los comandos disponibles
- **[docker-bake.hcl](docker-bake.hcl)** - Configuración de Docker Bake

## 🤝 Contribuir

Este es un POC, pero las mejoras son bienvenidas:

1. Fork el proyecto
2. Crea una rama (`git checkout -b feature/amazing`)
3. Commit cambios (`git commit -m 'Add amazing feature'`)
4. Push (`git push origin feature/amazing`)
5. Abre un Pull Request

## 📄 Licencia

Este proyecto es un POC para propósitos de aprendizaje y testing.

---

**Hecho con ❤️ usando Go, RabbitMQ, Docker, y Task**
