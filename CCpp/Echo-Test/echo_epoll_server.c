#include <stdio.h>

#include <sys/socket.h>
#include <sys/epoll.h>
#include <arpa/inet.h>
#include <string.h>
#include <unistd.h>
#include <stdlib.h>

#include <errno.h>



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

    struct epoll_event ev, events[20];
    int epfd = epoll_create(256);

    ev.data.fd = listenfd;
    ev.events = EPOLLIN;
    epoll_ctl(epfd, EPOLL_CTL_ADD, listenfd, &ev);

    ev.data.fd = udpfd;
    ev.events = EPOLLIN;
    epoll_ctl(epfd, EPOLL_CTL_ADD, udpfd, &ev);

    for (;;){

        int nfds = epoll_wait(epfd, events, 20, -1);
        for (int i = 0; i < nfds; i++){
            if (events[i].data.fd == listenfd){
                int connfd = accept(listenfd, NULL, NULL);
                if (connfd < 0){
                    continue;
                }

                ev.data.fd = connfd;
                ev.events = EPOLLIN;
                epoll_ctl(epfd, EPOLL_CTL_ADD, connfd, &ev);
            }
            else if (events[i].data.fd == udpfd){
                struct sockaddr_in u_clientaddr;
                memset(&u_clientaddr, 0, sizeof(u_clientaddr));

                int len;
                char u_buffer[1024] = {0};
                int n = recvfrom(udpfd, u_buffer, 1024, 0, (struct sockaddr *)&u_clientaddr, (socklen_t*)&len);
                printf("\nUDP Recev : ");
                puts(u_buffer);

                sendto(udpfd, u_buffer, n, 0, (struct sockaddr *)&u_clientaddr, sizeof(u_clientaddr));
            }
            else if (events[i].events & EPOLLIN){
                char buffer[1024] = {0};
                int retn = recv(events[i].data.fd, buffer, 1024, 0);
                if (retn <= 0){
                    printf("Client Logout!");
                    close(events[i].data.fd);
                    continue;
                }

                printf("Tcp Recev : %s", buffer);
                send(events[i].data.fd, buffer, retn, 0);

                //EWOULDBLOCK
            }
        }
    }


    return 0;
}