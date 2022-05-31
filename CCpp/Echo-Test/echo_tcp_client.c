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

    if ((sockfd = socket(AF_INET, SOCK_STREAM, 0)) < 0){
        printf("socket creation failed!\n");
        exit(0);
    }

    memset(&servaddr, 0, sizeof(servaddr));
    servaddr.sin_family = AF_INET;
    servaddr.sin_port = htons(port);
    servaddr.sin_addr.s_addr = inet_addr("127.0.0.1");

    if (connect(sockfd, (struct sockaddr *)&servaddr, sizeof(servaddr)) < 0){
        printf("Failed to connect to server!\n");
        exit(0);
    }

    while (1){
        memset(&buffer, 0, sizeof(buffer));

        write(sockfd, (const char*)message, strlen(message));

        printf("Message TCP server: ");

        int n, len;
        n = read(sockfd, (char*)buffer, 1024);

        //printf("recvfrom %d", n);

        puts(buffer);

        sleep(10);
    }

    close(sockfd);
    return 0;
}