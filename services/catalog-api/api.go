package catalogapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/rezafaghani/vantrel/libs/go/catalog"
)

type Authorizer func(*http.Request, string) bool

type API struct {
	store catalog.Store
	allow Authorizer
}

type seriesWrite struct {
	Series    catalog.Series `json:"series"`
	ChangedBy string         `json:"changed_by"`
	Reason    string         `json:"reason"`
}

func New(store catalog.Store, allow Authorizer) *API {
	if allow == nil {
		allow = func(*http.Request, string) bool { return true }
	}
	return &API{store: store, allow: allow}
}

func (a *API) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("GET /v1/series", a.searchSeries)
	mux.HandleFunc("POST /v1/series", a.createSeries)
	mux.HandleFunc("GET /v1/series/{series_id}", a.getSeries)
	mux.HandleFunc("PUT /v1/series/{series_id}", a.updateSeries)
	mux.HandleFunc("GET /v1/series/{series_id}/versions", a.getSeriesVersions)
	mux.HandleFunc("GET /v1/series/{series_id}/relationships", a.getRelationships)
	mux.HandleFunc("POST /v1/relationships", a.addRelationship)
	mux.HandleFunc("POST /v1/providers", a.saveProvider)
	mux.HandleFunc("POST /v1/sources", a.saveSource)
	mux.HandleFunc("POST /v1/data-products", a.saveDataProduct)
	mux.HandleFunc("GET /v1/data-products/{product_id}", a.getDataProduct)
	mux.HandleFunc("POST /v1/data-products/{product_id}/members", a.addDataProductMember)
	return mux
}

func (a *API) searchSeries(w http.ResponseWriter, r *http.Request) {
	series, err := a.store.SearchSeries(r.URL.Query().Get("q"))
	writeResult(w, series, err)
}

func (a *API) createSeries(w http.ResponseWriter, r *http.Request) {
	if !a.allow(r, "catalog.series.write") {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}
	var in seriesWrite
	if !readJSON(w, r, &in) {
		return
	}
	version, err := a.store.SaveSeries(in.Series, actor(r, in.ChangedBy), in.Reason)
	writeResult(w, version, err)
}

func (a *API) updateSeries(w http.ResponseWriter, r *http.Request) {
	if !a.allow(r, "catalog.series.write") {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}
	var in seriesWrite
	if !readJSON(w, r, &in) {
		return
	}
	in.Series.ID = r.PathValue("series_id")
	version, err := a.store.SaveSeries(in.Series, actor(r, in.ChangedBy), in.Reason)
	writeResult(w, version, err)
}

func (a *API) getSeries(w http.ResponseWriter, r *http.Request) {
	series, err := a.store.GetSeries(r.PathValue("series_id"))
	writeResult(w, series, err)
}

func (a *API) getSeriesVersions(w http.ResponseWriter, r *http.Request) {
	versions, err := a.store.GetSeriesVersions(r.PathValue("series_id"))
	writeResult(w, versions, err)
}

func (a *API) getRelationships(w http.ResponseWriter, r *http.Request) {
	relationships, err := a.store.GetRelationships(r.PathValue("series_id"))
	writeResult(w, relationships, err)
}

func (a *API) addRelationship(w http.ResponseWriter, r *http.Request) {
	if !a.allow(r, "catalog.relationship.write") {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}
	var relationship catalog.SeriesRelationship
	if readJSON(w, r, &relationship) {
		writeResult(w, map[string]string{"status": "ok"}, a.store.AddRelationship(relationship))
	}
}

func (a *API) saveProvider(w http.ResponseWriter, r *http.Request) {
	if !a.allow(r, "catalog.provider.write") {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}
	var provider catalog.Provider
	if readJSON(w, r, &provider) {
		writeResult(w, map[string]string{"status": "ok"}, a.store.SaveProvider(provider))
	}
}

func (a *API) saveSource(w http.ResponseWriter, r *http.Request) {
	if !a.allow(r, "catalog.source.write") {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}
	var source catalog.Source
	if readJSON(w, r, &source) {
		writeResult(w, map[string]string{"status": "ok"}, a.store.SaveSource(source))
	}
}

func (a *API) saveDataProduct(w http.ResponseWriter, r *http.Request) {
	if !a.allow(r, "catalog.data_product.write") {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}
	var product catalog.DataProduct
	if readJSON(w, r, &product) {
		writeResult(w, map[string]string{"status": "ok"}, a.store.SaveDataProduct(product))
	}
}

func (a *API) getDataProduct(w http.ResponseWriter, r *http.Request) {
	product, members, err := a.store.GetDataProduct(r.PathValue("product_id"))
	writeResult(w, map[string]any{"data_product": product, "members": members}, err)
}

func (a *API) addDataProductMember(w http.ResponseWriter, r *http.Request) {
	if !a.allow(r, "catalog.data_product.write") {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}
	var member catalog.DataProductMember
	if !readJSON(w, r, &member) {
		return
	}
	member.DataProductID = r.PathValue("product_id")
	writeResult(w, map[string]string{"status": "ok"}, a.store.AddDataProductMember(member))
}

func writeResult(w http.ResponseWriter, value any, err error) {
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, catalog.ErrNotFound) {
			status = http.StatusNotFound
		}
		writeError(w, status, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, value)
}

func readJSON(w http.ResponseWriter, r *http.Request, v any) bool {
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return false
	}
	return true
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func actor(r *http.Request, fallback string) string {
	if v := strings.TrimSpace(r.Header.Get("X-Vantrel-Actor")); v != "" {
		return v
	}
	return fallback
}
