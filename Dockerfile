FROM alpine:edge

MAINTAINER Chuanjian Wang <me@ckeyer.com>

ADD bin/projectionist /bin/projectionist

EXPOSE 80
WORKDIR /video

CMD /bin/projectionist
