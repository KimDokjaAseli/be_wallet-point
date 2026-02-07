# be_wallet-point

Backend for Wallet Point system built with Go.

## Docker Setup

This project is fully dockerized for easy development and deployment.
It includes:
- **App Service**: The Go application running on port 8080.
- **Database Service**: MySQL 8.0 running on port 3306.
- **phpMyAdmin**: Database management tool running on port 8081.

### Prerequisites

- Docker
- Docker Compose

### How to Run

1.  **Start the services:**
    ```bash
    docker-compose up --build -d
    ```

2.  **Access the Application:**
    - API: `http://localhost:8080`
    - Swagger Docs (if enabled): `http://localhost:8080/swagger/index.html`

3.  **Manage Database:**
    - URL: `http://localhost:8081`
    - Server: `db`
    - Username: `myuser`
    - Password: `mypassword`
    (Or login as `root` with password `rootpassword`)

4.  **Stopping the services:**
    ```bash
    docker-compose down
    ```

### Environment Variables

The default configuration is set in `docker-compose.yml`. You can modify it there or create a `.env` file to override values if you switch to `env_file` configuration.