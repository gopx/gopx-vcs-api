#!/usr/bin/env bash

# Export the necessary environment variables
export $(cat scripts/.env | xargs)

# Install the API server executable
go install ./cmd/gopx-vcs-api

# Run the server
gopx-vcs-api