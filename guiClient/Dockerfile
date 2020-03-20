FROM python:latest

WORKDIR /guiClientApp

COPY requirements.txt /guiClientApp/requirements.txt
RUN pip install -r requirements.txt

COPY . /guiClientApp

EXPOSE 8000

CMD ["python", "manage.py", "runserver", "0.0.0.0:8000"]