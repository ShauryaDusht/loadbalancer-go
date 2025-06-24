# Load Balancer in Go

This is a simple load balancer written in Go. It can distribute incoming HTTP requests to multiple backend servers using round-robin scheduling and random selection.

## Round Robin Load Balancer

Uncomments `chosenAlgorithm := "round_robin"` in the `main` function to use the round-robin algorithm

## Random Selection Load Balancer

Uncomments `chosenAlgorithm := "random"` in the `main` function to use the random selection algorithm.

## Usage

1. Clone the repository:
    ```bash
    git clone https://github.com/your-username/load-balancer-go.git
    cd load-balancer-go
    ```
2. Install the required dependencies:
    ```bash
    go mod tidy
    ```
3. Run the load balancer:
    ```bash
    go run main.go
    ```
4. Open new terminal and run the backend server instances:
If no port is specified, the backend server will run on port 8080 by default.
    ```bash
    go run backend.go 8081
    go run backend.go 8082
    go run backend.go 8083
    ```
    5. Now run the test client module:
    ```bash
    go run test.go
    ```

## Notes
- The server for backend provide a simple delay to simulate processing time.
- The load balancer will distribute requests to the backend servers based on the chosen algorithm.
- Currently, the load balancer supports two algorithms: round-robin and random selection.
- After completion of tests, images will be saved in the `images` directory.