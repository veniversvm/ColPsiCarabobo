package service

import (
	"context"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
	"golang.org/x/crypto/bcrypt"

	domain "github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/utils"
	// Importa tus utilidades (ajusta la ruta según tu proyecto)
	// "github.com/veniversvm/ColPsiCarabobo/api/internal/utils"
)

// getValorSeguro evita "index out of range" al leer filas que terminan vacías en el Excel
func getValorSeguro(row []string, index int) string {
	if index < len(row) {
		return strings.TrimSpace(row[index])
	}
	return ""
}

func (s *PsiService) ImportFromXLSX(ctx context.Context, reader io.Reader, adminID uuid.UUID) (int, []map[string]string) {
	f, err := excelize.OpenReader(reader)
	if err != nil {
		return 0, []map[string]string{
			{"error": "no se pudo abrir el archivo Excel: " + err.Error()},
		}
	}
	defer f.Close()

	// Asumiendo que la hoja se llama "BD ColPsiCarabobo 2026"
	rows, err := f.GetRows("BD ColPsiCarabobo 2026")
	if err != nil {
		return 0, []map[string]string{
			{"error": "no se pudo leer la hoja especificada: " + err.Error()},
		}
	}

	successCount := 0
	var failedRecords []map[string]string

	// OPTIMIZACIÓN: Como todos tendrán la misma clave inicial, generamos el hash UNA SOLA VEZ fuera del bucle.
	defaultPassword := utils.GenerateSecureRandomString(12)
	hashedPasswordBytes, _ := bcrypt.GenerateFromPassword([]byte(defaultPassword), bcrypt.DefaultCost)
	hashedPassword := string(hashedPasswordBytes)

	for i, row := range rows {
		// Omitir fila 0 y fila 1 (si la 2 es el encabezado)
		if i < 2 {
			continue
		}

		// Row num real en Excel (i + 1) para los reportes de error
		excelRow := fmt.Sprintf("%d", i+1)

		numFPV := getValorSeguro(row, 3)
		ciStr := getValorSeguro(row, 6)
		firstName := getValorSeguro(row, 7)
		lastName := getValorSeguro(row, 9)
		fullName := firstName + " " + lastName

		if numFPV == "" && ciStr == "" {
			// Probablemente una fila vacía al final del Excel, la ignoramos.
			continue
		}

		// ── Geo-normalización ─────────────────────────────────────────────
		var municipioCarabobo string
		if raw := getValorSeguro(row, 16); raw != "" && raw != "-" {
			mun, ok := utils.NormalizeMunicipioCarabobo(raw)
			if !ok {
				failedRecords = append(failedRecords, map[string]string{
					"fila":   excelRow,
					"nombre": fullName,
					"ci":     ciStr,
					"fpv":    numFPV,
					"error":  fmt.Sprintf("municipio de Carabobo inválido: %q", raw),
				})
				continue
			}
			municipioCarabobo = mun
		}

		var estadoOutside string
		if raw := getValorSeguro(row, 19); raw != "" && raw != "-" {
			estado, ok := utils.NormalizeEstadoVenezuela(raw)
			if !ok {
				estadoOutside = raw
			} else {
				estadoOutside = estado
			}
		}

		psiID := uuid.Must(uuid.NewV7())
		audit := domain.AuditModel{
			CreateById: &adminID,
			CreateBy:   "Admin_XLSX_Import",
			UpdateById: &adminID,
			UpdateBy:   "Admin_XLSX_Import",
		}

		// ── Generación de Username ─────────────────────────────────────────
		email := getValorSeguro(row, 15)
		var username string
		if strings.Contains(email, "@") {
			email_split := strings.Split(email, "@")
			username = email_split[0] + numFPV
		} else {
			// Fallback si no hay email o no tiene @
			username = "psi" + numFPV + ciStr
		}

		// ── Modelo Principal (PsiUserModel) ────────────────────────────────
		psi := &domain.PsiUserModel{
			ID:         psiID,
			Key:        uuid.Must(uuid.NewV7()).String(),
			AuditModel: audit,

			// Credenciales de acceso
			IsActive: getValorSeguro(row, 45) == "Activo",
			Username: username,
			Email:    email,
			Password: hashedPassword,

			// Identidad legal
			FirstName:      firstName,
			SecondName:     cleanDash(getValorSeguro(row, 8)),
			LastName:       lastName,
			SecondLastName: cleanDash(getValorSeguro(row, 10)),
			FPV:            parseInt(numFPV),
			CI:             parseInt(ciStr),
			Nationality:    getValorSeguro(row, 5),
			BornDate:       parseDate(getValorSeguro(row, 11)),
			Genre:          getValorSeguro(row, 13),

			// Estado gremial
			ProofOfLife: strings.ToLower(getValorSeguro(row, 14)) != "fallecido",
			Solvent:     getValorSeguro(row, 45) == "Activo",

			// Contacto interno del gremio
			ContactPhone:     cleanDash(getValorSeguro(row, 17)),
			ContactCellPhone: cleanDash(getValorSeguro(row, 18)),

			// Contacto público y privacidad
			ContactEmail:     "",
			ShowContactEmail: false,

			// Ubicación: Carabobo
			MunicipalityCarabobo:     municipioCarabobo,
			ShowMunicipalityCarabobo: false,
			PhoneCarabobo:            "",
			ShowPhoneCarabobo:        false,
			CelPhoneCarabobo:         "",
			ShowCelPhoneCarabobo:     false,

			// Ubicación: Fuera de Carabobo
			StateOutside:                cleanDash(estadoOutside),
			MunicipalityOutSideCarabobo: cleanDash(getValorSeguro(row, 20)),
			PhoneOutSideCarabobo:        cleanDash(getValorSeguro(row, 21)),
			CelPhoneOutSideCarabobo:     cleanDash(getValorSeguro(row, 22)),

			// Ubicación: Fuera del país
			Country:                   cleanDash(getValorSeguro(row, 23)),
			PhoneOutSideVenezuela:     cleanDash(getValorSeguro(row, 24)),
			ShowPhoneOutSideVenezuela: false,
		}

		// ── Datos Colegiales (PsiUserColData) ──────────────────────────────
		colData := &domain.PsiUserColData{
			ID:                   uuid.Must(uuid.NewV7()),
			PsiUserModelID:       psiID,
			AuditModel:           audit,
			GuildInscriptionDate: parseDate(getValorSeguro(row, 4)),

			// Pregrado
			UniversityUndergraduate: getValorSeguro(row, 25),
			GraduateDate:            parseDate(getValorSeguro(row, 26)),
			MentionUndergraduate:    getValorSeguro(row, 27),

			// Registro legal del título
			RegisterTitleState: getValorSeguro(row, 28),
			RegisterTitleDate:  parseDate(getValorSeguro(row, 29)),
			RegisterNumber:     parseInt(getValorSeguro(row, 30)),
			RegisterFolio:      cleanDash(getValorSeguro(row, 31)),
			RegisterTome:       cleanDash(getValorSeguro(row, 32)),

			// Flags gremiales
			GuildDirector:       len(getValorSeguro(row, 38)) > 0,
			SixtyFiveOrPlus:     len(getValorSeguro(row, 39)) > 0,
			GuildCollaborator:   len(getValorSeguro(row, 40)) > 0,
			PublicEmployee:      len(getValorSeguro(row, 41)) > 0,
			Discapacity:         len(getValorSeguro(row, 42)) > 0,
			UniversityProfessor: len(getValorSeguro(row, 43)) > 0,

			// Solvencia y membresías
			DateOfLastSolvency:  parseDate(getValorSeguro(row, 44)),
			DoubleGuild:         len(getValorSeguro(row, 46)) > 0,
			DoubleGuildLocation: getValorSeguro(row, 46),
			CPSM:                strings.ToLower(getValorSeguro(row, 47)) == "aprobado",
		}

		// ── Postgrados (PsiUserPostGrade) ──────────────────────────────────
		var postgrades []domain.PsiUserPostGrade // Nota: usando valores o punteros según tu ORM

		if val := getValorSeguro(row, 33); len(val) > 0 && val != "-" {
			postgrades = append(postgrades, domain.PsiUserPostGrade{
				PsiUserID: psiID,
				Type:      domain.Diplomado,
				Title:     val,
			})
		}
		if val := getValorSeguro(row, 34); len(val) > 0 && val != "-" {
			postgrades = append(postgrades, domain.PsiUserPostGrade{
				PsiUserID: psiID,
				Type:      domain.Especializacion,
				Title:     val,
			})
		}
		if val := getValorSeguro(row, 35); len(val) > 0 && val != "-" {
			postgrades = append(postgrades, domain.PsiUserPostGrade{
				PsiUserID: psiID,
				Type:      domain.Maestria,
				Title:     val,
			})
		}
		if val := getValorSeguro(row, 36); len(val) > 0 && val != "-" { // Columna 36 = Doctorado
			postgrades = append(postgrades, domain.PsiUserPostGrade{
				PsiUserID: psiID,
				Type:      domain.Doctorado,
				Title:     val,
			})
		}

		// ── Solvencias (Mantenido de tu código anterior) ───────────────────
		solvencies := createSolvencieModel(parseDate(getValorSeguro(row, 44)), psi.ID, audit)

		// ── Persistencia transaccional ─────────────────────────────────────
		// IMPORTANTE: Asegúrate de que este método en tu Repositorio ahora acepte 'postgrades'
		// Ej: CreateFullProfile(ctx, psi, colData, solvencies, postgrades)
		if err := s.repo.CreateWithColData(ctx, psi, colData, solvencies, postgrades); err != nil {
			failedRecords = append(failedRecords, map[string]string{
				"fila":   excelRow,
				"nombre": fullName,
				"ci":     ciStr,
				"fpv":    numFPV,
				"error":  MapDBError(err).Error(), // asumiendo que MapDBError existe en tu código
			})
			continue // Saltamos al siguiente registro
		}

		// ── Notificación de bienvenida (no bloqueante) ────────────────────
		if email != "" && strings.Contains(email, "@") {
			mailData := map[string]interface{}{
				"Name":     psi.FirstName,
				"Email":    psi.Email,
				"Password": defaultPassword, // Se envía la clave temporal limpia al correo
			}
			if err := s.mailService.SendEmail(psi.Email, "Bienvenido a la plataforma Colegio de Psicólogos", "welcome_psi", mailData); err != nil {
				log.Printf("⚠️ Error al enviar correo de bienvenida a %s [%s]: %v", psi.Email, psi.Username, err)
			}
		}

		successCount++
	}

	return successCount, failedRecords
}
