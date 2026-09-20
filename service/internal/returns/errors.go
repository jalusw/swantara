package returns

import "errors"

var (
	ErrRMANotFound        = errors.New("rma not found")
	ErrRMAState           = errors.New("rma is not in the required state")
	ErrRMAType            = errors.New("invalid rma type")
	ErrRMAOrder           = errors.New("origin order not found or not ready")
	ErrRMAContactMismatch = errors.New("rma contact does not match the origin order")
	ErrRMALines           = errors.New("rma has no lines")
	ErrRMALineProduct     = errors.New("rma line item is required")
	ErrRMALineQty         = errors.New("rma line quantity must be greater than zero")
	ErrRMADisposition     = errors.New("invalid rma line disposition")
	ErrRMAMoveNotFound    = errors.New("no stock movement found for the origin order")
	ErrRMACost            = errors.New("no unit cost found for the original movement")
	ErrRMAScrapLocation   = errors.New("no scrap location found for this organization")
	ErrRMAInvoice         = errors.New("no posted invoice found for the origin order")
	ErrRMAOrganization    = errors.New("organization is required")
)
