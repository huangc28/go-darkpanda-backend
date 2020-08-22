package inquiry

import (
	"fmt"

	"github.com/huangc28/go-darkpanda-backend/internal/models"
	"github.com/looplab/fsm"
)

type InquiryActions string

var (
	Cancel InquiryActions = "cancel"
	Book   InquiryActions = "book"
	Expire InquiryActions = "expire"
)

func (a *InquiryActions) ToString() string {
	return string(*a)
}

func NewInquiryFSM(initial models.InquiryStatus) (*fsm.FSM, error) {
	if !initial.IsValid() {
		return nil, fmt.Errorf("The initial inquiry state: %s is invalid", initial.ToString())
	}

	f := fsm.NewFSM(
		initial.ToString(),
		fsm.Events{
			{
				Name: Cancel.ToString(),
				Src: []string{
					string(models.InquiryStatusInquiring),
				},
				Dst: string(models.InquiryStatusCanceled),
			},
			{
				Name: Book.ToString(),
				Src: []string{
					string(models.InquiryStatusInquiring),
				},
				Dst: string(models.InquiryStatusBooked),
			},
			{
				Name: Expire.ToString(),
				Src: []string{
					string(models.InquiryStatusInquiring),
				},
				Dst: string(models.InquiryStatusExpired),
			},
		},
		fsm.Callbacks{},
	)

	return f, nil
}
