package service

import (
	"log"
	"strings"

	"github.com/google/uuid"
	domain "github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"github.com/xuri/excelize/v2"
	"golang.org/x/crypto/bcrypt"
)

type Colegiado struct {
	Numero                  string // [Nº]
	Colegio                 string // [Colegio]
	NumArchivador           string // [N° de Archivador]
	NumFPV                  string // [Nº FPV]
	FechaColegiatura        string // [Fecha Colegiatura]
	Nacionalidad            string // [Nacionalidad]
	CI                      string // [Nº C.I.]
	PrimerNombre            string // [1er Nombre]
	SegundoNombre           string // [2do Nombre]
	PrimerApellido          string // [1er Apellido]
	SegundoApellido         string // [2do Apellido]
	FechaNacimiento         string // [Fecha Nacimiento]
	Edad                    string // [Edad (actualizada al 2025]
	Genero                  string // [Género]
	FeDeVida                string // [Fe de Vida]
	Correo                  string // [Correo electrónico]
	Municipio               string // [Municipio]
	TelefonoLocal           string // [Teléfono Local]
	TelefonoMovil           string // [Teléfono Móvil]
	Estado                  string // [Estado]
	MunicipioTemp           string // [Municipio DATO TEMPORAL]
	TelefonoLocalTemp       string // [Teléfono Local] (Segundo)
	TelefonoMovilTemp       string // [Teléfono Móvil] (Segundo)
	Pais                    string // [Pais]
	TlfExterior             string // [Tlf Exterior]
	Universidad             string // [Univ]
	FechaGrado              string // [Fecha de Grado]
	Mencion                 string // [Mención]
	RegistroPrincipal       string // [Registro Principal del Estado]
	FechaRegistro           string // [Fecha Registro]
	NumRegistro             string // [N°] (Segundo)
	Folio                   string // [Folio]
	Tomo                    string // [Tomo]
	Diplomados              string // [Diplomados]
	Especializacion         string // [Especialización]
	Maestria                string // [Maestría]
	Doctorado               string // [Doctorado]
	EjercicioProfesional    string // [Ejercicio Profesional]
	Exoneracion100_1        string // [100% Exoneración de Solvencia] (Primera)
	Exoneracion100_2        string // [100% Exoneración de Solvencia] (Segunda)
	Exoneracion100_3        string // [100 % Exoneración de Solvencia] (Tercera)
	Exoneracion50_1         string // [50 % Exoneración de Solvencia] (Primera)
	Exoneracion50_2         string // [50 % Exoneración de Solvencia] (Segunda)
	Exoneracion50_3         string // [50 % Exoneración de Solvencia] (Tercera)
	FechaSolvencia          string // [Fecha hasta la cual esta solvente con ColPsiCarabobo]
	Estatus                 string // [Estatus]
	DobleColegiatura        string // [Doble Colegiatura]
	CertificacionTallerCPSM string // [Certificación Taller CPSM]
	Observaciones           string // [Observaciones]
	Deontologia             string // [Deontología]
}

func getValorSeguro(row []string, index int) string {
	if index < len(row) {
		return row[index]
	}
	return ""
}

func read_sheet(adminID uuid.UUID) {
	f, err := excelize.OpenFile("file.xlsx")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	// Asumiendo que la hoja se llama "Sheet1" o "Hoja1"
	rows, err := f.GetRows("BD ColPsiCarabobo 2026")
	if err != nil {
		log.Fatal(err)
	}

	var colegiados []domain.PsiUserModel

	// Si dices que la fila 2 tiene los encabezados, en Excel eso es la fila 2.
	// En el slice de Go (que empieza en 0), la fila 1 es index 0, fila 2 es index 1.
	// Por lo tanto, los datos REALES comienzan en la fila 3 de Excel (index 2 en Go).

	for i, row := range rows {
		// Omitir fila 0 y fila 1 (si la 2 es el encabezado)
		if i < 2 {
			continue
		}

		psiID := uuid.New()
		audit := domain.AuditModel{
			CreateById: &adminID,
			CreateBy:   "Admin_CSV_Import",
			UpdateById: &adminID,
			UpdateBy:   "Admin_CSV_Import",
		}

		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("Colpsi2025!"), bcrypt.DefaultCost)

		email := getValorSeguro(row, 15) // Campo Correo
		numFPV := getValorSeguro(row, 3) // numero de fpv

		var username string

		if strings.Contains(email, "@") {
			email_split := strings.Split(email, "@")
			username = email_split[0] + numFPV
		}

		// Mapeamos columna por columna a nuestro struct
		c := &domain.PsiUserModel{
			ID: psiID,
			// ── Credenciales de acceso ────────────────────────────────────────────
			AuditModel: audit,
			IsActive:   getValorSeguro(row, 45) == "Activo",
			Username:   username,
			Email:      email,
			Password:   string(hashedPassword),
			Key:        uuid.New().String(),

			// ── Identidad legal ───────────────────────────────────────────────────
			FirstName:      getValorSeguro(row, 7),
			SecondName:     getValorSeguro(row, 8),
			LastName:       getValorSeguro(row, 9),
			SecondLastName: getValorSeguro(row, 10),
			FPV:            parseInt(numFPV),
			CI:             parseInt(getValorSeguro(row, 6)),
			Nationality:    getValorSeguro(row, 5),
			BornDate:       parseDate(getValorSeguro(row, 11)),
			Genre:          getValorSeguro(row, 13),

			// ── Estado gremial y multimedia ───────────────────────────────────────

			ProofOfLife: strings.ToLower(getValorSeguro(row, 14)) != "fallecido",
			Solvent:     getValorSeguro(row, 45) == "Activo",

			// ── Contacto interno del gremio ────────────────────────────────────
			ContactPhone:     getValorSeguro(row, 17),
			ContactCellPhone: getValorSeguro(row, 18),

			// ── Contacto público y privacidad ────────────────────────────────────
			ContactEmail:     "",
			ShowContactEmail: false,

			// Contacto Publico dentro de Carabobo
			MunicipalityCarabobo:     getValorSeguro(row, 16),
			ShowMunicipalityCarabobo: false,
			PhoneCarabobo:            "",
			ShowPhoneCarabobo:        false,
			CelPhoneCarabobo:         "",
			ShowCelPhoneCarabobo:     false,

			// Informacion Fuera de Carabobo
			StateOutside:                getValorSeguro(row, 19),
			MunicipalityOutSideCarabobo: getValorSeguro(row, 20),
			PhoneOutSideCarabobo:        getValorSeguro(row, 21),
			CelPhoneOutSideCarabobo:     getValorSeguro(row, 22),

			// Fuera del pais
			Country:                   getValorSeguro(row, 23),
			PhoneOutSideVenezuela:     getValorSeguro(row, 24),
			ShowPhoneOutSideVenezuela: false,
			// celular y direccion quedan en blanco
		}

		colData := &domain.PsiUserColData{
			ID:                   uuid.New(),
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
			RegisterFolio:      getValorSeguro(row, 31),
			RegisterTome:       getValorSeguro(row, 32),

			// // Especializaciones
			// Diplomados:      getValorSeguro(row, 33),
			// Especializacion: getValorSeguro(row, 34),
			// Maestria:        getValorSeguro(row, 35),
			// Doctorado:       getValorSeguro(row, 36),
			// // EjercicioProfesional: getValorSeguro(row, 37),

			// ── Flags gremiales ───────────────────────────────────────────────────
			GuildDirector:       len(getValorSeguro(row, 38)) > 0,
			SixtyFiveOrPlus:     len(getValorSeguro(row, 39)) > 0, // crear funcion
			GuildCollaborator:   len(getValorSeguro(row, 40)) > 0,
			PublicEmployee:      len(getValorSeguro(row, 41)) > 0,
			Discapacity:         len(getValorSeguro(row, 42)) > 0,
			UniversityProfessor: len(getValorSeguro(row, 43)) > 0,

			// ── Solvencia y membresías ────────────────────────────────────────────
			DateOfLastSolvency:  parseDate(getValorSeguro(row, 44)),
			DoubleGuild:         len(getValorSeguro(row, 46)) > 0,
			DoubleGuildLocation: getValorSeguro(row, 46),
			CPSM:                strings.ToLower(getValorSeguro(row, 47)) == "aprobado",
			// Observaciones:        getValorSeguro(row, 48),
			// Deontologia:          getValorSeguro(row, 49),
		}

		// Agregamos este colegiado a la lista
		colegiados = append(colegiados, c)
	}

}
