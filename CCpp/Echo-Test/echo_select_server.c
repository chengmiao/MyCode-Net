#include <stdio.h>

#include <sys/socket.h>
#include <arpa/inet.h>
#include <unistd.h>
#include <stdlib.h>
#include <string.h>


int main(int argc, char **argv){

    if (argc < 2){
        return 0;
    }

    int port = atoi(argv[1]);
    if (port < 0){
        return 0;
    }

    int listenfd, udpfd;
    struct sockaddr_in servaddr;

    memset(&servaddr, 0, sizeof(servaddr));
    servaddr.sin_family = AF_INET;
    servaddr.sin_addr.s_addr = INADDR_ANY;
    servaddr.sin_port = htons(port);

    listenfd = socket(AF_INET, SOCK_STREAM, 0);
    bind(listenfd, (struct sockaddr *)&servaddr, sizeof(servaddr));
    listen(listenfd, 10);

    udpfd = socket(AF_INET, SOCK_DGRAM, 0);
    bind(udpfd, (struct sockaddr *)&servaddr, sizeof(servaddr));

    
    fd_set rfds;
    int maxfd;
    int clientfd[50];
    memset(&clientfd, -1, sizeof(clientfd));

    maxfd = listenfd > udpfd ? listenfd : udpfd;

    for (;;){

        FD_ZERO(&rfds);
        FD_SET(listenfd, &rfds);
        FD_SET(udpfd, &rfds);

        for (int i = 0; i < 50; i++){
            if (clientfd[i] == listenfd || clientfd[i] == udpfd){
                continue;
            }

            if (clientfd[i] > 0){
                if (maxfd < clientfd[i]){
                    maxfd = clientfd[i];
                }

                FD_SET(clientfd[i], &rfds);
            }
        }

        if (select(maxfd + 1, &rfds, NULL, NULL, NULL) < 0){
            printf("Select Error!");
            return 0;
        }

        if (FD_ISSET(listenfd, &rfds)){
            int fd;
            if ((fd = accept(listenfd, NULL, NULL)) < 0){
                printf("Accept Error!");
                return 0;
            }

            if (fd >= 50){
                printf("Too many connections!");
                close(fd);
            }else{
                clientfd[fd] == -1 ? clientfd[fd] = fd : close(fd);
            }
        }

        if (FD_ISSET(udpfd, &rfds)){
            struct sockaddr_in u_clientaddr;
            memset(&u_clientaddr, 0, sizeof(u_clientaddr));

            int len;
            char u_buffer[1024] = {0};
            int n = recvfrom(udpfd, u_buffer, 1024, 0, (struct sockaddr *)&u_clientaddr, (socklen_t*)&len);
            printf("\nUDP Recev : ");
            puts(u_buffer);

            sendto(udpfd, u_buffer, n, 0, (struct sockaddr *)&u_clientaddr, sizeof(u_clientaddr));
        }

        for (int i = 0; i < 50; i++){
            if (i == listenfd || i == udpfd){
                continue;
            }

            if (FD_ISSET(i, &rfds)){
                char buffer[1024] = {0};
                int retn = recv(i, buffer, 1024, 0);
                if (retn <= 0){
                    printf("Client Logout!");
                    clientfd[i] = -1;
                    close(i);
                    continue;
                }

                printf("Tcp Recev : %s", buffer);
                send(i, buffer, retn, 0);
            }
        }
    }

    return 0;
}