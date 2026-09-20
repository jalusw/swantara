package handler

import (
	"time"

	"github.com/jalusw/swantara/apps/service/internal/contacts"
)

type ContactResponse struct {
	ID             uint64    `json:"id"`
	OrganizationID *uint64   `json:"organization_id"`
	Name           string    `json:"name"`
	DisplayName    *string   `json:"display_name"`
	IsOrganization bool      `json:"is_organization"`
	ParentID       *uint64   `json:"parent_id"`
	Email          *string   `json:"email"`
	Phone          *string   `json:"phone"`
	Mobile         *string   `json:"mobile"`
	Website        *string   `json:"website"`
	TaxID          *string   `json:"tax_id"`
	Industry       *string   `json:"industry"`
	CurrencyCode   *string   `json:"currency_code"`
	Lang           string    `json:"lang"`
	Active         bool      `json:"active"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func newContactResponse(contact *contacts.Contact) ContactResponse {
	return ContactResponse{
		ID:             contact.ID,
		OrganizationID: contact.OrganizationID,
		Name:           contact.Name,
		DisplayName:    contact.DisplayName,
		IsOrganization: contact.IsOrganization,
		ParentID:       contact.ParentID,
		Email:          contact.Email,
		Phone:          contact.Phone,
		Mobile:         contact.Mobile,
		Website:        contact.Website,
		TaxID:          contact.TaxID,
		Industry:       contact.Industry,
		CurrencyCode:   contact.CurrencyCode,
		Lang:           contact.Lang,
		Active:         contact.Active,
		CreatedAt:      contact.CreatedAt,
		UpdatedAt:      contact.UpdatedAt,
	}
}

var contactQueryAllowlist = map[string]struct{}{
	"organization_id": {},
	"name":            {},
	"email":           {},
	"phone":           {},
	"tax_id":          {},
	"industry":        {},
	"is_organization": {},
	"currency_code":   {},
	"active":          {},
	"created_at":      {},
	"updated_at":      {},
}
