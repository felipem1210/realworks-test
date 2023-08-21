#From python alpine slime
FROM python:3.7-alpine
# Label Maintainer
LABEL maintainer="Felipe Macias <felipem1210@gmail.com>"

ADD ./ /app
WORKDIR /app

# Expose port 8080
EXPOSE 8080

# Run the application
CMD ["python", "src/app.py"]