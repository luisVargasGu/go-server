# Go Chat Application

This is a simple chat application written in Go(lang) with PostgreSQL as the database backend. It allows users to register, create channels, join rooms, and exchange messages in real-time.

## Features

- **User Registration:** Users can register with a unique username and password.
- **Channel Creation:** Users can create channels for different topics.
- **Room Creation:** Users can join different rooms within channels to chat with other users.
- **Real-time Messaging:** Users can exchange messages in real-time within rooms.

## Technologies Used

- **Go(lang):** The backend is built using Go, a powerful and efficient programming language.
- **PostgreSQL:** The database backend is powered by PostgreSQL, providing robust data storage capabilities.
- **Gorilla/mux:** The Gorilla web toolkit is used for routing HTTP requests in the Go application.
- **dgrijalva/jwt-go:** JWT package for Go(lang) is used for handling authentication tokens.
- **React/SCSS/JavaScript:** Frontend components are compiled into HTML, CSS, and JavaScript for user interaction.

## Setup back-end

1. **Install Go:** Make sure you have Go installed on your system. You can download it from the [official Go website](https://golang.org/).
2. **Install PostgreSQL:** Install PostgreSQL on your system and set up the database according to the schema provided in `Chat app.sql`.
3. **Clone the Repository:** Clone this repository to your local machine.
4. **Configure Environment Variables:** Set up environment variables for your PostgreSQL database connection in a `.env` file. For example:
    ```dotenv
    DB_HOST=localhost
    DB_PORT=5432
    DB_USER=admin
    DB_PASSWORD=password
    DB_NAME=chat_app
    ```
5. **Build and Run:** Navigate to the project directory and run the following commands:
    ```bash
    go build
    ./go-chat-app ||
    go run .
    ```

## Setup front-end

1. **Install Node:** Make sure you have Node installed on your system. You can download it from the [official Node website](https://nodejs.org/).
2. **Clone the Repository:** Clone [this](https://github.com/luisVargasGu/react-portfolio) repository to your local machine.
3. **Install Dependencies:** Navigate to the project directory and run `npm i`.
4. **Start the development server:** `npm run dev`
5. **Access the App:** Once the server is running, access the application in your web browser at `http://localhost:3000`.

## User Flow

![User chat flow](https://github.com/luisVargasGu/go-server/blob/main/assets/Chat.png)
