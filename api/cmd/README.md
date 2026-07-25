# 🚀 Puntos de Entrada (cmd/)

El directorio `cmd/` contiene los puntos de entrada de la aplicación, siguiendo la convención de layout de proyectos en Go donde todos los ejecutables residen bajo `cmd/`.

## Estructura

```
cmd/
├── api/                    # Servidor API principal (Fiber)
│   ├── main.go
│   ├── README.md
│   └── .ai-context.md
└── exp/                    # Herramientas experimentales
    ├── migrate/
    │   ├── main.go
    │   ├── README.md
    │   └── .ai-context.md
    ├── README.md
    └── .ai-context.md
```

## Ejecutables

### `cmd/api/`

El servidor API principal de la aplicación. Utiliza **Fiber v2** como framework HTTP y expone todos los endpoints del sistema sobre el puerto `:8080`. Inicializa la base de datos, migraciones, S3, y monta la cadena completa de middleware.

**Uso:**
```bash
go run cmd/api/main.go
```

### `cmd/exp/migrate/`

Herramienta experimental para generar el esquema SQL a partir de los modelos GORM. Crea una base de datos SQLite temporal, ejecuta `AutoMigrate` con todos los modelos del dominio y vuelca el esquema resultante a stdout. Este esquema se utiliza como referencia para las migraciones de **Atlas**.

**Uso:**
```bash
go run cmd/exp/migrate/main.go > schema.sql
```

## Convención de Layout

Siguiendo la convención estándar de proyectos Go, todos los ejecutables viven bajo `cmd/`. Cada subdirectorio contiene su propio `main.go` con una función `main()` independiente. Esto permite mantener múltiples binarios en un solo repositorio sin conflicto de dependencias.

Los paquetes internos (`internal/`, `pkg/`, o paquetes de dominio en la raíz) se comparten entre ejecutables, pero cada punto de entrada tiene su propia cadena de inicialización.
