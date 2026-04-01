package types

import "github.com/fishkaoff/ts-backend/internal/domain/lib/utils"

type ProductUnit string

const Unit ProductUnit = "штука"
const Set ProductUnit = "компл"

func ParseUnit(u string) ProductUnit {
	switch utils.NormalizeString(u) {
	case "шт":
		return Unit
	case "компл":
		return Set
	default:
		return Unit
	}
}
