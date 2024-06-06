FROM ubuntu:24.10

RUN apt-get update && apt-get install -y cowsay 
RUN ln -s /usr/games/cowsay /usr/local/bin/cowsay

CMD [ "cowsay" ]
