#include <stdio.h>

#include <string.h>
#include <sys/socket.h>
#include <stdlib.h>
#include <arpa/inet.h>
#include <unistd.h>


int main(int argc, char **argv){
    
    if (argc != 2){
        return 0;
    }

    int port = atoi(argv[1]);

    int sockfd;
    char buffer[1024];
    //char tmp[] = "123456";
    char message[] = "Hello, Server";
    struct sockaddr_in servaddr;

    if ((sockfd = socket(AF_INET, SOCK_DGRAM, 0)) < 0){
        printf("socket creation failed!\n");
        exit(0);
    }

    memset(&servaddr, 0, sizeof(servaddr));
    servaddr.sin_family = AF_INET;
    servaddr.sin_port = htons(port);
    servaddr.sin_addr.s_addr = inet_addr("127.0.0.1");

    while (1){
        memset(&buffer, 0, sizeof(buffer));

        sendto(sockfd, (const char*)message, strlen(message),
        0, (const struct sockaddr*)&servaddr,
        sizeof(servaddr));

        printf("Message UDP server: ");

        int n, len;
        n = recvfrom(sockfd, (char*)buffer, 1024, 0,
            (struct sockaddr*)&servaddr, (socklen_t*)&len);

        //printf("recvfrom %d", n);

        puts(buffer);

        sleep(10);
    }

    close(sockfd);
    return 0;
}