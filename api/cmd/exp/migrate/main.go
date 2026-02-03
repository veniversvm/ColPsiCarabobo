package main

import (
	"io"
	"log"
	"os"

	"ariga.io/atlas-provider-gorm/gormschema"
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
		&domain.Post{},
	)
	if err != nil {
		log.Fatalf("failed to load gorm schema: %v", err)
	}
	// Esto imprime el esquema SQL que Atlas comparará
	io.WriteString(os.Stdout, stmts)
}
