module github.com/rezafaghani/vantrel/libs/go/rawarchive

go 1.24.0

require (
	github.com/rezafaghani/vantrel/libs/go/ingestion v0.0.0
	github.com/rezafaghani/vantrel/libs/go/mrx v0.0.0
)

replace github.com/rezafaghani/vantrel/libs/go/ingestion => ../ingestion

replace github.com/rezafaghani/vantrel/libs/go/mrx => ../mrx
