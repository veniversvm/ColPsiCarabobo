package service

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
	"golang.org/x/crypto/bcrypt"

	"github.com/veniversvm/ColPsiCarabobo/api/internal/config"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/utils"
)

// ImportFromCSV importa registros de psicólogos desde el Excel maestro del Colegio
// ("BD ColPsiCarabobo 2026"), creando usuarios y enviando emails de bienvenida.
//
// Mapeo de columnas (índice 0-based, filas de datos a partir de la 3):
//
//	0  Nº (número de control interno)        3  Nº FPV         4  Fecha Colegiatura
//	5  Nacionalidad                          6  Nº C.I.        7  1er Nombre
//	8  2do Nombre                            9  1er Apellido  10  2do Apellido
//	11 Fecha Nacimiento                     13  Género        14  Fe de Vida
//	15 Correo electrónico                   16  Municipio     17  Teléfono Local
//	18 Teléfono Móvil                       19  Estado        20  Municipio (fuera)
//	21 Tel. Local (fuera)                   22  Tel. Móvil (fuera)
//	23 País (exterior)                      24  Tlf. Exterior
//	25 Univ                                 26  Fecha de Grado 27 Mención
//	28 Registro Principal del Estado        29  Fecha Registro
//	30 N° registro                          31  Folio          32 Tomo
//	33 Diplomados                           34  Especialización
//	35 Maestría                             36  Doctorado
//	37 Ejercicio Profesional                38-43 Flags gremiales
//	44 Fecha hasta la cual está solvente    46  Doble Colegiatura
//	47 Certificación Taller CPSM
//
// Las fechas pueden venir como seriales de Excel (días desde 1899-12-30) o como
// texto; parseDate resuelve ambos formatos.
func (s *PsiService) ImportFromCSV(ctx context.Context, reader io.Reader, adminID uuid.UUID) (int, []map[string]string) {
	_ = os.Mkdir("logs", 0755)
	logFileName := fmt.Sprintf("logs/import_%s.log", time.Now().Format("2006-01-02_15-04-05"))
	logFile, _ := os.OpenFile(logFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	defer logFile.Close()
	auditLogger := log.New(logFile, "", log.LstdFlags)

	f, err := excelize.OpenReader(reader)
	if err != nil {
		return 0, []map[string]string{{"error": "archivo inválido"}}
	}
	defer f.Close()

	rows, err := f.Rows("BD ColPsiCarabobo 2026")
	if err != nil {
		return 0, []map[string]string{{"error": "no se pudo leer la hoja del archivo"}}
	}
	defer rows.Close()

	successCount := 0
	var failedRecords []map[string]string
	var defaultPassword string
	if config.Envs.Environment == "development" {
		defaultPassword = "Colpsi2025!"
	} else {
		defaultPassword = utils.GenerateSecureRandomString(16)
	}
	hashedPasswordBytes, _ := bcrypt.GenerateFromPassword([]byte(defaultPassword), bcrypt.DefaultCost)
	hashedPassword := string(hashedPasswordBytes)

	audit := domain.AuditModel{
		CreateById: &adminID, CreateBy: "Admin_XLSX_Import",
		UpdateById: &adminID, UpdateBy: "Admin_XLSX_Import",
	}

	rowIdx := 0
	for rows.Next() {
		rowIdx++

		// Si el contexto se canceló (timeout del import o request abortada),
		// abortamos el resto del archivo: cada transacción restante fallaría
		// igual con el ctx cancelado y solo quemaríamos CPU/lectura.
		if ctx.Err() != nil {
			failedRecords = append(failedRecords, map[string]string{
				"error": "importación cancelada por timeout",
			})
			break
		}

		row, _ := rows.Columns()
		if rowIdx <= 2 {
			continue
		}

		rawControl := getValorSeguro(row, 0)
		rawFPV := getValorSeguro(row, 3)
		rawCI := getValorSeguro(row, 6)
		firstName := getValorSeguro(row, 7)
		lastName := getValorSeguro(row, 9)
		fullName := firstName + " " + lastName

		// Detección de filas fantasma (formato heredado sin datos al final del libro).
		if rawControl == "" && rawFPV == "" && rawCI == "" && firstName == "" && lastName == "" {
			continue
		}

		fpvInt := parseInt(rawFPV)
		ciInt := parseInt(rawCI)

		if fpvInt == 0 || ciInt == 0 || firstName == "" || lastName == "" {
			failedRecords = append(failedRecords, map[string]string{"fila": strconv.Itoa(rowIdx), "nombre": fullName, "error": "Datos incompletos"})
			continue
		}

		email := getValorSeguro(row, 15)
		emailToProcess := email
		validEmail := true
		if email == "" || !strings.Contains(email, "@") {
			emailToProcess = fmt.Sprintf("%d.sincorreo@colpsi.com", fpvInt)
			validEmail = false
		}

		psiID := uuid.Must(uuid.NewV7())
		sessionKey := uuid.Must(uuid.NewV7()).String()

		// Geo-normalización tolerante: si el texto no calza con el catálogo se
		// conserva tal cual (la importación masiva no debe abortar por un típico).
		municipioCarabobo := getValorSeguro(row, 16)
		if mun, ok := utils.NormalizeMunicipioCarabobo(municipioCarabobo); ok {
			municipioCarabobo = mun
		}
		estadoOutside := getValorSeguro(row, 19)
		if estado, ok := utils.NormalizeEstadoVenezuela(estadoOutside); ok {
			estadoOutside = estado
		}

		bornDate := parseDate(getValorSeguro(row, 11))
		inscriptionDate := parseDate(getValorSeguro(row, 4))
		graduateDate := parseDate(getValorSeguro(row, 26))
		registerDate := parseDate(getValorSeguro(row, 29))
		lastSolvencyDate := parseDate(getValorSeguro(row, 44))

		psi := &domain.PsiUserModel{
			ID: psiID, AuditModel: audit,
			Credentials: domain.Credentials{
				Key:      sessionKey,
				Username: generateSecureUsername(emailToProcess, strconv.Itoa(fpvInt), firstName),
				Email:    emailToProcess, Password: hashedPassword,
				// Para el lanzamiento todos los registros importados quedan activos.
				IsActive:           true,
				MustChangePassword: true,
			},
			AudioBookShellId: psiID.String(),
			FirstName:        firstName, SecondName: cleanDash(getValorSeguro(row, 8)),
			LastName: lastName, SecondLastName: cleanDash(getValorSeguro(row, 10)),
			FPV: fpvInt, CI: ciInt,
			Nationality:   parseNationality(getValorSeguro(row, 5)),
			ControlNumber: cleanDash(rawControl),
			BornDate:      bornDate,
			Genre:         getValorSeguro(row, 13),
			// Solvente solo si la última solvencia registrada es del año vigente.
			Solvent:                     isSolventByLastSolvency(lastSolvencyDate),
			ProofOfLife:                 isProofOfLife(getValorSeguro(row, 14)),
			ContactPhone:                cleanDash(getValorSeguro(row, 17)),
			ContactCellPhone:            cleanDash(getValorSeguro(row, 18)),
			ContactEmail:                email,
			MunicipalityCarabobo:        municipioCarabobo,
			StateOutside:                cleanDash(estadoOutside),
			MunicipalityOutSideCarabobo: cleanDash(getValorSeguro(row, 20)),
			PhoneOutSideCarabobo:        cleanDash(getValorSeguro(row, 21)),
			CelPhoneOutSideCarabobo:     cleanDash(getValorSeguro(row, 22)),
			Country:                     cleanDash(getValorSeguro(row, 23)),
			PhoneOutSideVenezuela:       cleanDash(getValorSeguro(row, 24)),
			PrimaryWorkArea:             getValorSeguro(row, 37),
		}

		colData := &domain.PsiUserColData{
			ID: uuid.Must(uuid.NewV7()), PsiUserModelID: psiID, AuditModel: audit,
			GuildInscriptionDate:    inscriptionDate,
			UniversityUndergraduate: getValorSeguro(row, 25),
			GraduateDate:            graduateDate,
			MentionUndergraduate:    getValorSeguro(row, 27),
			RegisterTitleState:      getValorSeguro(row, 28),
			RegisterTitleDate:       registerDate,
			RegisterNumber:          parseInt(getValorSeguro(row, 30)),
			RegisterFolio:           cleanDash(getValorSeguro(row, 31)),
			RegisterTome:            cleanDash(getValorSeguro(row, 32)),
			GuildDirector:           len(getValorSeguro(row, 38)) > 0,
			SixtyFiveOrPlus:         len(getValorSeguro(row, 39)) > 0,
			GuildCollaborator:       len(getValorSeguro(row, 40)) > 0,
			PublicEmployee:          len(getValorSeguro(row, 41)) > 0,
			Discapacity:             len(getValorSeguro(row, 42)) > 0,
			UniversityProfessor:     len(getValorSeguro(row, 43)) > 0,
			DateOfLastSolvency:      lastSolvencyDate,
			DoubleGuild:             len(getValorSeguro(row, 46)) > 0,
			DoubleGuildLocation:     cleanDash(getValorSeguro(row, 46)),
			CPSM:                    strings.ToLower(getValorSeguro(row, 47)) == "aprobado",
		}

		var postgrades []domain.PsiUserPostGrade
		if val := getValorSeguro(row, 33); val != "" && val != "-" {
			postgrades = append(postgrades, domain.PsiUserPostGrade{PsiUserID: psiID, Type: domain.Diplomado, Title: val})
		}
		if val := getValorSeguro(row, 34); val != "" && val != "-" {
			postgrades = append(postgrades, domain.PsiUserPostGrade{PsiUserID: psiID, Type: domain.Especializacion, Title: val})
		}
		if val := getValorSeguro(row, 35); val != "" && val != "-" {
			postgrades = append(postgrades, domain.PsiUserPostGrade{PsiUserID: psiID, Type: domain.Maestria, Title: val})
		}
		if val := getValorSeguro(row, 36); val != "" && val != "-" {
			postgrades = append(postgrades, domain.PsiUserPostGrade{PsiUserID: psiID, Type: domain.Doctorado, Title: val})
		}

		// Historial de solvencias anuales desde la colegiatura (mín. 2024) hasta
		// el año de la última solvencia. Vació si no hay fecha válida.
		solvencies := buildSolvencyHistory(lastSolvencyDate, inscriptionDate.Year(), psiID, audit)

		if err := s.repo.CreateWithColData(ctx, psi, colData, solvencies, postgrades); err != nil {
			humanError := MapDBError(err).Error()
			auditLogger.Printf("[ERROR] FILA %d | %s | %v", rowIdx, fullName, humanError)
			failedRecords = append(failedRecords, map[string]string{"fila": strconv.Itoa(rowIdx), "nombre": fullName, "error": humanError})
			continue
		}

		// Envío NO BLOQUEANTE: SendEmail solo encola en memoria; el worker
		// (SMTP o Resend) despacha en background con su propio throttling.
		if s.mailService != nil && psi.ProofOfLife && validEmail {
			if err := s.mailService.SendEmail(
				psi.Email,
				"Bienvenido(a) a la plataforma COLPSI Carabobo",
				"welcome_psi",
				map[string]interface{}{
					"Name":     psi.FirstName,
					"Email":    psi.Email,
					"Password": defaultPassword,
				},
			); err != nil {
				log.Printf("❌ Error al encolar email para %s: %v", maskEmail(psi.Email), err)
			}
		}

		successCount++
	}

	return successCount, failedRecords
}

func parseInt(val string) int {
	if val == "" || val == "-" {
		return 0
	}
	clean := ""
	for _, r := range val {
		if r >= '0' && r <= '9' {
			clean += string(r)
		} else if r == ',' || r == '.' {
			continue
		}
	}

	i, _ := strconv.Atoi(clean)
	return i
}

func generateSecureUsername(email, fpv, name string) string {
	base := ""
	if strings.Contains(email, "@") {
		base = strings.Split(email, "@")[0]
	} else {
		base = strings.ReplaceAll(strings.ToLower(name), " ", "")
	}

	cleanFPV := strings.NewReplacer(",", "", ".", "", " ", "").Replace(fpv)

	combined := base + cleanFPV
	if len(combined) > 25 {
		maxBase := 25 - len(cleanFPV)
		if maxBase > 0 {
			combined = base[:maxBase] + cleanFPV
		} else {
			combined = combined[:25]
		}
	}
	return combined
}

func parseBool(s string) bool {
	s = strings.ToLower(strings.TrimSpace(s))
	return s == "true" || s == "1" || s == "v" || s == "s"
}

// parseNationality normaliza la nacionalidad a "V" o "E", con "V" por defecto.
// En el Excel maestro la columna no incluye la letra junto a la cédula, pero la
// columna 'Nacionalidad' puede venir vacía en archivos antiguos.
func parseNationality(val string) string {
	switch strings.ToUpper(strings.TrimSpace(val)) {
	case "E":
		return "E"
	default:
		return "V"
	}
}

// isProofOfLife interpreta la columna 'Fe de Vida' (SI/NO). Solo valores
// afirmativos producen true; "NO", "Fallecido" o vacío devuelven false.
func isProofOfLife(val string) bool {
	switch strings.ToLower(strings.TrimSpace(val)) {
	case "si", "sí", "s", "1", "true", "x", "si.", "sí.", "vigente":
		return true
	default:
		return false
	}
}

// isSolventByLastSolvency define la regla de solvencia del lanzamiento: un
// colegiado es solvente únicamente si su última solvencia registrada es del año
// vigente (o posterior).
func isSolventByLastSolvency(last time.Time) bool {
	return !last.IsZero() && last.Year() >= time.Now().Year()
}

// parseDate interpreta fechas en formato texto (dd/mm/aaaa, aaaa-mm-dd, etc.) o
// como serial de Excel (días transcurridos desde 1899-12-30). El Excel maestro
// del Colegio guarda las fechas como seriales numéricos con formato "General".
func parseDate(val string) time.Time {
	val = strings.TrimSpace(val)
	if val == "" || val == "-" || val == "0" {
		return time.Time{}
	}

	// Serial de Excel: 45852 → 2025-07-14.
	if serial, err := strconv.ParseFloat(val, 64); err == nil && serial >= 1 && serial < 100000 {
		return time.Date(1899, 12, 30, 0, 0, 0, 0, time.UTC).AddDate(0, 0, int(serial))
	}

	layouts := []string{
		"1/2/2006", "1-2-2006", "2006-01-02",
		"02/01/2006", "02-01-2006",
		"1/2/06", "1-2-06", "01-02-06",
	}
	for _, l := range layouts {
		t, err := time.Parse(l, val)
		if err == nil {
			return t
		}
	}
	return time.Time{}
}

func cleanDash(val string) string {
	if val == "-" || val == "0" {
		return ""
	}
	return val
}
