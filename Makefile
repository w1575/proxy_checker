ifeq ($(OS),Windows_NT)
    TOUCH = type nul >
    COPY = copy
    compile = GOOS=windows go build -a -o build
else
    TOUCH = touch
    COPY = cp
    compile = go build -a -o build
endif

run:
	$(compile)
	$(TOUCH) ./build/raw.txt
	$(TOUCH) ./build/valid.txt
	$(TOUCH) ./build/invalid.txt
	$(TOUCH) ./build/sources.txt
	$(COPY) ./settings.yml ./build/settings.yml
	$(COPY) ./sources.txt.example ./build/sources.txt