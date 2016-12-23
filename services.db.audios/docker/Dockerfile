FROM scratch
MAINTAINER Vmintam <vmintam@gmail.com>
ADD services.db.audios /
EXPOSE 36010 35010
#ENTRYPOINT ["/tmp/services.db.audios"]
CMD ["/main" , "-port" , "36010" , "-debug" , "35010" , "-audio_db" , "192.168.9.93:9501"]
