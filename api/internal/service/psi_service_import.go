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
	"golang.org/x/crypto/bcrypt"
	"github.com/xuri/excelize/v2"

	"github.com/veniversvm/ColPsiCarabobo/api/internal/config"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/utils"
)

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

	rows, _ := f.Rows("BD ColPsiCarabobo 2026")

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
		row, _ := rows.Columns()
		if rowIdx <= 2 {
			continue
		}

		rawFPV := getValorSeguro(row, 3)
		rawCI := getValorSeguro(row, 6)
		firstName := getValorSeguro(row, 7)
		lastName := getValorSeguro(row, 9)
		fullName := firstName + " " + lastName

		fpvInt := parseInt(rawFPV)
		ciInt := parseInt(rawCI)

		if fpvInt == 0 || ciInt == 0 || firstName == "" {
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

		psi := &domain.PsiUserModel{
			ID: psiID, AuditModel: audit,
			Credentials: domain.Credentials{
				Key:                sessionKey,
				Username:           generateSecureUsername(emailToProcess, strconv.Itoa(fpvInt), firstName),
				Email:              emailToProcess, Password: hashedPassword,
				IsActive:           getValorSeguro(row, 45) == "Activo",
				MustChangePassword: true,
			},
			AudioBookShellId: psiID.String(),
			FirstName:        firstName, LastName: lastName,
			FPV: fpvInt, CI: ciInt, BornDate: parseDate(getValorSeguro(row, 11)),
			Genre:          getValorSeguro(row, 13),
			Solvent:              getValorSeguro(row, 45) == "Activo",
			ProofOfLife:          strings.ToLower(getValorSeguro(row, 14)) != "fallecido",
			ContactPhone:         cleanDash(getValorSeguro(row, 17)),
			ContactCellPhone:     cleanDash(getValorSeguro(row, 18)),
			ContactEmail:         email,
			MunicipalityCarabobo: getValorSeguro(row, 16),
		}

		colData := &domain.PsiUserColData{
			ID: uuid.Must(uuid.NewV7()), PsiUserModelID: psiID, AuditModel: audit,
			GuildInscriptionDate:    parseDate(getValorSeguro(row, 4)),
			UniversityUndergraduate: getValorSeguro(row, 25),
			GraduateDate:            parseDate(getValorSeguro(row, 26)),
			MentionUndergraduate:    getValorSeguro(row, 27),
			RegisterNumber:          parseInt(getValorSeguro(row, 30)),
			DateOfLastSolvency:      parseDate(getValorSeguro(row, 44)),
		}

		solvency := domain.PsiUserSolvency{
			ID: uuid.Must(uuid.NewV7()), PsiUserModelID: psiID, AuditModel: audit, Date: colData.DateOfLastSolvency,
		}

		if err := s.repo.CreateWithColData(ctx, psi, colData, solvency, []domain.PsiUserPostGrade{}); err != nil {
			humanError := MapDBError(err).Error()
			auditLogger.Printf("[ERROR] FILA %d | %s | %v", rowIdx, fullName, humanError)
			failedRecords = append(failedRecords, map[string]string{"fila": strconv.Itoa(rowIdx), "nombre": fullName, "error": humanError})
			continue
		}

		if psi.ProofOfLife && validEmail {
			go s.mailService.SendEmail(
				psi.Email,
				"Bienvenido(a) a la plataforma COLPSI Carabobo",
				"welcome_psi",
				map[string]interface{}{
					"Name":     psi.FirstName,
					"Email":    psi.Email,
					"Password": defaultPassword,
				},
			)
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

func parseDate(val string) time.Time {
	if val == "" || val == "-" || val == "0" {
		return time.Time{}
	}
	layouts := []string{
		"02/01/2006", "02-01-2006", "2006-01-02",
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
