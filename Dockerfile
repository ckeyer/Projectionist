FROM alpine:edge

MAINTAINER Chuanjian Wang <me@ckeyer.com>
EXPOSE 80

ADD bin/projectionist /projectionist

WORKDIR /video
CMD ["/projectionist"]
