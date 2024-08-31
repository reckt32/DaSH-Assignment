#!/bin/bash

go run server.go &

# Wait for the server to start
sleep 2

# Navigate to client directory and start clients
cd ../clients

# Number of clients
NUM_CLIENTS=3

# Start clients
for ((i=0; i<NUM_CLIENTS; i++)); do
    go run client.go $i $NUM_CLIENTS &
done

# Wait for all clients to finish
wait

echo "All clients have finished."
