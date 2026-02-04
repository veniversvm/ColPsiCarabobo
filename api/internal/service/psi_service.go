package service

import (
	"context"
	"encoding/csv"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"github.com/veniversvm/ColPsiCarabobo/api/pkg/s3"
	"golang.org/x/crypto/bcrypt"
)

type PsiService struct {
	repo     domain.PsiUserRepository
	s3Client *s3.S3Client
}

func NewPsiService(repo domain.PsiUserRepository, s3Client *s3.S3Client) *PsiService {
	return &PsiService{
		repo:     repo,
		s3Client: s3Client,
	}
}

func (s *PsiService) ImportFromCSV(ctx context.Context, reader io.Reader, adminID uuid.UUID) (int, []map[string]string) {
	csvReader := csv.NewReader(reader)
	_, _ = csvReader.Read() // Saltar cabeceras

	successCount := 0
	var failedRecords []map[string]string

	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}

		// 1. Hash de la contraseña (Seguridad Senior)
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(record[2]), bcrypt.DefaultCost)

		// 2. Mapeo del Modelo Principal
		psi := &domain.PsiUserModel{
			AuditModel: domain.AuditModel{
				CreateById: &adminID,
				CreateBy:   "Admin_CSV_Import",
			},
			Username:                 record[0],
			Email:                    record[1],
			Password:                 string(hashedPassword),
			FirstName:                record[3],
			SecondName:               record[4],
			LastName:                 record[5],
			SecondLastName:           record[6],
			FPV:                      parseInt(record[7]),
			CI:                       parseInt(record[8]),
			Nationality:              record[10],
			BornDate:                 parseDate(record[11]),
			Genre:                    record[12],
			ContactEmail:             record[13],
			ShowContactEmail:         parseBool(record[14]),
			PublicPhone:              record[15],
			ShowPublicPhone:          parseBool(record[16]),
			ServiceAddress:           record[17],
			ShowPublicServiceAddress: parseBool(record[18]),
			Solvent:                  parseBool(record[19]),
		}

		// 3. Mapeo de Datos Colegiales
		colData := &domain.PsiUserColData{
			AuditModel: domain.AuditModel{
				CreateById: &adminID,
			},
			UniversityUndergraduate:     record[27],
			ShowUniversityUndergraduate: parseBool(record[14]), // Ejemplo de reuso de lógica
			GraduateDate:                parseDate(record[28]),
			MentionUndergraduate:        record[29],
			RegisterTitleState:          record[30],
			RegisterTitleDate:           parseDate(record[31]),
			RegisterNumber:              parseInt(record[32]),
			RegisterFolio:               record[33],
			RegisterTome:                record[34],
			GuildDirector:               parseBool(record[35]),
			SixtyFiveOrPlus:             parseBool(record[36]),
			DateOfLastSolvency:          parseDate(record[40]),
		}

		// 4. Persistencia mediante el Repositorio (Transaccional)
		err = s.repo.CreateWithColData(ctx, psi, colData)
		if err != nil {
			failedRecords = append(failedRecords, map[string]string{
				"fila":      record[0],
				"identidad": record[8], // CI
				"error":     err.Error(),
			})
			continue
		}
		successCount++
	}

	return successCount, failedRecords
}

// --- HELPERS DE CONVERSIÓN (Privados) ---

func parseInt(s string) int {
	val, _ := strconv.Atoi(strings.TrimSpace(s))
	return val
}

func parseBool(s string) bool {
	s = strings.ToLower(strings.TrimSpace(s))
	return s == "true" || s == "1" || s == "v" || s == "s"
}

func parseDate(s string) time.Time {
	layout := "2006-01-02" // Formato estándar del CSV que pasaste
	t, err := time.Parse(layout, strings.TrimSpace(s))
	if err != nil {
		return time.Time{} // Fecha cero si falla
	}
	return t
}
