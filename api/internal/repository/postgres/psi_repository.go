// api/internal/repository/postgres/psi_repository.go

// Package postgres provee la implementación concreta de los repositorios usando PostgreSQL y GORM.
package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/request_structs"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// psiRepo implementa la interfaz domain.PsiUserRepository.
// Centraliza todas las interacciones con la base de datos relacionadas con los psicólogos,
// incluyendo su perfil, historial académico, solvencia y redes sociales.
type psiRepo struct {
	db *gorm.DB
}

// NewPsiRepository inyecta la conexión de GORM y devuelve la interfaz del dominio.
func NewPsiRepository(db *gorm.DB) domain.PsiUserRepository {
	return &psiRepo{db: db}
}

// =========================================================================
// GESTIÓN CORE DEL PSICÓLOGO
// =========================================================================

// CreateWithColData realiza una inserción atómica de usuario y sus datos colegiales.
// Utiliza una transacción para asegurar que no se cree un usuario sin sus datos base asociados.
func (r *psiRepo) CreateWithColData(
	ctx context.Context,
	psi *domain.PsiUserModel,
	colData *domain.PsiUserColData,
	solvencies []domain.PsiUserSolvency,
	postgrades []domain.PsiUserPostGrade,
) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Crear una bio vacía para satisfacer la FK (siempre requerida)
		emptyBio := &domain.TextModel{
			// Ajusta los campos requeridos de tu TextModel aquí
			Content: "",
		}
		if err := tx.Create(emptyBio).Error; err != nil {
			return fmt.Errorf("error creating empty bio: %w", err)
		}

		// 2. Asignar el ID de la bio al usuario ANTES de insertarlo
		psi.BioTextID = emptyBio.ID

		// 3. Crear el Psicólogo (ahora bio_text_id apunta a un registro real)
		if err := tx.Create(psi).Error; err != nil {
			return fmt.Errorf("error creating psi user: %w", err)
		}

		// 4. Vincular los datos colegiales
		colData.PsiUserModelID = psi.ID
		if err := tx.Create(colData).Error; err != nil {
			return fmt.Errorf("error creating col data: %w", err)
		}

		// 5. Crear solvencias (historial anual completo)
		if len(solvencies) > 0 {
			for i := range solvencies {
				solvencies[i].PsiUserModelID = psi.ID
			}
			if err := tx.Create(&solvencies).Error; err != nil {
				return fmt.Errorf("error creating solvency data: %w", err)
			}
		}

		// ── 6. Crear los postgrados (NUEVO) ───────────────────────────
		if len(postgrades) > 0 {
			// Nos aseguramos de que todos los postgrados tengan el ID del usuario recién creado
			for i := range postgrades {
				postgrades[i].PsiUserID = psi.ID
			}

			// GORM hará un Bulk Insert (inserción múltiple) automáticamente
			if err := tx.Create(&postgrades).Error; err != nil {
				return fmt.Errorf("error creating postgrades: %w", err)
			}
		}

		return nil
	})
}

// GetByID recupera un psicólogo incluyendo sus relaciones mediante Eager Loading.
// Optimiza la carga de postgrados ordenándolos cronológicamente de forma descendente.
func (r *psiRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.PsiUserModel, error) {
	var psi domain.PsiUserModel

	err := r.db.WithContext(ctx).
		Preload("ColData").
		Preload("PostGrades", func(db *gorm.DB) *gorm.DB {
			return db.Order("graduation_year DESC")
		}).
		Preload("SocialNetworks").
		Preload("FullBio").
		First(&psi, "id = ?", id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("psicólogo no encontrado: %w", err)
		}
		return nil, err
	}

	return &psi, nil
}

// GetByFPV recupera un psicólogo incluyendo sus relaciones mediante Eager Loading.
// Optimiza la carga de postgrados ordenándolos cronológicamente de forma descendente.
func (r *psiRepo) GetByFPV(ctx context.Context, id int) (domain.PsiUserModel, error) {
	var psi domain.PsiUserModel

	err := r.db.WithContext(ctx).
		Preload("ColData").
		Preload("PostGrades", func(db *gorm.DB) *gorm.DB {
			return db.Order("graduation_year DESC")
		}).
		Preload("SocialNetworks").
		Preload("FullBio").
		First(&psi, "fpv = ?", id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return psi, fmt.Errorf("psicólogo no encontrado: %w", err)
		}
		return psi, err
	}

	return psi, nil
}

// GetByIdentifier busca un psicólogo por su nombre de usuario o correo electrónico.
// Es una función crítica para los procesos de autenticación (Login).
func (r *psiRepo) GetByIdentifier(ctx context.Context, identifier string) (*domain.PsiUserModel, error) {
	var psi domain.PsiUserModel
	err := r.db.WithContext(ctx).
		Where("username = ? OR email = ?", identifier, identifier).
		First(&psi).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("usuario no encontrado")
		}
		return nil, err
	}
	return &psi, nil
}

// Delete realiza un borrado lógico (Soft Delete) del psicólogo.
// Al existir el campo DeletedAt en el modelo, GORM ejecuta un UPDATE en lugar de un DELETE físico.
func (r *psiRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.PsiUserModel{}, "id = ?", id).Error
}

// =========================================================================
// ACTUALIZACIONES (MUTACIONES)
// =========================================================================

// Update actualiza tanto el perfil como los datos colegiales dentro de una transacción.
// Generalmente utilizado por administradores para ediciones globales.
func (r *psiRepo) Update(
	ctx context.Context,
	psi *domain.PsiUserModel,
	colData *domain.PsiUserColData,
	bioText *domain.TextModel,
	solvencies []domain.PsiUserSolvency,
) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {

		// 1. Guardar Bio Extensa
		if bioText != nil {
			if err := tx.Model(bioText).Updates(map[string]interface{}{
				"content": bioText.Content,
			}).Error; err != nil {
				return err
			}
			psi.BioTextID = bioText.ID
		}

		// 2. Actualizar perfil principal
		updateMap := map[string]interface{}{
			// ── Credenciales ──────────────────────────────────────────────
			"username":            psi.Username,
			"email":               psi.Email,
			"password":            psi.Password,
			"key":                 psi.Key,
			"audio_book_shell_id": psi.AudioBookShellId,

			// ── Identidad y Filiación (solo admin) ───────────────────────
			"first_name":       psi.FirstName,
			"second_name":      psi.SecondName,
			"last_name":        psi.LastName,
			"second_last_name": psi.SecondLastName,
			"ci":               psi.CI,
			"fpv":              psi.FPV,
			"born_date":        psi.BornDate,
			"genre":            psi.Genre,
			"nationality":      psi.Nationality,

			// ── Estatus Administrativo (solo admin) ──────────────────────
			"is_active":     gorm.Expr("?", psi.IsActive),
			"solvent":       gorm.Expr("?", psi.Solvent),
			"proof_of_life": gorm.Expr("?", psi.ProofOfLife),

			// ── Contacto ─────────────────────────────────────────────────
			"contact_email":      psi.ContactEmail,
			"contact_phone":      psi.ContactPhone,     // Reemplaza a public_phone
			"contact_cell_phone": psi.ContactCellPhone, // Nuevo
			"service_address":    psi.ServiceAddress,

			// ── Ubicación: Carabobo ───────────────────────────────────────
			"municipality_carabobo": psi.MunicipalityCarabobo,
			"phone_carabobo":        psi.PhoneCarabobo,
			"cel_phone_carabobo":    psi.CelPhoneCarabobo,

			// ── Ubicación: Fuera de Carabobo (Venezuela) ─────────────────
			"state_outside":                     psi.StateOutside,
			"municipality_out_side_carabobo":    psi.MunicipalityOutSideCarabobo,
			"phone_out_side_carabobo":           psi.PhoneOutSideCarabobo,
			"cel_phone_out_side_carabobo":       psi.CelPhoneOutSideCarabobo,
			"service_address_out_side_carabobo": psi.ServiceAddressOutSideCarabobo,

			// ── Ubicación: Fuera de Venezuela ─────────────────────────────
			"country":                            psi.Country,
			"phone_out_side_venezuela":           psi.PhoneOutSideVenezuela,
			"cell_phone_out_side_venezuela":      psi.CellPhoneOutSideVenezuela, // Nuevo
			"service_address_out_side_venezuela": psi.ServiceAddressOutSideVenezuela,

			// ── Perfil Profesional ────────────────────────────────────────
			"primary_work_area":      psi.PrimaryWorkArea,
			"secondary_work_area":    psi.SecondaryWorkArea,
			"primary_specialty_id":   psi.PrimarySpecialtyID,
			"secondary_specialty_id": psi.SecondarySpecialtyID,
			"mini_bio":               psi.MiniBio,
			"bio_text_id":            psi.BioTextID,

			// ── Imagen de perfil ──────────────────────────────────────────
			"profile_picture_s3_key": psi.ProfilePictureS3Key,

			// ── Auditoría ─────────────────────────────────────────────────
			"update_by":    psi.UpdateBy,
			"update_by_id": psi.UpdateById,

			// ── Privacidad: Contacto principal ────────────────────────────
			"show_contact_email":          gorm.Expr("?", psi.ShowContactEmail),
			"show_public_service_address": gorm.Expr("?", psi.ShowPublicServiceAddress),

			// ── Privacidad: Carabobo ─────────────────────────────────────
			"show_municipality_carabobo": gorm.Expr("?", psi.ShowMunicipalityCarabobo),
			"show_phone_carabobo":        gorm.Expr("?", psi.ShowPhoneCarabobo),
			"show_cel_phone_carabobo":    gorm.Expr("?", psi.ShowCelPhoneCarabobo),

			// ── Privacidad: Fuera de Carabobo ────────────────────────────
			"show_state_outside":                            gorm.Expr("?", psi.ShowStateOutside),
			"show_municipality_out_side_carabobo":           gorm.Expr("?", psi.ShowMunicipalityOutSideCarabobo),
			"show_phone_out_side_carabobo":                  gorm.Expr("?", psi.ShowPhoneOutSideCarabobo),
			"show_cell_phone_out_side_carabobo":             gorm.Expr("?", psi.ShowCellPhoneOutSideCarabobo),
			"show_public_service_address_out_side_carabobo": gorm.Expr("?", psi.ShowPublicServiceAddressOutSideCarabobo),

			// ── Privacidad: Fuera de Venezuela ────────────────────────────
			"show_phone_out_side_venezuela":                  gorm.Expr("?", psi.ShowPhoneOutSideVenezuela),
			"show_cell_phone_out_side_venezuela":             gorm.Expr("?", psi.ShowCellPhoneOutSideVenezuela),
			"show_public_service_address_out_side_venezuela": gorm.Expr("?", psi.ShowPublicServiceAddressOutSideVenezuela),
		}

		if err := tx.Model(&domain.PsiUserModel{}).
			Where("id = ?", psi.ID).
			Omit("created_at", "create_by", "create_by_id").
			Updates(updateMap).Error; err != nil {
			return err
		}

		// 3. Actualizar Datos Colegiales
		if colData != nil {
			colDataMap := map[string]interface{}{
				// ── Visibilidad ───────────────────────────────────────────
				"show_university_undergraduate": gorm.Expr("?", colData.ShowUniversityUndergraduate),
				"show_graduate_date":            gorm.Expr("?", colData.ShowGraduateDate),
				"show_mention_undergraduate":    gorm.Expr("?", colData.ShowMentionUndergraduate),

				// ── Datos Académicos (solo admin) ─────────────────────────
				"guild_inscription_date":   colData.GuildInscriptionDate, // Nuevo
				"university_undergraduate": colData.UniversityUndergraduate,
				"graduate_date":            colData.GraduateDate,
				"mention_undergraduate":    colData.MentionUndergraduate,

				// ── Registro de Título (solo admin) ───────────────────────
				"register_number":      colData.RegisterNumber,
				"register_title_state": colData.RegisterTitleState,
				"register_title_date":  colData.RegisterTitleDate,
				"register_folio":       colData.RegisterFolio,
				"register_tome":        colData.RegisterTome,

				// ── Banderas Profesionales (solo admin) ───────────────────
				"guild_director":        gorm.Expr("?", colData.GuildDirector),
				"sixty_five_or_plus":    gorm.Expr("?", colData.SixtyFiveOrPlus),
				"guild_collaborator":    gorm.Expr("?", colData.GuildCollaborator),
				"public_employee":       gorm.Expr("?", colData.PublicEmployee),
				"discapacity":           gorm.Expr("?", colData.Discapacity), // Nuevo
				"university_professor":  gorm.Expr("?", colData.UniversityProfessor),
				"double_guild":          gorm.Expr("?", colData.DoubleGuild),
				"double_guild_location": colData.DoubleGuildLocation, // Nuevo
				"cpsm":                  gorm.Expr("?", colData.CPSM),
				"date_of_last_solvency": colData.DateOfLastSolvency,

				// ── Imágenes de Títulos ───────────────────────────────────
				"title_image_one_s3_key":   colData.TitleImageOneS3Key,
				"title_image_two_s3_key":   colData.TitleImageTwoS3Key,
				"title_image_three_s3_key": colData.TitleImageThreeS3Key,

				// ── Auditoría ─────────────────────────────────────────────
				"update_by":    colData.UpdateBy,
				"update_by_id": colData.UpdateById,
			}

			if err := tx.Model(&domain.PsiUserColData{}).
				Where("id = ?", colData.ID).
				Omit("created_at", "create_by", "create_by_id").
				Updates(colDataMap).Error; err != nil {
				return err
			}
		}

		if len(solvencies) > 0 {
			// Usamos tx.Create pasándole el PUNTERO al slice
			err := tx.Clauses(clause.OnConflict{
				// Columnas de conflicto (deben tener un UNIQUE INDEX en la DB)
				Columns: []clause.Column{{Name: "psi_user_model_id"}, {Name: "date"}},
				// Si hay conflicto, actualizamos estos campos
				DoUpdates: clause.AssignmentColumns([]string{"updated_at", "update_by", "update_by_id"}),
			}).Create(&solvencies).Error

			if err != nil {
				// Si esto falla, toda la transacción (incluyendo psi y colData) hará Rollback
				return fmt.Errorf("error en on-conflict solvencias: %w", err)
			}
		}

		return nil
	})
}

// UpdatePublicProfile actualiza los datos permitidos para edición por parte del usuario.
// Usa tx.Omit("ColData") para prevenir que GORM intente actualizar asociaciones no deseadas.
// api/internal/repository/postgres/psi_repository.go
// PROBLEMA: GORM ignora zero values (false, "", 0) tanto en Updates(struct)
// como en Updates(map) cuando el valor es un tipo nativo de Go.
//
// SOLUCIÓN: Usar gorm.Expr() para los campos bool, lo que fuerza a GORM
// a escribir el valor literal en el SQL sin importar si es true o false.

func (r *psiRepo) UpdatePublicProfile(
	ctx context.Context,
	psi *domain.PsiUserModel,
	colData *domain.PsiUserColData,
	bioText *domain.TextModel,
) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {

		// 1. Guardar Bio Extensa
		if bioText != nil {
			if err := tx.Model(bioText).Updates(map[string]interface{}{
				"content": bioText.Content,
			}).Error; err != nil {
				return err
			}
			psi.BioTextID = bioText.ID
		}

		// 2. Actualizar perfil principal
		updateMap := map[string]interface{}{
			// Texto Base
			"username":           psi.Username,
			"contact_email":      psi.ContactEmail,
			"contact_phone":      psi.ContactPhone,     // Reemplaza a public_phone
			"contact_cell_phone": psi.ContactCellPhone, // Nuevo campo
			"service_address":    psi.ServiceAddress,

			// Ubicación: Carabobo
			"municipality_carabobo": psi.MunicipalityCarabobo,
			"phone_carabobo":        psi.PhoneCarabobo,
			"cel_phone_carabobo":    psi.CelPhoneCarabobo,

			// Ubicación: Fuera de Carabobo (Venezuela)
			"state_outside":                     psi.StateOutside,
			"municipality_out_side_carabobo":    psi.MunicipalityOutSideCarabobo,
			"phone_out_side_carabobo":           psi.PhoneOutSideCarabobo,
			"cel_phone_out_side_carabobo":       psi.CelPhoneOutSideCarabobo,
			"service_address_out_side_carabobo": psi.ServiceAddressOutSideCarabobo,

			// Ubicación: Exterior
			"country":                            psi.Country,
			"phone_out_side_venezuela":           psi.PhoneOutSideVenezuela,
			"cell_phone_out_side_venezuela":      psi.CellPhoneOutSideVenezuela, // Nuevo
			"service_address_out_side_venezuela": psi.ServiceAddressOutSideVenezuela,

			// Profesional
			"primary_work_area":      psi.PrimaryWorkArea,
			"secondary_work_area":    psi.SecondaryWorkArea,
			"primary_specialty_id":   psi.PrimarySpecialtyID,
			"secondary_specialty_id": psi.SecondarySpecialtyID,
			"mini_bio":               psi.MiniBio,
			"bio_text_id":            psi.BioTextID,
			"profile_picture_s3_key": psi.ProfilePictureS3Key,

			// Password & Key
			"password": psi.Password,
			"key":      psi.Key,

			// Auditoría
			"update_by":    psi.UpdateBy,
			"update_by_id": psi.UpdateById,

			// ── PRIVACIDAD (BOOLS) ──
			// Contacto principal
			"show_contact_email":          gorm.Expr("?", psi.ShowContactEmail),
			"show_public_service_address": gorm.Expr("?", psi.ShowPublicServiceAddress),
			// Nota: show_public_phone se eliminó, la privacidad telefónica ahora es por zona

			// Privacidad: Carabobo
			"show_municipality_carabobo": gorm.Expr("?", psi.ShowMunicipalityCarabobo),
			"show_phone_carabobo":        gorm.Expr("?", psi.ShowPhoneCarabobo),
			"show_cel_phone_carabobo":    gorm.Expr("?", psi.ShowCelPhoneCarabobo),

			// Privacidad: Fuera de Carabobo
			"show_state_outside":                            gorm.Expr("?", psi.ShowStateOutside),
			"show_municipality_out_side_carabobo":           gorm.Expr("?", psi.ShowMunicipalityOutSideCarabobo),
			"show_phone_out_side_carabobo":                  gorm.Expr("?", psi.ShowPhoneOutSideCarabobo),
			"show_cell_phone_out_side_carabobo":             gorm.Expr("?", psi.ShowCellPhoneOutSideCarabobo),
			"show_public_service_address_out_side_carabobo": gorm.Expr("?", psi.ShowPublicServiceAddressOutSideCarabobo),

			// Privacidad: Exterior
			"show_phone_out_side_venezuela":                  gorm.Expr("?", psi.ShowPhoneOutSideVenezuela),
			"show_cell_phone_out_side_venezuela":             gorm.Expr("?", psi.ShowCellPhoneOutSideVenezuela),
			"show_public_service_address_out_side_venezuela": gorm.Expr("?", psi.ShowPublicServiceAddressOutSideVenezuela),
		}

		if err := tx.Model(psi).
			Where("id = ?", psi.ID).
			Omit("ci", "fpv", "is_active", "solvent", "created_at", "create_by", "create_by_id").
			Updates(updateMap).Error; err != nil {
			return err
		}

		// 3. Actualizar Datos Colegiales
		if colData != nil {
			colDataMap := map[string]interface{}{
				"show_university_undergraduate": gorm.Expr("?", colData.ShowUniversityUndergraduate),
				"show_graduate_date":            gorm.Expr("?", colData.ShowGraduateDate),
				"show_mention_undergraduate":    gorm.Expr("?", colData.ShowMentionUndergraduate),

				"title_image_one_s3_key":   colData.TitleImageOneS3Key,
				"title_image_two_s3_key":   colData.TitleImageTwoS3Key,
				"title_image_three_s3_key": colData.TitleImageThreeS3Key,

				"update_by":    colData.UpdateBy,
				"update_by_id": colData.UpdateById,
			}

			if err := tx.Model(colData).
				Where("id = ?", colData.ID).
				Updates(colDataMap).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

// UpdateKey actualiza únicamente los campos relacionados con la sesión y auditoría.
// Mejora el rendimiento al usar Select() para limitar las columnas en la sentencia SQL.
func (r *psiRepo) UpdateKey(ctx context.Context, psi *domain.PsiUserModel) error {
	return r.db.WithContext(ctx).Model(psi).
		Select("Key", "UpdatedAt", "UpdateBy", "UpdateById").
		Updates(psi).Error
}

// GetPsiUserColData recupera exclusivamente la información colegial de un psicólogo.
// Útil para vistas ligeras donde no se requiere el perfil completo del usuario.
func (r *psiRepo) GetPsiUserColData(ctx context.Context, psiID uuid.UUID) (*domain.PsiUserColData, error) {
	var colData domain.PsiUserColData
	err := r.db.WithContext(ctx).First(&colData, "psi_user_model_id = ?", psiID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("datos colegiales no encontrados para el psicólogo: %w", err)
		}
		return nil, err
	}
	return &colData, nil
}

// =========================================================================
// MOTORES DE BÚSQUEDA Y ESTADÍSTICAS
// =========================================================================

// SearchDirectory implementa la lógica de búsqueda para el directorio público.
// Diferencia entre búsqueda por "Identidad" (exacta por CI/FPV) y "Navegación" (filtro por solvencia).
func (r *psiRepo) SearchDirectory(ctx context.Context, filter request_structs.PsiDirectoryFilterDTO) ([]domain.PsiUserModel, int64, error) {
	var users []domain.PsiUserModel
	var total int64

	// 1. Base: Siempre ACTIVOS
	query := r.db.WithContext(ctx).Model(&domain.PsiUserModel{}).
		Select("id, first_name, last_name, ci, fpv, profile_picture_s3_key, mini_bio, solvent, primary_work_area, secondary_work_area, primary_specialty_id, secondary_specialty_id, updated_at").
		Where("is_active = ?", true)

	// 2. Lógica de Búsqueda por Identidad
	if filter.SearchTerm != "" {
		// Dividimos la búsqueda en palabras (tokens). Ej: "Francisco Hernandez" -> ["Francisco", "Hernandez"]
		words := strings.Fields(strings.TrimSpace(filter.SearchTerm))

		for _, word := range words {
			w := "%" + word + "%"
			// Cada palabra debe aparecer en algún campo del registro.
			// Se usa unaccent() individual por campo para aprovechar los expression indexes:
			//   idx_psi_users_unaccent_first_name, idx_psi_users_unaccent_last_name
			query = query.Where(
				r.db.Where("unaccent(first_name) ILIKE unaccent(?)", w).
					Or("unaccent(second_name) ILIKE unaccent(?)", w).
					Or("unaccent(last_name) ILIKE unaccent(?)", w).
					Or("unaccent(second_last_name) ILIKE unaccent(?)", w).
					Or("CAST(ci AS TEXT) LIKE ?", w).
					Or("CAST(fpv AS TEXT) LIKE ?", w),
			)
		}
	} else {
		// Si no hay búsqueda de texto, solo mostramos los SOLVENTES (Navegación general)
		query = query.Where("solvent = ?", true)
	}

	// 3. Filtro por Área de Desempeño (Especialidad) — FK directa
	if filter.SpecialtyID > 0 {
		query = query.Where("primary_specialty_id = ? OR secondary_specialty_id = ?",
			filter.SpecialtyID, filter.SpecialtyID)
	}

	// 4. Filtro de Ubicación (Respetando Privacidad)
	if filter.Location != "" {
		loc := "%" + strings.TrimSpace(filter.Location) + "%"
		query = query.Where(
			r.db.Where("unaccent(municipality_carabobo) ILIKE unaccent(?) AND show_municipality_carabobo = ?", loc, true).
				Or("unaccent(state_outside) ILIKE unaccent(?) AND show_state_outside = ?", loc, true).
				Or("unaccent(municipality_out_side_carabobo) ILIKE unaccent(?) AND show_municipality_out_side_carabobo = ?", loc, true).
				Or("unaccent(country) ILIKE unaccent(?)", loc),
		)
	}

	// 5. Conteo y Ejecución
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (filter.Page - 1) * filter.Limit

	// Orden: Solventes primero, luego fotos, luego apellido
	err := query.Order("solvent DESC, profile_picture_s3_key DESC, last_name ASC").
		Limit(filter.Limit).
		Offset(offset).
		Find(&users).Error

	return users, total, err
}

// SearchAdmin provee una búsqueda sin restricciones para el panel administrativo.
// Ignora estados de solvencia o privacidad para permitir la gestión total.
func (r *psiRepo) SearchAdmin(ctx context.Context, filter request_structs.PsiDirectoryFilterDTO) ([]domain.PsiUserModel, int64, error) {
	var users []domain.PsiUserModel
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.PsiUserModel{}).
		Select("id, first_name, last_name, ci, fpv, email, solvent, is_active, primary_work_area, secondary_work_area, primary_specialty_id, secondary_specialty_id")

	if filter.SearchTerm != "" {
		// Limpiamos espacios
		term := "%" + strings.TrimSpace(filter.SearchTerm) + "%"

		// Usamos unaccent() individual por campo para aprovechar los expression indexes:
		//   idx_psi_users_unaccent_first_name, idx_psi_users_unaccent_last_name
		query = query.Where(
			r.db.Where("unaccent(first_name) ILIKE unaccent(?)", term).
				Or("unaccent(last_name) ILIKE unaccent(?)", term).
				Or("unaccent(first_name || ' ' || last_name) ILIKE unaccent(?)", term).
				Or("unaccent(last_name || ' ' || first_name) ILIKE unaccent(?)", term).
				Or("CAST(ci AS TEXT) LIKE ?", term).
				Or("CAST(fpv AS TEXT) LIKE ?", term),
		)
	}

	// Filtro por Área de Desempeño — FK directa
	if filter.SpecialtyID > 0 {
		query = query.Where("primary_specialty_id = ? OR secondary_specialty_id = ?",
			filter.SpecialtyID, filter.SpecialtyID)
	}

	// Filtro por Ubicación (también con unaccent, muy útil para nombres de municipios)
	if filter.Location != "" {
		loc := "%" + filter.Location + "%"
		query = query.Where(
			r.db.Where("unaccent(municipality_carabobo) ILIKE unaccent(?)", loc).
				Or("unaccent(state_outside) ILIKE unaccent(?)", loc).
				Or("unaccent(municipality_out_side_carabobo) ILIKE unaccent(?)", loc),
		)
	}

	if filter.Gender != "" {
		query = query.Where("genre = ?", filter.Gender)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (filter.Page - 1) * filter.Limit
	err := query.Order("created_at DESC").
		Limit(filter.Limit).
		Offset(offset).
		Find(&users).Error

	return users, total, err
}

// Count devuelve el número total de psicólogos, permitiendo filtrar por estado activo.
func (r *psiRepo) Count(ctx context.Context, active *bool) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&domain.PsiUserModel{})

	if active != nil {
		query = query.Where("is_active = ?", *active)
	}

	if err := query.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("repo.Count: %w", err)
	}

	return count, nil
}

// Search implementa una búsqueda genérica mediante un mapa de filtros dinámicos.
func (r *psiRepo) Search(ctx context.Context, filters map[string]interface{}, page, pageSize int) ([]domain.PsiUserModel, int64, error) {
	var psis []domain.PsiUserModel
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.PsiUserModel{})

	if ci, ok := filters["ci"]; ok && ci != "" {
		query = query.Where("ci = ?", ci)
	}
	if fpv, ok := filters["fpv"]; ok && fpv != "" {
		query = query.Where("fpv = ?", fpv)
	}
	if name, ok := filters["name"]; ok && name != "" {
		search := "%" + name.(string) + "%"
		query = query.Where("unaccent(first_name) ILIKE unaccent(?) OR unaccent(last_name) ILIKE unaccent(?)", search, search)
	}
	if active, ok := filters["active"]; ok && active != nil {
		query = query.Where("active = ?", active)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Offset(offset).
		Limit(pageSize).
		Preload("ColData").
		Order("created_at DESC").
		Find(&psis).Error

	return psis, total, err
}

// =========================================================================
// MÓDULO ACADÉMICO (POSTGRADOS)
// =========================================================================

// CreatePostGrade registra un nuevo título o estudio de postgrado.
func (r *psiRepo) CreatePostGrade(ctx context.Context, pg *domain.PsiUserPostGrade) error {
	return r.db.WithContext(ctx).Create(pg).Error
}

// GetPostGradeByID recupera la información de un postgrado específico.
func (r *psiRepo) GetPostGradeByID(ctx context.Context, id uuid.UUID) (*domain.PsiUserPostGrade, error) {
	var pg domain.PsiUserPostGrade
	err := r.db.WithContext(ctx).First(&pg, "id = ?", id).Error
	return &pg, err
}

// UpdatePostGrade actualiza los datos de un registro académico existente.
// Utiliza Updates() con mapa explícito para proteger booleanos e int contra zero-values.
func (r *psiRepo) UpdatePostGrade(ctx context.Context, pg *domain.PsiUserPostGrade) error {
	return r.db.WithContext(ctx).Model(pg).Updates(map[string]interface{}{
		"type":             pg.Type,
		"title":            pg.Title,
		"university":       pg.University,
		"graduation_year":  pg.GraduationYear,
		"description":      pg.Description,
		"active":           gorm.Expr("?", pg.Active),
		"pic_one_s3_key":   pg.PicOneS3Key,
		"pic_two_s3_key":   pg.PicTwoS3Key,
		"pic_three_s3_key": pg.PicThreeS3Key,
		"update_by":        pg.UpdateBy,
		"update_by_id":     pg.UpdateById,
	}).Error
}

// =========================================================================
// GESTION DE SOLVENCIAS
// =========================================================================

func (r *psiRepo) CreateSolvency(ctx context.Context, pg *domain.PsiUserSolvency) error {
	return r.db.WithContext(ctx).Save(pg).Error
}

func (r *psiRepo) GetSolvencies(ctx context.Context, id uuid.UUID) ([]domain.PsiUserSolvency, error) {
	var pg []domain.PsiUserSolvency

	err := r.db.WithContext(ctx).Where("psi_user_model_id = ?", id).Find(&pg).Error
	return pg, err
}

func (r *psiRepo) CreateOrUpdateSolvencies(ctx context.Context, solvencies []domain.PsiUserSolvency) error {
	if len(solvencies) == 0 {
		return nil
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return tx.Clauses(clause.OnConflict{
			// Columnas que determinan el conflicto
			Columns: []clause.Column{{Name: "psi_user_model_id"}, {Name: "date"}},
			// Qué hacer cuando hay conflicto: actualizar todo excepto el ID original
			DoUpdates: clause.AssignmentColumns([]string{"updated_at", "update_by"}),
		}).Create(&solvencies).Error
	})
}

// =========================================================================
// PRESENCIA DIGITAL (REDES SOCIALES)
// =========================================================================

// CreateSocialNetwork vincula una nueva red social al perfil del psicólogo.
func (r *psiRepo) CreateSocialNetwork(ctx context.Context, sn *domain.PsiUserSocialNetwork) error {
	return r.db.WithContext(ctx).Create(sn).Error
}

// GetSocialNetworkByID busca un registro de red social por su ID único.
func (r *psiRepo) GetSocialNetworkByID(ctx context.Context, id uuid.UUID) (*domain.PsiUserSocialNetwork, error) {
	var sn domain.PsiUserSocialNetwork
	err := r.db.WithContext(ctx).First(&sn, "id = ?", id).Error
	return &sn, err
}

// UpdateSocialNetwork modifica el enlace o tipo de una red social existente.
// Utiliza Updates() con mapa explícito para proteger el booleano IsActive contra zero-values.
func (r *psiRepo) UpdateSocialNetwork(ctx context.Context, sn *domain.PsiUserSocialNetwork) error {
	return r.db.WithContext(ctx).Model(sn).Updates(map[string]interface{}{
		"name":         sn.Name,
		"url":          sn.URL,
		"is_active":    gorm.Expr("?", sn.IsActive),
		"update_by":    sn.UpdateBy,
		"update_by_id": sn.UpdateById,
	}).Error
}

// DeleteSocialNetwork elimina una red social (aplica Soft Delete si el modelo lo soporta).
func (r *psiRepo) DeleteSocialNetwork(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.PsiUserSocialNetwork{}, "id = ?", id).Error
}

// CountSocialNetworksByPsiID devuelve la cantidad de redes activas que tiene un usuario.
func (r *psiRepo) CountSocialNetworksByPsiID(ctx context.Context, psiID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.PsiUserSocialNetwork{}).
		Where("psi_user_id = ?", psiID).
		Count(&count).Error

	return count, err
}

// GetTextContentByID recupera el contenido de un TextModel dado su UUID.
// Usamos Select("content") para no traer toda la estructura de auditoría si no es necesaria,
// optimizando la transferencia de datos desde Postgres.
func (r *psiRepo) GetTextContentByID(ctx context.Context, id uuid.UUID) (string, error) {
	var textModel domain.TextModel

	err := r.db.WithContext(ctx).
		Model(&domain.TextModel{}).
		Select("content"). // Solo traemos el campo necesario
		Where("id = ?", id).
		First(&textModel).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", fmt.Errorf("biografía no encontrada: %w", err)
		}
		return "", err
	}

	return textModel.Content, nil
}

// ValidateUniqueCredentials comprueba ambos campos y retorna un error descriptivo si fallan.
// Usa comparación exacta (=) ya que username y email tienen UNIQUE constraints,
// lo que permite aprovechar los índices únicos directamente.
func (r *psiRepo) ValidateUniqueCredentials(ctx context.Context, username, email string, excludeID uuid.UUID) error {
	if username != "" {
		var count int64
		r.db.WithContext(ctx).Model(&domain.PsiUserModel{}).
			Where("username = ? AND id != ?", username, excludeID).
			Count(&count)
		if count > 0 {
			return errors.New("el nombre de usuario ya está en uso")
		}
	}
	if email != "" {
		var count int64
		r.db.WithContext(ctx).Model(&domain.PsiUserModel{}).
			Where("email = ? AND id != ?", email, excludeID).
			Count(&count)
		if count > 0 {
			return errors.New("el email ya está en uso")
		}
	}
	return nil
}

func (r *psiRepo) GetSitemapData(ctx context.Context) ([]domain.PsiUserModel, error) {
	var users []domain.PsiUserModel
	// Traemos solo los campos necesarios para el slug y solo los activos/solventes
	err := r.db.WithContext(ctx).
		Select("first_name, last_name, fpv").
		Where("is_active = ? AND solvent = ?", true, true).
		Find(&users).Error
	return users, err
}
