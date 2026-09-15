module github.com/rezafaghani/vantrel/services/local-pipeline

go 1.24.0

require (
	github.com/rezafaghani/vantrel/libs/go/catalog v0.0.0
	github.com/rezafaghani/vantrel/libs/go/ingestion v0.0.0
	github.com/rezafaghani/vantrel/libs/go/marketstore v0.0.0
	github.com/rezafaghani/vantrel/libs/go/mrx v0.0.0
	github.com/rezafaghani/vantrel/libs/go/rawarchive v0.0.0
	github.com/rezafaghani/vantrel/libs/go/synthetic v0.0.0
	github.com/rezafaghani/vantrel/libs/go/validation v0.0.0
)

replace github.com/rezafaghani/vantrel/libs/go/catalog => ../../libs/go/catalog

replace github.com/rezafaghani/vantrel/libs/go/ingestion => ../../libs/go/ingestion

replace github.com/rezafaghani/vantrel/libs/go/marketstore => ../../libs/go/marketstore

replace github.com/rezafaghani/vantrel/libs/go/mrx => ../../libs/go/mrx

replace github.com/rezafaghani/vantrel/libs/go/rawarchive => ../../libs/go/rawarchive

replace github.com/rezafaghani/vantrel/libs/go/synthetic => ../../libs/go/synthetic

replace github.com/rezafaghani/vantrel/libs/go/validation => ../../libs/go/validation
