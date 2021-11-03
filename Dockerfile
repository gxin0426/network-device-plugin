FROM golang:1.16.8
MAINTAINER gx <gaoxin123@xxx.com>
RUN apt update
RUN apt install net-tools
COPY ./build/easyalgo /root/easyalgo

CMD ["/root/easyalgo"]
