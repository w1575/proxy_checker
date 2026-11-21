ifeq ($(OS),Windows_NT)
    TOUCH = type nul >
    compile = GOOS=windows go build -a -o build
else
    TOUCH = touch
    compile = go build -a -o build
endif

run:
	$(compile)
	$(TOUCH) ./build/raw.txt
	$(TOUCH) ./build/valid.txt
	$(TOUCH) ./build/invalid.txt
	$(TOUCH) ./build/sources.txt

