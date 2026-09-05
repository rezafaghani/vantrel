package catalog

import (
	"fmt"
	"sync"
	"time"
)

type MemoryStore struct {
	mu            sync.RWMutex
	providers     map[string]Provider
	sources       map[string]Source
	series        map[string]Series
	versions      map[string][]SeriesVersion
	relationships []SeriesRelationship
	products      map[string]DataProduct
	members       map[string][]DataProductMember
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		providers: make(map[string]Provider),
		sources:   make(map[string]Source),
		series:    make(map[string]Series),
		versions:  make(map[string][]SeriesVersion),
		products:  make(map[string]DataProduct),
		members:   make(map[string][]DataProductMember),
	}
}

func (s *MemoryStore) SaveProvider(p Provider) error {
	if err := ValidateProvider(p); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.providers[p.ID] = p
	return nil
}

func (s *MemoryStore) SaveSource(src Source) error {
	if err := ValidateSource(src); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.providers[src.ProviderID]; !ok {
		return fmt.Errorf("%w: provider %q", ErrNotFound, src.ProviderID)
	}
	s.sources[src.ID] = src
	return nil
}

func (s *MemoryStore) SaveSeries(series Series, changedBy, reason string) (SeriesVersion, error) {
	if err := ValidateSeries(series); err != nil {
		return SeriesVersion{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.providers[series.ProviderID]; !ok {
		return SeriesVersion{}, fmt.Errorf("%w: provider %q", ErrNotFound, series.ProviderID)
	}
	src, ok := s.sources[series.SourceID]
	if !ok {
		return SeriesVersion{}, fmt.Errorf("%w: source %q", ErrNotFound, series.SourceID)
	}
	if src.ProviderID != series.ProviderID {
		return SeriesVersion{}, fmt.Errorf("%w: source %q does not belong to provider %q", ErrInvalid, src.ID, series.ProviderID)
	}
	version := SeriesVersion{
		SeriesID:  series.ID,
		Version:   uint64(len(s.versions[series.ID]) + 1),
		ChangedAt: time.Now().UTC(),
		ChangedBy: changedBy,
		Reason:    reason,
		Snapshot:  cloneSeries(series),
	}
	s.series[series.ID] = cloneSeries(series)
	s.versions[series.ID] = append(s.versions[series.ID], version)
	return version, nil
}

func (s *MemoryStore) AddRelationship(r SeriesRelationship) error {
	if err := ValidateRelationship(r); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.series[r.FromSeriesID]; !ok {
		return fmt.Errorf("%w: series %q", ErrNotFound, r.FromSeriesID)
	}
	if _, ok := s.series[r.ToSeriesID]; !ok {
		return fmt.Errorf("%w: series %q", ErrNotFound, r.ToSeriesID)
	}
	s.relationships = append(s.relationships, r)
	return nil
}

func (s *MemoryStore) SaveDataProduct(p DataProduct) error {
	if err := ValidateDataProduct(p); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.products[p.ID] = p
	return nil
}

func (s *MemoryStore) AddDataProductMember(m DataProductMember) error {
	if err := ValidateDataProductMember(m); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.products[m.DataProductID]; !ok {
		return fmt.Errorf("%w: data product %q", ErrNotFound, m.DataProductID)
	}
	if _, ok := s.series[m.SeriesID]; !ok {
		return fmt.Errorf("%w: series %q", ErrNotFound, m.SeriesID)
	}
	s.members[m.DataProductID] = append(s.members[m.DataProductID], m)
	return nil
}

func (s *MemoryStore) GetSeries(id string) (Series, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	series, ok := s.series[id]
	if !ok {
		return Series{}, fmt.Errorf("%w: series %q", ErrNotFound, id)
	}
	return cloneSeries(series), nil
}

func (s *MemoryStore) GetSeriesVersions(id string) ([]SeriesVersion, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	versions := s.versions[id]
	if len(versions) == 0 {
		return nil, fmt.Errorf("%w: series %q", ErrNotFound, id)
	}
	return append([]SeriesVersion(nil), versions...), nil
}

func (s *MemoryStore) GetRelationships(seriesID string) ([]SeriesRelationship, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if _, ok := s.series[seriesID]; !ok {
		return nil, fmt.Errorf("%w: series %q", ErrNotFound, seriesID)
	}
	var out []SeriesRelationship
	for _, r := range s.relationships {
		if r.FromSeriesID == seriesID || r.ToSeriesID == seriesID {
			out = append(out, r)
		}
	}
	return out, nil
}

func (s *MemoryStore) GetDataProduct(id string) (DataProduct, []DataProductMember, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	product, ok := s.products[id]
	if !ok {
		return DataProduct{}, nil, fmt.Errorf("%w: data product %q", ErrNotFound, id)
	}
	return product, append([]DataProductMember(nil), s.members[id]...), nil
}

func cloneSeries(s Series) Series {
	s.Purposes = append([]Purpose(nil), s.Purposes...)
	s.Tags = append([]string(nil), s.Tags...)
	return s
}
