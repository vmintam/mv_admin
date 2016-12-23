FROM scratch
MAINTAINER Vmintam <vmintam@gmail.com>
ADD services.db.users /
EXPOSE 36000 35000
CMD ["/main" , "-port" , "36000" , "-debug" , "35000" , "-audio_db" , "192.168.9.93:9500"]
