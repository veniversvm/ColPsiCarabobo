# FIX-15 Report — Pool de conexiones DB + GOMAXPROCS + connect_timeout

| Campo | Valor |
|-------|-------|
| **Fix ID** | FIX-15 |
| **Hallazgo original** | HIGH-08, LOW-15 |
| **Archivos modificados** | `pkg/database/postgres.go`, `cmd/api/main.go` |
| **Fecha de implementación** | 2026-07-25 |
| **Estado** | Completado |

---

## Problema

### Pool de conexiones (postgres.go)
No se configuraban `MaxOpenConns`, `MaxIdleConns`, `ConnMaxLifetime`. Defaults de Go (`MaxOpenConns=0` = ilimitado) podían saturar PgBouncer (`DEFAULT_POOL_SIZE=20`).

### DSN sin timeout (postgres.go)
El DSN no tenía `connect_timeout`. Si la DB no respondía, la app se colgaba indefinidamente.

### GOMAXPROCS hardcodeado (main.go)
`runtime.GOMAXPROCS(2)` limitaba la app a 2 núcleos sin importar el servidor.

---

## Corrección

### postgres.go — Pool + timeout

```go
// DSN con timeout:
dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable connect_timeout=5", ...)

// Pool de conexiones:
sqlDB, err := db.DB()
if err != nil {
    return nil, fmt.Errorf("error al obtener pool de conexiones: %w", err)
}
sqlDB.SetMaxOpenConns(25)
sqlDB.SetMaxIdleConns(10)
sqlDB.SetConnMaxLifetime(5 * time.Minute)
```

### main.go — GOMAXPROCS adaptativo

```go
numCPU := runtime.NumCPU()
if numCPU > 2 {
    runtime.GOMAXPROCS(numCPU / 2)
} else {
    runtime.GOMAXPROCS(numCPU)
}
```

| Servidor | Núcleos | GOMAXPROCS | Libres |
|----------|---------|------------|--------|
| Local | 2 | 2 | 0 |
| AWS mediano | 4 | 2 | 2 |
| AWS grande | 8 | 4 | 4 |

---

## Testing

- `go build ./...` — Pass
- `go vet ./...` — Pass
- Middleware tests: 15/15 Pass

---

## Archivos relacionados

| Archivo | Relación |
|---------|----------|
| `pkg/database/postgres.go` | Pool + timeout |
| `cmd/api/main.go:43-48` | GOMAXPROCS |
| `docker-compose.yml:37` | PgBouncer `DEFAULT_POOL_SIZE=20` |
