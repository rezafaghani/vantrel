module github.com/rezafaghani/vantrel/libs/go/marketstore

go 1.24.0

require (
	github.com/rezafaghani/vantrel/libs/go/ingestion v0.0.0
	github.com/rezafaghani/vantrel/libs/go/synthetic v0.0.0
	github.com/rezafaghani/vantrel/libs/go/validation v0.0.0
)

require github.com/rezafaghani/vantrel/libs/go/catalog v0.0.0 // indirect

replace github.com/rezafaghani/vantrel/libs/go/catalog => ../catalog

replace github.com/rezafaghani/vantrel/libs/go/ingestion => ../ingestion

replace github.com/rezafaghani/vantrel/libs/go/synthetic => ../synthetic

replace github.com/rezafaghani/vantrel/libs/go/validation => ../validation
