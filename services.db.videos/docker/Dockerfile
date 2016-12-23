FROM scratch
MAINTAINER Vmintam <vmintam@gmail.com>
ADD services.db.videos /
EXPOSE 36020 35020
CMD ["/main" , "-port" , "36020" , "-debug" , "35020" , "-audio_db" , "192.168.9.93:9502"]
