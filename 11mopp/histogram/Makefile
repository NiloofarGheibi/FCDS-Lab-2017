CC=gcc

FLAGS=-O3 -lm

all: histogram

histogram: histogram.c
	$(CC)  -fopenmp histogram.c $(FLAGS)  -o histogram

clean:
	rm -f histogram

# all: histogram

# histogram: histogram.go
# 	go build -o histogram histogram.go 

# clean:
# 	rm -f histogram
# 	go clean