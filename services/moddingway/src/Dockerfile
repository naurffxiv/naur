FROM python:3.11

# install requiremed packages
RUN --mount=type=bind,source=requirements.txt,target=/tmp/requirements.txt \
    pip install --requirement /tmp/requirements.txt

COPY . /opt/app
WORKDIR /opt/app

CMD ["python", "./src/main.py"]