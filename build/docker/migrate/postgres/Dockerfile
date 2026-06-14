FROM migrate/migrate

WORKDIR /opt/app

COPY db/ ./db
COPY build/docker/migrate/postgres/entrypoint.sh ./

RUN chmod +x entrypoint.sh


ENTRYPOINT [ "./entrypoint.sh"]