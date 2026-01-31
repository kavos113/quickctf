FROM ubuntu:24.04

RUN apt update && apt upgrade -y && apt install -y xinetd netcat-traditional strace
RUN groupadd -r user && useradd -r -g user user

COPY --chown=root:root ./ctf.conf /etc/xinetd.d/ctf
COPY --chown=root:user ./example /home/user/example
COPY --chown=root:user ./start.sh /home/user/start.sh
COPY --chown=root:user ./init.sh /home/user/init.sh

WORKDIR /home/user

RUN chmod 555 ./example && \
    chmod 555 ./start.sh && \
    chmod 444 /etc/xinetd.d/ctf && \
    chmod 555 ./init.sh
USER user
EXPOSE 39092

CMD ["./init.sh"]