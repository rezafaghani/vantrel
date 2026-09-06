package synthetic

import (
	"time"

	"github.com/rezafaghani/vantrel/libs/go/catalog"
)

const DataProductDK1PowerAndApple = "SYN_DK1_POWER_AND_APPLE_MARKET_SAMPLE"

func RegisterCatalog(store catalog.Store) error {
	if err := store.SaveProvider(catalog.Provider{ID: ProviderID, DisplayName: "Synthetic Market Source", License: "generated-test-data"}); err != nil {
		return err
	}
	if err := store.SaveSource(catalog.Source{ID: SourceID, ProviderID: ProviderID, Name: "Synthetic market source", Protocol: "in-memory", Endpoint: "synthetic://market"}); err != nil {
		return err
	}
	for _, series := range catalogSeries() {
		if _, err := store.SaveSeries(series, "synthetic-source", "initial synthetic market series"); err != nil {
			return err
		}
	}
	if err := store.AddRelationship(catalog.SeriesRelationship{
		FromSeriesID: SeriesDK1WindForecast,
		ToSeriesID:   SeriesDK1WindActual,
		Type:         catalog.RelationshipForecastOf,
		ValidFrom:    time.Unix(1704067200, 0).UTC(),
	}); err != nil {
		return err
	}
	product := catalog.DataProduct{
		ID:          DataProductDK1PowerAndApple,
		DisplayName: "Synthetic DK1 power and Apple market sample",
		Description: "Small generated data product for ingestion SDK and catalog integration tests.",
		Owner:       "vantrel",
		Purposes:    []catalog.Purpose{catalog.PurposeTraining, catalog.PurposeValidation},
		Active:      true,
	}
	if err := store.SaveDataProduct(product); err != nil {
		return err
	}
	for _, member := range []catalog.DataProductMember{
		{DataProductID: product.ID, SeriesID: SeriesAppleTradePrice, Role: "trade_price", Required: true},
		{DataProductID: product.ID, SeriesID: SeriesAppleQuote, Role: "quote", Required: true},
		{DataProductID: product.ID, SeriesID: SeriesDK1WindForecast, Role: "wind_forecast", Required: true},
		{DataProductID: product.ID, SeriesID: SeriesDK1WindActual, Role: "wind_actual", Required: true},
	} {
		if err := store.AddDataProductMember(member); err != nil {
			return err
		}
	}
	return nil
}

func catalogSeries() []catalog.Series {
	return []catalog.Series{
		{
			ID:               SeriesAppleTradePrice,
			CanonicalName:    "synthetic.equities.aapl.trade_price",
			DisplayName:      "Synthetic Apple trade price",
			Description:      "Generated Apple trade price observations for local development.",
			AssetClass:       "EQUITY",
			SeriesType:       "TRADE",
			ProviderID:       ProviderID,
			SourceID:         SourceID,
			ProviderSeriesID: "AAPL.TRADE",
			Market:           "US_EQUITIES",
			Instrument:       "AAPL",
			Currency:         "USD",
			Resolution:       "event",
			Timezone:         "UTC",
			Unit:             "USD",
			ValueType:        "PRICE",
			Purposes:         []catalog.Purpose{catalog.PurposeRealtime, catalog.PurposeTraining, catalog.PurposeValidation},
			Tags:             []string{"synthetic", "equity", "trade"},
			Active:           true,
		},
		{
			ID:               SeriesAppleQuote,
			CanonicalName:    "synthetic.equities.aapl.quote",
			DisplayName:      "Synthetic Apple quote",
			Description:      "Generated Apple best bid/ask quote observations for local development.",
			AssetClass:       "EQUITY",
			SeriesType:       "QUOTE",
			ProviderID:       ProviderID,
			SourceID:         SourceID,
			ProviderSeriesID: "AAPL.QUOTE",
			Market:           "US_EQUITIES",
			Instrument:       "AAPL",
			Currency:         "USD",
			Resolution:       "event",
			Timezone:         "UTC",
			Unit:             "USD",
			ValueType:        "BID_ASK",
			Purposes:         []catalog.Purpose{catalog.PurposeRealtime, catalog.PurposeTraining, catalog.PurposeValidation},
			Tags:             []string{"synthetic", "equity", "quote"},
			Active:           true,
		},
		{
			ID:               SeriesDK1WindForecast,
			CanonicalName:    "synthetic.energy.dk1.wind_forecast",
			DisplayName:      "Synthetic DK1 wind forecast",
			Description:      "Generated DK1 wind production forecast observations.",
			AssetClass:       "ENERGY",
			SeriesType:       "FORECAST",
			ProviderID:       ProviderID,
			SourceID:         SourceID,
			ProviderSeriesID: "DK1.WIND.FORECAST",
			Market:           "NORDIC_POWER",
			BiddingZone:      "DK1",
			Commodity:        "POWER",
			Region:           "DK",
			Resolution:       "PT30M",
			Timezone:         "UTC",
			Unit:             "MW",
			ValueType:        "POWER",
			Purposes:         []catalog.Purpose{catalog.PurposePreTrade, catalog.PurposeTraining, catalog.PurposeValidation},
			Tags:             []string{"synthetic", "energy", "forecast", "wind"},
			Active:           true,
		},
		{
			ID:               SeriesDK1WindActual,
			CanonicalName:    "synthetic.energy.dk1.wind_actual",
			DisplayName:      "Synthetic DK1 wind actual",
			Description:      "Generated DK1 realized wind production observations.",
			AssetClass:       "ENERGY",
			SeriesType:       "ACTUAL",
			ProviderID:       ProviderID,
			SourceID:         SourceID,
			ProviderSeriesID: "DK1.WIND.ACTUAL",
			Market:           "NORDIC_POWER",
			BiddingZone:      "DK1",
			Commodity:        "POWER",
			Region:           "DK",
			Resolution:       "PT30M",
			Timezone:         "UTC",
			Unit:             "MW",
			ValueType:        "POWER",
			Purposes:         []catalog.Purpose{catalog.PurposePostTrade, catalog.PurposeTraining, catalog.PurposeValidation},
			Tags:             []string{"synthetic", "energy", "actual", "wind"},
			Active:           true,
		},
	}
}
