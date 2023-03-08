package payment

import (
	"sample/domain"
)

type Repository interface {
	domain.Repository[Payment]
}
