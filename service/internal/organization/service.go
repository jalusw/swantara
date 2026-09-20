package organization

import (
	"context"
	"strings"

	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/queue"
	"github.com/jalusw/swantara/apps/service/internal/queue/tasks"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

type Service struct {
	orgs              OrgDAO
	currencies        CurrencyDAO
	members           MemberProvisioner
	paymentTermSeeder PaymentTermSeeder
	modules           ModuleStore
	enqueuer          queue.TaskEnqueuer
}

func NewOrganizationService(orgs OrgDAO, currencies CurrencyDAO, members MemberProvisioner) Service {
	return Service{orgs: orgs, currencies: currencies, members: members}
}

func NewOrganizationServiceWithPaymentTerms(orgs OrgDAO, currencies CurrencyDAO, members MemberProvisioner, seeder PaymentTermSeeder) Service {
	return Service{orgs: orgs, currencies: currencies, members: members, paymentTermSeeder: seeder}
}

func (s *Service) SetPaymentTermSeeder(seeder PaymentTermSeeder) {
	s.paymentTermSeeder = seeder
}

func (s *Service) SetModuleStore(store ModuleStore) {
	s.modules = store
}

func (s *Service) SetProvisionEnqueuer(enqueuer queue.TaskEnqueuer) {
	s.enqueuer = enqueuer
}

func (s Service) List(ctx context.Context, q *query.Query) (*query.Page[reference.Organization], error) {
	return s.orgs.List(ctx, q)
}

func (s Service) Find(ctx context.Context, id uint64) (*reference.Organization, error) {
	return s.orgs.Find(ctx, id)
}

func (s Service) Delete(ctx context.Context, id uint64) error {
	return s.orgs.Delete(ctx, id)
}

func (s Service) QuickCreate(ctx context.Context, name string, countryCode string, ownerUserID uint64) (*reference.Organization, error) {
	if err := s.validateName(name); err != nil {
		return nil, err
	}

	defaults, ok := lookupCountryDefaults(countryCode)
	if !ok {
		return nil, ErrCountryCodeInvalid
	}

	if err := s.validateBaseCurrency(ctx, defaults.Currency); err != nil {
		return nil, err
	}

	cc := countryCode
	org := &reference.Organization{
		Name:              name,
		BaseCurrency:      defaults.Currency,
		CountryCode:       &cc,
		Timezone:          defaults.Timezone,
		TaxYearStartMonth: 1,
	}

	created, err := s.orgs.Create(ctx, org)
	if err != nil {
		return nil, err
	}
	if err := s.members.ProvisionOwner(ctx, created.ID, ownerUserID); err != nil {
		return nil, err
	}
	if s.enqueuer == nil {
		if err := s.Provision(ctx, created.ID); err != nil {
			return nil, err
		}
		return created, nil
	}
	task, err := tasks.NewOrganizationProvisioningTask(created.ID, helper.RequestID(ctx))
	if err != nil {
		return nil, err
	}
	if _, err := s.enqueuer.Enqueue(task); err != nil {
		return nil, err
	}
	return created, nil
}

func (s Service) Provision(ctx context.Context, organizationID uint64) error {
	if s.paymentTermSeeder != nil {
		if err := s.paymentTermSeeder.SeedForOrganization(ctx, organizationID); err != nil {
			return err
		}
	}
	if s.modules != nil {
		if err := s.activateAllModules(ctx, organizationID); err != nil {
			return err
		}
	}
	return nil
}

func (s Service) Create(ctx context.Context, org *reference.Organization, ownerUserID uint64) (*reference.Organization, error) {
	if err := s.validateName(org.Name); err != nil {
		return nil, err
	}
	if err := s.validateBaseCurrency(ctx, org.BaseCurrency); err != nil {
		return nil, err
	}
	if err := s.validateParent(ctx, 0, org.ParentID); err != nil {
		return nil, err
	}

	if org.Timezone == "" {
		org.Timezone = "UTC"
	}
	if org.TaxYearStartMonth == 0 {
		org.TaxYearStartMonth = 1
	}

	created, err := s.orgs.Create(ctx, org)
	if err != nil {
		return nil, err
	}
	if err := s.members.ProvisionOwner(ctx, created.ID, ownerUserID); err != nil {
		return nil, err
	}
	if s.paymentTermSeeder != nil {
		if err := s.paymentTermSeeder.SeedForOrganization(ctx, created.ID); err != nil {
			return nil, err
		}
	}
	return created, nil
}

func (s Service) Update(ctx context.Context, orgID uint64, org *reference.Organization) (*reference.Organization, error) {
	existing, err := s.orgs.Find(ctx, orgID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, ErrOrganizationNotFound
	}

	if err := s.validateName(org.Name); err != nil {
		return nil, err
	}
	if err := s.validateParent(ctx, orgID, org.ParentID); err != nil {
		return nil, err
	}
	if org.BaseCurrency != "" {
		if err := s.validateBaseCurrency(ctx, org.BaseCurrency); err != nil {
			return nil, err
		}
		existing.BaseCurrency = org.BaseCurrency
	}

	existing.Name = org.Name
	existing.LegalName = org.LegalName
	existing.ParentID = org.ParentID
	existing.CountryCode = org.CountryCode
	existing.TaxID = org.TaxID
	if org.Timezone != "" {
		existing.Timezone = org.Timezone
	}
	if org.TaxYearStartMonth != 0 {
		existing.TaxYearStartMonth = org.TaxYearStartMonth
	}

	return s.orgs.Update(ctx, existing)
}

func (s Service) validateName(name string) error {
	if strings.TrimSpace(name) == "" {
		return ErrNameRequired
	}
	return nil
}

func (s Service) validateBaseCurrency(ctx context.Context, code string) error {
	currency, err := s.currencies.Search(ctx, "code", code)
	if err != nil {
		return err
	}
	if currency == nil {
		return ErrBaseCurrencyNotFound
	}
	return nil
}

func (s Service) validateParent(ctx context.Context, orgID uint64, parentID *uint64) error {
	if parentID == nil {
		return nil
	}
	if *parentID == orgID {
		return ErrParentCycle
	}

	parent, err := s.orgs.Find(ctx, *parentID)
	if err != nil {
		return err
	}
	if parent == nil {
		return ErrParentNotFound
	}

	current := parent
	for current.ParentID != nil {
		if *current.ParentID == orgID {
			return ErrParentCycle
		}
		next, err := s.orgs.Find(ctx, *current.ParentID)
		if err != nil {
			return err
		}
		if next == nil {
			return ErrParentNotFound
		}
		current = next
	}
	return nil
}
