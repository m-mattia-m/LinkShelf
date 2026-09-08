.PHONY: test cover cover-full cover-html clean-coverage

# Quick local test run, no coverage - what you'd run while iterating.
test:
	go test ./...

# Fast coverage: unit tests only, no Docker required. Good enough for a
# pre-push sanity check and for GoLand's gutter highlighting day-to-day -
# see "Show Coverage Data" in the Run menu, point it at coverage.unit.out.
cover:
	go list ./... | grep -Ev '/(mocks|config|cmd|migrations)' | xargs go test -coverprofile=coverage.unit.out
	@echo "-------------------------------"
	@go tool cover -func=coverage.unit.out | tail -1
	@echo "-------------------------------"
	@echo "Wrote coverage.unit.out - open it in GoLand via Run > Show Coverage Data."

# Full coverage: mirrors ci.yaml exactly - unit + sqlmock-tagged integration
# tests + the repository package against a real Postgres AND a real MySQL
# container (needs Docker running). This is the number CI's 80% gate checks.
cover-full:
	go list ./... | grep -Ev '/(mocks|config|cmd|migrations)' | xargs go test -coverprofile=coverage.unit.out
	go list -tags=integration ./... | grep -Ev '/(mocks|config|cmd|migrations)' | xargs go test -tags=integration -coverprofile=coverage.integration.out
	APP_DATABASE_ENGINE=postgres go test -tags=realdb -run RealDB -count=1 -coverprofile=coverage.realdb.postgres.out backend/internal/infrastructure/repository
	APP_DATABASE_ENGINE=mysql go test -tags=realdb -run RealDB -count=1 -coverprofile=coverage.realdb.mysql.out backend/internal/infrastructure/repository
	echo "mode: atomic" > coverage.out
	grep -h -v "^mode:" coverage.unit.out coverage.integration.out coverage.realdb.postgres.out coverage.realdb.mysql.out >> coverage.out
	@echo "-------------------------------"
	@go tool cover -func=coverage.out | tail -1
	@echo "-------------------------------"
	@echo "Wrote coverage.out - open it in GoLand via Run > Show Coverage Data."

# Opens the merged profile as an HTML report in your browser - the fastest
# way to actually see which lines are uncovered, line by line.
cover-html: cover-full
	go tool cover -html=coverage.out

clean-coverage:
	rm -f coverage.unit.out coverage.integration.out coverage.realdb.postgres.out coverage.realdb.mysql.out coverage.out
