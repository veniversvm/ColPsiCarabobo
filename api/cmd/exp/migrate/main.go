// api/cmd/exp/migrate/main.go
package main

import (
	"io"
	"os"

	"ariga.io/atlas-provider-gorm/gormschema"
	"github.com/rs/zerolog/log"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
)

func main() {
	// Aquí cargamos todos los modelos que definimos en domain/models.go
	stmts, err := gormschema.New("postgres").Load(
		&domain.TextModel{},
		&domain.UserAdmin{},
		&domain.PsiUserModel{},
		&domain.PsiUserColData{},
		&domain.PsiUserPostGrade{},
		&domain.PsiUserDocument{},
		&domain.PsiUserSolvency{},
		&domain.PsiInscriptionRequest{},
		&domain.PsiInscriptionDocument{},
		&domain.Post{},
		&domain.PsiSpecialtyModel{},
		&domain.PsiUserSocialNetwork{},
		&domain.PsiODeontologia{},
		&domain.PsiObservations{},
		&domain.LoginEvent{},
		&domain.PageView{},
		&domain.SearchEvent{},
		&domain.ProfileView{},
		&domain.ActiveSession{},
		&domain.Notification{},
		&domain.NotificationTarget{},
		&domain.NotificationFilter{},
		&domain.NotificationAttach{},
		&domain.KanbanProject{},
		&domain.KanbanMember{},
		&domain.KanbanColumn{},
		&domain.KanbanCard{},
		&domain.KanbanNote{},
&domain.TicketMotivo{},
		&domain.TicketEstado{},
		&domain.Ticket{},
		&domain.TicketStatusLog{},
		&domain.TicketMensaje{},
		&domain.TicketAdjunto{},
		&domain.AdminPermissionLog{},
		&domain.AppSetting{},
		&domain.SettingsAuditLog{},
	)
	if err != nil {
		log.Fatal().Err(err).Str("component", "migrate").Msg("Failed to load gorm schema")
	}
	// Esto imprime el esquema SQL que Atlas comparará
	io.WriteString(os.Stdout, stmts)
}
