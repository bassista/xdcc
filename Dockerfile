FROM arm64v8/python:3-alpine

ENTRYPOINT []

ENV PYTHONWARNINGS=ignore

RUN apk update && apk add --update make && apk add --update py3-setuptools && pip3 install setuptools

WORKDIR /root/

COPY puffotter /root/puffotter
RUN cd /root/puffotter && python setup.py install && rm -rf /root/puffotter

COPY xdcc-dl /root/xdcc
RUN cd /root/xdcc && python setup.py install && rm -rf /root/xdcc

#installo le dipendenze a mano perchè setup.py non lo fa più, bho
RUN pip install bs4 && pip install requests && pip install cfscrape && pip install typing && pip install colorama && pip install irc && pip install sentry-sdk && pip install names && pip uninstall urllib3 -y && pip install --use-deprecated=legacy-resolver urllib3==1.26.20

#ssh server
#RUN apk add --update --no-cache  openssh
#RUN mkdir /var/run/sshd
#RUN ssh-keygen -A
#COPY sshd_config /etc/ssh/sshd_config
#RUN echo 'root:insane1980' | chpasswd
#CMD ["/usr/sbin/sshd", "-D"]
#end ssh server

CMD ["tail", "-f", "/dev/null"]
