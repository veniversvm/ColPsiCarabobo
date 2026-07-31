package service

import (
	"context"
	"fmt"
	"github.com/rs/zerolog/log"
	"io"
	"strings"

	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
	"golang.org/x/crypto/bcrypt"

	domain "github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/utils"
)

// getValorSeguro es una utilidad defensiva para el parseo de matrices.
// Previene el Panic ("index out of range") que ocurre comúnmente al parsear
// archivos CSV/XLSX si las últimas columnas de una fila están totalmente vacías
// y la librería las recorta de la longitud del array.
func getValorSeguro(row []string, index int) string {
	if index < len(row) {
		return strings.TrimSpace(row[index])
	}
	return ""
}

// ImportFromXLSX es el motor de ingesta masiva (Data Ingestion Engine) del sistema.
//
// Arquitectura de Tolerancia a Fallos (Fault Tolerance):
// Está diseñado para no detenerse ante errores individuales. Si la fila 5 falla
// por un error de formato o un FPV duplicado, el proceso la registra en `failedRecords`
// y continúa ininterrumpidamente con la fila 6. Al final, retorna un reporte
// consolidado para que el administrador pueda corregir los errores manualmente.
//
// Retorna:
// - int: Cantidad de registros importados exitosamente.
// - []map[string]string: Lista detallada de filas que fallaron con su respectivo motivo.
func (s *PsiService) ImportFromXLSX(ctx context.Context, reader io.Reader, adminID uuid.UUID) (int, []map[string]string) {
	// 1. Apertura del Stream Binario en Memoria (Virtual File System)
	f, err := excelize.OpenReader(reader)
	if err != nil {
		return 0, []map[string]string{
			{"error": "no se pudo abrir el archivo Excel: " + err.Error()},
		}
	}
	// Se garantiza la liberación de la memoria del archivo (GC) al terminar
	defer f.Close()

	// 2. Extracción de la Hoja Maestra (Hardcoded Name Requirement)
	rows, err := f.GetRows("BD ColPsiCarabobo 2026")
	if err != nil {
		return 0, []map[string]string{
			{"error": "no se pudo leer la hoja especificada: " + err.Error()},
		}
	}

	successCount := 0
	var failedRecords []map[string]string

	// 3. OPTIMIZACIÓN CRIPTOGRÁFICA (Performance Tuning)
	// bcrypt.GenerateFromPassword es una función diseñada matemáticamente para ser
	// extremadamente lenta (Mitigación de ataques Timing/Bruteforce).
	// Si un Excel tiene 5,000 psicólogos y generáramos la clave dentro del bucle for,
	// el proceso tardaría minutos. Al generar una única clave temporal predeterminada
	// *antes* del bucle, la importación masiva se reduce de minutos a segundos.
	defaultPassword := utils.GenerateSecureRandomString(12)
	hashedPasswordBytes, _ := bcrypt.GenerateFromPassword([]byte(defaultPassword), bcrypt.DefaultCost)
	hashedPassword := string(hashedPasswordBytes)

	// 4. Iteración y Parseo de Filas
	for i, row := range rows {
		// Omitir fila 0 (Título general) y fila 1 (Encabezados de columnas)
		if i < 2 {
			continue
		}

		// Registro visual de la fila para facilitar la auditoría de errores al usuario
		excelRow := fmt.Sprintf("%d", i+1)

		numFPV := getValorSeguro(row, 3)
		ciStr := getValorSeguro(row, 6)
		firstName := getValorSeguro(row, 7)
		lastName := getValorSeguro(row, 9)
		fullName := firstName + " " + lastName

		// Detección de Colas Vacías (Ghost Rows):
		// Los usuarios de Excel suelen dejar filas con formato pero sin datos al final del documento.
		if numFPV == "" && ciStr == "" {
			continue
		}

		// ── Sanitización Semántica: Geo-normalización ───────────────────────
		// Transforma entradas de texto libre humano (ej. "valencia", "VALENCIA ")
		// en IDs o nomenclaturas estandarizadas por el sistema.
		var municipioCarabobo string
		if raw := getValorSeguro(row, 16); raw != "" && raw != "-" {
			mun, ok := utils.NormalizeMunicipioCarabobo(raw)
			if !ok {
				// Fallo de Integridad de Datos: Se registra y aborta esta fila
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

		// ── Asignación de Trazabilidad Criptográfica ─────────────────────────
		psiID := uuid.Must(uuid.NewV7())
		audit := domain.AuditModel{
			CreateById: &adminID,
			CreateBy:   "Admin_XLSX_Import", // Firma identificativa del proceso
			UpdateById: &adminID,
			UpdateBy:   "Admin_XLSX_Import",
		}

		// ── Estrategia de Resolución de Identidad (Username Generation) ───────
		email := getValorSeguro(row, 15)
		var username string
		if strings.Contains(email, "@") {
			// Previene colisiones usando la combinación de (nombre de correo + FPV)
			email_split := strings.Split(email, "@")
			username = email_split[0] + numFPV
		} else {
			// Fallback determinístico si el registro no tiene email
			username = "psi" + numFPV + ciStr
		}

		// ── Ensamblaje del Grafo del Modelo de Dominio ──────────────────────

		// 1. Modelo de Identidad Principal
		psi := &domain.PsiUserModel{
			ID:         psiID,
			AuditModel: audit,
			Credentials: domain.Credentials{
				Key:                uuid.Must(uuid.NewV7()).String(),
				IsActive:           true,
				Username:           username,
				Email:              email,
				Password:           hashedPassword,
				MustChangePassword: true,
			},

			FirstName:      firstName,
			SecondName:     cleanDash(getValorSeguro(row, 8)),
			LastName:       lastName,
			SecondLastName: cleanDash(getValorSeguro(row, 10)),
			FPV:            parseInt(numFPV),
			CI:             parseInt(ciStr),
			Nationality:    parseNationality(getValorSeguro(row, 5)),
			ControlNumber:  cleanDash(getValorSeguro(row, 0)),
			BornDate:       parseDate(getValorSeguro(row, 11)),
			Genre:          getValorSeguro(row, 13),

			ProofOfLife: isProofOfLife(getValorSeguro(row, 14)),
			Solvent:     isSolventByLastSolvency(parseDate(getValorSeguro(row, 44))),

			ContactPhone:     cleanDash(getValorSeguro(row, 17)),
			ContactCellPhone: cleanDash(getValorSeguro(row, 18)),

			// Por defecto de privacidad (Opt-in privacy), todos los escudos están activos (false)
			ContactEmail:     "",
			ShowContactEmail: false,

			MunicipalityCarabobo:     municipioCarabobo,
			ShowMunicipalityCarabobo: false,
			PhoneCarabobo:            "",
			ShowPhoneCarabobo:        false,
			CelPhoneCarabobo:         "",
			ShowCelPhoneCarabobo:     false,

			StateOutside:                cleanDash(estadoOutside),
			MunicipalityOutSideCarabobo: cleanDash(getValorSeguro(row, 20)),
			PhoneOutSideCarabobo:        cleanDash(getValorSeguro(row, 21)),
			CelPhoneOutSideCarabobo:     cleanDash(getValorSeguro(row, 22)),

			Country:                   cleanDash(getValorSeguro(row, 23)),
			PhoneOutSideVenezuela:     cleanDash(getValorSeguro(row, 24)),
			ShowPhoneOutSideVenezuela: false,
		}

		// 2. Modelo de Trazabilidad Gremial (Datos Institucionales)
		colData := &domain.PsiUserColData{
			ID:                   uuid.Must(uuid.NewV7()),
			PsiUserModelID:       psiID,
			AuditModel:           audit,
			GuildInscriptionDate: parseDate(getValorSeguro(row, 4)),

			UniversityUndergraduate: getValorSeguro(row, 25),
			GraduateDate:            parseDate(getValorSeguro(row, 26)),
			MentionUndergraduate:    getValorSeguro(row, 27),

			RegisterTitleState: getValorSeguro(row, 28),
			RegisterTitleDate:  parseDate(getValorSeguro(row, 29)),
			RegisterNumber:     parseInt(getValorSeguro(row, 30)),
			RegisterFolio:      cleanDash(getValorSeguro(row, 31)),
			RegisterTome:       cleanDash(getValorSeguro(row, 32)),

			GuildDirector:       len(getValorSeguro(row, 38)) > 0,
			SixtyFiveOrPlus:     len(getValorSeguro(row, 39)) > 0,
			GuildCollaborator:   len(getValorSeguro(row, 40)) > 0,
			PublicEmployee:      len(getValorSeguro(row, 41)) > 0,
			Discapacity:         len(getValorSeguro(row, 42)) > 0,
			UniversityProfessor: len(getValorSeguro(row, 43)) > 0,

			DateOfLastSolvency:  parseDate(getValorSeguro(row, 44)),
			DoubleGuild:         len(getValorSeguro(row, 46)) > 0,
			DoubleGuildLocation: getValorSeguro(row, 46),
			CPSM:                strings.ToLower(getValorSeguro(row, 47)) == "aprobado",
		}

		// 3. Relaciones 1:N (Postgrados Académicos)
		var postgrades []domain.PsiUserPostGrade

		// Evaluación condicional para inyectar al array solo si existe información en la celda
		if val := getValorSeguro(row, 33); len(val) > 0 && val != "-" {
			postgrades = append(postgrades, domain.PsiUserPostGrade{
				PsiUserID: psiID, Type: domain.Diplomado, Title: val,
			})
		}
		if val := getValorSeguro(row, 34); len(val) > 0 && val != "-" {
			postgrades = append(postgrades, domain.PsiUserPostGrade{
				PsiUserID: psiID, Type: domain.Especializacion, Title: val,
			})
		}
		if val := getValorSeguro(row, 35); len(val) > 0 && val != "-" {
			postgrades = append(postgrades, domain.PsiUserPostGrade{
				PsiUserID: psiID, Type: domain.Maestria, Title: val,
			})
		}
		if val := getValorSeguro(row, 36); len(val) > 0 && val != "-" {
			postgrades = append(postgrades, domain.PsiUserPostGrade{
				PsiUserID: psiID, Type: domain.Doctorado, Title: val,
			})
		}

		// 4. Historial Financiero
		solvencies := buildSolvencyHistory(parseDate(getValorSeguro(row, 44)), parseDate(getValorSeguro(row, 4)).Year(), psi.ID, audit)

		// ── Persistencia Transaccional Estricta (ACID) ──────────────────────
		// Ejecuta un "Bulk Insert" relacional. Si ocurre un fallo (ej. Violación de restricción UNIQUE en FPV),
		// la base de datos realiza un Rollback atómico de este usuario específico.
		if err := s.repo.CreateWithColData(ctx, psi, colData, solvencies, postgrades); err != nil {
			// El MapDBError actúa como escudo para no exponer metadatos de Postgres al Frontend.
			failedRecords = append(failedRecords, map[string]string{
				"fila":   excelRow,
				"nombre": fullName,
				"ci":     ciStr,
				"fpv":    numFPV,
				"error":  MapDBError(err).Error(),
			})
			continue // Aborta la iteración de esta fila y salta a la siguiente sin pánico.
		}

		// ── Evento Asíncrono de Onboarding ──────────────────────────────────
		// Si la transacción fue exitosa y el usuario posee correo válido, se le envía su acceso.
		// Al ser gestionado por un Worker asíncrono (s.mailService), esto no ralentiza el ciclo FOR de lectura.
		if email != "" && strings.Contains(email, "@") {
			mailData := map[string]interface{}{
				"Name":     psi.FirstName,
				"Email":    psi.Email,
				"Password": defaultPassword, // Envío de credencial temporal en plano (Plain-text transitorio)
			}
			if err := s.mailService.SendEmail(psi.Email, "Bienvenido a la plataforma Colegio de Psicólogos", "welcome_psi", mailData); err != nil {
				log.Warn().Err(err).Str("component", "psi_service_xlsx").Str("user_id", psi.ID.String()).Msg("Error sending welcome email")
			}
		}

		// Conteo exitoso finalizado para este ciclo
		successCount++
	}

	return successCount, failedRecords
}
