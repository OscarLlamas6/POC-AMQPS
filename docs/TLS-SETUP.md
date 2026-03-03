# Configuración TLS/SSL para RabbitMQ

Esta guía explica cómo configurar TLS/SSL para RabbitMQ usando certificados de DigiCert (wildcard) y un dominio de GoDaddy.

## 📋 Prerequisitos

- Certificado wildcard de DigiCert (bundle completo)
- Dominio configurado en GoDaddy
- Acceso SSH a tu VPS

## 🔐 Estructura de Certificados DigiCert

Cuando descargas un certificado de DigiCert, típicamente recibes:

```
digicert-bundle/
├── your_domain.crt          # Certificado del servidor
├── DigiCertCA.crt           # Certificado intermedio
├── TrustedRoot.crt          # Certificado raíz
└── your_domain.key          # Clave privada (si la generaste con DigiCert)
```

## 📁 Ubicación de Certificados en el Proyecto

```
POC-AMQPS/
└── certs/
    ├── wildcard_bundle.crt  # Tu certificado bundle (cert + intermedios)
    └── wildcard.key         # Tu clave privada
```

**Nota:** Si tu bundle DigiCert ya incluye el certificado + intermedios (como en Nginx), solo necesitas estos 2 archivos.

## 🔧 Paso 1: Preparar los Certificados

### Opción A: Si tienes el bundle de DigiCert (Formato Nginx - Recomendado)

Si ya usas estos certificados en Nginx, simplemente cópialos:

```bash
# Crear directorio de certificados
mkdir -p certs

# 1. Copiar el bundle (certificado + intermedios)
cp /path/to/your_wildcard_bundle.crt certs/wildcard_bundle.crt

# 2. Copiar la clave privada
cp /path/to/your_wildcard.key certs/wildcard.key

# 3. Verificar permisos
chmod 600 certs/wildcard.key
chmod 644 certs/wildcard_bundle.crt

# 4. Verificar contenido del bundle
openssl x509 -in certs/wildcard_bundle.crt -text -noout
```

**Nota:** El archivo bundle ya contiene el certificado del servidor + certificados intermedios, por eso solo necesitas 2 archivos.

### Opción B: Si tienes certificados separados (sin bundle)

Si tienes los archivos separados y necesitas crear el bundle:

```bash
# Crear el bundle combinando certificado + intermedios
cat your_domain.crt DigiCertCA.crt > certs/wildcard_bundle.crt

# Copiar la clave privada
cp your_domain.key certs/wildcard.key

# Verificar permisos
chmod 600 certs/wildcard.key
chmod 644 certs/wildcard_bundle.crt
```

### Opción C: Generar certificados self-signed para testing

```bash
# Usar el comando de Taskfile (genera con nombres correctos)
task tls:generate

# O manualmente:
mkdir -p certs

# Generar certificado self-signed (bundle todo-en-uno)
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout certs/wildcard.key \
  -out certs/wildcard_bundle.crt \
  -subj "/C=US/ST=State/L=City/O=MyOrg/CN=*.yourdomain.com"

# Verificar permisos
chmod 600 certs/wildcard.key
chmod 644 certs/wildcard_bundle.crt
```

## 🌐 Paso 2: Configurar DNS en GoDaddy

### Configuración para Subdominios Específicos

Agrega los siguientes registros A en GoDaddy para tu dominio:

```
Tipo    Nombre      Valor              TTL      Descripción
A       pocmq       <IP-de-tu-VPS>    600      RabbitMQ AMQPS
A       pocapi      <IP-de-tu-VPS>    600      API REST
```

Esto creará:
- **RabbitMQ:** `pocmq.yourdomain.com:5671` (AMQPS)
- **API:** `pocapi.yourdomain.com:8080` (HTTP/HTTPS)
- **Management UI:** `pocmq.yourdomain.com:15672` (HTTP)

### Verificar DNS

Después de configurar, verifica que los registros se resuelvan:

```bash
# Verificar RabbitMQ
nslookup pocmq.yourdomain.com

# Verificar API
nslookup pocapi.yourdomain.com
```

**Nota:** Los cambios DNS pueden tardar hasta 48 horas en propagarse, pero típicamente toman 5-15 minutos.

## 🐰 Paso 3: Configurar RabbitMQ con TLS

El archivo `rabbitmq.conf` ya está configurado:

```conf
listeners.tcp.default = 5672
listeners.ssl.default = 5671

ssl_options.cacertfile = /etc/rabbitmq/ssl/ca-cert.pem
ssl_options.certfile   = /etc/rabbitmq/ssl/server-cert.pem
ssl_options.keyfile    = /etc/rabbitmq/ssl/server-key.pem
ssl_options.verify     = verify_peer
ssl_options.fail_if_no_peer_cert = false
```

## 🚀 Paso 4: Deployment en VPS

### Usando Docker Compose (Recomendado)

```bash
# 1. Subir el proyecto al VPS
scp -r POC-AMQPS root@your-vps-ip:/opt/

# 2. SSH al VPS
ssh root@your-vps-ip

# 3. Navegar al proyecto
cd /opt/POC-AMQPS

# 4. Copiar tus certificados DigiCert al directorio certs/
# (desde tu máquina local)
scp /path/to/your_wildcard_bundle.crt \
    root@your-vps-ip:/opt/POC-AMQPS/certs/wildcard_bundle.crt

scp /path/to/your_wildcard.key \
    root@your-vps-ip:/opt/POC-AMQPS/certs/wildcard.key

# 5. Configurar variables de entorno
cp .env.prod .env
nano .env

# Actualizar con tus credenciales:
# RABBITMQ_USER=admin
# RABBITMQ_PASS=tu_password_seguro
# RABBITMQ_URL=amqps://admin:tu_password_seguro@pocmq.yourdomain.com:5671/
# QUEUE_NAME=messages
# API_DOMAIN=pocapi.yourdomain.com

# 6. Build de imágenes con Docker Bake
docker buildx bake -f docker-bake.hcl prod

# 7. Levantar servicios
docker compose -f docker-compose.prod.yml up -d

# 8. Verificar logs
docker compose -f docker-compose.prod.yml logs -f
```

### Verificar la configuración TLS

```bash
# Test de conexión TLS
openssl s_client -connect your-vps-ip:5671

# Verificar certificado bundle
openssl x509 -in certs/wildcard_bundle.crt -text -noout

# Verificar que la clave privada coincide con el certificado
openssl x509 -noout -modulus -in certs/wildcard_bundle.crt | openssl md5
openssl rsa -noout -modulus -in certs/wildcard.key | openssl md5
# Los hash MD5 deben coincidir

# Test desde el cliente
curl -X POST https://pocapi.yourdomain.com/api/messages \
  -H "Content-Type: application/json" \
  -d '{"content":"Test TLS","metadata":{"source":"test","type":"tls"}}'
```

## 🔒 Paso 5: Configurar Cliente Local con TLS

### Actualizar variables de entorno del cliente

```bash
# En tu máquina local
export RABBITMQ_URL=amqps://admin:tu_password_seguro@pocmq.yourdomain.com:5671/
export QUEUE_NAME=messages

# Ejecutar cliente
go run cmd/client/main.go
```

### Si usas certificado self-signed (solo para testing)

Necesitarás copiar el `ca-cert.pem` a tu máquina local y configurar Go para confiar en él:

```bash
# Copiar CA cert
scp root@your-vps-ip:/opt/POC-AMQPS/certs/ca-cert.pem ./certs/

# Go automáticamente usará los certificados del sistema
# Para certificados custom, necesitas modificar el código del cliente
```

## 🔐 Seguridad y Mejores Prácticas

### 1. Permisos de Archivos

```bash
# En el VPS
chmod 600 certs/wildcard.key
chmod 644 certs/wildcard_bundle.crt
chown root:root certs/*
```

### 2. Firewall

```bash
# Cerrar puerto no-TLS en producción
sudo ufw deny 5672/tcp

# Permitir solo TLS
sudo ufw allow 5671/tcp
sudo ufw allow 443/tcp
```

### 3. Renovación de Certificados

DigiCert típicamente emite certificados con validez de 1-2 años. Configura recordatorios:

```bash
# Verificar fecha de expiración
openssl x509 -in certs/wildcard_bundle.crt -noout -enddate

# Crear cron job para verificar (opcional)
0 0 1 * * /usr/bin/openssl x509 -in /opt/POC-AMQPS/certs/wildcard_bundle.crt -noout -checkend 2592000 && echo "Certificate expiring soon" | mail -s "Cert Alert" admin@yourdomain.com
```

### 4. Variables de Entorno Seguras

```bash
# NUNCA commitear .env.prod
# Usar secrets management en producción (Vault, AWS Secrets Manager, etc.)

# Para Docker Swarm/Kubernetes:
docker secret create rabbitmq_user admin
docker secret create rabbitmq_pass <password>
```

## 🧪 Testing Local con TLS

Para probar TLS localmente antes de deployment:

```bash
# 1. Generar certificados self-signed
task tls:generate

# 2. Actualizar docker-compose.dev.yml para usar TLS
# (ya está configurado para montar ./certs)

# 3. Levantar ambiente de desarrollo
task dev:up

# 4. Probar conexión
curl -k https://localhost:8080/health
```

## 📊 Troubleshooting

### Error: "certificate verify failed"

```bash
# Verificar certificado bundle
openssl x509 -in certs/wildcard_bundle.crt -text -noout | grep -A2 "Validity"

# Verificar que el bundle contiene la cadena completa
openssl crl2pkcs7 -nocrl -certfile certs/wildcard_bundle.crt | \
  openssl pkcs7 -print_certs -noout
```

### Error: "connection refused" en puerto 5671

```bash
# Verificar que RabbitMQ está escuchando en 5671
docker exec rabbitmq-prod rabbitmq-diagnostics listeners

# Verificar logs
docker logs rabbitmq-prod
```

### Error: "tls: bad certificate"

```bash
# Verificar que el certificado coincide con la clave privada
openssl x509 -noout -modulus -in certs/wildcard_bundle.crt | openssl md5
openssl rsa -noout -modulus -in certs/wildcard.key | openssl md5

# Los hash MD5 deben coincidir
```

## 📚 Referencias

- [RabbitMQ TLS Support](https://www.rabbitmq.com/ssl.html)
- [DigiCert Certificate Installation](https://www.digicert.com/kb/ssl-certificate-installation.htm)
- [OpenSSL Commands](https://www.openssl.org/docs/man1.1.1/man1/)
