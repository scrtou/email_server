#!/bin/bash
# This script ensures the backend is run from the correct directory (project root)
# and that the .env file is loaded properly.

# Change to the directory where the script is located to get a consistent starting point
cd "$(dirname "$0")"

# Export variables from the root .env file
if [ -f .env ]; then
  echo "Loading environment variables from .env file..."
  set -a # automatically export all variables
  source .env
  set +a
  # --- DEBUG ---
  echo "ENCRYPTION_KEY from shell is: '$ENCRYPTION_KEY'"
  echo "Length: ${#ENCRYPTION_KEY}"
  # --- END DEBUG ---
fi

# Now, change to the backend directory which is the Go module root
cd src/backend

# Run the main Go program
echo "Starting backend server from Go module root ($(pwd))..."
go run .