# Makefile for running tests with coverage and generating an HTML report

# Variables
PACKAGE = .
COVERAGE_FILE = coverage.out
COVERAGE_HTML = coverage.html

# Default target
all: test coverage html

# Run tests with coverage
test:
	go test -coverprofile=$(COVERAGE_FILE) $(PACKAGE)

# # Generate coverage report in text format
# coverage: test
# 	go tool cover -func=$(COVERAGE_FILE)

# Generate coverage report in HTML format
show-html: test
	go tool cover -html=$(COVERAGE_FILE)

gen-html-output: test
	go tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)

# Clean up coverage files
clean:
	rm -f $(COVERAGE_FILE) $(COVERAGE_HTML)

.PHONY: all test coverage html clean