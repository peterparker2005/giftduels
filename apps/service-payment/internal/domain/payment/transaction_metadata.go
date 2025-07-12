package payment

type TransactionMetadata struct {
	Gift *TransactionMetadata_GiftDetails `json:"gift,omitempty"`
}

type TransactionMetadata_GiftDetails struct {
	GiftID string `json:"gift_id"`
	Title  string `json:"title"`
	Slug   string `json:"slug"`
}

func NewGiftWithdrawalCommissionMetadata(giftID, title, slug string) *TransactionMetadata {
	return &TransactionMetadata{
		Gift: &TransactionMetadata_GiftDetails{
			GiftID: giftID,
			Title:  title,
			Slug:   slug,
		},
	}
}
