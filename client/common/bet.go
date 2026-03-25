package common

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
)

type BetInfo struct {
	nombre     string
	apellido   string
	dni        string
	nacimiento string
	numero     string
}

var nacimientoRegex = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)

func NewBetFromEnv() (BetInfo, error) {
	nombre := os.Getenv("NOMBRE")
	apellido := os.Getenv("APELLIDO")
	dni := os.Getenv("DOCUMENTO")
	nacimiento := os.Getenv("NACIMIENTO")
	numero := os.Getenv("NUMERO")

	if _, err := strconv.Atoi(dni); err != nil {
		return BetInfo{}, fmt.Errorf("DOCUMENTO inválido: %q", dni)
	}
	if _, err := strconv.Atoi(numero); err != nil {
		return BetInfo{}, fmt.Errorf("NUMERO debe ser un entero: %q", numero)
	}
	if !nacimientoRegex.MatchString(nacimiento) {
		return BetInfo{}, fmt.Errorf("NACIMIENTO debe estar en formato YYYY-MM-DD: %q", nacimiento)
	}

	return BetInfo{
		nombre:     nombre,
		apellido:   apellido,
		dni:        dni,
		nacimiento: nacimiento,
		numero:     numero,
	}, nil
}
