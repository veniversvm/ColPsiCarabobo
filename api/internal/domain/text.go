package domain

import "github.com/google/uuid"

// TextModel representa el almacenamiento de contenido extenso de forma aislada.
// Se utiliza para guardar biografías detalladas, artículos de blog o descripciones
// largas, permitiendo que la tabla principal se mantenga ligera y eficiente.
// Esta separación facilita la implementación de filtros de sanitización (anti-XSS)
// sin afectar el rendimiento de las consultas de listado.
type TextModel struct {
	// AuditModel hereda campos de identificación y auditoría.
	AuditModel

	// Content almacena el texto enriquecido o extenso.
	// Se define como tipo 'text' en Postgres para permitir longitud ilimitada.
	Content string `gorm:"type:text" json:"content"`
}

// Post representa una publicación o noticia dentro de la plataforma del Colegio.
// Diseñado para soportar tanto anuncios públicos como contenido exclusivo para colegiados.
type Post struct {
	// AuditModel proporciona la trazabilidad de quién creó/editó la publicación.
	AuditModel

	// Title es el encabezado de la publicación.
	// Limitado a 100 caracteres para optimizar la visualización en el frontend.
	Title string `gorm:"size:100;not null" json:"title"`

	// ShortDescription actúa como un 'snippet' o resumen para las vistas de lista.
	// Ayuda al SEO y a la experiencia de usuario antes de entrar al detalle.
	ShortDescription string `gorm:"size:250" json:"short_description"`

	// Type define el alcance de la publicación.
	// Valores comunes: "public" (abierto a todos), "psi" (solo psicólogos logueados).
	Type string `gorm:"type:varchar(20);default:public" json:"type"`

	// Relación con el texto largo.
	// TextID es la clave foránea que apunta al contenido extenso en TextModel.
	TextID uuid.UUID `json:"text_id"`

	// Text es la instancia cargada del contenido.
	// GORM permite cargar este campo mediante 'Preload' cuando se requiere el detalle completo.
	Text TextModel `gorm:"foreignKey:TextID" json:"text,omitempty"`

	// ImageS3Key almacena la ruta (Key) del archivo de imagen en el bucket de AWS S3/MinIO.
	// No guardamos la URL completa para mantener flexibilidad si cambia el dominio del bucket.
	ImageS3Key string `gorm:"size:512" json:"image_url"`

	// IsActive permite el control de visibilidad (Borrador/Publicado).
	IsActive bool `gorm:"default:true" json:"is_active"`
}

/*

Seguridad de Contenido: Al tener el Content en una tabla separada (TextModel), podemos aplicar reglas de seguridad específicas. Por ejemplo, al guardar, podemos pasar ese string por una librería como bluemonday para eliminar cualquier tag <script> malicioso sin riesgo de "romper" los metadatos del Post.
Rendimiento de Base de Datos: Las bases de datos SQL leen "páginas" de datos. Si una fila de Post tuviera 10KB de texto, cabrían pocas filas por página. Al separar el texto, el listado de noticias (títulos y fechas) es extremadamente veloz porque las filas son pequeñas y uniformes.
Integración con S3: El campo ImageS3Key sigue nuestra estrategia de no comprometer la base de datos con virus. Al subir la imagen a S3 y solo guardar la llave, delegamos la seguridad binaria a AWS y mantenemos la DB limpia.


*/
