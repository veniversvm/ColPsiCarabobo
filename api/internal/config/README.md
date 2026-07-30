# ⚙️ Configuración (config/)

> **[⬆ internal](../)** — `api/internal/config/`

Configuración centralizada del sistema. Usa **singleton pattern** para garantizar una sola instancia de configuración en toda la aplicación.

## Patrón Singleton

```go
// Al inicio de la app (main.go)
cfg := config.InitConfig()

// En cualquier otro archivo
cfg := config.Config{} // Accede a la misma instancia
```

`InitConfig()` se llama una sola vez al arrancar el servidor. Todas las demás llamadas acceden a la misma instancia. Lee variables de entorno del SO (no usa archivos `.env`).

## Lectura de Variables

```go
func getEnv(key, defaultValue string) string {
    if value, exists := os.LookupEnv(key); exists {
        return value
    }
    return defaultValue
}
```

Si la variable de entorno no existe, se usa el valor por defecto. Esto permite que la app funcione sin configurar todas las variables (útil en desarrollo).

## Estructura Config

### 🖥️ Server

| Campo         | Variable de Entorno | Default       | Descripción                    |
|---------------|---------------------|---------------|--------------------------------|
| `Port`        | `PORT`              | `8080`        | Puerto del servidor HTTP       |
| `Environment` | `ENVIRONMENT`       | `development` | Entorno de ejecución           |

### 🗄️ Database

| Campo      | Variable de Entorno  | Default       | Descripción              |
|------------|----------------------|---------------|--------------------------|
| `Host`     | `DB_HOST`            | `localhost`   | Host de PostgreSQL       |
| `Port`     | `DB_PORT`            | `5432`        | Puerto de PostgreSQL     |
| `User`     | `DB_USER`            | `postgres`    | Usuario de DB            |
| `Password` | `DB_PASSWORD`        | *(requerido)* | Contraseña de DB         |
| `DBName`   | `DB_NAME`            | `colpsi_db`   | Nombre de la base datos  |
| `DBDriver` | `DB_DRIVER`          | `postgres`    | Driver de DB             |

### ☁️ Storage (AWS S3)

| Campo            | Variable de Entorno   | Default       | Descripción                    |
|------------------|-----------------------|---------------|--------------------------------|
| `AWSRegion`      | `AWS_REGION`          | `us-east-1`   | Región de AWS                  |
| `AWSAccessKey`   | `AWS_ACCESS_KEY`      | *(requerido)* | Access Key de AWS              |
| `AWSSecretKey`   | `AWS_SECRET_KEY`      | *(requerido)* | Secret Key de AWS              |
| `AWSBucket`      | `AWS_BUCKET`          | *(requerido)* | Nombre del bucket S3           |

### 📧 Email (SMTP)

| Campo      | Variable de Entorno | Default       | Descripción              |
|------------|---------------------|---------------|--------------------------|
| `SMTPHost` | `SMTP_HOST`         | *(requerido)* | Host del servidor SMTP   |
| `SMTPPort` | `SMTP_PORT`         | `587`         | Puerto SMTP              |
| `SMTPUser` | `SMTP_USER`         | *(requerido)* | Usuario SMTP             |
| `SMTPPass` | `SMTP_PASS`         | *(requerido)* | Contraseña SMTP          |
| `SMTPFrom` | `SMTP_FROM`         | *(requerido)* | Email remitente          |

### 🌍 Environment

| Campo   | Variable de Entorno | Default       | Descripción                   |
|---------|---------------------|---------------|-------------------------------|
| `GoEnv` | `GOENV`             | `development` | Entorno Go (development/prod) |

## Variables de Entorno Requeridas

Las siguientes variables **deben** estar configuradas para que la app funcione en producción:

```bash
# Database (REQUIRED)
DB_PASSWORD=tu_password_aqui

# AWS S3 (REQUIRED)
AWS_ACCESS_KEY=tu_access_key
AWS_SECRET_KEY=tu_secret_key
AWS_BUCKET=tu_bucket_name

# Email SMTP (REQUIRED)
SMTP_HOST=smtp.gmail.com
SMTP_USER=tu_email@gmail.com
SMTP_PASS=tu_password_app
SMTP_FROM=tu_email@gmail.com
```

## Variables Opcionales (con defaults)

```bash
# Server
PORT=8080
ENVIRONMENT=development

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_NAME=colpsi_db
DB_DRIVER=postgres

# AWS
AWS_REGION=us-east-1

# Email
SMTP_PORT=587

# Environment
GOENV=development
```

## Configuración en Docker

```yaml
# docker-compose.yml
services:
  api:
    environment:
      - DB_HOST=postgres
      - DB_PASSWORD=secret
      - AWS_ACCESS_KEY=${AWS_ACCESS_KEY}
      - AWS_SECRET_KEY=${AWS_SECRET_KEY}
      - AWS_BUCKET=colpsi-production
      - SMTP_HOST=smtp.gmail.com
      - SMTP_USER=admin@colpsi.com
      - SMTP_PASS=${SMTP_PASS}
      - SMTP_FROM=admin@colpsi.com
      - ENVIRONMENT=production
      - GOENV=production
```

## Uso

```go
package main

import "ColPsiCarabobo/api/internal/config"

func main() {
    cfg := config.InitConfig()

    // Usar en cualquier parte
    fmt.Println(cfg.Server.Port)           // "8080"
    fmt.Println(cfg.Database.Host)         // "localhost"
    fmt.Println(cfg.Storage.AWSRegion)     // "us-east-1"
    fmt.Println(cfg.Email.SMTPHost)        // "smtp.gmail.com"
}
```

## Archivos

| Archivo      | Descripción                              |
|--------------|------------------------------------------|
| `config.go`  | Struct Config + InitConfig() + getEnv()  |

**[⬆ Volver a internal](../)**
