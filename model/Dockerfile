# Use the official Python image from the Docker Hub
FROM python:3.12-slim

# Set the working directory in the container
WORKDIR /model

# Copy the requirements file into the container
COPY requirements.txt .
COPY text_classification_model.h5 .
COPY word2vec_model.model .
COPY label_encoder.pkl .
COPY model.py .

# Install the required Python packages
# RUN pip install --no-cache-dir -r requirements.txt
RUN pip install -r requirements.txt

# Copy the rest of the application code into the container
# COPY . .

ARG DEFAULT_PORT=8081
ENV PORT $DEFAULT_PORT

EXPOSE $PORT

# Command to run the application
CMD ["python", "model.py"]