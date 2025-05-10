package models

import (
	"time"

	"promoservice/proto/promo"

	"google.golang.org/protobuf/types/known/timestamppb"
)

type Promo struct {
	ID             string
	Name           string
	Description    string
	CreatorID      string
	DiscountAmount float64
	Code           string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	IsActive       bool
	ValidUntil     time.Time
	MaxUses        int32
	CurrentUses    int32
}

func (p *Promo) ToProto() *promo.Promo {
	return &promo.Promo{
		Id:             p.ID,
		Name:           p.Name,
		Description:    p.Description,
		CreatorId:      p.CreatorID,
		DiscountAmount: p.DiscountAmount,
		Code:           p.Code,
		CreatedAt:      timestamppb.New(p.CreatedAt),
		UpdatedAt:      timestamppb.New(p.UpdatedAt),
		IsActive:       p.IsActive,
		ValidUntil:     timestamppb.New(p.ValidUntil),
		MaxUses:        p.MaxUses,
		CurrentUses:    p.CurrentUses,
	}
}

func PromoFromProto(p *promo.Promo) *Promo {
	return &Promo{
		ID:             p.Id,
		Name:           p.Name,
		Description:    p.Description,
		CreatorID:      p.CreatorId,
		DiscountAmount: p.DiscountAmount,
		Code:           p.Code,
		CreatedAt:      p.CreatedAt.AsTime(),
		UpdatedAt:      p.UpdatedAt.AsTime(),
		IsActive:       p.IsActive,
		ValidUntil:     p.ValidUntil.AsTime(),
		MaxUses:        p.MaxUses,
		CurrentUses:    p.CurrentUses,
	}
}
