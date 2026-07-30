# cmd/api — Servidor API Principal

> **[⬆ cmd](../)** — `api/cmd/api/`

Punto de entrada principal de la aplicación. `main.go` orquesta toda la inicialización del sistema y arranca el servidor HTTP con Fiber.

## Flujo de Arranque

El servidor ejecuta las siguientes fases en orden estricto durante el arranque:

```
┌─────────────────────────────────────────────────────────────────┐
│                        ARRANQUE (main.go)                       │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  1. config.InitConfig()          ← Carga variables de entorno   │
│           ↓                                                     │
│  2. database.InitDatabase()      ← Conexión a PostgreSQL        │
│           ↓                                                     │
│  3. database.SeedAdminUsers()    ← Crea usuarios admin si       │
│           ↓                               no existen            │
│  4. database.InitMigrations()    ← Ejecuta migraciones GORM     │
│           ↓                                                     │
│  5. s3.InitS3()                  ← Configura cliente S3/MinIO   │
│           ↓                                                     │
│  6. Fiber App + Middleware       ← Compresión, CORS, Logger     │
│           ↓                                                     │
│  7. router.InitRouter()          ← Monta todas las rutas        │
│           ↓                                                     │
│  8. app.Listen(":8080")          ← Arranca el servidor          │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

## Cadena de Middleware

El orden de ejecución de los middleware es crítico. Cada request atraviesa la cadena en el siguiente orden:

```
Request entrante
      │
      ▼
┌──────────────┐
│  Analytics   │  ← Registra métricas de la request
└──────┬───────┘
       │
       ▼
┌──────────────┐
│  RequestID   │  ← Asigna un ID único a cada request
└──────┬───────┘
       │
       ▼
┌──────────────┐
│     CORS     │  ← Maneja headers de Cross-Origin
└──────┬───────┘
       │
       ▼
┌──────────────┐
│ Rate Limiter │  ← Controla tasa de requests por IP
└──────┬───────┘
       │
       ▼
┌──────────────┐
│ Idempotency  │  ← Previene procesamiento duplicado
└──────┬───────┘
       │
       ▼
┌──────────────┐
│   Handler    │  ← Lógica de negocio / ruta específica
└──────┬───────┘
       │
       ▼
   Response
```

## Configuración del Servidor

| Parámetro | Valor |
|-----------|-------|
| Framework | Fiber v2 |
| Puerto | `:8080` |
| Compresión | Habilitada |
| CORS | Habilitado |
| Logger | Habilitado |
| Graceful Shutdown | 10 segundos de timeout |

## Graceful Shutdown

El servidor escucha las señales `SIGINT` y `SIGTERM` para realizar un apagado ordenado. Al recibir una señal:

1. Detiene de aceptar nuevas conexiones
2. Espera hasta 10 segundos a que las requests en vuelo terminen
3. Cierra la conexión a la base de datos
4. Finaliza el proceso

```go
// Patrón de shutdown graceful
quit := make(chan os.Signal, 1)
signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
<-quit

ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
app.ShutdownWithContext(ctx)
```

## Variables Globales

Las siguientes variables se inicializan en `main.go` y se pasan como dependencias a los repositorios y servicios:

- `db` — Instancia de la conexión a PostgreSQL (GORM)
- `S3` — Cliente de almacenamiento S3/MinIO
- `analyticsService` — Servicio de analíticas

> **Nota:** La base de datos y el cliente S3 se pasan como constructores de repositorio, no como globals. Esto facilita el testing y la separación de concerns.

**[⬆ Volver a cmd](../)**
