#!/bin/sh
# Bootstrap del usuario root de Audiobookshelf (servicio `abs-init`).
#
# POR QUÉ EXISTE
# --------------
# Al crear su base por primera vez, Audiobookshelf genera la contraseña del
# usuario `root` y la muestra una sola vez en los logs. Si ese log se pierde
# (recreate del contenedor, logrotate, un server nuevo) la contraseña queda
# desconocida y no hay forma de recuperarla por la vía normal. Como
# `ABS_ADMIN_TOKEN` es un gate duro de arranque (cmd/api/main.go:75 → log.Fatal),
# ese bloqueo tumba la API completa.
#
# Este script crea el root SOLO si no existe, usando POST /init — el propio
# endpoint de bootstrap de ABS. Es idempotente por diseño: el handler rechaza
# con 500 cuando `Database.hasRootUser` ya es true, así que nunca pisa una clave
# existente ni deja el arranque a medias.
#
# Idempotencia verificada en dos capas: primero se consulta GET /status (que
# expone `isInit`) y solo se llama a /init si viene `false`.
#
# Variables de entorno (las inyecta docker-compose.yml):
#   ABS_URL   base URL interna de ABS   (http://audiobookshelf:80)
#   ABS_USER  usuario a crear           (root)
#   ABS_PASS  contraseña a asignar       (ABS_ADMIN_PASSWORD del .env)

set -eu

ABS_URL="${ABS_URL:-http://audiobookshelf:80}"
ABS_USER="${ABS_USER:-root}"
ABS_PASS="${ABS_PASS:-}"

# Espera a que ABS acepte conexiones. /status responde incluso sin root, así que
# sirve de sonda de "la API interna ya está viva".
echo "abs-init: esperando a Audiobookshelf en ${ABS_URL} ..."
until wget -q -O /dev/null "${ABS_URL}/status" 2>/dev/null; do
  sleep 2
done

# Sondeo de estado. -O- sin -q porque el cuerpo es lo que vamos a inspeccionar.
status_json="$(wget -q -O- "${ABS_URL}/status" 2>/dev/null || true)"

if echo "${status_json}" | grep -q '"isInit":true'; then
  echo "abs-init: Audiobookshelf ya tiene usuario root: se deja intacto."
  exit 0
fi

if [ -z "${ABS_PASS}" ]; then
  echo "abs-init: ERROR — falta ABS_ADMIN_PASSWORD en el .env; no se puede crear el root." >&2
  exit 1
fi

echo "abs-init: creando usuario '${ABS_USER}' de Audiobookshelf ..."
if wget -q -O /dev/null \
  --header='Content-Type: application/json' \
  --post-data="{\"newRoot\":{\"username\":\"${ABS_USER}\",\"password\":\"${ABS_PASS}\"}}" \
  "${ABS_URL}/init"; then
  echo "abs-init: usuario root creado."
  exit 0
fi

echo "abs-init: FALLO al crear el root (POST /init)." >&2
exit 1
