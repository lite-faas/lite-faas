FROM python:3.11-alpine

WORKDIR /app

COPY function.py .
COPY requirements.txt .

RUN pip install -r requirements.txt

EXPOSE 8080

CMD ["python", "function.py"]
