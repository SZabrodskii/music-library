# Music Library

## Prerequisites

- Docker
- Docker Compose

## Setup

1. Clone the repository:

    ```sh
    git clone https://github.com/yourusername/yourproject.git
    cd yourproject
    ```

2. Create a `.env` file in the root directory with the following content:

    ```env
    DB_HOST=localhost
    DB_PORT=5432
    DB_USER=admin
    DB_PASSWORD=password
    DB_NAME=library
    API_URL=https://library/info
    RABBITMQ_URL=amqp://guest:guest@localhost:5672/
    REDIS_URL=localhost:6379
    ```

3. Build and run the services using Docker Compose:

    ```sh
    docker-compose up --build
    ```

4. The services will be available at the following URLs:
   - Gateway: `http://localhost:8080`
   - Song Service: `http://localhost:8081`

5. Swagger documentation will be available at:
   - Gateway: `http://localhost:8080/swagger/index.html`

## API Endpoints

### Songs

- **GET /api/v1/songs**: Get songs with filtering and pagination
- **GET /api/v1/songs/:songId/text**: Get song text with pagination by verses
- **DELETE /api/v1/songs/:songId**: Delete a song
- **PATCH /api/v1/songs/:songId**: Update a song
- **POST /api/v1/songs**: Add a new song

### Models

#### Song

```json
{
  "group": "Muse",
  "song": "Supermassive Black Hole"
}