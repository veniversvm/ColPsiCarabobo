// api/internal/service/birthdays.go

// Package service implementa la lógica de negocio central de la aplicación.
//
// Este archivo gestiona el aviso de cumpleaños al personal administrativo:
// dado el opt-in del psicólogo (birthday_notification), el admin recibe
// un recordatorio de quiénes cumplen años hoy o en los próximos días.
package service

import (
	"context"
	"time"
)

// BirthdayInfoProjection agrupa la información mínima que se envía al admin
// en el aviso de cumpleaños (sin exponer datos sensibles de contacto).
type BirthdayInfoProjection struct {
	ID        string `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	FPV       int    `json:"fpv"`
	IsActive  bool   `json:"is_active"`
	// Mes y día del cumpleaños (para mostrarlos como fecha sin exponer el año real).
	Month int `json:"month"`
	Day   int `json:"day"`
}

// GetBirthdaysByRange recupera los agremiados que cumplen años en el rango [from;to]
// y que autorizaron el aviso. Solo devuelve metadatos mínimos para el banner del admin.
func (s *PsiService) GetBirthdaysByRange(ctx context.Context, from, to time.Time) ([]BirthdayInfoProjection, error) {
	users, err := s.repo.GetBirthdays(ctx, from, to)
	if err != nil {
		return nil, err
	}

	out := make([]BirthdayInfoProjection, 0, len(users))
	for _, u := range users {
		out = append(out, BirthdayInfoProjection{
			ID:        u.ID.String(),
			FirstName: u.FirstName,
			LastName:  u.LastName,
			FPV:       u.FPV,
			IsActive:  u.IsActive,
			Month:     int(u.BornDate.Month()),
			Day:       u.BornDate.Day(),
		})
	}
	return out, nil
}
