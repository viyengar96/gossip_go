#include <stdio.h>
#include <stdlib.h>
#include <mpi.h>

void parseArgs(int arg, char ** argv, int *counterT, int *tableT, int *failT);

int main(int argv, char** argc) {
    int counterT, tableT, failT;
    parseArgs(argv, argc, &counterT, &tableT, &failT);
    printf("HB Counter Period: %d\nHB Table period: %d\nNode Fail Perod: %d\n", counterT, tableT, failT);
    return 0;
}

void parseArgs(int argv, char** argc, int *counterT, int *tableT, int *failT){

}
