FROM python:3.10-alpine
WORKDIR /scripts
COPY pyproject.toml .
COPY poetry.lock .

RUN apk update && apk add gcc libc-dev git-lfs
RUN pip install poetry==1.2.2 && \
    poetry config virtualenvs.create false && poetry install 

ENV GIT_PYTHON_REFRESH=quiet
COPY . /scripts
RUN chmod +x /scripts/run.sh

ENTRYPOINT ["sh", "/scripts/run.sh"]
