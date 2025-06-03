# Stage 1: Build the Vue.js application
FROM node:20-alpine AS builder

# Set the working directory
WORKDIR /app

# Copy package.json and package-lock.json (or yarn.lock)
COPY package.json ./
COPY package-lock.json ./

# Install project dependencies
RUN npm install

# Copy the rest of the application code
COPY . .

# Build the application for production
# Inject build-time environment variables
ARG VUE_APP_API_BASE_URL
ENV VUE_APP_API_BASE_URL=${VUE_APP_API_BASE_URL}

RUN npm run build

# Stage 2: Serve the application with a simple static server
FROM node:20-alpine

# Install serve globally for serving static files
RUN npm install -g serve

# Set the working directory
WORKDIR /app

# Copy the build output from the builder stage
COPY --from=builder /app/dist ./dist

# Expose port 80
EXPOSE 80

# Start the static server
CMD ["serve", "-s", "dist", "-l", "80"]