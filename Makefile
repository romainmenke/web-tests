.DEFAULT_GOAL := all

feature_dirs := $(wildcard ./specifications/*/*/*)

features: $(feature_dirs)

$(feature_dirs):
	@$(MAKE) -C $@

scripts: 
	@$(MAKE) -C ./scripts

all: features scripts

.PHONY: all features $(feature_dirs) scripts
