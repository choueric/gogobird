EXEC = gogobird

all:
	@go build -o $(EXEC)

clean:
	@rm -rfv $(EXEC)

run: all
	@./$(EXEC)
