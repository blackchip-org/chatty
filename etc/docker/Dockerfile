FROM ubuntu:xenial

RUN apt-get update
RUN apt-get install -y ircd-irc2 telnet

CMD ircd -t

COPY ircd.conf /etc/ircd/ircd.conf 

EXPOSE 6667:6667
