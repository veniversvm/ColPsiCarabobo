package domain

import "errors"

var (
	// ErrPasswordIncorrect is returned when the provided password does not match.
	ErrPasswordIncorrect = errors.New("contraseña actual incorrecta")
	// ErrInvalidCredentials is returned when login credentials are invalid.
	ErrInvalidCredentials = errors.New("credenciales inválidas")
	// ErrAccountInactive is returned when the user account is suspended or inactive.
	ErrAccountInactive = errors.New("cuenta inactiva o suspendida")
)

var (
	// ErrPermissionDenied is returned when the user lacks permission to edit a resource.
	ErrPermissionDenied = errors.New("no tienes permiso para editar este registro")
	// ErrInsufficientPerms is returned when the user has insufficient role permissions.
	ErrInsufficientPerms = errors.New("permisos insuficientes")
)

var (
	// ErrPsiNotFound is returned when a psychologist record cannot be found.
	ErrPsiNotFound = errors.New("psicólogo no encontrado")
)

var (
	// ErrMaxSocialNetworks is returned when the social network limit per profile is reached.
	ErrMaxSocialNetworks = errors.New("límite máximo de redes sociales alcanzado")
	// ErrSocialPermDenied is returned when the user cannot edit another user's social network.
	ErrSocialPermDenied = errors.New("no tienes permiso para editar esta red social")
	// ErrSocialOwnDenied is returned when the user cannot delete a social network they don't own.
	ErrSocialOwnDenied = errors.New("no puedes borrar una red social que no te pertenece")
	// ErrPostPermDenied is returned when the user lacks publishing permissions.
	ErrPostPermDenied = errors.New("no tienes permiso para publicar")
	// ErrUniqueViolation is returned when a unique constraint is violated.
	ErrUniqueViolation = errors.New("registro duplicado")
	// ErrSudoExists is returned when a SUDO user already exists.
	ErrSudoExists = errors.New("ya existe un usuario SUDO")
	// ErrInvalidRequest is returned when a request payload fails basic validation.
	ErrInvalidRequest = errors.New("solicitud inválida")
	// ErrDeontologiaNotFound is returned when a deontological record cannot be found.
	ErrDeontologiaNotFound = errors.New("entrada deontológica no encontrada")
	// ErrObservacionesNotFound is returned when an internal observation cannot be found.
	ErrObservacionesNotFound = errors.New("observación no encontrada")
	// ErrDocumentNotFound is returned when a digital document cannot be found.
	ErrDocumentNotFound = errors.New("documento no encontrado")
)

var (
	// ErrNotificationNotFound is returned when a notification cannot be found.
	ErrNotificationNotFound = errors.New("notificación no encontrada")
	// ErrNotificationPermDenied is returned when an admin lacks notification permissions.
	ErrNotificationPermDenied = errors.New("no tienes permiso para gestionar notificaciones")
	// ErrNotificationCannotCancel is returned when a notification is not in a cancellable state.
	ErrNotificationCannotCancel = errors.New("solo se pueden cancelar notificaciones programadas pendientes")
	// ErrNotificationTargetNotOwned is returned when a user tries to mark a notification they don't own.
	ErrNotificationTargetNotOwned = errors.New("no puedes marcar una notificación que no te pertenece")
	// ErrNotificationInvalidSchedule is returned when scheduled_at is in the past.
	ErrNotificationInvalidSchedule = errors.New("la fecha programada debe ser futura")
	// ErrNotificationInvalidTargetType is returned when target_type es inválido.
	ErrNotificationInvalidTargetType = errors.New("target_type inválido")
	// ErrAttachmentNotFound is returned when a notification attachment cannot be found.
	ErrAttachmentNotFound = errors.New("adjunto no encontrado")
)

var (
	// ErrProjectNotFound is returned when a kanban project cannot be found.
	ErrProjectNotFound = errors.New("proyecto no encontrado")
	// ErrColumnNotFound is returned when a kanban column cannot be found.
	ErrColumnNotFound = errors.New("columna no encontrada")
	// ErrCardNotFound is returned when a kanban card cannot be found.
	ErrCardNotFound = errors.New("tarjeta no encontrada")
	// ErrNoteNotFound is returned when a kanban note cannot be found.
	ErrNoteNotFound = errors.New("nota no encontrada")
	// ErrNoteLimitReached is returned when a card already has its 10 notes.
	ErrNoteLimitReached = errors.New("la tarjeta ya tiene el máximo de 10 notas")
	// ErrNoteTooLong is returned when a note exceeds 500 characters.
	ErrNoteTooLong = errors.New("la nota no puede superar los 500 caracteres")
	// ErrNotProjectMember is returned when the admin is not part of the project.
	ErrNotProjectMember = errors.New("no perteneces a este proyecto")
	// ErrMemberAlreadyExists is returned when adding a member that is already part of the project.
	ErrMemberAlreadyExists = errors.New("ese administrador ya es miembro del proyecto")
	// ErrInvalidMemberRole is returned when a member role is not valid.
	ErrInvalidMemberRole = errors.New("rol inválido para el proyecto")
)
