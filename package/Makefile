.PHONY: fmt

# go install -v github.com/incu6us/goimports-reviser/v3@latest
fmt:
	@echo "Running goimports-reviser ..."
	find . -type f -name "*.go" \
	  ! -path "./cmd/*" \
	  ! -path "./tools/*" \
	  ! -path "./generated/*" \
	  ! -path "./vendor/*" \
	  -exec goimports-reviser \
	   -company-prefixes github.com/wundergraph \
	   -rm-unused -set-alias -format {} \;
	@echo "goimports-reviser done!"