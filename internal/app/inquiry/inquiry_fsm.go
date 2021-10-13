package inquiry

import (
	"fmt"

	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
	"github.com/looplab/fsm"
)

type InquiryActions string

var (
	Cancel      InquiryActions = "cancel"
	Expire      InquiryActions = "expire"
	Pickup      InquiryActions = "pickup"
	AgreePickup InquiryActions = "agree_pickup"
	Skip        InquiryActions = "skip"
	GirlApprove InquiryActions = "girl_approve"
	Book        InquiryActions = "book"
	RevertChat  InquiryActions = "revert_chat"
	Disagree    InquiryActions = "disagree"
)

func (a *InquiryActions) ToString() string {
	return string(*a)
}

func NewInquiryFSM(initial models.InquiryStatus) (*fsm.FSM, error) {
	if !initial.IsValid() {
		return nil, fmt.Errorf("the initial inquiry state: %s is invalid", initial.ToString())
	}

	f := fsm.NewFSM(
		initial.ToString(),
		fsm.Events{
			{
				Name: RevertChat.ToString(),
				Src: []string{
					string(models.InquiryStatusChatting),
				},
				Dst: string(models.InquiryStatusInquiring),
			},
			{
				Name: Cancel.ToString(),
				Src: []string{
					string(models.InquiryStatusInquiring),

					// Only girl can cancel a direct inquiry
					string(models.InquiryStatusAsking),
				},
				Dst: string(models.InquiryStatusCanceled),
			},
			{
				Name: Pickup.ToString(),
				Src: []string{
					string(models.InquiryStatusInquiring),
				},
				Dst: string(models.InquiryStatusAsking),
			},
			{
				Name: AgreePickup.ToString(),
				Src: []string{
					string(models.InquiryStatusAsking),
				},
				Dst: string(models.InquiryStatusChatting),
			},
			{
				Name: Skip.ToString(),
				Src: []string{
					string(models.InquiryStatusAsking),
				},
				Dst: string(models.InquiryStatusInquiring),
			},
			{
				Name: GirlApprove.ToString(),
				Src: []string{
					string(models.InquiryStatusChatting),
				},
				Dst: string(models.InquiryStatusWaitForInquirerApprove),
			},
			{
				Name: Disagree.ToString(),
				Src: []string{
					string(models.InquiryStatusWaitForInquirerApprove),
				},
				Dst: string(models.InquiryStatusChatting),
			},
			{
				Name: Book.ToString(),
				Src: []string{
					string(models.InquiryStatusWaitForInquirerApprove),
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
