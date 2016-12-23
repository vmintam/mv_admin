FROM scratch
MAINTAINER Vmintam <vmintam@gmail.com>
ADD services.db.comments /
EXPOSE 36030 35030
CMD ["/main" , "-port" , "36030" , "-debug" , "35030" , "-audio_db" , "192.168.9.93:9503"]
