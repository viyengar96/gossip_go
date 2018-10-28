exec = gossip

gossip: gossip.o
	mpicc -std=c99 -g -Wall -o gossip gossip.o

gossip.o: gossip.c
	mpicc -std=c99 -g -Wall -c gossip.c

clean:
	rm -rf *.dSYM *.o $(exec)