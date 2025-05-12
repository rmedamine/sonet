#!/bin/sh

# Start the frontend
cd /app/frontend
npm start &

# Start the backend in the background
cd /app/backend
./main

