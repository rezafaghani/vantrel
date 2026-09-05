package catalog

import (
	"errors"
	"fmt"
	"time"
)

var (
	ErrInvalid  = errors.New("catalog: invalid entity")
	ErrNotFound = errors.New("catalog: not found")
	ErrConflict = errors.New("catalog: conflict")
)

type Purpose string

const (
	PurposePreTrade   Purpose = "PRE_TRADE"
	PurposeRealtime   Purpose = "REALTIME"
	PurposePostTrade  Purpose = "POST_TRADE"
	PurposeSettlement Purpose = "SETTLEMENT"
	PurposeReference  Purpose = "REFERENCE"
	PurposeTraining   Purpose = "TRAINING"
	PurposeValidation Purpose = "VALIDATION"
	PurposeCompliance Purpose = "COMPLIANCE"
)

type RelationshipType string

const (
	RelationshipForecastOf     RelationshipType = "FORECAST_OF"
	RelationshipActualOf       RelationshipType = "ACTUAL_OF"
	RelationshipDerivedFrom    RelationshipType = "DERIVED_FROM"
	RelationshipAggregatedFrom RelationshipType = "AGGREGATED_FROM"
	RelationshipInputTo        RelationshipType = "INPUT_TO"
	RelationshipOutputOf       RelationshipType = "OUTPUT_OF"
	RelationshipRelatedTo      RelationshipType = "RELATED_TO"
	RelationshipCorrelatedWith RelationshipType = "CORRELATED_WITH"
	RelationshipReplacedBy     RelationshipType = "REPLACED_BY"
	RelationshipInfluencedBy   RelationshipType = "INFLUENCED_BY"
)

type Provider struct {
	ID          string
	DisplayName string
	License     string
}

type Source struct {
	ID         string
	ProviderID string
	Name       string
	Protocol   string
	Endpoint   string
}

type Series struct {
	ID                     string
	CanonicalName          string
	DisplayName            string
	Description            string
	AssetClass             string
	SeriesType             string
	ProviderID             string
	SourceID               string
	ProviderSeriesID       string
	Market                 string
	Venue                  string
	BiddingZone            string
	Instrument             string
	Commodity              string
	Currency               string
	Region                 string
	Resolution             string
	Timezone               string
	Unit                   string
	ValueType              string
	AggregationSemantics   string
	ValidationProfileID    string
	NormalizationProfileID string
	Purposes               []Purpose
	Tags                   []string
	Active                 bool
}

type SeriesVersion struct {
	SeriesID  string
	Version   uint64
	ChangedAt time.Time
	ChangedBy string
	Reason    string
	Snapshot  Series
}

type SeriesRelationship struct {
	FromSeriesID string
	ToSeriesID   string
	Type         RelationshipType
	ValidFrom    time.Time
	ValidTo      *time.Time
}

type DataProduct struct {
	ID          string
	DisplayName string
	Description string
	Owner       string
	Purposes    []Purpose
	Active      bool
}

type DataProductMember struct {
	DataProductID string
	SeriesID      string
	Role          string
	Required      bool
}

type Store interface {
	SaveProvider(Provider) error
	SaveSource(Source) error
	SaveSeries(Series, string, string) (SeriesVersion, error)
	AddRelationship(SeriesRelationship) error
	SaveDataProduct(DataProduct) error
	AddDataProductMember(DataProductMember) error
	GetSeries(string) (Series, error)
	GetSeriesVersions(string) ([]SeriesVersion, error)
	GetRelationships(string) ([]SeriesRelationship, error)
	GetDataProduct(string) (DataProduct, []DataProductMember, error)
}

func ValidateProvider(p Provider) error {
	if p.ID == "" || p.DisplayName == "" {
		return fmt.Errorf("%w: provider id and display name are required", ErrInvalid)
	}
	return nil
}

func ValidateSource(s Source) error {
	if s.ID == "" || s.ProviderID == "" || s.Name == "" || s.Protocol == "" {
		return fmt.Errorf("%w: source id, provider id, name and protocol are required", ErrInvalid)
	}
	return nil
}

func ValidateSeries(s Series) error {
	switch {
	case s.ID == "":
		return fmt.Errorf("%w: series id is required", ErrInvalid)
	case s.CanonicalName == "":
		return fmt.Errorf("%w: canonical name is required", ErrInvalid)
	case s.ProviderID == "":
		return fmt.Errorf("%w: provider id is required", ErrInvalid)
	case s.SourceID == "":
		return fmt.Errorf("%w: source id is required", ErrInvalid)
	case len(s.Purposes) == 0:
		return fmt.Errorf("%w: at least one purpose is required", ErrInvalid)
	}
	for _, purpose := range s.Purposes {
		if !validPurpose(purpose) {
			return fmt.Errorf("%w: unknown purpose %q", ErrInvalid, purpose)
		}
	}
	return nil
}

func ValidateRelationship(r SeriesRelationship) error {
	if r.FromSeriesID == "" || r.ToSeriesID == "" || r.Type == "" {
		return fmt.Errorf("%w: relationship endpoints and type are required", ErrInvalid)
	}
	if r.FromSeriesID == r.ToSeriesID {
		return fmt.Errorf("%w: relationship cannot point to itself", ErrInvalid)
	}
	if !validRelationshipType(r.Type) {
		return fmt.Errorf("%w: unknown relationship type %q", ErrInvalid, r.Type)
	}
	return nil
}

func ValidateDataProduct(p DataProduct) error {
	if p.ID == "" || p.DisplayName == "" {
		return fmt.Errorf("%w: data product id and display name are required", ErrInvalid)
	}
	for _, purpose := range p.Purposes {
		if !validPurpose(purpose) {
			return fmt.Errorf("%w: unknown purpose %q", ErrInvalid, purpose)
		}
	}
	return nil
}

func ValidateDataProductMember(m DataProductMember) error {
	if m.DataProductID == "" || m.SeriesID == "" {
		return fmt.Errorf("%w: data product id and series id are required", ErrInvalid)
	}
	return nil
}

func validPurpose(p Purpose) bool {
	switch p {
	case PurposePreTrade, PurposeRealtime, PurposePostTrade, PurposeSettlement,
		PurposeReference, PurposeTraining, PurposeValidation, PurposeCompliance:
		return true
	default:
		return false
	}
}

func validRelationshipType(t RelationshipType) bool {
	switch t {
	case RelationshipForecastOf, RelationshipActualOf, RelationshipDerivedFrom,
		RelationshipAggregatedFrom, RelationshipInputTo, RelationshipOutputOf,
		RelationshipRelatedTo, RelationshipCorrelatedWith, RelationshipReplacedBy,
		RelationshipInfluencedBy:
		return true
	default:
		return false
	}
}
