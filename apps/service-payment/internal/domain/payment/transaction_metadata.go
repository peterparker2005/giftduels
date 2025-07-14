package payment

type TransactionMetadata struct {
	Gift *TransactionMetadataGiftDetails `json:"gift,omitempty"`
}

type TransactionMetadataGiftDetails struct {
	GiftID string `json:"gift_id"`
	Title  string `json:"title"`
	Slug   string `json:"slug"`
}

func NewGiftWithdrawalCommissionMetadata(giftID, title, slug string) *TransactionMetadata {
	return &TransactionMetadata{
		Gift: &TransactionMetadataGiftDetails{
			GiftID: giftID,
			Title:  title,
			Slug:   slug,
		},
	}
}
