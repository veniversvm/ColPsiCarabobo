# FIX-15b Report — pg_hba.conf restringido a red Docker

| Campo | Valor |
|-------|-------|
| **Fix ID** | FIX-15b |
| **Archivos modificados** | `docker-compose.yml`, `init-db/pg_hba.conf` (NUEVO) |
| **Fecha de implementación** | 2026-07-25 |
| **Estado** | Completado |

---

## Problema

El `pg_hba.conf` default de PostgreSQL acepta conexiones desde **cualquier IP** (`0.0.0.0/0`). En producción esto permite intentos de conexión desde redes externas.

---

## Corrección

Crear `init-db/pg_hba.conf` restringido y montarlo en el contenedor PostgreSQL:

```
# pg_hba.conf
local   all       all                        scram-sha-256
host    all       all        172.16.0.0/12    scram-sha-256
host    all       all        ::/0             reject
```

```yaml
# docker-compose.yml
db:
  volumes:
    - ./init-db/pg_hba.conf:/etc/postgresql/pg_hba.conf
  command: postgres -c hba_file=/etc/postgresql/pg_hba.conf
```

| Regla | Efecto |
|-------|--------|
| `local all all scram-sha-256` | Local socket requiere password |
| `host all all 172.16.0.0/12 scram-sha-256` | Solo red Docker puede conectarse |
| `host all all ::/0 reject` | Todo lo demás rechazado |

---

## Testing

- Docker compose up: todos los servicios arrancan correctamente
- PgBouncer conecta a PostgreSQL: OK
- Migrador Atlas conecta: OK

---

## Archivos relacionados

| Archivo | Relación |
|---------|----------|
| `init-db/pg_hba.conf` | NUEVO — reglas de autenticación |
| `docker-compose.yml` | Montaje + command flag |
