# Makefile for running tests with coverage and generating an HTML report

# Variables
PACKAGE = .
COVERAGE_FILE = ./coverage/coverage.out
COVERAGE_REPORT = ./coverage/coverage-report.txt

# Default target
all: test coverage html

# Run tests with coverage
test:
	go test -coverprofile=$(COVERAGE_FILE) $(PACKAGE)

# # Generate coverage report in text format
coverage: test
	go tool cover -func=$(COVERAGE_FILE) -o ${COVERAGE_REPORT}

# Clean up coverage files
clean:
	rm -f $(COVERAGE_FILE) $(COVERAGE_REPORT)

# Generate coverage report in HTML format
show-coverage-html: test
	go tool cover -html=$(COVERAGE_FILE)

.PHONY: all test coverage html clean