# Music Library API

## Prerequisites

- Docker
- Docker Compose

## Getting Started

1. Clone the repository:
    ```sh
    git clone <repository-url>
    cd music-library
    ```

2. Build and run the application using Docker Compose:
    ```sh
    docker-compose up --build
    ```

3. Access the API documentation at:
    ```
    http://localhost:8080/swagger/index.html
    ```

## Environment Variables

Create a `.env` file in the root directory with the following content:
```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=youruser
DB_PASSWORD=yourpassword
DB_NAME=yourdbname
API_URL=https://library/info
RABBITMQ_URL=amqp://guest:guest@localhost:5672/
REDIS_URL=localhost:6379