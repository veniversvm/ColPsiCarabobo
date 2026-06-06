# ⚙️ Módulo de Configuración (Config)

[⬅ Volver al Inicio](../../README.md)

Este módulo es el encargado de la **Gestión de Entorno**. Su propósito principal es desacoplar el código fuente de los valores sensibles (credenciales) y de los parámetros que cambian según el lugar donde se ejecute la aplicación (Local, Docker, Producción).

## 🎯 ¿Por qué existe este módulo?

1. **Seguridad (12-Factor App):** Sigue las mejores prácticas de la industria al no "hardcodear" (escribir directamente) claves secretas en el código. Las credenciales viajan a través de variables de entorno.
2. **Portabilidad:** Permite que la misma aplicación corra en tu PC con `localhost:5432` y en un servidor de producción con un RDS de AWS sin cambiar una sola línea de código.
3. **Centralización:** Si mañana el servicio de correo cambia de puerto, solo se modifica el archivo `.env` o este struct, evitando buscar y reemplazar en todo el proyecto.

---

## 🛠️ Estructura de Datos: `Config`

La estructura `Config` agrupa los parámetros en 5 categorías lógicas:

| Categoría               | Variables Clave                                | Propósito                                                                           |
| :----------------------- | :--------------------------------------------- | :----------------------------------------------------------------------------------- |
| **Servidor**       | `Port`, `AllowedOrigins`                   | Define dónde escucha la API y quién tiene permiso de acceder (CORS).               |
| **Base de Datos**  | `DBHost`, `DBUser`, `DBPass`, `DBName` | Datos de conexión para el driver de PostgreSQL.                                     |
| **Almacenamiento** | `S3Bucket`, `S3Endpoint`, `S3Region`     | Configuración para AWS S3 o MinIO (para manejo de imágenes de perfiles).           |
| **Email (SMTP)**   | `SMTPHost`, `SMTPPort`, `SMTPUser`       | Parámetros para el envío de notificaciones y bienvenidas.                          |
| **Entorno**        | `Environment`                                | Define si estamos en `development` o `production` para ajustar el nivel de logs. |

---

## 🛠️ Funciones del Módulo

### `InitConfig()`

Es la función "maestra" de inicialización.

- **Qué hace:** Intenta cargar un archivo `.env` usando la librería `godotenv`. Si no lo encuentra, no se detiene (no lanza `Fatal`), ya que en entornos de **Docker** o **Kubernetes**, las variables ya están inyectadas en el sistema.
- **Singleton:** Instancia la variable global `Envs` para que el resto de la aplicación pueda consultar la configuración de forma rápida.

### `getEnv(key, fallback)`

Es una función auxiliar de seguridad y robustez.

- **`key`**: El nombre de la variable que buscamos en el sistema.
- **`fallback`**: El valor por defecto que usará la aplicación si la variable no existe.
- **Importancia:** Evita que la aplicación falle por falta de una variable no crítica, garantizando valores por defecto sensatos (ej: usar el puerto `8080` si no se especifica uno).

---

## 📋 Listado de Variables de Entorno Soportadas

| Variable            | Descripción        | Valor Default (Fallback)   |
| :------------------ | :------------------ | :------------------------- |
| `PORT`            | Puerto de la API    | `8080`                   |
| `DB_PASSWORD`     | Clave de Postgres   | `postgres`               |
| `AWS_REGION`      | Región de S3       | `us-east-1`              |
| `SMTP_PORT`       | Puerto de correo    | `1025` (MailHog/Mailpit) |
| `ALLOWED_ORIGINS` | Dominios permitidos | `http://localhost:3000`  |

---

## 🔗 Navegación

- [Ir al Core de Dominio ➡](../domain/README.md)
- [Ir a Utilidades ➡](../utils/README.md)
