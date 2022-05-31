#include <iostream>

//#include <malloc_extension.h>

int leak(){
    int* p = new int();
    return 0;
}


int main(int argc, char **argv){

    //MallocExtension::instance()->SetMemoryReleaseRate(10.0);

    leak();
    return 0;
}