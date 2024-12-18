# Use the official PostgreSQL image from Docker Hub
FROM postgres:14

# Set environment variables
ENV POSTGRES_USER=pagamov
ENV POSTGRES_PASSWORD=multipass
ENV POSTGRES_DB=database

# Copy initialization scripts (optional)
# Place your SQL scripts in a directory named 'init-scripts'
COPY init-scripts/ /docker-entrypoint-initdb.d/

# Expose the PostgreSQL port
EXPOSE 5432