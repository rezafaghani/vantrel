FROM docker.io/library/golang:1.24-alpine AS build

WORKDIR /src
COPY libs/go ./libs/go
COPY services/local-pipeline ./services/local-pipeline
RUN cd services/local-pipeline && go build -o /out/vantrel-local-pipeline .

FROM docker.io/library/alpine:3.22
WORKDIR /app
COPY --from=build /out/vantrel-local-pipeline /app/vantrel-local-pipeline
COPY contracts/questdb/market_observations.sql /app/contracts/questdb/market_observations.sql
EXPOSE 8080
ENTRYPOINT ["/app/vantrel-local-pipeline"]
