module github.com/rezafaghani/vantrel/libs/go/synthetic

go 1.24.0

require (
	github.com/rezafaghani/vantrel/libs/go/catalog v0.0.0
	github.com/rezafaghani/vantrel/libs/go/ingestion v0.0.0
)

replace github.com/rezafaghani/vantrel/libs/go/catalog => ../catalog

replace github.com/rezafaghani/vantrel/libs/go/ingestion => ../ingestion
