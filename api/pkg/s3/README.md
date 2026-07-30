# 📦 pkg/s3/

> **[⬆ pkg](../)** — `api/pkg/s3/`

## Descripción

Proporciona una **abstracción de almacenamiento de archivos** en S3 (AWS o MinIO para desarrollo local). Maneja la inicialización del cliente, subida de imágenes con validación, generación de nombres únicos y detección automática de content-type.

---

## 📁 Archivos

| Archivo | Funciones | Descripción |
|---------|-----------|-------------|
| `s3.go` | `InitS3()` | Inicializa el cliente AWS S3 desde variables de entorno |
| `upload.go` | `UploadImage()`, `UploadImageToS3()` | Sube archivos con validación y naming UUID |

---

## 🔌 s3.go — Inicialización del Cliente

`InitS3()` configura y retorna un cliente S3:

1. Lee las credenciales y configuración desde variables de entorno
2. Carga las credencialesAWS
3. Crea la sesión con la región especificada
4. Retorna un cliente `*s3.S3`

### Variables de entorno requeridas

| Variable | Descripción | Ejemplo |
|----------|-------------|---------|
| `AWS_REGION` | Región de AWS | `us-east-1` |
| `AWS_ACCESS_KEY` | Clave de acceso | `AKIA...` |
| `AWS_SECRET_KEY` | Clave secreta | `wJal...` |
| `AWS_BUCKET` | Nombre del bucket | `colpsi-uploads` |

### Soporte MinIO (desarrollo local)

Para desarrollo local con MinIO, configurar las variables apuntando a la instancia MinIO:

```
AWS_REGION=us-east-1
AWS_ACCESS_KEY=minioadmin
AWS_SECRET_KEY=minioadmin
AWS_BUCKET=colpsi-local
```

> MinIO implementa la API compatible con S3, por lo que el mismo cliente funciona sin cambios.

---

## 📤 upload.go — Subida de Archivos

### UploadImage(file, folder)

Función de alto nivel que:

1. **Valida el tamaño** del archivo (máximo 10 MB)
2. **Detecta el content-type** automáticamente
3. **Genera un nombre único** con UUID
4. **Sube a S3** con el prefijo de carpeta
5. **Retorna la URL** completa del archivo

```go
url, err := pkgs3.UploadImage(file, "avatars")
// → "https://bucket.s3.region.amazonaws.com/avatars/a1b2c3d4-..."
```

### UploadImageToS3(file, folder)

Función interna que realiza la subida efectiva:

1. Lee el contenido del archivo en memoria
2. Determina el `Content-Type` via `http.DetectContentType`
3. Genera nombre: `{folder}/{uuid}.{ext}`
4. Ejecuta `PutObject` con el bucket configurado
5. Construye y retorna la URL del objeto

### Límites y validaciones

| Regla | Valor | Detalle |
|-------|-------|---------|
| Tamaño máximo | 10 MB | Archivos más grandes son rechazados |
| Naming | UUID | Previene path traversal y colisiones |
| Content-Type | Detectado | Vía `http.DetectContentType` |
| Folder scope | Obligatorio | Cada upload debe especificar una carpeta |

---

## 🔒 Notas de Seguridad

### Credenciales
- Las credenciales de AWS **solo** se leen de variables de entorno.
- **Nunca** se hardcodean credenciales en el código fuente.
- En producción, se recomienda usar IAM Roles en vez de access keys.

### Integridad de archivos
- Los nombres UUID **previenen path traversal** (no se usa el nombre original del archivo).
- El content-type se detecta del contenido real, no del nombre del archivo.
- La validación de tamaño ocurre **antes** de intentar la subida a S3.

### Acceso
- Los objetos subidos deben tener configurado un bucket policy o ACL apropiado.
- Para assets públicos, usar bucket policy con `s3:GetObject`.

---

## 🏗️ Uso

```go
// Inicializar el cliente (una vez al inicio)
s3Client := pkgs3.InitS3()

// Subir imagen desde un handler
func UploadHandler(c *gin.Context) {
    file, err := c.FormFile("image")
    if err != nil {
        c.JSON(400, gin.H{"error": "Archivo requerido"})
        return
    }

    url, err := pkgs3.UploadImage(file, "posts")
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }

    c.JSON(200, gin.H{"url": url})
}
```

---

## 👥 Consumidores

- `internal/service/post_service.go` — Upload de imágenes para publicaciones
- `internal/service/psi_service.go` — Upload de imágenes de perfil

**[⬆ Volver a pkg](../)**
