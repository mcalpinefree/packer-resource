FROM docker

RUN apk update && apk add wget iptables ca-certificates

RUN wget https://releases.hashicorp.com/packer/0.12.3/packer_0.12.3_linux_amd64.zip && \
	unzip packer_0.12.3_linux_amd64.zip && \
	mv packer /usr/local/bin/ && \
	rm packer_0.12.3_linux_amd64.zip

RUN mkdir -p /etc/docker && echo '{}' > /etc/docker/daemon.json

RUN mkdir -p /opt/resource
ADD ./check/check /opt/resource/
ADD ./out/out /opt/resource/
ADD ./in/in /opt/resource/
